/*=============================================================================
#     FileName: rwstream.go
#       Author: sunminghong, allen.fantasy@gmail.com, http://weibo.com/5d13
#         Team: http://1201.us
#   LastChange: 2015-08-25 16:09:39
#      History:
=============================================================================*/


/*

*/

package gotools

import (
    "encoding/binary"
    "errors"
    "sync"

    "fmt"
)


const (
    BigEndian    = 0
    LittleEndian = 1
)

func GetEndianer(endian int) IEndianer {
    if endian == BigEndian {
        return binary.BigEndian
    } else {
        return binary.LittleEndian
    }
}



//switch bigEndianer or littleEndianer
type IEndianer interface {
    Uint16(b []byte) uint16
    PutUint16(b []byte, v uint16)

    Uint32(b []byte) uint32
    PutUint32(b []byte, v uint32)

    Uint64(b []byte) uint64
    PutUint64(b []byte, v uint64)
}

// A Buffer is a variable-sized buffer of bytes with Read and Write methods.
// The zero value for Buffer is an empty buffer ready to use.
type RWStream struct {
    buffSize int

    //Endian   int //default to false, means that is littleEdian
    Endianer IEndianer

    buf []byte
    rlock *sync.RWMutex

    end  int
    last int
}

func NewRWStream(buf interface{}, endianer IEndianer) *RWStream {
    b := &RWStream{Endianer: endianer, rlock:new(sync.RWMutex)}

    b.Endianer = endianer
    /*
    b.Endian = endian
    if endian == BigEndian {
        b.Endianer = binary.BigEndian
    } else {
        b.Endianer = binary.LittleEndian
    }
    */

    b.Init(buf)
    return b
}

// ErrTooLarge is passed to panic if memory cannot be allocated to store data in a buffer.
var ErrTooLarge = errors.New("net.RWStream: too large")
var ErrIndex = errors.New("net.RWStream: index over range")

func (b *RWStream) DebugOut() (string,int,[]byte,int,int) { return "rwstream.buf:",b.buffSize,b.buf[:],b.last,b.end}

func (b *RWStream) Bytes() []byte {
    if b.last <= b.end {
        p := b.buf[b.last : b.end]
        return p
    }

    //分段读取
    buf := make([]byte, b.Len())
    copy(buf, b.buf[b.last:])
    copy(buf[b.buffSize - b.last:], b.buf[:b.end])

    return buf
}

func (b *RWStream) Len() int {
    if b.end > b.last {
        return b.end - b.last
    } else {
        return b.buffSize + b.end - b.last
    }
}

func (b *RWStream) Init(params ...interface{}) {
    if len(params) > 0 {
        buf := params[0]

        switch tmp := buf.(type) {
        case int:
            b.buffSize = tmp
            b.buf = make([]byte, b.buffSize)
            b.last = 0
            b.end = 0
        case []byte:
            b.buf = tmp[:]
            b.buffSize = len(tmp)
            b.last = 0
            b.end = len(tmp)
        default:
            b.buffSize = 1024
            b.buf = make([]byte, b.buffSize)
            b.last = 0
            b.end = 0
        }
    } else {
        b.last = 0
        b.end = 0
        b.buffSize = 1024
        b.buf = make([]byte, b.buffSize)
    }
}

//call Reset before each use this Buffer
func (b *RWStream) Reset() {
    b.end = 0
    b.last = 0
    b.buffSize = 1024
    b.buf = make([]byte, b.buffSize)
}


func (b *RWStream) Write(p []byte) (i int) {
    n := len(p)
    m := b.Len()
    max := b.buffSize

    if b.end + n <= max {
        copy(b.buf[b.end:], p)
        b.end = (b.end + n) % max
        return n
    }

    if m + n > max {
        b.rlock.Lock()
        // not enough space anywhere
        //icap := m + n - max
        icap := m + n
        if icap < 10240 {
            icap = icap * 2
        }
        fmt.Printf("makeslice makeslice makeslice", b.last,b.end, icap,n, "\n")
        tmp := make([]byte, icap)
        //tmp := make([]byte, icap + max)
        if b.last < b.end {
            copy(tmp, b.buf[b.last:b.end])
        } else {
            copy(tmp, b.buf[b.last:b.buffSize])
            copy(tmp[b.buffSize - b.last:], b.buf[:b.end])
        }
        copy(tmp[m:], p)

        fmt.Printf("|||% X\n",b.buf)
        fmt.Printf("|||% X\n",tmp)

        b.buffSize = len(tmp)
        b.last = 0
        b.end = n + m
        b.buf = tmp
        b.rlock.Unlock()

        return n
    }

    if b.last > b.end {
        copy(b.buf[b.end:], p)
        b.end = (b.end + n) % max
        return n
    }

    //分段写入
    copy(b.buf[b.end:], p[:max - b.end])
    copy(b.buf[0:], p[max - b.end:])
    b.end = n + b.end - max

    return n
}

func (b *RWStream) Read(n int) ([]byte,int) {
    end := b.end
    if b.last <= b.end {
        if b.last + n <= end {
            p := b.buf[b.last : b.last+n]
            b.last += n
            return p, n
        } else {
            return nil,0
        }
    }

    max := b.buffSize
    if b.last + n < max {
        p := b.buf[b.last : b.last+n]
        b.last += n
        return p, n
    }

    if b.last + n > b.end + max {
        return nil,0
    }

    //分段读取
    buf := make([]byte, n)
    copy(buf, b.buf[b.last:])
    copy(buf[max - b.last:], b.buf[:n + b.last - max])
    b.last = n - max + b.last

    return buf, n
}

func (b *RWStream) GetPos() int {
    return b.last
}

func (b *RWStream) SetPos(pos int) {
    if pos < 0 {
        b.last += pos
        if b.last < 0 {
            b.last += b.buffSize
        }
        return
    }

    if b.end < b.last {
        la := (b.last + pos) % b.buffSize
        if la > b.end {
            b.last = b.end
        } else {
            b.last = la
        }
    } else if b.last + pos > b.end {
        b.last = b.end
    } else {
        b.last += pos
    }
}

// WriteString appends the contents of s to the buffer.  The return
// value n is the length of s; err is always nil.
// If the buffer becomes too large, WriteString will panic with
// ErrTooLarge.
func (b *RWStream) WriteString(s string) int {
    b.WriteUint(uint(len(s)))
    return b.Write([]byte(s))
}

func (b *RWStream) WriteStringU32(s string) int {
    b.WriteUint32(uint32(len(s)))
    return b.Write([]byte(s))
}


func (b *RWStream) WriteByte(c byte) int {
    b.Write([]byte{c})
    return 1
}

func (b *RWStream) WriteUint16(x uint16) int {
    var buf = make([]byte, 2)
    b.Endianer.PutUint16(buf, x)
    return b.Write(buf)
}

func (b *RWStream) WriteUint32(x uint32) int {
    var buf = make([]byte, 4)
    b.Endianer.PutUint32(buf, x)
    return b.Write(buf)
}

func (b *RWStream) WriteUint64(x uint64) int {
    var buf = make([]byte, 8)
    b.Endianer.PutUint64(buf, x)
    return b.Write(buf)
}

func (b *RWStream) ReadByte() (byte, error) {
    buf, n := b.Read(1)
    if n < 1 {
        return 0, ErrIndex
    }
    return buf[0], nil
}

func (b *RWStream) ReadUint16() (uint16, error) {
    buf, n := b.Read(2)
    if n < 2 {
        return 0, ErrIndex
    }
    x := b.Endianer.Uint16(buf)
    return x, nil
}

func (b *RWStream) ReadUint32() (uint32, error) {
    buf, n := b.Read(4)
    if n < 4 {
        return 0, ErrIndex
    }
    x := b.Endianer.Uint32(buf)
    return x, nil
}

func (b *RWStream) ReadUint64() (uint64, error) {
    buf, n := b.Read(8)
    if n < 8 {
        return 0, ErrIndex
    }
    x := b.Endianer.Uint64(buf)
    return x, nil
}

func (b *RWStream) ReadUint() (uint, error) {
    if b.Len() < 1 {
        return 0, ErrIndex
    }

    var x uint
    var s uint
    for {
        i := 0
        b, err := b.ReadByte()
        if err != nil {
            break
        }

        if b < 0x80 {
            if i > 9 || i == 9 && b > 1 {
                return 0, ErrTooLarge
            }
            return x | uint(b)<<s, nil
        }
        x |= uint(b & 0x7f) << s
        s += 7
        i += 1
    }
    return 0, ErrTooLarge
}

func (b *RWStream) ReadInt() (int, error) {
    ux, err := b.ReadUint() // ok to continue in presence of error
    if err != nil {
        return 0, err
    }

    x := int(ux >> 1)
    if ux&1 != 0 {
        x = ^x
    }
    return x, nil
}

func (b *RWStream) WriteUint(x uint) int {
    buf := [8]byte{}
    i := 0
    for x >= 0x80 {
        buf[i] = byte(x) | 0x80
        x >>= 7
        i++
    }
    buf[i] = byte(x)

    b.Write(buf[0 : i+1])
    return i + 1
}

func (b *RWStream) WriteInt(x int) int {
    ux := uint(x) << 1
    if x < 0 {
        ux = ^ux
    }
    return b.WriteUint(ux)
}

func (b *RWStream) ReadStringU32() (string, error) {
    l, err := b.ReadUint32()
    if err != nil {
        return "", err
    }

    ll := int(l)
    buf, n := b.Read(ll)
    if n < ll {
        return "", ErrIndex
    }

    return string(buf), nil
}

func (b *RWStream) ReadString() (string, error) {
    l, err := b.ReadUint()
    if err != nil {
        return "", err
    }

    ll := int(l)
    buf, n := b.Read(ll)
    if n < ll {
        return "", ErrIndex
    }

    return string(buf), nil
}

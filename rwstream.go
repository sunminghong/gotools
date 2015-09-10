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

func GetEndianer(writePointian int) IEndianer {
    if writePointian == BigEndian {
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
    bufSize int
    initSize int

    //Endian   int //default to false, means that is littleEdian
    Endianer IEndianer

    buf []byte
    rlock *sync.RWMutex

    writePoint  int
    readPoint int
}

func NewRWStream(buf interface{}, writePointianer IEndianer) *RWStream {
    b := &RWStream{Endianer: writePointianer, rlock:new(sync.RWMutex)}

    b.Endianer = writePointianer
    /*
    b.Endian = writePointian
    if writePointian == BigEndian {
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

func (b *RWStream) DebugOut() (string,int,[]byte,int,int) { return "rwstream.buf:",b.bufSize,b.buf[:],b.readPoint,b.writePoint}

func (b *RWStream) Bytes() []byte {
    Trace("Bytes() read:%d, %d, %d", b.readPoint, b.writePoint, b.bufSize)
    if b.readPoint <= b.writePoint {
        p := b.buf[b.readPoint : b.writePoint]
        return p
    }

    Trace("Bytes() segment read")
    //分段读取
    buf := make([]byte, b.Len())
    copy(buf, b.buf[b.readPoint:])
    copy(buf[b.bufSize - b.readPoint:], b.buf[:b.writePoint])

    return buf
}

func (b *RWStream) Len() int {
    if b.writePoint >= b.readPoint {
        return b.writePoint - b.readPoint
    } else {
        return b.bufSize + b.writePoint - b.readPoint
    }
}

func (b *RWStream) Init(params ...interface{}) {
    if len(params) > 0 {
        buf := params[0]

        switch tmp := buf.(type) {
        case int:
            b.initSize = tmp
            b.bufSize = tmp
            b.buf = make([]byte, b.bufSize)
            b.readPoint = 0
            b.writePoint = 0
        case []byte:
            b.buf = tmp[:]
            b.initSize = len(tmp)
            b.bufSize = b.initSize
            b.readPoint = 0
            b.writePoint = len(tmp)
        default:
            b.initSize = 20480
            b.bufSize = 20480
            b.buf = make([]byte, b.bufSize)
            b.readPoint = 0
            b.writePoint = 0
        }
    } else {
        b.readPoint = 0
        b.writePoint = 0
        b.initSize = 20480
        b.bufSize = 20480
        b.buf = make([]byte, b.bufSize)
    }
}

//call Reset before each use this Buffer
func (b *RWStream) Reset() {
    b.writePoint = 0
    b.readPoint = 0
    b.bufSize = b.initSize
    b.buf = make([]byte, b.bufSize)
}


func (b *RWStream) Write(p []byte) (i int) {
    n := len(p)
    m := b.Len()
    max := b.bufSize

    if m + n >= max {
        //是否需要扩展环容量
        b.rlock.Lock()
        // not enough space anywhere
        //icap := m + n - max
        icap := m + n
        if icap < 20480 {
            icap = icap * 2
        }
        Trace("makeslice:readPoint:%d,writePoint:%d,icap:%d,m:%d,n:%d", b.readPoint,b.writePoint,icap,m,n)
        tmp := make([]byte, icap)
        //tmp := make([]byte, icap + max)
        if b.readPoint < b.writePoint {
            copy(tmp, b.buf[b.readPoint:b.writePoint])
        } else {
            copy(tmp, b.buf[b.readPoint:max])
            copy(tmp[max - b.readPoint:], b.buf[:b.writePoint])
        }
        copy(tmp[m:], p)

        fmt.Printf("|||% X\n",b.buf)
        fmt.Printf("|||% X\n",tmp)

        b.bufSize = len(tmp)
        b.readPoint = 0
        b.writePoint = n + m
        b.buf = tmp
        b.rlock.Unlock()

        return n
    }

    if b.writePoint + n <= max {
        copy(b.buf[b.writePoint:], p)
        Trace("rwstream.write one:%d, %d, %d", b.writePoint, n, max)
        b.writePoint = (b.writePoint + n) % max
        return n
    }

    //分段写入
    Trace("rwstream.write segment", max)
    copy(b.buf[b.writePoint:], p[:max - b.writePoint])
    copy(b.buf[0:], p[max - b.writePoint:])
    b.writePoint = n + b.writePoint - max

    return n
}

func (b *RWStream) Read(n int) ([]byte,int) {
    if n==0 || b.Len() < n {
        return nil,0
    }

    max := b.bufSize
    writePoint := b.writePoint
    if b.readPoint < writePoint || b.readPoint + n <= max {
        p := b.buf[b.readPoint : b.readPoint+n]
        b.readPoint += n
        return p, n
    }

    //分段读取
    Trace("read segment", b.readPoint, n, b.writePoint, max)
    buf := make([]byte, n)
    copy(buf, b.buf[b.readPoint:])
    copy(buf[max - b.readPoint:], b.buf[:n + b.readPoint - max])
    b.readPoint = n - max + b.readPoint

    return buf, n
}

func (b *RWStream) GetPos() int {
    return b.readPoint
}

func (b *RWStream) SetPos(pos int) {
    if pos < 0 {
        b.readPoint += pos
        if b.readPoint < 0 {
            b.readPoint += b.bufSize
        }
        return
    }

    if b.writePoint < b.readPoint {
        la := (b.readPoint + pos) % b.bufSize
        if la > b.writePoint {
            b.readPoint = b.writePoint
        } else {
            b.readPoint = la
        }
    } else if b.readPoint + pos > b.writePoint {
        b.readPoint = b.writePoint
    } else {
        b.readPoint += pos
    }
}

// WriteString appwritePoints the contents of s to the buffer.  The return
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

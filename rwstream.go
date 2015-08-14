/*=============================================================================
#     FileName: rwstream.go
#         Desc: RWStream struct
#       Author: sunminghong
#        Email: allen.fantasy@gmail.com
#     HomePage: http://weibo.com/5d13
#      Version: 0.0.1
#   LastChange: 2015-08-14 11:37:10
#      History:
=============================================================================*/
package gotools

import (
    "encoding/binary"
    "errors"
    //"fmt"
)

const (
    BigEndian    = 0
    LittleEndian = 1
)

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

    Endian   int //default to false, means that is littleEdian
    Endianer IEndianer

    buf []byte // contents are the bytes buf[off:len(buf)]

    off  int // read at &buf[off], write at &buf[len(buf)]
    end  int // data end pos, data = buf[off,end]
    last int // last read operation, so that Unread* can work correctly.
}

func NewRWStream(buf interface{}, endian int) *RWStream {
    b := &RWStream{Endian: endian}

    b.Endian = endian
    if endian == BigEndian {
        b.Endianer = binary.BigEndian
    } else {
        b.Endianer = binary.LittleEndian
    }

    b.Init(buf)
    return b
}

// ErrTooLarge is passed to panic if memory cannot be allocated to store data in a buffer.
var ErrTooLarge = errors.New("net.RWStream: too large")
var ErrIndex = errors.New("net.RWStream: index over range")

func (b *RWStream) DebugOut() (string,int,[]byte,int,int,int) { return "rwstream.buf:",len(b.buf),b.buf[:],b.off,b.last,b.end}
func (b *RWStream) Bytes() []byte { return b.buf[b.off:b.end] }

func (b *RWStream) Len() int { return b.end - b.off }

func (b *RWStream) Init(params ...interface{}) {
    if len(params) > 0 {
        buf := params[0]

        switch tmp := buf.(type) {
        case int:
            b.buffSize = tmp
            b.buf = make([]byte, b.buffSize)
            b.last = 0
            b.off = 0
            b.end = 0
        case []byte:
            b.buf = tmp[:]
            b.buffSize = len(tmp)
            b.last = 0
            b.off = 0
            b.end = len(tmp)
        default:
            b.buffSize = 1024
            b.buf = make([]byte, b.buffSize)
            b.last = 0
            b.end = 0
            b.off = 0
        }
    } else {
        b.last = 0
        b.end = 0
        b.off = 0
    }
}

//call Reset before each use this Buffer
func (b *RWStream) Reset() {
    b.off = b.end
    b.last = b.off
}

// makeSlice allocates a slice of size n. If the allocation fails, it panics
// with ErrTooLarge.
func makeSlice(n int) []byte {
    // If the make fails, give a known error.
    defer func() {
        if recover() != nil {
            panic(ErrTooLarge)
        }
    }()
    return make([]byte, n*2)
}

// grow grows the buffer to guarantee space for n more bytes.
// It returns the index where bytes should be written.
// If the buffer can't grow it will panic with ErrTooLarge.
func (b *RWStream) grow(n int) int {
    m := b.Len()
    x := len(b.buf)

    if b.end+n > x {
        if m+n > x {
            var buf []byte
            // not enough space anywhere
            buf = makeSlice(m + n)
            copy(buf, b.buf[b.off:])
            b.buf = buf
        } else {
            copy(b.buf[0:], b.buf[b.off:b.off+m])
        }
        b.last -= b.off
        b.off = 0
        b.end = m
    //} else {
        //if x > b.buffSize {
            //b.buf = b.buf[b.off : b.off+m]
            //b.last -= b.off
            //b.off = 0
            //b.end = m
        //}
    }
    return b.end
}

// Write appends the contents of p to the buffer.  The return
// value n is the length of p; err is always nil.
// If the buffer becomes too large, Write will panic with
// ErrTooLarge.
func (b *RWStream) Write(p []byte) (n int) {
    //defer func(){
		//if x := recover(); x != nil {
            //fmt.Printf("rwstream write() recover:%s,cap=%d,end=%d,len=%d,newlen=%d\n", x,cap(b.buf), b.end,b.Len(),len(p))

        //panic("write is err!")
		//}
    //}()

    //fmt.Printf("2rwstream write(),cap=%d,len(b.buf)=%d,off=%d,end=%d,m=%d,len=%d,newlen=%d\n", cap(b.buf),len(b.buf), b.off, b.end,m,b.Len(),len(p))
    nn := len(p)
    m := b.grow(nn)
    b.end += nn

    //if len(b.buf) > 2000 {
        //fmt.Printf("1rwstream write(),cap=%d,len(b.buf)=%d,off=%d,end=%d,len=%d,newlen=%d\n", cap(b.buf),len(b.buf), b.off, b.end,b.Len(),len(p))
        //panic("rwstream.Write() is error:")
    //}
    return copy(b.buf[m:], p)
}

func (b *RWStream) GetPos() int {
    return b.last - b.off
}

func (b *RWStream) SetPos(pos int) {
    if pos < 0 {
        b.last += pos
        if b.last < b.off {
            b.last = b.off
        }
        return
    }

    if pos+b.off > b.end {
        b.last = b.end
    } else {
        b.last = pos + b.off
    }
}

func (b *RWStream) Read(n int) ([]byte,int) {
    if b.last+n > b.end {
        return nil,0
        //n = b.end - b.last
    }
    //if n<0 {
    //    return 0,nil
    //}
    p := b.buf[b.last : b.last+n]
    b.last += n
    return p, n
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
    m := b.grow(1)
    b.buf[m] = c
    b.end += 1
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
    if b.last >= b.end {
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
        x |= uint(b&0x7f) << s
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

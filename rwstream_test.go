/*=============================================================================
#     FileName: rwstream_test.go
#       Author: sunminghong, allen.fantasy@gmail.com, http://weibo.com/5d13
#         Team: http://1201.us
#   LastChange: 2015-08-25 19:21:49
#      History:
=============================================================================*/


/*

*/
package gotools

import (
//    "encoding/binary"
//    "errors"
    "testing"
    "bytes"
    "fmt"
)


func Test_NewRWStream(t *testing.T){
    bys :=[]byte{1,2,3,4,5,6,7,8,9,10}

    b := NewRWStream(bys,GetEndianer(BigEndian))

    _bs := b.Bytes()
    if !bytes.Equal(bys,_bs) {
        t.Error("func Bytes is error:",_bs,bys)
    }

    return
}

func Test_Init(t *testing.T) {
    bytes :=[]byte{1,2,3,4,5,6,7,8,9,10}

    b := NewRWStream(bytes,GetEndianer(BigEndian))
    b.Init()

    if (b.readPoint !=0) || (b.writePoint != 0) {
        t.Error("init() is error:readPoint is wrong(0)",b.readPoint)
    }
}

func Test_RW(t *testing.T) {
    bytes :=[]byte{1,2,3,4,5,6,7,8,9,10}

    b := NewRWStream(bytes,GetEndianer(BigEndian))
    //fmt.Println("b.buf Len(),off,writePoint,readPoint=",ii,b.Len(),b.off,b.writePoint,b.readPoint)
    b.Init(31)
    //fmt.Println("b.buf Len(),off,writePoint,readPoint=",ii,b.Len(),b.off,b.writePoint,b.readPoint)

    h,i,j,k,l,m := 1,16,3232,646426464,7777777,-77777777
    fmt.Printf("///% X\n", b.Bytes())

    for ii:=0;ii < 10000;ii++ {
        if ii ==2 {
            //b.Reset()
        }
        fmt.Printf("/%d//\n",ii)
        //fmt.Printf("/%d//% X\n",ii, b.Bytes())
        //fmt.Printf("|%d||%q\n",ii, b.Bytes())
        b.WriteByte(byte(h))
        b.WriteUint16(uint16(i))
        b.WriteUint32(uint32(j))
        b.WriteUint64(uint64(k))
        b.WriteUint(uint(l))
        b.WriteInt(m)


        //fmt.Printf("|%d||% X\n",ii, b.buf)

        h1,err := b.ReadByte()
        if err != nil || int(h1) != h {
            t.Error("ReadByte() error h1=",ii,h1,h)
        }

        i1,err := b.ReadUint16()
        if err != nil || int(i1) != i {
            t.Error("ReadByte() error h1=",ii,i1,i)
        }

        j1,err := b.ReadUint32()
        println("j1:", j1)
        if err != nil || j1 != uint32(j) {
            t.Error("ReadByte() error h1=",ii,j1,j,err)
        }

        k1,err := b.ReadUint64()
        println("k1:", k1)
        if err != nil || int(k1) !=k {
            t.Error("ReadByte() error k1=",ii,k1,k)
        }

        l1,err := b.ReadUint()
        println("l1:", l1)
        if err != nil || int(l1) !=l {
            t.Error("ReadByte() error l1=",ii,l1,l)
        }

        m1,err := b.ReadInt()
        println("m1:", m1)
        if err != nil || int(m1) !=m {
            t.Error("ReadByte() error m1=",ii,m1,m)
        }

        s := "abcdefghijk"
        b.WriteString(s)
        b.WriteString(s)

        s1,err := b.ReadString()
        println("s1:", s1)
        if err != nil || s1 !=s {
            t.Error("ReadByte() error s1=",ii,s1,s)
        }

        s1,err = b.ReadString()
        println("s1:", s1)
        if err != nil || s1 !=s {
            t.Error("ReadByte() error s1=",ii,s1,s)
        }
    }

}



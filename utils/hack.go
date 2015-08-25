/*=============================================================================
#     FileName: hack.go
#         Desc: RWStream struct
#       Author: sunminghong
#        Email: allen.fantasy@gmail.com
#     HomePage: http://weibo.com/5d13
#      Version: 0.0.1
#   LastChange: 2015-08-18 14:38:18
#      History:
=============================================================================*/
package utils

import (
    "unsafe"
    "reflect"
)

// convert b to string without copy
func ByteString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// returns &s[0], which is not allowed in go
func StringPointer(s string) unsafe.Pointer {
	p := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return unsafe.Pointer(p.Data)
}

// returns &b[0], which is not allowed in go
func BytePointer(b []byte) unsafe.Pointer {
	p := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	return unsafe.Pointer(p.Data)
}


/*
将一个struct 数据转换成[]byte，用于保存，比json、gob高效n倍
type Struct struct {
	A int
	B int
}

func StructToBytes(s *Struct) []byte {
    var sizeOfStruct = int(unsafe.Sizeof(Struct{}))
	var x reflect.SliceHeader
	x.Len = sizeOfStruct
	x.Cap = sizeOfStruct
	x.Data = uintptr(unsafe.Pointer(s))
	return *(*[]byte)(unsafe.Pointer(&x))
}

func BytesToStruct(b []byte) *Struct {
	return (*Struct)(unsafe.Pointer(
		(*reflect.SliceHeader)(unsafe.Pointer(&b)).Data,
	))
}
*/

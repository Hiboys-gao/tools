package utils

import (
	"reflect"
	"unsafe"
)

// String2Bytes string to []byte
func String2Bytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

// Bytes2String []byte to string
func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

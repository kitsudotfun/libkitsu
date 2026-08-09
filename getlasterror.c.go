package main

import "C"

var LastErr error

//export KtGetLastError
func KtGetLastError() *C.char {
	if LastErr == nil {
		return nil
	}

	return C.CString(LastErr.Error())
}

//export GMS_KtGetLastError
func GMS_KtGetLastError() *C.char {
	if LastErr == nil {
		return nil
	}

	str := C.CString(LastErr.Error())
	LastString = str

	return str
}

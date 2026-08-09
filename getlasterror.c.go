package main

import "C"

var LastError error

//export KtGetLastError
func KtGetLastError() *C.char {
	if LastError == nil {
		return nil
	}

	return C.CString(LastError.Error())
}

//export GMS_KtGetLastError
func GMS_KtGetLastError() *C.char {
	if LastError == nil {
		return nil
	}

	str := C.CString(LastError.Error())
	LastString = str

	return str
}

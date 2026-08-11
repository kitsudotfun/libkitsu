package main

import "C"

var lastError error

//export KtGetLastError
func KtGetLastError() *C.char {
	if lastError == nil {
		return nil
	}

	return C.CString(lastError.Error())
}

//export GMS_KtGetLastError
func GMS_KtGetLastError() *C.char {
	if lastError == nil {
		return nil
	}

	str := C.CString(lastError.Error())
	lastString = str

	return str
}

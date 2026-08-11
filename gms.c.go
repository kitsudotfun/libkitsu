package main

/*
#include <stdlib.h>
*/
import "C"
import "unsafe"

var lastString *C.char

//export GMS_KtFreeLastString
func GMS_KtFreeLastString() C.double {
	if lastString == nil {
		return 0
	}

	C.free(unsafe.Pointer(lastString))
	lastString = nil

	return 1
}

package main

/*
#include <stdlib.h>
*/
import "C"
import "unsafe"

var LastString *C.char

//export GMS_KtFreeLastString
func GMS_KtFreeLastString() C.double {
	if LastString == nil {
		return 0
	}

	C.free(unsafe.Pointer(LastString))
	LastString = nil

	return 1
}

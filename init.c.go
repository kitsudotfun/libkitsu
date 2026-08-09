package main

import (
	. "github.com/kitsudotfun/libkitsu/internal"
)

/*
#include <stdbool.h>
*/
import "C"

//export KtInit
func KtInit(game *C.char, key *C.char) *C.char {
	id, err := IKtInit(C.GoString(game), C.GoString(key))
	if err != nil {
		LastErr = err
		return nil
	}

	return C.CString(id.String())
}

//export GMS_KtInit
func GMS_KtInit(game *C.char, key *C.char) *C.char {
	id, err := IKtInit(C.GoString(game), C.GoString(key))
	if err != nil {
		LastErr = err
		return nil
	}

	str := C.CString(id.String())
	LastString = str

	return str
}

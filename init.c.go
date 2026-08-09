package main

import (
	. "github.com/kitsudotfun/libkitsu/internal"
)

/*
#include <stdbool.h>
*/
import "C"

//export KtInit
func KtInit(game *C.char, key *C.char) C.bool {
	err := IKtInit(C.GoString(game), C.GoString(key))
	if err != nil {
		LastErr = err
		return false
	}

	return true
}

//export GMS_KtInit
func GMS_KtInit(game *C.char, key *C.char) C.double {
	err := IKtInit(C.GoString(game), C.GoString(key))
	if err != nil {
		LastErr = err
		return 0
	}

	return 1
}

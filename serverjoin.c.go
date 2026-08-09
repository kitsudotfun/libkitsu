package main

import (
	. "github.com/kitsudotfun/libkitsu/internal"
)

/*
#include <stdbool.h>
*/
import "C"

//export KtServerJoin
func KtServerJoin(id *C.char, password *C.char) C.bool {
	err := IKtServerJoin(C.GoString(id), C.GoString(password))
	if err != nil {
		LastError = err
		return false
	}

	return true
}

//export GMS_KtServerJoin
func GMS_KtServerJoin(id *C.char, password *C.char) C.double {
	err := IKtServerJoin(C.GoString(id), C.GoString(password))
	if err != nil {
		LastError = err
		return 0
	}

	return 1
}

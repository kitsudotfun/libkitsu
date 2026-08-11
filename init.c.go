package main

import (
	. "github.com/kitsudotfun/libkitsu/internal"
)

/*
#include <stdbool.h>
#include <stdint.h>
*/
import "C"

//export KtInit
func KtInit(game *C.char, key *C.char) C.uint32_t {
	id, err := IKtInit(C.GoString(game), C.GoString(key))
	if err != nil {
		lastError = err
		return 0
	}

	return C.uint32_t(id)
}

//export GMS_KtInit
func GMS_KtInit(game *C.char, key *C.char) C.double {
	id, err := IKtInit(C.GoString(game), C.GoString(key))
	if err != nil {
		lastError = err
		return 0
	}

	return C.double(id)
}

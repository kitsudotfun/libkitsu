package main

import (
	. "github.com/kitsudotfun/kyuubi/api/defs"
	. "github.com/kitsudotfun/libkitsu/internal"
)

/*
#include <stdint.h>
#include <stdbool.h>
*/
import "C"

//export KtServerJoin
func KtServerJoin(id C.uint32_t, password *C.char) C.bool {
	err := IKtServerJoin(PeerID(id), C.GoString(password))
	if err != nil {
		lastError = err
		return false
	}

	return true
}

//export GMS_KtServerJoin
func GMS_KtServerJoin(id C.double, password *C.char) C.double {
	err := IKtServerJoin(PeerID(id), C.GoString(password))
	if err != nil {
		lastError = err
		return 0
	}

	return 1
}

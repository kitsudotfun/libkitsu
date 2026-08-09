package main

import (
	. "github.com/kitsudotfun/libkitsu/internal"
)

/*
#include <stdbool.h>
*/
import "C"

//export KtPeerDisconnect
func KtPeerDisconnect(id *C.char) C.bool {
	err := IKtPeerDisconnect(C.GoString(id))
	if err != nil {
		LastError = err
		return false
	}

	return true
}

//export GMS_KtPeerDisconnect
func GMS_KtPeerDisconnect(id *C.char) C.double {
	err := IKtPeerDisconnect(C.GoString(id))
	if err != nil {
		LastError = err
		return 0
	}

	return 1
}

package main

import (
	. "github.com/kitsudotfun/kyuubi/api/defs"
	. "github.com/kitsudotfun/libkitsu/internal"
)

/*
#include <stdbool.h>
#include <stdint.h>
*/
import "C"

//export KtPeerDisconnect
func KtPeerDisconnect(id C.uint32_t) C.bool {
	err := IKtPeerDisconnect(PeerID(id))
	if err != nil {
		LastError = err
		return false
	}

	return true
}

//export GMS_KtPeerDisconnect
func GMS_KtPeerDisconnect(id C.double) C.double {
	err := IKtPeerDisconnect(PeerID(id))
	if err != nil {
		LastError = err
		return 0
	}

	return 1
}

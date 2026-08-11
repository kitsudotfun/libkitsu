package main

import (
	"encoding/base64"
	"unsafe"

	. "github.com/kitsudotfun/kyuubi/api/defs"
	. "github.com/kitsudotfun/libkitsu/internal"
)

/*
#include <stdlib.h>
#include <stdbool.h>
#include <stdint.h>
*/
import "C"

//export KtPeerMsgSend
func KtPeerMsgSend(id C.uint32_t, msg *C.char, msgLen C.int) C.bool {
	err := IKtPeerMsgSend(PeerID(id), C.GoBytes(unsafe.Pointer(msg), msgLen))
	if err != nil {
		LastError = err
		return false
	}

	return true
}

//export GMS_KtPeerMsgSend
func GMS_KtPeerMsgSend(id C.double, msg *C.char) C.double {
	msgBytes, err := base64.StdEncoding.DecodeString(C.GoString(msg))
	if err != nil {
		LastError = err
		return 0
	}

	err = IKtPeerMsgSend(PeerID(id), msgBytes)
	if err != nil {
		LastError = err
		return 0
	}

	return 1
}

//export KtPeerMsgSendAll
func KtPeerMsgSendAll(msg *C.char, msgLen C.int) C.bool {
	err := IKtPeerMsgSendAll(C.GoBytes(unsafe.Pointer(msg), msgLen))
	if err != nil {
		LastError = err
		return false
	}

	return true
}

//export GMS_KtPeerMsgSendAll
func GMS_KtPeerMsgSendAll(msg *C.char) C.double {
	msgBytes, err := base64.StdEncoding.DecodeString(C.GoString(msg))
	if err != nil {
		LastError = err
		return 0
	}

	err = IKtPeerMsgSendAll(msgBytes)
	if err != nil {
		LastError = err
		return 0
	}

	return 1
}

//export KtPeerMsgRecv
func KtPeerMsgRecv(id C.uint32_t, blocking C.bool, dst *C.char, dstLen C.int) C.int {
	s, err := IKtPeerMsgRecv(PeerID(id), blocking == true)
	if err != nil {
		LastError = err
		return -1
	}
	if len(s) > int(dstLen) {
		LastError = ErrBufferTooSmall
		return -1
	}
	if len(s) > 0 {
		copy(unsafe.Slice((*byte)(unsafe.Pointer(dst)), len(s)), s)
	}

	return C.int(len(s))
}

//export GMS_KtPeerMsgRecv
func GMS_KtPeerMsgRecv(id C.double, blocking C.double) *C.char {
	s, err := IKtPeerMsgRecv(PeerID(id), blocking != 0)
	if err != nil {
		LastError = err
		return nil
	}

	str := C.CString(base64.StdEncoding.EncodeToString(s))
	LastString = str

	return str
}

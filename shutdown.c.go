package main

import (
	. "github.com/kitsudotfun/libkitsu/internal"
)

/*
#include <stdbool.h>
*/
import "C"

//export KtShutdown
func KtShutdown() C.bool {
	err := IKtShutdown()
	if err != nil {
		LastError = err
		return false
	}

	return true
}

//export GMS_KtShutdown
func GMS_KtShutdown() C.double {
	err := IKtShutdown()
	if err != nil {
		LastError = err
		return 0
	}

	return 1
}

package main

import (
	"encoding/json"

	. "github.com/kitsudotfun/libkitsu/internal"

	. "github.com/kitsudotfun/kyuubi/api/defs"
)

/*
#include <stdint.h>
#include <stdbool.h>

typedef struct {
    const char *name;
	const char *password;

    bool hidden;

    uint8_t players;
    uint8_t max_players;

	const char *data;
} KtServerData;
*/
import "C"

//export KtServerAnnounce
func KtServerAnnounce(data *C.KtServerData) C.bool {
	err := IKtServerAnnounce(Server{
		Name:       C.GoString(data.name),
		Password:   C.GoString(data.password),
		Hidden:     bool(data.hidden),
		Players:    int(data.players),
		MaxPlayers: int(data.max_players),
		Data:       C.GoString(data.data),
	})
	if err != nil {
		LastErr = err
		return false
	}

	return true
}

//export GMS_KtServerAnnounce
func GMS_KtServerAnnounce(data *C.char) C.double {
	var server Server
	err := json.Unmarshal([]byte(C.GoString(data)), &server)
	if err != nil {
		LastErr = err
		return 0
	}

	err = IKtServerAnnounce(server)
	if err != nil {
		LastErr = err
		return 0
	}

	return 1
}

//export KtServerShutdown
func KtServerShutdown() C.bool {
	err := IKtServerShutdown()
	if err != nil {
		LastErr = err
		return false
	}

	return true
}

//export GMS_KtServerShutdown
func GMS_KtServerShutdown() C.double {
	err := IKtServerShutdown()
	if err != nil {
		LastErr = err
		return 0
	}

	return 1
}

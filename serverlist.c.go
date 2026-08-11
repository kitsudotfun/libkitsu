package main

import (
	"encoding/json"
	"unsafe"

	. "github.com/kitsudotfun/libkitsu/internal"
)

/*
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>

typedef struct {
	uint32_t id;
    const char *name;

	bool password;

    uint8_t players;
    uint8_t max_players;

	const char *data;
} KtServerListItem;
*/
import "C"

//export KtServerList
func KtServerList(count *C.int) *C.KtServerListItem {
	list, err := IKtServerList()
	if err != nil {
		lastError = err
		*count = -1
		return nil
	}
	if len(list) == 0 {
		*count = 0
		return nil
	}

	ptr := C.malloc(C.size_t(len(list)) * C.size_t(unsafe.Sizeof(C.KtServerListItem{})))
	if ptr == nil {
		lastError = ErrMallocFailed
		*count = -1
		return nil
	}

	servers := unsafe.Slice((*C.KtServerListItem)(ptr), len(list))
	for i, server := range list {
		servers[i].id = C.uint32_t(server.ID)
		servers[i].name = C.CString(server.Name)

		servers[i].password = server.Password != ""

		servers[i].players = C.uint8_t(server.Players)
		servers[i].max_players = C.uint8_t(server.MaxPlayers)

		data, _ := json.Marshal(server.Data)

		servers[i].data = C.CString(string(data))
	}

	*count = C.int(len(list))

	return (*C.KtServerListItem)(ptr)
}

//export GMS_KtServerList
func GMS_KtServerList() *C.char {
	list, err := IKtServerList()
	if err != nil {
		lastError = err
		return nil
	}
	if len(list) == 0 {
		return nil
	}

	b, _ := json.Marshal(list)

	str := C.CString(string(b))
	lastString = str

	return str
}

//export KtServerListFree
func KtServerListFree(ptr *C.KtServerListItem, count C.int) {
	if ptr == nil {
		return
	}

	servers := unsafe.Slice(ptr, int(count))
	for i := range servers {
		C.free(unsafe.Pointer(servers[i].name))
		C.free(unsafe.Pointer(servers[i].data))
	}

	C.free(unsafe.Pointer(ptr))
}

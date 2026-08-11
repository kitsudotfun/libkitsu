package main

import (
	"encoding/json"
	"unsafe"

	. "github.com/kitsudotfun/libkitsu/internal"

	. "github.com/kitsudotfun/kyuubi/api/defs"
)

/*
#include <stdlib.h>
#include <stdint.h>

typedef struct {
	uint32_t id;
	uint8_t state;
} KtPeerListItem;
*/
import "C"

//export KtPeerList
func KtPeerList(count *C.int) *C.KtPeerListItem {
	cm, err := GetCM()
	if err != nil {
		*count = -1
		return nil
	}

	peers := cm.GetPeers()

	if len(peers) == 0 {
		*count = 0
		return nil
	}

	ptr := C.malloc(C.size_t(len(peers)) * C.size_t(unsafe.Sizeof(C.KtPeerListItem{})))
	if ptr == nil {
		lastError = ErrMallocFailed
		*count = -1
		return nil
	}

	list := unsafe.Slice((*C.KtPeerListItem)(ptr), len(peers))
	for i, peer := range peers {
		list[i].id = C.uint32_t(peer.ID)
		list[i].state = C.uint8_t(peer.State)
	}

	*count = C.int(len(peers))

	return (*C.KtPeerListItem)(ptr)
}

//export GMS_KtPeerList
func GMS_KtPeerList() *C.char {
	type PeerInfo struct {
		ID    PeerID `json:"id"`
		State int    `json:"state"`
	}

	cm, err := GetCM()
	if err != nil {
		lastError = err
		return nil
	}

	var peers []PeerInfo
	for _, peer := range cm.GetPeers() {
		peers = append(peers, PeerInfo{ID: peer.ID, State: peer.State})
	}

	b, _ := json.Marshal(peers)

	str := C.CString(string(b))
	lastString = str

	return str
}

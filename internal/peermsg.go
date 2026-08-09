package internal

import (
	"errors"

	. "github.com/kitsudotfun/kyuubi/api/defs"
)

var (
	ErrBufferTooSmall   = errors.New("buffer too small")
	ErrPeerNotConnected = errors.New("peer not connected")
)

func IKtPeerMsgSend(id string, msg []byte) error {
	var sid SessionID
	err := sid.FromString(id)
	if err != nil {
		return err
	}

	peer, err := cm.GetPeerByID(sid)
	if err != nil {
		return err
	}
	if peer.State != Connected {
		return ErrPeerNotConnected
	}

	err = peer.Send(msg)
	if err != nil {
		return err
	}

	return nil
}

func IKtPeerMsgSendAll(msg []byte) error {
	for _, peer := range cm.peers {
		if peer.State != Connected {
			continue
		}

		err := peer.Send(msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func IKtPeerMsgRecv(id string) ([]byte, error) {
	var sid SessionID
	err := sid.FromString(id)
	if err != nil {
		return nil, err
	}

	peer, err := cm.GetPeerByID(sid)
	if err != nil {
		return nil, err
	}
	if peer.State != Connected {
		return nil, ErrPeerNotConnected
	}

	return peer.Receive(), nil
}

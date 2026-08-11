package internal

import (
	"errors"

	. "github.com/kitsudotfun/kyuubi/api/defs"
)

var (
	ErrBufferTooSmall   = errors.New("buffer too small")
	ErrPeerNotConnected = errors.New("peer not connected")
)

func IKtPeerMsgSend(id PeerID, msg []byte) error {
	peer, err := cm.GetPeerByID(id)
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

func IKtPeerMsgRecv(id PeerID, blocking bool) ([]byte, error) {
	peer, err := cm.GetPeerByID(id)
	if err != nil {
		return nil, err
	}
	if peer.State != Connected {
		return nil, ErrPeerNotConnected
	}

	return peer.Receive(blocking), nil
}

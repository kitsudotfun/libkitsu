package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"net/netip"
	"slices"
	"time"

	. "github.com/kitsudotfun/natneg/defs"
)

var (
	ErrNotInitialized = errors.New("not initialized")
)

type ConnectionManager struct {
	conn *net.UDPConn

	natneg Peer // TODO: add to peers eventually?
	peers  []*Peer
}

func (cm *ConnectionManager) init() error {
	var err error
	cm.conn, err = net.ListenUDP("udp4", nil)
	if err != nil {
		return err
	}

	go cm.readLoop()

	var discover DiscoverResponse
	err = natnegCall(Discover, DiscoverRequest{Token: sessionToken}, &discover)
	if err != nil {
		return err
	}

	attestToken = discover.Token

	go cm.keepAliveLoop()

	return nil
}

func (cm *ConnectionManager) shutdown() error {
	if serverData == nil { // client
		// disconnect from peers (server) manually
		for _, peer := range cm.peers {
			err := cm.deletePeer(peer.ID)
			if err != nil {
				return err
			}
		}
	} else { // server
		err := IKtServerShutdown()
		if err != nil {
			return err
		}
	}

	err := cm.conn.Close()
	if err != nil {
		return err
	}

	close(cm.natneg.inbox)

	return nil
}

func (cm *ConnectionManager) keepAliveLoop() {
	for range time.NewTicker(time.Second * 30).C {
		err := natnegCall(KeepAlive, KeepAliveRequest{}, &KeepAliveResponse{})
		if err != nil {
			break
		}
	}
}

func (cm *ConnectionManager) readLoop() {
	buf := make([]byte, 1400)
	for {
		n, addr, err := cm.conn.ReadFromUDPAddrPort(buf)
		if err != nil {
			break
		}

		err = cm.handlePacket(addr, slices.Clone(buf[:n]))
		if err != nil {
			// TODO: log this
			continue
		}
	}
}

func (cm *ConnectionManager) handlePacket(addr netip.AddrPort, data []byte) error {
	b, internal := bytes.CutPrefix(data, []byte(PeerMagic))
	if len(b) < 1 {
		return nil
	}

	// internal NATNEG message
	if internal && addr == cm.natneg.addr {
		if b[0] == JoinNotify {
			// TODO: handle error from this somehow
			go cm.handleJoinNotify(b[1:])
			return nil
		}

		// this is a response that natnegCall is waiting on
		cm.natneg.addMessage(b)
		return nil
	}

	// peer message
	peer, err := cm.getPeerByAddr(addr)
	if err != nil {
		return err
	}

	peer.lastPacket = time.Now()

	// internal peer message
	if internal {
		switch b[0] {
		case PeerKeepAlive:
			err = peer.send(append([]byte(PeerMagic), PeerKeepAliveAck))
			if err != nil {
				return err
			}

			return nil
		case PeerKeepAliveAck:
			// nothing to do
			return nil
		case PeerDisconnect:
			err = cm.deletePeer(peer.ID)
			if err != nil {
				return err
			}

			return nil
		}

		// other internal messages should be received by AddMessage
	}

	peer.addMessage(data)
	return nil
}

func (cm *ConnectionManager) handleJoinNotify(data []byte) error {
	var notify JoinNotifyResponse
	err := json.Unmarshal(data, &notify)
	if err != nil {
		return err
	}

	err = cm.addPeer(notify.ClientID, notify.ClientAddr)
	if err != nil {
		return err
	}

	return nil
}

func (cm *ConnectionManager) GetPeers() []*Peer {
	return cm.peers
}

func GetCM() (*ConnectionManager, error) {
	if cm.conn == nil {
		return nil, ErrNotInitialized
	}

	return &cm, nil
}

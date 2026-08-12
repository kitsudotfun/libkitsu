package internal

import (
	"bytes"
	"errors"
	"net/netip"
	"slices"
	"time"

	. "github.com/kitsudotfun/kyuubi/api/defs"
)

var (
	ErrPeerUnknown           = errors.New("peer unknown")
	ErrPeerExists            = errors.New("peer exists")
	ErrPeerAlreadyConnecting = errors.New("peer already connecting")
	ErrPeerConnectFailed     = errors.New("peer connection failed")
)

const PeerMagic = "KTsu"

// peer packet types
const (
	PeerKeepAlive byte = iota
	PeerKeepAliveAck

	PeerConnect
	PeerConnectAck

	PeerDisconnect
)

// peer state types
const (
	Uncontacted = iota // default
	Connecting
	ConnectingAck
	Connected
)

type Peer struct {
	cm *ConnectionManager

	ID    PeerID
	State int

	addr  netip.AddrPort
	inbox chan []byte

	lastPacket time.Time
}

// add a Peer to the ConnectionManager and connects to them
func (cm *ConnectionManager) addPeer(id PeerID, addr netip.AddrPort) error {
	for _, peer := range cm.peers {
		if peer.ID != id {
			continue
		}

		return ErrPeerExists
	}

	peer := Peer{cm: cm, ID: id, addr: addr, inbox: make(chan []byte, 32), lastPacket: time.Now()}

	// add to peers so AddPeerMessage can write into its inbox
	cm.peers = append(cm.peers, &peer)

	err := peer.connect()
	if err != nil {
		cm.deletePeer(peer.ID)
		return err
	}

	return nil
}

func (cm *ConnectionManager) deletePeer(id PeerID) error {
	msg := append([]byte(PeerMagic), PeerDisconnect)
	for i, peer := range cm.peers {
		if peer.ID != id {
			continue
		}

		cm.peers = slices.Delete(cm.peers, i, i+1)

		close(peer.inbox)

		err := peer.send(msg)
		if err != nil {
			return err
		}

		return nil
	}

	return ErrPeerUnknown
}

func (cm *ConnectionManager) getPeerByID(id PeerID) (*Peer, error) {
	for _, peer := range cm.peers {
		if peer.ID != id {
			continue
		}

		return peer, nil
	}

	return nil, ErrPeerUnknown
}

func (cm *ConnectionManager) getPeerByAddr(addr netip.AddrPort) (*Peer, error) {
	for _, peer := range cm.peers {
		if peer.addr != addr {
			continue
		}

		return peer, nil
	}

	return nil, ErrPeerUnknown
}

func (p *Peer) addMessage(msg []byte) {
	select {
	case p.inbox <- msg:
	default:
		_ = <-p.inbox
		p.inbox <- msg
	}
}

// establish a p2p connection to the Peer
func (p *Peer) connect() error {
	if p.State != Uncontacted {
		return ErrPeerAlreadyConnecting
	}

	p.State = Connecting
	packetType := PeerConnect

	for attempt := 0; attempt < 5 && p.State != Connected; {
		err := p.send(append([]byte(PeerMagic), packetType))
		if err != nil {
			return err
		}

		timeout := time.NewTimer(time.Second)

		var data []byte
		select {
		case data = <-p.inbox:
		case <-timeout.C:
			// do nothing, will be handled by prefix check
		}

		var found bool
		data, found = bytes.CutPrefix(data, []byte(PeerMagic))
		if !found || len(data) < 1 {
			attempt++
			continue
		}

		switch p.State {
		case Connecting:
			if data[0] == PeerConnect || data[0] == PeerConnectAck {
				p.State = ConnectingAck
				packetType = PeerConnectAck
				continue
			}
		case ConnectingAck:
			if data[0] == PeerConnectAck {
				p.State = Connected
			}
		}
	}
	if p.State != Connected {
		return ErrPeerConnectFailed
	}

	go p.keepAliveLoop()

	return nil
}

// like ConnectionManager keepAliveLoop but for the Peer connection
func (p *Peer) keepAliveLoop() {
	msg := append([]byte(PeerMagic), PeerKeepAlive)
	for range time.NewTicker(time.Second * 30).C {
		err := p.send(msg)
		if err != nil {
			break
		}

		if !p.lastPacket.IsZero() && time.Since(p.lastPacket) > time.Minute {
			err = p.cm.deletePeer(p.ID)
			if err != nil {
				// TODO: log this
			}
		}
	}
}

func (p *Peer) send(b []byte) error {
	_, err := p.cm.conn.WriteToUDPAddrPort(b, p.addr)
	if err != nil {
		return err
	}

	return nil
}

func (p *Peer) receive(blocking bool) []byte {
	if blocking {
		return <-p.inbox
	}

	select {
	case msg := <-p.inbox:
		return msg
	default:
	}

	return nil
}

func IKtPeerDisconnect(id PeerID) error {
	err := cm.deletePeer(id)
	if err != nil {
		return err
	}

	return nil
}

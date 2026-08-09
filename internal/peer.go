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
	TimedOut // TODO: return this at some point
)

type Peer struct {
	cm *ConnectionManager

	ID    SessionID
	State int

	addr  netip.AddrPort
	inbox chan []byte

	lastKeepAlive time.Time
}

func GetPeers() []*Peer {
	return cm.peers
}

// add a Peer to the ConnectionManager and connects to them
func (cm *ConnectionManager) AddPeer(id SessionID, addr netip.AddrPort) error {
	for _, peer := range cm.peers {
		if peer.ID != id {
			continue
		}

		return ErrPeerExists
	}

	peer := Peer{cm: cm, ID: id, addr: addr, inbox: make(chan []byte)}

	// add to peers so AddPeerMessage can write into its inbox
	cm.peers = append(cm.peers, &peer)

	err := peer.Connect()
	if err != nil {
		return err
	}

	return nil
}

func (cm *ConnectionManager) DeletePeer(id SessionID) error {
	for i, peer := range cm.peers {
		if peer.ID != id {
			continue
		}

		cm.peers = slices.Delete(cm.peers, i, i+1)

		close(peer.inbox)

		err := peer.Send(append([]byte(PeerMagic), PeerDisconnect))
		if err != nil {
			return err
		}

		return nil
	}

	return ErrPeerUnknown
}

// add a received message from the Peer to its inbox
func (cm *ConnectionManager) AddPeerMessage(addr netip.AddrPort, msg []byte) error {
	for _, peer := range cm.peers {
		if peer.addr != addr {
			continue
		}

		peer.inbox <- msg
		return nil
	}

	return ErrPeerUnknown
}

func (cm *ConnectionManager) GetPeerByID(id SessionID) (*Peer, error) {
	for _, peer := range cm.peers {
		if peer.ID != id {
			continue
		}

		return peer, nil
	}

	return nil, ErrPeerUnknown
}

func (cm *ConnectionManager) GetPeerByAddr(addr netip.AddrPort) (*Peer, error) {
	for _, peer := range cm.peers {
		if peer.addr != addr {
			continue
		}

		return peer, nil
	}

	return nil, ErrPeerUnknown
}

// establish a p2p connection to the Peer
func (p *Peer) Connect() error {
	if p.State != Uncontacted {
		return ErrPeerAlreadyConnecting
	}

	p.State = Connecting

	packetType := PeerConnect
	for {
		timeout := time.NewTimer(time.Second * 3)

		err := p.Send(append([]byte(PeerMagic), packetType))
		if err != nil {
			return err
		}

		if p.State == Connected {
			// PeerConnectAck sent, break now
			break
		}

		var recv []byte
		select {
		case in := <-p.inbox:
			recv = in
		case <-timeout.C:
			// do nothing, will be handled by prefix check
		}

		var found bool
		recv, found = bytes.CutPrefix(recv, []byte(PeerMagic))
		if !found {
			// unexpected
			continue
		}

		if len(recv) < 1 {
			// unexpected
			continue
		}

		switch recv[0] {
		case PeerConnect:
			p.State = ConnectingAck
			packetType = PeerConnectAck
		case PeerConnectAck:
			p.State = Connected
			// continue so PeerConnectAck can be sent
		}
	}

	go p.KeepAliveSender()

	return nil
}

// like ConnectionManager KeepAliveSender but for the Peer connection
func (p *Peer) KeepAliveSender() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		<-ticker.C
		err := p.Send(append([]byte(PeerMagic), PeerKeepAlive))
		if err != nil {
			break
		}
	}
}

func (p *Peer) Send(b []byte) error {
	_, err := p.cm.conn.WriteToUDPAddrPort(b, p.addr)
	if err != nil {
		return err
	}

	return nil
}

func (p *Peer) Receive() []byte {
	return <-p.inbox
}

func IKtPeerDisconnect(id string) error {
	var sid SessionID
	err := sid.FromString(id)
	if err != nil {
		return err
	}

	err = cm.DeletePeer(sid)
	if err != nil {
		return err
	}

	return nil
}

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

	ID    PeerID
	State int

	addr  netip.AddrPort
	inbox chan []byte

	lastKeepAlive time.Time
}

// add a Peer to the ConnectionManager and connects to them
func (cm *ConnectionManager) AddPeer(id PeerID, addr netip.AddrPort) error {
	for _, peer := range cm.peers {
		if peer.ID != id {
			continue
		}

		return ErrPeerExists
	}

	peer := Peer{cm: cm, ID: id, addr: addr, inbox: make(chan []byte, 32)}

	// add to peers so AddPeerMessage can write into its inbox
	cm.peers = append(cm.peers, &peer)

	err := peer.Connect()
	if err != nil {
		return err
	}

	return nil
}

func (cm *ConnectionManager) DeletePeer(id PeerID) error {
	var buf bytes.Buffer
	buf.WriteString(PeerMagic)
	buf.WriteByte(PeerDisconnect)

	for i, peer := range cm.peers {
		if peer.ID != id {
			continue
		}

		cm.peers = slices.Delete(cm.peers, i, i+1)

		close(peer.inbox)

		err := peer.Send(buf.Bytes())
		if err != nil {
			return err
		}

		return nil
	}

	return ErrPeerUnknown
}

func (cm *ConnectionManager) GetPeerByID(id PeerID) (*Peer, error) {
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

func (p *Peer) AddMessage(msg []byte) {
	select {
	case p.inbox <- msg:
	default:
		_ = <-p.inbox
		p.inbox <- msg
	}
}

// establish a p2p connection to the Peer
func (p *Peer) Connect() error {
	if p.State != Uncontacted {
		return ErrPeerAlreadyConnecting
	}

	p.State = Connecting
	packetType := PeerConnect

	var tries int
	for {
		var buf bytes.Buffer
		buf.WriteString(PeerMagic)
		buf.WriteByte(packetType)
		err := p.Send(buf.Bytes())
		if err != nil {
			return err
		}

		timeout := time.NewTimer(time.Second)

		if p.State == Connected {
			// PeerConnectAck sent, break now
			break
		}

		var data []byte
		select {
		case data = <-p.inbox:
		case <-timeout.C:
			if tries >= 5 {
				return ErrTimedOut
			}

			tries++
			// do nothing, will be handled by prefix check
		}

		var found bool
		data, found = bytes.CutPrefix(data, []byte(PeerMagic))
		if !found {
			// unexpected
			continue
		}

		if len(data) < 1 {
			// unexpected
			continue
		}

		switch data[0] {
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

	var buf bytes.Buffer
	buf.WriteString(PeerMagic)
	buf.WriteByte(PeerKeepAlive)

	for {
		<-ticker.C

		err := p.Send(buf.Bytes())
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

func (p *Peer) Receive(blocking bool) []byte {
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
	err := cm.DeletePeer(id)
	if err != nil {
		return err
	}

	return nil
}

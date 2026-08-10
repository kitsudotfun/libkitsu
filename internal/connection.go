package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/netip"
	"slices"
	"strconv"
	"time"

	. "github.com/kitsudotfun/natneg/defs"
)

var (
	ErrNotInitialized = errors.New("not initialized")
)

type ConnectionManager struct {
	conn *net.UDPConn

	natnegAddr  netip.AddrPort
	natnegInbox chan []byte

	peers []*Peer
}

func (cm *ConnectionManager) Init(natnegAddr string) error {
	var err error
	cm.conn, err = net.ListenUDP("udp4", nil)
	if err != nil {
		return err
	}

	host, port, err := net.SplitHostPort(natnegAddr)
	if err != nil {
		return err
	}
	addrs, err := net.DefaultResolver.LookupNetIP(context.Background(), "ip4", host)
	if err != nil {
		return err
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	cm.natnegAddr = netip.AddrPortFrom(addrs[0], uint16(portInt))

	go cm.Reader()

	var discover DiscoverResponse
	err = natnegCall(Discover, DiscoverRequest{Token: sessionToken}, &discover)
	if err != nil {
		return err
	}

	attestToken = discover.Token

	go cm.KeepAliveSender()

	return nil
}

func (cm *ConnectionManager) Shutdown() error {
	if ServerData == nil { // client
		// disconnect from peers (server) manually
		for _, peer := range cm.peers {
			err := cm.DeletePeer(peer.ID)
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

	close(cm.natnegInbox)

	return nil
}

func (cm *ConnectionManager) KeepAliveSender() {
	ticker := time.NewTicker(time.Second * 30)
	for {
		<-ticker.C
		err := natnegCall(KeepAlive, KeepAliveRequest{}, &KeepAliveResponse{})
		if err != nil {
			break
		}
	}
}

func (cm *ConnectionManager) Reader() {
	buf := make([]byte, 1400)
	for {
		n, addr, err := cm.conn.ReadFromUDPAddrPort(buf)
		if err != nil {
			break
		}

		data := slices.Clone(buf[:n])

		// NATNEG message
		b, natneg := bytes.CutPrefix(data, []byte(NatnegMagic))
		if natneg && addr == cm.natnegAddr && len(b) > 0 {
			if b[0] == JoinNotify {
				// TODO: handle error from this somehow
				go cm.handleJoinNotify(b[1:])
				continue
			}

			// this is a response that natnegCall is waiting on
			cm.natnegInbox <- b
			continue
		}

		// peer message
		peer, err := cm.GetPeerByAddr(addr)
		if err != nil {
			// TODO: log this
			continue
		}

		// check for internal message
		b, internal := bytes.CutPrefix(data, []byte(PeerMagic))
		if internal && len(b) > 0 {
			switch b[0] {
			case PeerKeepAlive:
				peer.lastKeepAlive = time.Now()

				var buf bytes.Buffer
				buf.WriteString(PeerMagic)
				buf.WriteByte(PeerKeepAliveAck)
				err = peer.Send(buf.Bytes())
				if err != nil {
					// TODO: log this
					continue
				}

				continue
			case PeerKeepAliveAck:
				// nothing to do
				continue
			case PeerDisconnect:
				err = cm.DeletePeer(peer.ID)
				if err != nil {
					// TODO: log this
					continue
				}

				continue
			}

			// other internal messages should be received by AddMessage
		}

		err = peer.AddMessage(data)
		if err != nil {
			// TODO: log this
			continue
		}
	}
}

func (cm *ConnectionManager) handleJoinNotify(data []byte) error {
	var notify JoinNotifyResponse
	err := json.Unmarshal(data, &notify)
	if err != nil {
		return err
	}

	err = cm.AddPeer(notify.ClientID, notify.ClientAddr)
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

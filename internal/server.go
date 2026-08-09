package internal

import (
	"time"

	. "github.com/kitsudotfun/kyuubi/api/defs"
)

var ServerData *Server

func IKtServerAnnounce(server Server) error {
	ServerData = &server

	err := sendHeartbeat()
	if err != nil {
		return err
	}

	return nil
}

func IKtServerShutdown() error {
	ServerData = nil

	err := apiCall("/server/delete", sessionToken, ServerDeleteRequest{}, &ServerDeleteResponse{})
	if err != nil {
		return err
	}

	for _, peer := range cm.peers {
		err = cm.DeletePeer(peer.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func HeartbeatSender() {
	ticker := time.NewTicker(time.Minute * 4)
	for {
		<-ticker.C
		if ServerData == nil {
			continue
		}

		sendHeartbeat()
	}
}

func sendHeartbeat() error {
	if ServerData == nil {
		return nil
	}

	err := apiCall("/server/heartbeat", sessionToken, ServerHeartbeatRequest{
		Server: *ServerData,
		Token:  attestToken,
	}, &ServerHeartbeatResponse{})
	if err != nil {
		return err
	}

	return nil
}

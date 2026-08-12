package internal

import (
	"time"

	. "github.com/kitsudotfun/kyuubi/api/defs"
)

var serverData *Server

func IKtServerAnnounce(server Server) error {
	serverData = &server

	err := sendHeartbeat()
	if err != nil {
		return err
	}

	return nil
}

func IKtServerShutdown() error {
	serverData = nil

	err := apiCall("/server/delete", sessionToken, ServerDeleteRequest{}, &ServerDeleteResponse{})
	if err != nil {
		return err
	}

	for _, peer := range cm.peers {
		err = cm.deletePeer(peer.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func heartbeatLoop() {
	for range time.NewTicker(time.Minute * 5).C {
		if serverData == nil {
			continue
		}

		sendHeartbeat()
	}
}

func sendHeartbeat() error {
	if serverData == nil {
		return nil
	}

	cm, err := GetCM()
	if err != nil {
		return err
	}

	serverData.Players = len(cm.GetPeers())

	err = apiCall("/server/heartbeat", sessionToken, ServerHeartbeatRequest{
		Server: *serverData,
		Token:  attestToken,
	}, &ServerHeartbeatResponse{})
	if err != nil {
		return err
	}

	return nil
}

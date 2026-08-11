package internal

import (
	. "github.com/kitsudotfun/kyuubi/api/defs"
	. "github.com/kitsudotfun/natneg/defs"
)

func IKtServerJoin(id PeerID, password string) error {
	var serverJoin ServerJoinResponse
	err := apiCall("/server/join", sessionToken, ServerJoinRequest{
		ServerID: id,
		Password: password,
	}, &serverJoin)
	if err != nil {
		return err
	}

	var join JoinResponse
	err = natnegCall(Join, JoinRequest{Token: serverJoin.Token}, &join)
	if err != nil {
		return err
	}

	err = cm.addPeer(join.ServerID, join.ServerAddr)
	if err != nil {
		return err
	}

	return nil
}

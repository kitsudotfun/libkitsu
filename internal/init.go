package internal

import (
	"encoding/base64"
	"errors"

	. "github.com/kitsudotfun/kyuubi/api/defs"
)

var (
	ErrMallocFailed = errors.New("malloc failed")
)

var (
	cm ConnectionManager

	sessionToken string
	attestToken  string
)

func IKtInit(game string, key string) (PeerID, error) {
	var new SessionNewResponse
	err := apiCall("/session/new", "", SessionNewRequest{GameID: game}, &new)
	if err != nil {
		return 0, err
	}

	proofKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return 0, err
	}

	var verify SessionVerifyResponse
	err = apiCall("/session/verify", "", SessionVerifyRequest{
		Token: new.Token,
		Proof: getProofSolution(proofKey, new.Salt[:], new.Difficulty),
	}, &verify)
	if err != nil {
		return 0, err
	}

	sessionToken = verify.Token

	cm = ConnectionManager{natnegAddr: verify.NatNegAddr, natnegInbox: make(chan []byte)}
	err = cm.init()
	if err != nil {
		return 0, err
	}

	go heartbeatLoop()

	return verify.ID, nil
}

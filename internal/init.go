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

func IKtInit(game string, key string) error {
	var new SessionNewResponse
	err := apiCall("/session/new", "", SessionNewRequest{GameID: game}, &new)
	if err != nil {
		return err
	}

	proofKey, err := base64.RawStdEncoding.DecodeString(key)
	if err != nil {
		return err
	}

	var verify SessionVerifyResponse
	err = apiCall("/session/verify", "", SessionVerifyRequest{
		Token: new.Token,
		Proof: getProofSolution(proofKey, new.Salt[:], new.Difficulty),
	}, &verify)
	if err != nil {
		return err
	}

	sessionToken = verify.Token

	cm = ConnectionManager{natnegInbox: make(chan []byte)}
	err = cm.Init(verify.NatNegServer)
	if err != nil {
		return err
	}

	go HeartbeatSender()

	return nil
}

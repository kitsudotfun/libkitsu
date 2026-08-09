package internal

import (
	. "github.com/kitsudotfun/kyuubi/api/defs"
)

func IKtServerList() ([]Server, error) {
	var list ServerListResponse
	err := apiCall("/server/list", sessionToken, ServerListRequest{}, &list)
	if err != nil {
		return nil, err
	}

	return list.Servers, nil
}

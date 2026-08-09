package internal

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"math"
	"math/bits"
	"net/http"
	"runtime"
	"slices"
	"sync"

	. "github.com/kitsudotfun/natneg/defs"
)

var (
	ErrUnexpectedResponse = errors.New("unexpected response")
)

const apiUrl = "https://kyuubi.kitsu.fun/dev"

func apiCall[reqT any, resT any](endpoint string, auth string, req reqT, res *resT) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(req)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", apiUrl+endpoint, &buf)
	if err != nil {
		return err
	}

	if auth != "" {
		request.Header.Set("Authorization", "Bearer "+auth)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return ErrNonOkStatus
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(res)
	if err != nil {
		return err
	}

	return nil
}

func natnegCall[reqT any, resT any](packetType byte, req reqT, res *resT) error {
	var buf bytes.Buffer
	buf.WriteByte(packetType)
	err := json.NewEncoder(&buf).Encode(req)
	if err != nil {
		return err
	}

	_, err = cm.conn.WriteToUDPAddrPort(buf.Bytes(), cm.natnegAddr)
	if err != nil {
		return err
	}

	data := <-cm.natnegInbox

	if data[0] != packetType {
		return ErrUnexpectedResponse
	}

	err = json.Unmarshal(data[1:], res)
	if err != nil {
		return err
	}

	return nil
}

func getProofSolution(key []byte, salt []byte, difficulty int) []byte {
	prefix := make([]byte, 0, len(key)+len(salt)+4)
	prefix = append(prefix, key...)
	prefix = append(prefix, salt...)

	prefixLen := len(prefix)

	prefix = append(prefix, make([]byte, 4)...)

	var wg sync.WaitGroup
	result := make(chan uint32, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workers := runtime.GOMAXPROCS(0)
	for worker := range workers {
		wg.Go(func() {
			buf := slices.Clone(prefix)

			for i := uint64(worker); i <= math.MaxUint32; i += uint64(workers) {
				select {
				case <-ctx.Done():
					return
				default:
				}

				binary.BigEndian.PutUint32(buf[prefixLen:], uint32(i))

				sum := sha256.Sum256(buf)
				if bits.LeadingZeros64(binary.BigEndian.Uint64(sum[:8])) < difficulty {
					continue
				}

				select {
				case result <- uint32(i):
					cancel()
				default:
				}
			}
		})
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	solution, ok := <-result
	if !ok {
		return nil
	}

	guess := make([]byte, 4)
	binary.BigEndian.PutUint32(guess, solution)

	return guess
}

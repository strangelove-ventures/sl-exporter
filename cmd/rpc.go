package cmd

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
)

type SyncInfo struct {
	LatestBlockHeight string `json:"latest_block_height"`
}

type Result struct {
	SyncInfo SyncInfo `json:"sync_info"`
}

type Response struct {
	Result Result `json:"result"`
}

// fetchRPCNodeHeight fetches node height from the rpcURL
func fetchRPCNodeHeight(rpcURL string) (uint64, error) {
	// Make a GET request to the REST endpoint
	resp, err := http.Get(rpcURL + "/status")
	if err != nil {
		return 0, fmt.Errorf("error making GET request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	limitedReader := &io.LimitedReader{R: resp.Body, N: 4 * 1024}
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return 0, fmt.Errorf("error reading response body: %v", err)
	}

	// Unmarshal JSON data into Response struct
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, fmt.Errorf("error unmarshaling JSON data: %v", err)
	}

	// Extract the latest_block_height as a number
	latestBlockHeightStr := response.Result.SyncInfo.LatestBlockHeight
	latestBlockHeight, err := strconv.ParseUint(latestBlockHeightStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error converting latest_block_height to a number: %v", err)
	}

	log.Debugf("Latest block height [%s]: %d\n", rpcURL, latestBlockHeight)

	return latestBlockHeight, nil
}

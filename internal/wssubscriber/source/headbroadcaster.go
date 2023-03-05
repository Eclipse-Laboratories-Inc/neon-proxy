package source

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/broadcaster"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
)

// error code is returned when specific slot is skipped, we check this and jump over skipped blocks during process
const (
	slotWasSkippedErrorCode = -32007
)

// define block header that we broadcast to users
type BlockHeader struct {
	Result struct {
		BlockHeight       int    `json:"blockHeight"`
		BlockTime         int    `json:"blockTime"`
		Blockhash         string `json:"blockhash"`
		ParentSlot        int    `json:"parentSlot"`
		PreviousBlockhash string `json:"previousBlockhash"`
	} `json:"result"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// RegisterNewHeadBroadcasterSources passes data and error channels where new incoming data (block heads) will be pushed and redirected to broadcaster
func RegisterNewHeadBroadcasterSources(ctx *context.Context, log logger.Logger, solanaWebsocketEndpoint string, broadcaster *broadcaster.Broadcaster) error {
	log.Info().Msg("block pulling from rpc started ... ")

	// declare sources to be set
	source := make(chan interface{})
	sourceError := make(chan error)

	// register given sources
	broadcaster.SetSources(source, sourceError)

	// subscribe to every result coming into the channel
	go func() {
		// store latest processed block slot
		var latestProcessedBlockSlot uint64

		// subscribe to every result coming into the channel
		for {
			// Calling Sleep method
			time.Sleep(1 * time.Second)

			// get latest block number
			slotResponse, err := jsonRPC([]byte(`{"jsonrpc":"2.0","id":1, "method":"getSlot", "params":[{"commitment":"finalized"}]}`), solanaWebsocketEndpoint, "POST")
			if err != nil {
				log.Error().Err(err).Msg("Error on rpc call for checking the latest slot on chain")
				sourceError <- err
				continue
			}

			// declare response json
			var blockSlot struct {
				Jsonrpc string `json:"jsonrpc"`
				Result  uint64 `json:"result"`
				ID      int    `json:"id"`
				Error   struct {
					Code    int    `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}

			// unrmarshal latest block slot
			err = json.Unmarshal([]byte(slotResponse), &blockSlot)
			if err != nil {
				sourceError <- err
				log.Error().Err(err).Msg("Error on unmarshaling slot response from rpc endpoint")
				continue
			}

			// error from rpc
			if blockSlot.Error.Message != "" {
				log.Error().Err(err).Msg("Error from rpc endpoint")
				sourceError <- err
				continue
			}

			// check the first iteration
			if latestProcessedBlockSlot == 0 {
				latestProcessedBlockSlot = blockSlot.Result
				continue
			}

			// process the blocks in given interval
			log.Info().Msg("processing blocks from " + strconv.FormatUint(latestProcessedBlockSlot, 10) + " to " + strconv.FormatUint(blockSlot.Result, 10))
			if err := processBlocks(ctx, solanaWebsocketEndpoint, log, source, sourceError, &latestProcessedBlockSlot, blockSlot.Result); err != nil {
				log.Error().Err(err).Msg("Error on processing blocks")
				sourceError <- err
				continue
			}
		}
	}()

	return nil
}

// request each block from "from" to "to" from the rpc endpoint and broadcast to user
func processBlocks(ctx *context.Context, solanaWebsocketEndpoint string, log logger.Logger, broadcaster chan interface{}, broadcasterErr chan error, from *uint64, to uint64) error {
	// process each block sequentially, broadcasting to peers
	for *from < to {
		// get block with given slot
		response, err := jsonRPC([]byte(`{
      "jsonrpc": "2.0","id":1,
      "method":"getBlock",
      "params": [
        `+strconv.FormatUint(*from, 10)+`,
        {
          "encoding": "json",
          "maxSupportedTransactionVersion":0,
          "transactionDetails":"none", "commitment" :"finalized",
          "rewards":false
        }
      ]
    }`), solanaWebsocketEndpoint, "POST")

		// check err
		if err != nil {
			return err
		}

		var blockHeader BlockHeader

		// unrmarshal latest block slot
		err = json.Unmarshal([]byte(response), &blockHeader)
		if err != nil {
			return err
		}

		// check rpc error
		if blockHeader.Error != nil {
			// check if the given slot was skipped if so we jump over it and continue processing blocks from the next slot
			if blockHeader.Error.Code == slotWasSkippedErrorCode {
				*from++
				continue
			}
			return errors.New(blockHeader.Error.Message)
		} else {
			// insert new block to be broadcasted
			clientResponse, err := json.Marshal(blockHeader.Result)

			// check json marshaling error
			if err != nil {
				log.Error().Err(err).Msg(fmt.Sprintf("marshalling response output: %v", err))
				return err
			}

			broadcaster <- clientResponse
			*from++
		}
	}

	return nil
}

// execute rpc call to the solana rpc endpoint
func jsonRPC(jsonStr []byte, url string, requestType string) ([]byte, error) {
	req, err := http.NewRequest(requestType, url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return body, nil
}

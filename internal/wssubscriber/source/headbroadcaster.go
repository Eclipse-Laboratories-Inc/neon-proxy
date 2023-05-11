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
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/utils"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
)

// error code is returned when specific slot is skipped, we check this and jump over skipped blocks during process
const (
	slotWasSkippedErrorCode    = -32007
	blockNotAvailableErrorCode = -32004
)

type BlockData struct {
	BlockHeight       string `json:"number"`
	BlockTime         string `json:"timestamp"`
	Blockhash         string `json:"hash"`
	PreviousBlockhash string `json:"parentHash"`
	Difficulty        string `json:"difficulty"`
	ExtraData         string `json:"extraData"`
	LogsBloom         string `json:"logsBloom"`
	GasLimit          string `json:"gasLimit"`
	TransactionsRoot  string `json:"transactionsRoot"`
	ReceiptsRoot      string `json:"receiptsRoot"`
	StateRoot         string `json:"stateRoot"`
	Sha3Uncles        string `json:"sha3Uncles"`
	Miner             string `json:"miner"`
	Nonce             string `json:"nonce"`
	MixHash           string `json:"mixHash"`
	GasUsed           string `json:"gasUsed"`
	Signature         string `json:"signature"`
}

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

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type BlockSlot struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  uint64 `json:"result"`
	ID      int    `json:"id"`
	Error   *Error `json:"error,omitempty"`
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

		// store latest processed block hash
		var lastProcessedBlockHash string

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
			var blockSlot BlockSlot

			// unrmarshal latest block slot
			err = json.Unmarshal(slotResponse, &blockSlot)
			if err != nil {
				log.Error().Err(err).Msg("Error on unmarshaling slot response from rpc endpoint")
				sourceError <- err
				continue
			}

			// error from rpc
			if blockSlot.Error != nil && blockSlot.Error.Message != "" {
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
			if err := processBlocks(ctx, solanaWebsocketEndpoint, log, source, sourceError, &lastProcessedBlockHash, &latestProcessedBlockSlot, blockSlot.Result); err != nil {
				log.Error().Err(err).Msg("Error on processing blocks")
				sourceError <- err
				continue
			}
		}
	}()

	return nil
}

// request each block from "from" to "to" from the rpc endpoint and broadcast to user
func processBlocks(ctx *context.Context, solanaWebsocketEndpoint string, log logger.Logger, broadcaster chan interface{}, broadcasterErr chan error, lastProcessedBlockHash *string, from *uint64, to uint64) error {
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

			if blockHeader.Error.Code == blockNotAvailableErrorCode {
				continue
			}
			return errors.New(blockHeader.Error.Message)
		} else {
			// marshal and send into broadcaster
			if err := MarshalBlockData(&blockHeader, broadcaster, log, *lastProcessedBlockHash); err != nil {
				return err
			} else {
				*from++
				*lastProcessedBlockHash = blockHeader.Result.Blockhash
			}

		}
	}

	return nil
}

// extract fields from block header and broadcast
func MarshalBlockData(blockHeader *BlockHeader, broadcaster chan interface{}, log logger.Logger, lastProcessedBlockHash string) error {
	// declare broadcasted struct
	var blockData BlockData

	// it's the first block we are processing so we skip as we don't have previous block hash
	if lastProcessedBlockHash == "" {
		return nil
	}

	// fill block data
	blockData.BlockHeight = "0x" + fmt.Sprintf("%x", blockHeader.Result.BlockHeight)
	blockData.BlockTime = "0x" + fmt.Sprintf("%x", blockHeader.Result.BlockTime)
	blockData.Blockhash = utils.Base64stringToHex(blockHeader.Result.Blockhash)
	blockData.Difficulty = "0x0"
	blockData.ExtraData = "0x0000000000000000000000000000000000000000000000000000000000000001"
	blockData.GasLimit = "0xec8563e271ac"
	blockData.GasUsed = "0xec8563e271ac" //TODO
	blockData.LogsBloom = "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	blockData.Miner = "0x0000000000000000000000000000000000000000"
	blockData.MixHash = "0x0000000000000000000000000000000000000000000000000000000000000001"
	blockData.Nonce = "0x0000000000000000"
	blockData.PreviousBlockhash = utils.Base64stringToHex(lastProcessedBlockHash)
	blockData.ReceiptsRoot = "0x0000000000000000000000000000000000000000000000000000000000000001"
	blockData.StateRoot = "0x0000000000000000000000000000000000000000000000000000000000000001"
	blockData.TransactionsRoot = "0x0000000000000000000000000000000000000000000000000000000000000001"
	blockData.Sha3Uncles = "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"
	blockData.Signature = "0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"

	// insert new block to be broadcasted
	clientResponse, err := json.Marshal(blockData)

	// check json marshaling error
	if err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("marshalling response output: %v", err))
		return err
	}
	broadcaster <- clientResponse
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

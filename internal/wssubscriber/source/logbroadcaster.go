package source

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/broadcaster"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"log"
	"strconv"
	"time"
)

type Block struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		BlockHeight       int         `json:"blockHeight"`
		BlockTime         interface{} `json:"blockTime"`
		Blockhash         string      `json:"blockhash"`
		ParentSlot        int         `json:"parentSlot"`
		PreviousBlockhash string      `json:"previousBlockhash"`
		Transactions      []struct {
			Meta struct {
				Err               interface{}   `json:"err"`
				Fee               int           `json:"fee"`
				InnerInstructions []interface{} `json:"innerInstructions"`
				LogMessages       []string      `json:"logMessages"`
				PostBalances      []int64       `json:"postBalances"`
				PostTokenBalances []interface{} `json:"postTokenBalances"`
				PreBalances       []int64       `json:"preBalances"`
				PreTokenBalances  []interface{} `json:"preTokenBalances"`
				Rewards           interface{}   `json:"rewards"`
				Status            struct {
					Ok interface{} `json:"Ok"`
				} `json:"status"`
			} `json:"meta"`
			Transaction struct {
				Message struct {
					AccountKeys []string `json:"accountKeys"`
					Header      struct {
						NumReadonlySignedAccounts   int `json:"numReadonlySignedAccounts"`
						NumReadonlyUnsignedAccounts int `json:"numReadonlyUnsignedAccounts"`
						NumRequiredSignatures       int `json:"numRequiredSignatures"`
					} `json:"header"`
					Instructions []struct {
						Accounts       []int  `json:"accounts"`
						Data           string `json:"data"`
						ProgramIdIndex int    `json:"programIdIndex"`
					} `json:"instructions"`
					RecentBlockhash string `json:"recentBlockhash"`
				} `json:"message"`
				Signatures []string `json:"signatures"`
			} `json:"transaction"`
		} `json:"transactions"`
	} `json:"result"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
	Id int `json:"id"`
}

// RegisterNewHeadBroadcasterSources passes data and error channels where new incoming data (block heads) will be pushed and redirected to broadcaster
func RegisterLogsBroadcasterSources(ctx *context.Context, log logger.Logger, endpoint string, broadcaster *broadcaster.Broadcaster) error {
	log.Info().Msg("pending transaction pulling from mempool started ... ")

	// declare sources to be set
	source := make(chan interface{})
	sourceError := make(chan error)

	// register given sources
	broadcaster.SetSources(source, sourceError)

	go func() {
		var latestProcessedBlockSlot uint64
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {

			// get the latest block number
			slotResponse, err := jsonRPC([]byte(`{"jsonrpc":"2.0","id":1, "method":"getSlot", "params":[{"commitment":"finalized"}]}`), endpoint, "POST")
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
			err = json.Unmarshal(slotResponse, &blockSlot)
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

			log.Info().Msg("logParser: processing blocks from " + strconv.FormatUint(latestProcessedBlockSlot, 10) + " to " + strconv.FormatUint(blockSlot.Result, 10))
			err = getLogsFromBlocks(endpoint, source, &latestProcessedBlockSlot, blockSlot.Result)
			// check unmarshaling error
			if err != nil {
				log.Error().Err(err).Msg("Error on processing blocks")
				sourceError <- err
				continue
			}
		}
	}()

	return nil
}

func getLogsFromBlocks(solanaWebsocketEndpoint string, logParser chan interface{}, from *uint64, to uint64) error {
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
          "transactionDetails":"full", "commitment" :"finalized",
          "rewards":false
        }
      ]
    }`), solanaWebsocketEndpoint, "POST")

		// check err
		if err != nil {
			return err
		}

		var block Block

		// unrmarshal latest block slot
		if err = json.Unmarshal([]byte(response), &block); err != nil {
			return err
		}

		// check rpc error
		if block.Error != nil {
			// check if the given slot was skipped if so we jump over it and continue processing blocks from the next slot
			if block.Error.Code == slotWasSkippedErrorCode {
				*from++
				continue
			}
			return errors.New(block.Error.Message)
		} else {
			logs := removeDuplicates(block)

			for _, l := range logs {
				log.Println(l)
			}

			clientResponse, err := json.Marshal(logs)
			if err != nil {
				return err
			}
			logParser <- clientResponse
			*from++
		}
	}

	return nil
}

func removeDuplicates(block Block) []string {
	logsMap := make(map[string]struct{}, 0)
	for _, tx := range block.Result.Transactions {
		for _, log := range tx.Meta.LogMessages {
			logsMap[log] = struct{}{}
		}
	}

	logs := make([]string, len(logsMap))
	var i int64
	for log := range logsMap {
		logs[i] = log
		i++
	}
	return logs
}

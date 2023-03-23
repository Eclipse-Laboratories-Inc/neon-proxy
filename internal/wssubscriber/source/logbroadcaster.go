package source

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/broadcaster"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"strconv"
	"time"
)

// RegisterNewHeadBroadcasterSources passes data and error channels where new incoming data (block heads) will be pushed and redirected to broadcaster
func RegisterLogsBroadcasterSources(_ *context.Context, log logger.Logger, endpoint string, broadcaster *broadcaster.Broadcaster) error {
	log.Info().Msg("pending transaction pulling from mempool started ... ")

	// declare sources to be set
	logsSource := make(chan interface{})
	logsSourceError := make(chan error)

	// register given sources
	broadcaster.SetSources(logsSource, logsSourceError)

	go func() {
		var latestProcessedBlockSlot uint64
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {

			// get the latest block number
			slotResponse, err := jsonRPC([]byte(`{"jsonrpc":"2.0","id":1, "method":"getSlot", "params":[{"commitment":"finalized"}]}`), endpoint, "POST")
			if err != nil {
				log.Error().Err(err).Msg("Error on rpc call for checking the latest slot on chain")
				logsSourceError <- err
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
				log.Error().Err(err).Msg("Error on unmarshaling slot response from rpc endpoint")
				logsSourceError <- err
				continue
			}

			// error from rpc
			if blockSlot.Error.Message != "" {
				log.Error().Err(err).Msg("Error from rpc endpoint")
				logsSourceError <- err
				continue
			}

			// check the first iteration
			if latestProcessedBlockSlot == 0 {
				latestProcessedBlockSlot = blockSlot.Result
				continue
			}

			log.Info().Msg("logParser: processing blocks from " + strconv.FormatUint(latestProcessedBlockSlot, 10) + " to " + strconv.FormatUint(blockSlot.Result, 10))
			if err := getLogsFromBlocks(endpoint, logsSource, &latestProcessedBlockSlot, blockSlot.Result); err != nil {
				log.Error().Err(err).Msg("Error on processing blocks")
				logsSourceError <- err
			}
		}
	}()

	return nil
}

func getLogsFromBlocks(solanaWebsocketEndpoint string, logParserSource chan interface{}, from *uint64, to uint64) error {
	// process each block sequentially, broadcasting to peers
	// using "transactionDetails":"full", we are getting full info about finalized executed transactions
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
          "transactionDetails":"full", 
          "commitment" :"finalized",
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
		if err = json.Unmarshal(response, &block); err != nil {
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
			clientResponse, err := json.Marshal(logs)
			if err != nil {
				return err
			}
			logParserSource <- clientResponse

			if err := getELogs(getTransactionsSignatures(block)); err != nil {
				return err
			}
			*from++
		}
	}

	return nil
}

// getting transaction signatures
// https://docs.solana.com/api/http#transaction-structure
func getTransactionsSignatures(block Block) TransactionSignaturesResponse {
	txs := block.Result.Transactions
	resp := TransactionSignaturesResponse{}
	for _, tx := range txs {
		resp.Signatures = append(resp.Signatures, tx.Transaction.Signatures...)
	}

	return resp
}

// TODO implement getting e-logs using transaction signatures
func getELogs(signatures TransactionSignaturesResponse) error {
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

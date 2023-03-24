package source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

			var blockSlot BlockSlot

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

			signatures, err := getTransactionsSignatures(solanaWebsocketEndpoint, *from, block)
			if err != nil {
				return err
			}
			if err := getELogs(signatures); err != nil {
				return err
			}
			*from++
		}
	}

	return nil
}

// getting transaction signatures
// https://docs.solana.com/api/http#transaction-structure
func getTransactionsSignatures(solanaWebsocketEndpoint string, fromSlot uint64, block Block) (*TransactionSignaturesResponse, error) {
	accountKeys := map[string]struct{}{}
	txs := block.Result.Transactions
	resp := &TransactionSignaturesResponse{}

	for _, tx := range txs {
		for _, key := range tx.Transaction.Message.AccountKeys {
			accountKeys[key] = struct{}{}
		}
	}

	for key, _ := range accountKeys {
		var txSignaturesFromSlot GetTransactionSignatureByAccountKeyResp

		fmt.Println(key)
		req := `{
		  "jsonrpc": "2.0","id":1,
		  "method":"getSignaturesForAddress",
		  "params": [
		    "` + key + `",
		    {
		      "commitment" :"finalized"
		    }
		  ]
		}`

		response, err := jsonRPC([]byte(req), solanaWebsocketEndpoint, "POST")
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(response, &txSignaturesFromSlot); err != nil {
			return nil, err
		}

		if txSignaturesFromSlot.Error != nil {
			return nil, errors.New(txSignaturesFromSlot.Error.Message)
		}

		for _, res := range txSignaturesFromSlot.Result {
			// transactions are already sorted from newest to oldest
			// if slot is smaller, than current, than this tx is too old
			if res.Slot < int64(fromSlot) {
				break
			}
			if res.Slot == int64(fromSlot) {
				resp.Signatures = append(resp.Signatures, res.Signature)
			}
		}
	}

	return resp, nil
}

// TODO implement getting e-logs using transaction signatures
func getELogs(signatures *TransactionSignaturesResponse) error {
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

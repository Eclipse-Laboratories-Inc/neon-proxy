package source

import (
	"context"
	"encoding/json"
	"strconv"
	"errors"
	"fmt"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/broadcaster"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/utils"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"time"
)

const MaxProcessedTransactionsBatch = 100

// eth event structure
type Event struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
}

// defines eth log structure for each transaction
type EthLog struct {
	BlockNumber      string   `json:blockNumber`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	BlockHash        string   `json:"blockHash"`
	Event
}

// RegisterLogsBroadcasterSources passes data and error channels where new incoming data (transaction logs) will be pushed and redirected to broadcaster
func RegisterLogsBroadcasterSources(ctx *context.Context, log logger.Logger, solanaWebsocketEndpoint, evmAddress string, broadcaster *broadcaster.Broadcaster) error {
	log.Info().Msg("logs pulling from evm address started ... ")

	// declare sources to be set
	logsSource := make(chan interface{})
	logsSourceError := make(chan error)

	// register given sources
	broadcaster.SetSources(logsSource, logsSourceError)

	// store latest processed block slot
	var latestProcessedBlockSlot uint64

	go func() {
		// declare time interval for checking new transactions
		ticker := time.NewTicker(1 * time.Second)

		for range ticker.C {
			log.Info().Msg("logParser: latest processed block slot signature " + fmt.Sprintf("%v", latestProcessedBlockSlot))

			// get latest block number
			slotResponse, err := jsonRPC([]byte(`{"jsonrpc":"2.0","id":1, "method":"getSlot", "params":[{"commitment":"finalized"}]}`), solanaWebsocketEndpoint, "POST")
			if err != nil {
				log.Error().Err(err).Msg("Error on rpc call for checking the latest slot on chain")
				logsSourceError <- err
				continue
			}

			// declare response json
			var blockSlot BlockSlot

			// unrmarshal latest block slot
			err = json.Unmarshal(slotResponse, &blockSlot)
			if err != nil {
				log.Error().Err(err).Msg("Error on unmarshaling slot response from rpc endpoint")
				logsSourceError <- err
				continue
			}

			// error from rpc
			if blockSlot.Error != nil && blockSlot.Error.Message != "" {
				log.Error().Err(err).Msg("Error from rpc endpoint")
				logsSourceError <- err
				continue
			}

			// check the first iteration
			if latestProcessedBlockSlot == 0 {
				latestProcessedBlockSlot = blockSlot.Result
				continue
			}

			// process the blocks in given interval
			log.Info().Msg("processing blocks from " + strconv.FormatUint(latestProcessedBlockSlot, 10) + " to " + strconv.FormatUint(blockSlot.Result, 10))
			if err := processBlockLogs(ctx, solanaWebsocketEndpoint, log, logsSource, logsSourceError, &latestProcessedBlockSlot, blockSlot.Result); err != nil {
				log.Error().Err(err).Msg("Error on processing blocks")
				logsSourceError <- err
				continue
			}
		}
	}()
	return nil
}

// request each block from "from" to "to" from the rpc endpoint and broadcast to user
func processBlockLogs(ctx *context.Context, solanaWebsocketEndpoint string, log logger.Logger, logsSource chan interface{}, logsSourceError chan error, from *uint64, to uint64) error {
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
		err = json.Unmarshal([]byte(response), &block)
		if err != nil {
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
		}

		// get eth logs from block
		ethlogs, err := getEthLogsFromBlock(&block, log, solanaWebsocketEndpoint)
		if err != nil {
			return err
		}

		// marshall each log and push into broadcaster
		for _, ethLog := range ethlogs {
			// insert new log to be broadcasted
			clientResponse, err := json.Marshal(ethLog)

			// check json marshaling error
			if err != nil {
				log.Error().Err(err).Msg(fmt.Sprintf("marshalling response output: %v", err))
				return err
			}

			logsSource <- clientResponse
		}
	}

	return nil
}

/*
For each given solana transaction signature we request the whole transaction object
from which we parse instruction data where eth transaction parameters (including logs) are stored
*/
func getEthLogsFromBlock(block *Block, log logger.Logger, solanaWebsocketEndpoint string) ([]EthLog, error) {
	// create response array of logs
	ethLogs := make([]EthLog, 0)

	// define log index value
	var logIndex int

	// process each transaction logs separately
	for k := len(block.Result.Transactions) - 1; k >= 0; k-- {
		// declare log instance to fill
		var ethLog EthLog

		// fill initial log values
		ethLog.BlockHash = utils.Base64stringToHex(block.Result.Blockhash)
		ethLog.BlockNumber = fmt.Sprintf("0x%x", block.Result.BlockHeight)
		ethLog.TransactionHash = block.Result.Transactions[k].Transaction.Signatures[0]
		ethLog.TransactionIndex = fmt.Sprintf("0x%x", len(block.Result.Transactions) - k - 1)

		// parse events from solana transaction logs
		events, err := getEvents(block.Result.Transactions[k].Meta.LogMessages, log, solanaWebsocketEndpoint)
		if err != nil {
			return nil, err
		}

		// fill events
		for _, event := range events {
			ethLog.Address = event.Address
			ethLog.Data = event.Data
			ethLog.LogIndex = fmt.Sprintf("0x%x", logIndex)
			// increment index for the next log/event
			logIndex++
			ethLog.Removed = event.Removed
			ethLog.Topics = event.Topics

			// append new log
			ethLogs = append(ethLogs, ethLog)
		}

	}
	// given log messages from solaan tx we parse eth logs
	return ethLogs, nil
}

// parses solana tx log messages and builds eth events
func getEvents(logMessages []string, log logger.Logger, solanaWebsocketEndpoint string) ([]Event, error) {
	return nil, nil
}

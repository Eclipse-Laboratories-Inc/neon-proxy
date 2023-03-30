package source

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/broadcaster"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/utils"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// Defines log internal type values
const (
	// log type value
	Log = 1

	// call numbers
	EnterCall         = 101
	EnterCallCode     = 102
	EnterStaticCall   = 103
	EnterDelegateCall = 104
	EnterCreate       = 105
	EnterCreate2      = 106

	// instruction values
	ExitStop         = 201
	ExitReturn       = 202
	ExitSelfDestruct = 203
	ExitRevert       = 204

	// result codes
	Return = 300
	Cancel = 301
)

// eth event log data parsed from solaan logs
type NeonLogTxEvent struct {
	address   []byte
	topicList []string
	data      []byte
}

// eth event structure
type Event struct {
	TransactionHash string   `json:"transactionHash"`
	Address         string   `json:"address"`
	Topics          []string `json:"topics"`
	Data            string   `json:"data"`
	LogIndex        string   `json:"logIndex"`
}

// defines eth log structure for each transaction
type EthLog struct {
	BlockNumber      string `json:blockNumber`
	TransactionIndex string `json:"transactionIndex"`
	BlockHash        string `json:"blockHash"`
	Event
}

// gas usage info parsed from logs
type NeonLogTxIx struct {
	gasUsed      int
	totalGasUsed int
}

// log return status
type NeonLogTxReturn struct {
	gasUsed int
	status  int
}

// RegisterLogsBroadcasterSources passes data and error channels where new incoming data (transaction logs) will be pushed and redirected to broadcaster
func RegisterLogsBroadcasterSources(ctx *context.Context, log logger.Logger, solanaWebsocketEndpoint, evmAddress string, broadcaster *broadcaster.Broadcaster) error {
	log.Info().Msg("logs pulling from blocks started ... ")

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
			log.Info().Msg("processing logs from block " + strconv.FormatUint(latestProcessedBlockSlot, 10) + " to " + strconv.FormatUint(blockSlot.Result, 10))
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
		response, err := jsonRPC([]byte(`{"jsonrpc": "2.0","id":1, "method":"getBlock", "params": [`+strconv.FormatUint(*from, 10)+`,{"encoding": "json", "maxSupportedTransactionVersion":0, "transactionDetails":"full", "commitment" :"finalized", "rewards":false}]}`), solanaWebsocketEndpoint, "POST")

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
		ethlogs, err := getEthLogsFromBlock(&block, from, log, solanaWebsocketEndpoint)
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

		// move to next block
		*from++
	}

	return nil
}

/*
For each given solana transaction signature we request the whole transaction object
from which we parse instruction data where eth transaction parameters (including logs) are stored
*/
func getEthLogsFromBlock(block *Block, blockNumber *uint64, log logger.Logger, solanaWebsocketEndpoint string) ([]EthLog, error) {
	// create response array of logs
	ethLogs := make([]EthLog, 0)

	// define log index value
	var logIndex int
	var transactionIndex int

	// process each transaction logs separately
	for k := len(block.Result.Transactions) - 1; k >= 0; k-- {
		// declare log instance to fill
		var ethLog EthLog
		// fill initial log values
		ethLog.BlockHash = utils.Base58stringToHex(block.Result.Blockhash)
		ethLog.BlockNumber = fmt.Sprintf("0x%x", int(*blockNumber))
		ethLog.TransactionIndex = fmt.Sprintf("0x%x", transactionIndex)
		// parse events from solana transaction logs
		events, err := GetEvents(block.Result.Transactions[k].Meta.LogMessages, &logIndex, log, solanaWebsocketEndpoint)
		if err != nil {
			return nil, err
		}

		// if we have eth transaction increase index
		if len(events) != 0 {
			transactionIndex++
		}

		// fill events
		for _, event := range events {
			ethLog.Address = event.Address
			ethLog.TransactionHash = event.TransactionHash
			ethLog.Data = event.Data
			ethLog.LogIndex = event.LogIndex
			// increment index for the next log/event
			ethLog.Topics = event.Topics

			// append new log
			ethLogs = append(ethLogs, ethLog)
		}
	}

	// given log messages from solaan tx we parse eth logs
	return ethLogs, nil
}

// decodes each solana tx log line
func decodeMnemonic(line string) (string, []string) {
	// split line into separate instruction words
	words := strings.Fields(line)
	if len(words) < 4 || words[0] != "Program" || words[1] != "data:" {
		return "", nil
	}

	// return line function and remaining data
	return utils.Base64stringDecodeToString(words[2]), words[3:]
}

// Decode eth transaction hash from solana logs
func DecodeNeonTxSig(dataList []string) ([]byte, error) {
	// data list for transaction hash must be 1 element
	if len(dataList) != 1 {
		return nil, errors.New("Failed to decode neon tx hash: dataList should be 1 element")
	}

	// decode transaction hash
	txSignature := utils.Base64stringToBytes(dataList[0])
	if len(txSignature) != 32 {
		return nil, errors.New("Failed to decode neon tx hash: wrong hash length")
	}

	return txSignature, nil
}

// parse exit code of the eth transaction
func DecodeNeonTxReturn(neonTxIx *NeonLogTxIx, dataList []string) (*NeonLogTxReturn, error) {
	// data list must have at lest 1 element
	if len(dataList) < 1 {
		return nil, errors.New("Failed to decode return data: less then 1 elements in data list")
	}

	// parse exit code
	var exitStatus byte
	err := binary.Read(bytes.NewReader(utils.Base64stringToBytes(dataList[0])), binary.LittleEndian, &exitStatus)
	if err != nil {
		return nil, err
	}

	// set exit code
	if exitStatus < 0xd0 {
		exitStatus = 0x1
	} else {
		exitStatus = 0x0
	}

	// check we have parsed gas usage info
	if neonTxIx == nil {
		return nil, errors.New("Total gas should be parsed before checking return value")
	}

	return &NeonLogTxReturn{gasUsed: neonTxIx.totalGasUsed, status: int(exitStatus)}, nil
}

// Decode gas usage of eth transaction
func DecodeNeonTxGas(dataList []string) (*NeonLogTxIx, error) {
	// data list must have 2 elements
	if len(dataList) != 2 {
		return nil, errors.New("Failed to decode neon ix gas : should be 2 element in datalist")
	}

	// decode used gas amount
	gasUsedInt := new(big.Int)
	gasUsedInt.SetString(utils.Base64stringToHex(dataList[0])[2:], 16)

	// decode total gas usage
	totalGasUsedInt := new(big.Int)
	totalGasUsedInt.SetString(utils.Base64stringToHex(dataList[0])[2:], 16)

	return &NeonLogTxIx{gasUsed: int(gasUsedInt.Int64()), totalGasUsed: int(totalGasUsedInt.Int64())}, nil
}

// decodes transaction event, returning address, topic list and data
func DecodeNeonTxEvent(logNum int, dataList []string) (*NeonLogTxEvent, error) {
	// declare return data
	var neonLogTxEvent NeonLogTxEvent
	// prepare topic list for parsed topics
	neonLogTxEvent.topicList = make([]string, 0)

	// check data list contains at least 3 items
	if len(dataList) < 3 {
		return nil, errors.New("Failed to decode events data: less 3 elements in datalist")
	}

	// check log num validity
	if logNum > 4 || logNum < 0 {
		return nil, errors.New("Failed to decode events data: count of topics not correct")
	}

	// check topic count and logNum to be equal
	topicCount := new(big.Int)
	topicCount.SetString(utils.Base64stringToHex(dataList[1])[2:], 16)
	if topicCount.Int64() != int64(logNum) {
		return nil, errors.New("Topic count and Lognum are not equal")
	}

	// decode address from data list
	neonLogTxEvent.address = utils.Base64stringToBytes(dataList[0])

	// decode each topic and insert into event data
	for k := 0; k < int(topicCount.Int64()); k++ {
		neonLogTxEvent.topicList = append(neonLogTxEvent.topicList, "0x"+hex.EncodeToString(utils.Base64stringToBytes(dataList[2+k])))
	}

	// decode log data
	if 2+int(topicCount.Int64()) < len(dataList) {
		neonLogTxEvent.data = utils.Base64stringToBytes(dataList[2+int(topicCount.Int64())])
	}

	return &neonLogTxEvent, nil
}

// parses solana tx log messages and builds eth events
func GetEvents(logMessages []string, logIndex *int, log logger.Logger, solanaWebsocketEndpoint string) ([]Event, error) {
	events := make([]Event, 0)
	// for parsing eth transaction hash
	var neonTxHash []byte
	// for parsed gas usage data
	var neonTxIx *NeonLogTxIx
	// for parsed transaction result code
	var neonTxReturn *NeonLogTxReturn
	// error decoding logs
	var err error
	neonTxEventList := make([]NeonLogTxEvent, 0)

	// for each solana transaction log message decode eth data
	for _, line := range logMessages {
		// decode log line
		name, dataList := decodeMnemonic(line)
		if len(name) == 0 {
			continue
		}

		// Use switch on the day variable.
		switch {
		case name == "HASH":
			if neonTxHash == nil {
				neonTxHash, err = DecodeNeonTxSig(dataList)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("transaction HASH encountered twice")
			}
		case name == "RETURN":
			if neonTxReturn == nil {
				neonTxReturn, err = DecodeNeonTxReturn(neonTxIx, dataList)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("transaction RETURN encountered twice")
			}
		case name == "GAS":
			if neonTxIx == nil {
				neonTxIx, err = DecodeNeonTxGas(dataList)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("transaction GAS encountered twice")
			}
		case len(name) >= 3 && name[0:3] == "LOG":
			logNum, err := strconv.Atoi(name[3:])
			if err != nil {
				return nil, err
			}
			// decode eth event log
			neonTxEvent, err := DecodeNeonTxEvent(logNum, dataList)
			if err != nil {
				return nil, err
			}
			// add new log to client response data
			neonTxEventList = append(neonTxEventList, *neonTxEvent)
		default:
			continue
		}
	}

	// loop through all the events from logs and only return contract event logs
	for _, e := range neonTxEventList {
		// set event field values
		events = append(events, Event{Address: "0x" + hex.EncodeToString(e.address), Data: "0x" + hex.EncodeToString(e.data), LogIndex: fmt.Sprintf("0x%x", *logIndex), Topics: e.topicList, TransactionHash: "0x" + hex.EncodeToString(neonTxHash)})

		// increase global counter for log index
		*logIndex++
	}

	return events, nil
}

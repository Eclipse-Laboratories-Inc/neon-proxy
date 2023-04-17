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

	// prepare logs for checking evm enter/exit points
	EvmInvocationLog        = "Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU invoke"
	EvmInvocationSuccessEnd = "Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU success"
	EvmInvocationFailEnd    = "Program eeLSJgWzzxrqKv1UxtRVVH8FX3qCQWUs9QuAjJpETGU fail"
)

// eth event log data parsed from solaan logs
type NeonLogTxEvent struct {
	eventType       int
	transactionHash []byte
	address         []byte
	topicList       []string
	data            []byte
	callLevel       int
}

// current log and transaction index
type Indexes struct {
	transactionIndex int
	logIndex         int
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
	var latestProcessedTransaction string

	go func() {
		// declare time interval for checking new transactions
		ticker := time.NewTicker(1 * time.Second)

		for range ticker.C {
			log.Info().Msg("logParser: latest processed transactionsignature " + latestProcessedTransaction)

			// for saving current tx and log indexes
			indexes := make(map[int]Indexes)

			// prepare curl parameter
			var untilParameter string
			if latestProcessedTransaction != "" {
				untilParameter = ", \"until\": \"" + latestProcessedTransaction + "\""
			}

			// get latest block number
			signaturesResponse, err := jsonRPC([]byte(`{"jsonrpc":"2.0","id":1, "method": "getSignaturesForAddress", "params":["`+evmAddress+`",{"commitment":"finalized" `+untilParameter+`}]}`), solanaWebsocketEndpoint, "POST")
			if err != nil {
				log.Error().Err(err).Msg("Error on rpc call for checking the latest slot on chain")
				logsSourceError <- err
				continue
			}

			// declare response json
			var transactionSignatures TransactionSignatures

			// unrmarshal latest block slot
			err = json.Unmarshal(signaturesResponse, &transactionSignatures)
			if err != nil {
				log.Error().Err(err).Msg("Error on unmarshaling slot response from rpc endpoint")
				logsSourceError <- err
				continue
			}

			// error from rpc
			if transactionSignatures.Error != nil && transactionSignatures.Error.Message != "" {
				log.Error().Err(err).Msg(transactionSignatures.Error.Message)
				logsSourceError <- err
				continue
			}

			// process the blocks in given interval
			log.Info().Msg("processing logs")
			if err := processTransactionLogs(ctx, solanaWebsocketEndpoint, log, logsSource, logsSourceError, &transactionSignatures, indexes); err != nil {
				log.Error().Err(err).Msg("Error on processing tx logs")
				logsSourceError <- err
				continue
			}

			// check the first iteration
			if len(transactionSignatures.Result) > 0 {
				latestProcessedTransaction = transactionSignatures.Result[0].Signature
			}

		}
	}()
	return nil
}

// request each block from "from" to "to" from the rpc endpoint and broadcast to user
func processTransactionLogs(ctx *context.Context, solanaWebsocketEndpoint string, log logger.Logger, logsSource chan interface{}, logsSourceError chan error, signatures *TransactionSignatures, indexes map[int]Indexes) error {
	// process each block sequentially, broadcasting to peers
	for k := len(signatures.Result) - 1; k >= 0; k-- {
		// get block with given slot
		response, err := jsonRPC([]byte(`{"jsonrpc": "2.0","id":1, "method":"getTransaction", "params": ["`+signatures.Result[k].Signature+`", {"encoding":"json", "maxSupportedTransactionVersion":0}]}`), solanaWebsocketEndpoint, "POST")

		// check err
		if err != nil {
			return err
		}

		var transaction Transaction

		// unrmarshal transaction data
		err = json.Unmarshal([]byte(response), &transaction)
		if err != nil {
			return err
		}

		// check rpc error
		if transaction.Error != nil {
			return errors.New(transaction.Error.Message)
		}

		// get eth logs from block
		ethlogs, err := getEthLogsFromTransaction(&transaction, indexes, log, solanaWebsocketEndpoint)
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

			fmt.Println(string(clientResponse))
			logsSource <- clientResponse
		}
	}

	return nil
}

/*
For each given solana transaction signature we request the whole transaction object
from which we parse instruction data where eth transaction parameters (including logs) are stored
*/
func getEthLogsFromTransaction(transaction *Transaction, indexes map[int]Indexes, log logger.Logger, solanaWebsocketEndpoint string) ([]EthLog, error) {
	// create response array of logs
	ethLogs := make([]EthLog, 0)

	// check if transaction failed
	if transaction.Result.Meta.Err != nil {
		// update tx index
		inds := indexes[transaction.Result.Slot]
		inds.transactionIndex++
		return ethLogs, nil
	}

	// declare log instance to fill
	var ethLog EthLog
	var err error

	// fill initial log values
	ethLog.BlockHash, err = GetBlockHash(transaction.Result.Slot, solanaWebsocketEndpoint)
	if err != nil {
		return nil, err
	}

	ethLog.BlockNumber = fmt.Sprintf("0x%x", transaction.Result.Slot)

	// parse events from solana transaction logs
	events, err := GetEvents(transaction.Result.Meta.LogMessages)
	if err != nil {
		return nil, err
	}

	ethLog.TransactionIndex = fmt.Sprintf("0x%x", indexes[transaction.Result.Slot].transactionIndex)

	// update tx index
	txinds := indexes[transaction.Result.Slot]
	txinds.transactionIndex++
	indexes[transaction.Result.Slot] = txinds

	// fill events
	for _, event := range events {
		ethLog.Address = event.Address
		ethLog.TransactionHash = event.TransactionHash
		ethLog.Data = event.Data
		ethLog.LogIndex = fmt.Sprintf("0x%x", indexes[transaction.Result.Slot].logIndex)

		// update log index
		inds := indexes[transaction.Result.Slot]
		inds.logIndex++
		indexes[transaction.Result.Slot] = inds

		// increment index for the next log/event
		ethLog.Topics = event.Topics

		// append new log
		ethLogs = append(ethLogs, ethLog)
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

// decode some non-evm program enter sites
func isNonEvmProgramInvoke(line string) bool {
	// split line into separate instruction words
	words := strings.Fields(line)
	if len(words) == 4 && words[0] == "Program" && words[2] == "invoke" && words[3][0] == '[' {
		return true
	}
	return false
}

// decode some non-evm program exit sites
func isNonEvmProgramExit(line string) bool {
	// split line into separate instruction words
	words := strings.Fields(line)
	if len(words) >= 3 && words[0] == "Program" && (words[2] == "success" || (len(words[2]) >=4 && words[2][0:4] == "fail")) {
		return true
	}
	return false
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
	if exitStatus > 0xd0 {
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

// Decode enter call from logs
func DecodeNeonTxEnter(dataList []string) (*NeonLogTxEvent, error) {
	// data list must have 2 elements
	if len(dataList) != 2 {
		return nil, errors.New("Failed to decode enter event, it should contain 2 elements")
	}

	// decode address
	address := utils.Base64stringToBytes(dataList[1])
	if len(address) != 20 {
		return nil, errors.New("Failed to decode enter event, address has wrong length")
	}

	// decode call type
	typeName := utils.Base64stringDecodeToString(dataList[0])

	// check call type
	switch {
	case typeName == "CALL":
		return &NeonLogTxEvent{eventType: EnterCall, address: address}, nil
	case typeName == "CALLCODE":
		return &NeonLogTxEvent{eventType: EnterCall, address: address}, nil
	case typeName == "STATICCALL":
		return &NeonLogTxEvent{eventType: EnterStaticCall, address: address}, nil
	case typeName == "DELEGATECALL":
		return &NeonLogTxEvent{eventType: EnterDelegateCall, address: address}, nil
	case typeName == "CREATE":
		return &NeonLogTxEvent{eventType: EnterCreate, address: address}, nil
	case typeName == "CREATE2":
		return &NeonLogTxEvent{eventType: EnterCreate2, address: address}, nil
	default:
		return nil, errors.New("Failed to decode enter event, wrong type")
	}
}

// decode exit code of eth function call
func DecodeNeonTxExit(dataList []string) (*NeonLogTxEvent, error) {
	// data list must have 1 elements
	if len(dataList) != 1 {
		return nil, errors.New("Failed to decode exit event, it should contain 1 element")
	}

	// decode call type
	typeName := utils.Base64stringDecodeToString(dataList[0])

	// check call type
	switch {
	case typeName == "STOP":
		return &NeonLogTxEvent{eventType: ExitStop}, nil
	case typeName == "RETURN":
		return &NeonLogTxEvent{eventType: ExitReturn}, nil
	case typeName == "SELFDESTRUCT":
		return &NeonLogTxEvent{eventType: ExitSelfDestruct}, nil
	case typeName == "REVERT":
		return &NeonLogTxEvent{eventType: ExitRevert}, nil
	default:
		return nil, errors.New("Failed to decode exit event, wrong type")
	}
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
	neonLogTxEvent.eventType = Log

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
func GetEvents(logMessages []string) ([]Event, error) {
	//final slice of eth event logs from program logs
	events := make([]Event, 0)
	ethEvents := make([]NeonLogTxEvent, 0)
	// find the call to evm program in the logs
	var k int
	for k < len(logMessages) {
		// check if the log indicates call to evm program
		if len(logMessages[k]) >= len(EvmInvocationLog) && logMessages[k][0:len(EvmInvocationLog)] == EvmInvocationLog {
			evmProgramStartIndex := k + 1
			k++
			// find the end of the evm program call
			for {
				// check if we find evm program end, if that call succeeded we parse logs inside the segment
				if len(logMessages[k]) >= len(EvmInvocationSuccessEnd) && logMessages[k][0:len(EvmInvocationSuccessEnd)] == EvmInvocationSuccessEnd {
					eventLogs, err := parseLogs(logMessages[evmProgramStartIndex:k])
					if err != nil {
						return nil, err
					}

					// append newly parsed logs
					ethEvents = append(ethEvents, eventLogs...)
					k++
					break
				}

				// check if we find evm program end, if that call failed we skip any logs inside the segment
				if len(logMessages[k]) >= len(EvmInvocationFailEnd) && logMessages[k][0:len(EvmInvocationFailEnd)] == EvmInvocationFailEnd {
					k++
					break
				}

				k++
			}
		} else {
			k++
		}
	}

	// loop through all the events from logs and only return contract event logs
	for _, e := range ethEvents {
		// set event field values
		events = append(events, Event{Address: "0x" + hex.EncodeToString(e.address), Data: "0x" + hex.EncodeToString(e.data), Topics: e.topicList, TransactionHash: "0x" + hex.EncodeToString(e.transactionHash)})
	}

	return events, nil
}

// parses logs from evm invocation segment in logs
func parseLogs(logMessages []string) ([]NeonLogTxEvent, error) {
	// for parsing eth transaction hash
	var neonTxHash []byte
	// for parsed gas usage data
	var neonTxIx *NeonLogTxIx
	// for parsed transaction result code
	var neonTxReturn *NeonLogTxReturn
	// error decoding logs
	var err error
	// we encounter non evm calls inside evm call segment so we remember the current depth of such calls
	var nonEvmCallDepth int
	// if we encounter enter instruction we increase call depth of evm contract
	var evmCallDepth int

	neonTxEventList := make([]NeonLogTxEvent, 0)

	// for each solana transaction log message decode eth data
	for _, line := range logMessages {
		// check if some non-evm program is called
		if isNonEvmProgramInvoke(line) {
			nonEvmCallDepth++
			continue
		}

		// check if some non-evm program exited here
		if isNonEvmProgramExit(line) {
			nonEvmCallDepth--
			continue
		}

		// if we are inside some other program's logs then skip those
		if nonEvmCallDepth != 0 {
			continue
		}

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
				// check transaction final status
				if neonTxReturn.status != 0 {
					return nil, errors.New("return status not ok")
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
		case name == "ENTER":
			// decode eth event log
			_, err := DecodeNeonTxEnter(dataList)
			if err != nil {
				return nil, err
			}
			// increase call depth
			evmCallDepth++
		case name == "EXIT":
			// decode eth event log
			neonTxEvent, err := DecodeNeonTxExit(dataList)
			if err != nil {
				return nil, err
			}

			// if the segment reverted remove those logs
			if neonTxEvent.eventType == ExitRevert {
				for len(neonTxEventList) != 0 && neonTxEventList[len(neonTxEventList) - 1].callLevel == evmCallDepth {
					neonTxEventList = neonTxEventList[:len(neonTxEventList)-1]
				}
			}

			// as we exit specific call depth decrease call depth level
			evmCallDepth--
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
			neonTxEvent.transactionHash = neonTxHash
			neonTxEvent.callLevel = evmCallDepth
			neonTxEventList = append(neonTxEventList, *neonTxEvent)
		default:
			continue
		}
	}

	// if the call stack is not finished skip tx
	if evmCallDepth != 0 || nonEvmCallDepth != 0 {
		return nil, nil
	}

	return neonTxEventList, nil
}

func GetBlockHash(slot int, solanaWebsocketEndpoint string) (string, error) {
	// get block with given slot
	response, err := jsonRPC([]byte(`{
		"jsonrpc": "2.0","id":1,
		"method":"getBlock",
		"params": [
			`+strconv.FormatUint(uint64(slot), 10)+`,
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
		return "", err
	}

	var blockHeader BlockHeader

	// unrmarshal latest block slot
	err = json.Unmarshal([]byte(response), &blockHeader)
	if err != nil {
		return "", err
	}

	// check rpc error
	if blockHeader.Error != nil {
		return "", errors.New(blockHeader.Error.Message)
	}

	return utils.Base58stringToHex(blockHeader.Result.Blockhash), nil
}

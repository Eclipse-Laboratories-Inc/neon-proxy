package source

import (
	"os"
	"bytes"
	"sort"
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/broadcaster"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/utils"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/config"
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

var (
	// prepare logs for checking evm enter/exit points
	EvmInvocationLog        = "Program " + os.Getenv(config.EvmAddress) + " invoke"
	EvmInvocationSuccessEnd = "Program " + os.Getenv(config.EvmAddress) + " success"
	EvmInvocationFailEnd    = "Program " + os.Getenv(config.EvmAddress) + " fail"
	// cache for transaction iterations until all the iterations are found and eth transaction processed
	splitTransactions map[string]IterationCache

)

// RegisterLogsBroadcasterSources passes data and error channels where new incoming data (transaction logs) will be pushed and redirected to broadcaster
func RegisterLogsBroadcasterSources(ctx *context.Context, log logger.Logger, solanaWebsocketEndpoint, evmAddress string, broadcaster *broadcaster.Broadcaster) error {
	log.Info().Msg("logs pulling from evm address started ... ")

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

			// for saving slot and corresponding block hash assiciation
			blockHashes := make(map[int]string)

			//for saving multiple-iteration transaction data (like a cache)
			splitTransactions = make(map[string]IterationCache)

			// prepare curl parameter
			var untilParameter string
			if latestProcessedTransaction != "" {
				untilParameter = ", \"until\": \"" + latestProcessedTransaction + "\""
			}

			// get latest block number
			signaturesResponse, err := jsonRPC([]byte(`{"jsonrpc":"2.0","id":1, "method": "getSignaturesForAddress", "params":["`+evmAddress+`",{"commitment":"finalized" `+untilParameter+`, "limit":100}]}`), solanaWebsocketEndpoint, "POST")
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
			if err := processTransactionLogs(ctx, solanaWebsocketEndpoint, log, logsSource, logsSourceError, &transactionSignatures, blockHashes, indexes); err != nil {
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
func processTransactionLogs(ctx *context.Context, solanaWebsocketEndpoint string, log logger.Logger, logsSource chan interface{}, logsSourceError chan error, signatures *TransactionSignatures, blockHashes map[int]string, indexes map[int]Indexes) error {
	// process each block sequentially, broadcasting to peers
	for k := len(signatures.Result) - 1; k >= 0; k-- {
		// get block with given slot
		response, err := jsonRPC([]byte(`{"jsonrpc": "2.0","id":1, "method":"getTransaction", "params": ["`+signatures.Result[k].Signature+`", {"encoding":"json", "maxSupportedTransactionVersion":0}]}`), solanaWebsocketEndpoint, "POST")

		// check err
		if err != nil {
			return err
		}

		// json defined transaction from rpc response
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
		ethlogs, err := getEthLogsFromTransaction(&transaction, blockHashes, indexes, log, solanaWebsocketEndpoint)
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

			// broadcast contract event log
			logsSource <- clientResponse
		}
	}

	return nil
}

/*
For each given solana transaction signature we request the whole transaction object
from which we parse instruction data where eth transaction parameters (including logs) are stored
*/
func getEthLogsFromTransaction(transaction *Transaction, blockHashes map[int]string, indexes map[int]Indexes, log logger.Logger, solanaWebsocketEndpoint string) ([]EthLog, error) {
	// create response array of logs
	ethLogs := make([]EthLog, 0)

	// declare log instance to fill
	var ethLog EthLog
	var err error

	// fill initial log values
	ethLog.BlockHash, err = GetBlockHash(blockHashes, transaction.Result.Slot, solanaWebsocketEndpoint)
	if err != nil {
		return nil, err
	}

	// set block hash
	ethLog.BlockNumber = fmt.Sprintf("0x%x", transaction.Result.Slot)

	// parse events from solana transaction logs
	events, err := GetEvents(transaction.Result.Meta.LogMessages)
	if err != nil {
		return nil, err
	}

	// if uninitialized we initialize the map for block slot which stores transaction hash and it's corresponding index in the given block
	if indexes[transaction.Result.Slot].processed == nil {
		v := indexes[transaction.Result.Slot]
		v.processed = make(map[string]int, 0)
		indexes[transaction.Result.Slot] = v
	}

	// fill events
	for _, event := range events {
		// set basic field values
		ethLog.Address = event.Address
		ethLog.TransactionHash = event.TransactionHash
		ethLog.Data = event.Data
		ethLog.LogIndex = fmt.Sprintf("0x%x", indexes[transaction.Result.Slot].logIndex)

		// update log index
		inds := indexes[transaction.Result.Slot]
		inds.logIndex++
		indexes[transaction.Result.Slot] = inds

		/*
			if we already have determined transaction index for given transaction hash we set that value
			otherwise it means we have a new/unknown eth transaction in the block and increase transaction index
		*/
		if _, ok := indexes[transaction.Result.Slot].processed[ethLog.TransactionHash]; ok == true {
			ethLog.TransactionIndex = fmt.Sprintf("0x%x", indexes[transaction.Result.Slot].processed[ethLog.TransactionHash])
		} else {
			indexes[transaction.Result.Slot].processed[ethLog.TransactionHash] = indexes[transaction.Result.Slot].transactionIndex
			ethLog.TransactionIndex = fmt.Sprintf("0x%x", indexes[transaction.Result.Slot].transactionIndex)
			v := indexes[transaction.Result.Slot]
			v.transactionIndex++
			indexes[transaction.Result.Slot] = v
		}

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
	if len(words) >= 3 && words[0] == "Program" && (words[2] == "success" || (len(words[2]) >= 4 && words[2][0:4] == "fail")) {
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

	return &NeonLogTxIx{gasUsed: int(binary.LittleEndian.Uint32(utils.Base64stringToBytes(dataList[0]))), totalGasUsed: int(binary.LittleEndian.Uint32(utils.Base64stringToBytes(dataList[1])))}, nil
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
					eventLogs, isIteration, txHash, usedGas, err := parseLogs(logMessages[evmProgramStartIndex:k])
					if err != nil {
						return nil, err
					}

					// if the transaction is one of the iteration process separately
					if isIteration {
						eventLogs, err = processSplitTransaction("0x" + hex.EncodeToString(txHash), usedGas, logMessages[evmProgramStartIndex:k])
						if err != nil {
							return nil, err
						}
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

func GetEnterExitCode(ind int, logMessages []string) (int, int) {
	// we encounter non evm calls inside evm call segment so we remember the current depth of such calls
	var nonEvmCallDepth int

	var callDepth int
	// we encounter non evm calls inside evm call segment so we remember the current depth of such calls
	for ind = ind; ind <= len(logMessages) - 1; ind++ {
		// check if some non-evm program is called
		if isNonEvmProgramInvoke(logMessages[ind]) {
			nonEvmCallDepth++
			continue
		}

		// check if some non-evm program exited here
		if isNonEvmProgramExit(logMessages[ind]) {
			nonEvmCallDepth--
			continue
		}

		// if we are inside some other program's logs then skip those
		if nonEvmCallDepth != 0 {
			continue
		}

		// decode log line
		name, dataList := decodeMnemonic(logMessages[ind])
		if len(name) == 0 {
			continue
		}

		// Use switch on the instruction name variable.
		switch {
		case name == "ENTER":
			// increase depth
			callDepth++
		case name == "EXIT":
			// reduce depths
			callDepth--

			// decode eth event log
			neonTxEvent, err := DecodeNeonTxExit(dataList)
			if err != nil {
				return 0, 0
			}

			// if we exit main call return it's exit status
			if callDepth == 0 {
				return neonTxEvent.eventType, ind
			}
		}
	}

	return 0, -1
}

// parses logs from evm invocation segment in logs
func parseLogs(logMessages []string) ([]NeonLogTxEvent, bool, []byte, *NeonLogTxIx, error) {
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
	// count the number of enter calls
	var evmCallDepth int


	neonTxEventList := make([]NeonLogTxEvent, 0)
	// for each solana transaction log message decode eth data
	for ind := 0; ind < len(logMessages); ind++ {
		// get line
		line := logMessages[ind]

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

		// Use switch on the instruction name variable.
		switch {
		case name == "HASH":
			if neonTxHash == nil {
				neonTxHash, err = DecodeNeonTxSig(dataList)
				if err != nil {
					return nil, false, neonTxHash, neonTxIx, err
				}
			}
		case name == "RETURN":
			if neonTxReturn == nil {
				neonTxReturn, err = DecodeNeonTxReturn(neonTxIx, dataList)
				if err != nil {
					return nil, false, neonTxHash, neonTxIx, err
				}
				// check transaction final status
				if neonTxReturn.status != 1 {
					return nil, false, neonTxHash, neonTxIx, errors.New("return status not ok")
				}
			} else {
				return nil, false, neonTxHash, neonTxIx, errors.New("transaction RETURN encountered twice")
			}
		case name == "GAS":
			if neonTxIx == nil {
				neonTxIx, err = DecodeNeonTxGas(dataList)
				if err != nil {
					return nil, false, neonTxHash, neonTxIx, err
				}
			}
		case name == "ENTER":
			evmCallDepth++
			// decode eth event log
			_, err := DecodeNeonTxEnter(dataList)
			if err != nil {
				return nil, false, neonTxHash, neonTxIx, err
			}
			// if the segment ends with revert, skip it
			var exitCode, endingInd int
			exitCode, endingInd = GetEnterExitCode(ind, logMessages);

			// omit inside segment if the segment ends with revert
			if exitCode == ExitRevert {
				ind = endingInd
			}
		case name == "EXIT":
			evmCallDepth--
			continue
		case len(name) >= 3 && name[0:3] == "LOG":
			logNum, err := strconv.Atoi(name[3:])
			if err != nil {
				return nil, false, neonTxHash, neonTxIx, err
			}
			// decode eth event log
			neonTxEvent, err := DecodeNeonTxEvent(logNum, dataList)
			if err != nil {
				return nil, false, neonTxHash, neonTxIx, err
			}
			// add new log to client response data
			neonTxEvent.transactionHash = neonTxHash
			neonTxEventList = append(neonTxEventList, *neonTxEvent)
		default:
			continue
		}
	}


	// check if transaction is one of the iterations of eth transaction
	/*
		case 1: the iteration is the last iteration with RETURN instruction and the total gas.
		Note that we use the totalGasUsed value from this transaction as the target.
		We sum up all the other iteration gas usage and when the sum is equal to the totalGasUsed from this iteration
		it means we have collected all the iterations of the eth transaction.
	*/
	if neonTxReturn != nil && neonTxIx.gasUsed != neonTxIx.totalGasUsed {
		/*
			set the gas usage target of the eth transaction.
			When the sum of all iteration gas usage reaches this value we know we have collected all the iterations.
		*/
		v := splitTransactions["0x" + hex.EncodeToString(neonTxHash)]
		v.targetTotalGas = neonTxIx.totalGasUsed
		splitTransactions["0x" + hex.EncodeToString(neonTxHash)] = v

		return nil, true, neonTxHash, neonTxIx, nil
	}

	/*
		case 2: one of the other iteration except the last iteration with RETURN instruction.
		In this case we have no RETURN instruction which basically means that it's already iteration type,
		but also we check that we have information about gas usage to be able to order
		the iterations according to gas usage in the and before processing them together.
  */
	if (neonTxReturn == nil && neonTxIx != nil) {
		return nil, true, neonTxHash, neonTxIx, nil
	}

	if evmCallDepth != 0 {
		return nil, false, neonTxHash, neonTxIx, errors.New("neon function call stack unfinished")
	}

	return neonTxEventList, false, neonTxHash, neonTxIx, nil
}

// at this point we know we have collected all the iteration log messages, we sort and process it as one eth transaction log
func processSplitTransaction(txHash string, usedGas *NeonLogTxIx, logMessages []string) ([]NeonLogTxEvent, error) {

	// append new iteration log messages
	logData := splitTransactions[txHash]
	logData.totalUsedGas += usedGas.gasUsed
	if logData.logIterations == nil {
		logData.logIterations = make([]LogData, 0)
	}
	logData.logIterations = append(logData.logIterations, LogData{usedGas: usedGas.totalGasUsed, logs: logMessages})
	splitTransactions[txHash] = logData

	// check condition if we have collected all the iterations
	if logData.totalUsedGas == logData.targetTotalGas {
		fmt.Println("processing iterations")
		// remove processed data from map
		defer delete(splitTransactions, txHash)

		// Sort by age, keeping original order or equal elements.
		sort.SliceStable(logData.logIterations, func(i, j int) bool {
			return logData.logIterations[i].usedGas < logData.logIterations[j].usedGas
		})

		// combine all the iteration logs
		logMessages := make([]string, 0)
		for _, messages := range logData.logIterations {
			logMessages = append(logMessages, messages.logs...)
		}

		// parse combined log messages from all the iterations
		logs, _, _, _, err := parseLogs(logMessages)
		if err != nil {
			return nil, err
		}

		return logs, nil
	}

	return nil, nil
}

// get block hash for given slot number to fill log data
func GetBlockHash(blockHashes map[int]string, slot int, solanaWebsocketEndpoint string) (string, error) {
	// check if we have already queried
	if val, ok := blockHashes[slot]; ok == true {
		return val, nil
	}

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

	// save in cache
	blockHashes[slot] = utils.Base58stringToHex(blockHeader.Result.Blockhash)

	return utils.Base58stringToHex(blockHeader.Result.Blockhash), nil
}

package source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/broadcaster"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"time"
)

const MaxProcessedTransactionsBatch = 100

// defines eth log structure for each transaction
type EthLog struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      string   `json:blockNumber`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	BlockHash        string   `json:"blockHash"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
}

// RegisterLogsBroadcasterSources passes data and error channels where new incoming data (transaction logs) will be pushed and redirected to broadcaster
func RegisterLogsBroadcasterSources(_ *context.Context, log logger.Logger, solanaWebsocketEndpoint, evmAddress string, broadcaster *broadcaster.Broadcaster) error {
	log.Info().Msg("logs pulling from evm address started ... ")

	// declare sources to be set
	logsSource := make(chan interface{})
	logsSourceError := make(chan error)

	// register given sources
	broadcaster.SetSources(logsSource, logsSourceError)

	go func() {
		// defines last transaction we parsed from evm
		var lastProcessedTransactionSignature string
		// construct until parameter for request
		var untilParam string

		// declare time interval for checking new transactions
		ticker := time.NewTicker(1 * time.Second)
		constString := fmt.Sprintf("%v", MaxProcessedTransactionsBatch)

		for range ticker.C {
			log.Info().Msg("logParser: latest processed transaction signature " + lastProcessedTransactionSignature)

			// if lastProcessedTransactionSignature is set,
			// return MaxProcessedTransactionsBatch transactions from newest to transaction, which signature
			// is set in lastProcessedTransactionSignature
			if len(lastProcessedTransactionSignature) != 0 {
				untilParam = `, "until": "` + lastProcessedTransactionSignature + `"`
			}

			req := `{
		  		"jsonrpc": "2.0","id":1,
		  		"method":"getSignaturesForAddress",
				"params": [
		    		"` + evmAddress + `",
		    		{
              			"limit": ` + constString + `,
		      			"commitment" :"finalized"` + untilParam + `
					}
				]
			}`

			// get evm transaction from node
			response, err := jsonRPC([]byte(req), solanaWebsocketEndpoint, "POST")
			if err != nil {
				log.Error().Err(err).Msg("Error on rpc call for getting batch of transactions signatures")
				logsSourceError <- err
			}

			// unmarshall response
			var txSignaturesFromSlot GetTransactionSignatureByAccountKeyResp
			if err := json.Unmarshal(response, &txSignaturesFromSlot); err != nil {
				log.Error().Err(err).Msg("Error on unmarshaling transaction signatures response from rpc endpoint")
				logsSourceError <- err
			}

			// check response error message
			if txSignaturesFromSlot.Error != nil {
				err = errors.New(txSignaturesFromSlot.Error.Message)
				log.Error().Err(err).Msg("Error on rpc call for getting batch of transactions signatures")
				logsSourceError <- err
			}

			// for each solana transaction signature parse ethereum data from logs and return the object
			for ind := len(txSignaturesFromSlot.Result); ind >= 0; ind-- {
				// get eth transaction by parsing solana transaction logs
				ethLog, err := getEthLogs(txSignaturesFromSlot.Result[ind].Signature, log, solanaWebsocketEndpoint)
				if err != nil {
					log.Error().Err(err).Msg("Cannot get transaction logs")
					logsSourceError <- err
				}

				// marshal eth object for broadcaster
				logData, err := json.Marshal(ethLog)
				if err != nil {
					log.Error().Err(err).Msg("Cannot marshal eth object")
					logsSourceError <- err
				}

				// push new log into broadcaster
				logsSource <- logData
			}

			// set latest processed tx
			if len(txSignaturesFromSlot.Result) != 0 {
				lastProcessedTransactionSignature = txSignaturesFromSlot.Result[0].Signature
			}
		}
	}()
	return nil
}

/*
For each given solana transaction signature we request the whole transaction object
from which we parse instruction data where eth transaction parameters (including logs) are stored
*/
func getEthLogs(signature string, log logger.Logger, solanaWebsocketEndpoint string) ([]EthLog, error) {
	// get latest block number
	txResponse, err := jsonRPC([]byte(`{"jsonrpc":"2.0","id":1, "method":"getTransaction", "params":["`+signature+`" ,"json"]}`), solanaWebsocketEndpoint, "POST")
	if err != nil {
		log.Error().Err(err).Msg("Error on rpc call for getting transaction")
		return nil, err
	}

	// declare response json
	var solanaTx SolanaTx

	// unrmarshal latest block slot
	err = json.Unmarshal(txResponse, &solanaTx)
	if err != nil {
		log.Error().Err(err).Msg("Error on unmarshaling transaction response from rpc endpoint")
		return nil, err
	}

	// error from rpc
	if solanaTx.Error != nil && solanaTx.Error.Message != "" {
		log.Error().Err(err).Msg("Error from rpc endpoint")
		return nil, err
	}

	// given log messages from solaan tx we parse eth logs
	return parseEthLogsFromLogMessages(solanaTx.Result.Meta.LogMessages, log, solanaWebsocketEndpoint)
}

// given log messages from solaan tx we parse eth logs
func parseEthLogsFromLogMessages(logMessages []string, log logger.Logger, solanaWebsocketEndpoint string) ([]EthLog, error) {
	return nil, nil
}

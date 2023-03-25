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

// RegisterLogsBroadcasterSources passes data and error channels where new incoming data (transaction logs) will be pushed and redirected to broadcaster
func RegisterLogsBroadcasterSources(_ *context.Context, log logger.Logger, solanaWebsocketEndpoint, evmAddress string, broadcaster *broadcaster.Broadcaster) error {
	log.Info().Msg("pending transaction pulling from mempool started ... ")

	// declare sources to be set
	logsSource := make(chan interface{})
	logsSourceError := make(chan error)

	// register given sources
	broadcaster.SetSources(logsSource, logsSourceError)

	go func() {
		var lastProcessedTransactionSignature string
		var untilParam string

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

			response, err := jsonRPC([]byte(req), solanaWebsocketEndpoint, "POST")
			if err != nil {
				log.Error().Err(err).Msg("Error on rpc call for getting batch of transactions signatures")
				logsSourceError <- err
			}

			var txSignaturesFromSlot GetTransactionSignatureByAccountKeyResp

			if err := json.Unmarshal(response, &txSignaturesFromSlot); err != nil {
				log.Error().Err(err).Msg("Error on unmarshaling transaction signatures response from rpc endpoint")
				logsSourceError <- err
			}

			if txSignaturesFromSlot.Error != nil {
				err = errors.New(txSignaturesFromSlot.Error.Message)
				log.Error().Err(err).Msg("Error on rpc call for getting batch of transactions signatures")
				logsSourceError <- err
			}

			// TODO implement getting e-logs using transaction signatures here

			if len(txSignaturesFromSlot.Result) > 0 {
				// save last processed transaction
				// transactions are returned from newest to oldest
				// https://docs.solana.com/api/http#getsignaturesforaddress
				lastProcessedTransactionSignature = txSignaturesFromSlot.Result[0].Signature
			}
		}
	}()
	return nil
}

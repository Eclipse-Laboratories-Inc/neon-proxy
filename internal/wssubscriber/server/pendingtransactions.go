package server

import (
	"encoding/json"

	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/utils"
)

// subscribes the client to pending transaction
func (c *Client) subscribeToNewPendingTransactions(requestRPC SubscribeJsonRPC, responseRPC *SubscribeJsonResponseRCP) {
	// if new head subscription is active skip another subscription
	c.pendingTransactionsLocker.Lock()
	defer c.pendingTransactionsLocker.Unlock()

	// check if subscription type for the client is active
	if c.pendingTransactionsIsActive {
		responseRPC.Error = "pendingTransactions subscription already active. Subscription ID: " + c.pendingTransactionsSubscriptionID
		return
	}

	// if not subscribe to broadcaster
	c.pendingTransactionsSource = c.pendingTransactionsBroadcaster.Subscribe()

	// generate subscription id
	responseRPC.Result = utils.NewID()
	responseRPC.ID = requestRPC.ID

	// register subscription id for client
	c.pendingTransactionsSubscriptionID = responseRPC.Result
	c.pendingTransactionsIsActive = true
	go c.CollectPendingTransactions()
}

// runs a separate go routine to collect incoming pending transactiosn from broadcaster and pusing them into client response buffer
func (c *Client) CollectPendingTransactions() {
	// listen for incoming pending transactions and send to user
	for {
		select {
		case tx, ok := <-c.pendingTransactionsSource:
			//channel has been closed
			if ok == false {
				return
			}
			// case when response to subscribe request isn't sent yet
			c.pendingTransactionsLocker.Lock()
			if c.pendingTransactionsIsActive == false {
				c.pendingTransactionsLocker.Unlock()
				continue
			}

			// construct response object for new event
			var clientResponse ClientResponse
			clientResponse.Jsonrpc = rpcVersion
			clientResponse.Method = methodSubscriptionName
			clientResponse.Params.Subscription = c.pendingTransactionsSubscriptionID
			clientResponse.Params.Result = []byte("\"" + tx.(string) + "\"")
			c.pendingTransactionsLocker.Unlock()

			// marshal to send it as a json
			response, err := json.Marshal(clientResponse)

			// check json marshaling error
			if err != nil {
				c.log.Error().Err(err).Msg(fmt.Sprintf("marshalling response output: %v", err))
				continue
			}

			c.clientResponseBuffer <- response
		}
	}
}

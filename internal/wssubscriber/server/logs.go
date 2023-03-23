package server

import (
	"encoding/json"
	"fmt"

	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/utils"
)

// to be implemented
func (c *Client) subscribeToNewLogs(requestRPC SubscribeJsonRPC, responseRPC *SubscribeJsonResponseRCP) {
	// if new logs subscription is active skip another subscription
	c.newLogsLocker.Lock()
	defer c.newLogsLocker.Unlock()

	// check if subscription type for the client is active
	if c.newLogsIsActive {
		responseRPC.Error = "newLogs subscription already active. Subscription ID: " + c.newLogsSubscriptionID
		return
	}

	// if not subscribe to broadcaster
	c.newLogsSource = c.newLogsBroadcaster.Subscribe()

	// generate subscription id
	responseRPC.Result = utils.NewID()
	responseRPC.ID = requestRPC.ID

	// register subscription id for client
	c.newLogsSubscriptionID = responseRPC.Result
	c.newLogsIsActive = true
	c.log.Info().Msg("NewLogs subscription succeeded with ID: " + responseRPC.Result)
	go c.CollectNewLogs()
}

// collects new logs coming from broadcaster and pushes the data into the client response buffer
func (c *Client) CollectNewLogs() {
	// listen for incoming logs and send to user
	for {
		select {
		case newLog, ok := <-c.newLogsSource:
			//channel has been closed
			if ok == false {
				return
			}
			// case when subscription response isn't sent yet, or it's not active anymore
			c.newLogsLocker.Lock()
			if c.newLogsIsActive == false {
				c.newLogsLocker.Unlock()
				continue
			}

			// construct response object for new event
			var clientResponse ClientResponse
			clientResponse.Jsonrpc = rpcVersion
			clientResponse.Method = methodSubscriptionName
			clientResponse.Params.Subscription = c.newLogsSubscriptionID
			clientResponse.Params.Result = newLog.([]byte)
			c.newLogsLocker.Unlock()

			// marshal to send it as a json
			response, err := json.Marshal(clientResponse)

			// check json marshaling error
			if err != nil {
				c.log.Error().Err(err).Msg(fmt.Sprintf("marshalling response output: %v", err))
				return
			}

			// push new response
			c.clientResponseBuffer <- response
		}
	}
}

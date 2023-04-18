package server

import (
	"encoding/json"
	"fmt"

	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/utils"
)

// subscribes to broadcaster for new heads
func (c *Client) subscribeToNewHeads(requestRPC SubscribeJsonRPC, responseRPC *SubscribeJsonResponseRCP) {
	// if new head subscription is active skip another subscription
	c.newHeadsLocker.Lock()
	defer c.newHeadsLocker.Unlock()

	// check if subscription type for the client is active
	if c.newHeadsIsActive {
		responseRPC.Error = "newHeads subscription already active. Subscription ID: " + c.newHeadSubscriptionID
		return
	}

	// if not subscribe to broadcaster
	c.newHeadsSource = c.newHeadsBroadcaster.Subscribe()

	// generate subscription id
	responseRPC.Result = utils.NewID()
	responseRPC.ID = requestRPC.ID

	// register subscription id for client
	c.newHeadSubscriptionID = responseRPC.Result
	c.newHeadsIsActive = true
	c.log.Info().Msg("NewHeads subscription succeeded with ID: " + responseRPC.Result)
	go c.CollectNewHeads()
}

// collects new heads coming from broadcaster and pushes the data into the client response buffer
func (c *Client) CollectNewHeads() {
	// listen for incoming heads and send to user
	for {
		select {
		case newHead, ok := <-c.newHeadsSource:
			//channel has been closed
			if ok == false {
				return
			}
			// case when subscription response isn't sent yet, or it's not active anymore
			c.newHeadsLocker.Lock()
			if c.newHeadsIsActive == false {
				c.newHeadsLocker.Unlock()
				continue
			}

			// construct response object for new event
			var clientResponse ClientResponse
			clientResponse.Jsonrpc = rpcVersion
			clientResponse.Method = methodSubscriptionName
			clientResponse.Params.Subscription = c.newHeadSubscriptionID
			clientResponse.Params.Result = newHead.([]byte)
			c.newHeadsLocker.Unlock()

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

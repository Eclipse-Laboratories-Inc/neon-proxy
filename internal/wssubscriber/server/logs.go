package server

import (
	"encoding/json"
	"fmt"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/source"

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

	if len(requestRPC.Params) > 1 {
		params, ok := requestRPC.Params[1].(SubscribeLogsFilterParams)
		if !ok {
			responseRPC.Error = "newLogs subscription accepts filters only by addresses and topics"
			return
		}

		c.buildFilters(params)
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

func (c *Client) buildFilters(params SubscribeLogsFilterParams) {
	c.logsFilters = &logsFilters{
		Addresses: make(map[string]struct{}),
		Topics:    make(map[string]struct{}),
	}

	for i := range params.Addresses {
		c.logsFilters.Addresses[params.Addresses[i]] = struct{}{}
	}

	for i := range params.Topics {
		c.logsFilters.Topics[params.Topics[i]] = struct{}{}
	}
}

// collects new logs coming from broadcaster and pushes the data into the client response buffer
func (c *Client) CollectNewLogs() {
	// listen for incoming logs and send to user
	for {
		select {
		case newLogs, ok := <-c.newLogsSource:
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

			if c.logsFilters != nil {
				logs, err := c.FilterLogs(newLogs.([]byte))
				if err != nil {
					c.log.Error().Err(err).Msg(fmt.Sprintf("filter logs output: %v", err))
					c.newLogsLocker.Unlock()
					return
				} else {
					clientResponse.Params.Result = logs
				}
			} else {
				clientResponse.Params.Result = newLogs.([]byte)
			}

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

func (c *Client) FilterLogs(newLogs []byte) ([]byte, error) {
	filteredLogs := make([]source.EthLog, 0)
	rowLogs := make([]source.EthLog, 0)
	if err := json.Unmarshal(newLogs, &rowLogs); err != nil {
		return nil, err
	}

	filters := c.logsFilters
	for i, rowLog := range rowLogs {
		// check if log contains address, which is set up as filter
		_, ok := filters.Addresses[rowLog.Address]
		if ok {
			filteredLogs = append(filteredLogs, rowLogs[i])
		}

		// check if log contains topic, which is set up as filter
		if rowLog.Topics != nil {
			for _, topic := range rowLog.Topics {
				_, ok := filters.Addresses[topic]
				if ok {
					filteredLogs = append(filteredLogs, rowLogs[i])
				}
			}
		}
	}

	return json.Marshal(filteredLogs)
}

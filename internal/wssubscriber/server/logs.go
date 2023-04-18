package server

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"

	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/source"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/utils"
)

var (
	ErrLogFilterInvalidDataTypes    = errors.New("invalid log filter: data types must start with 0x")
	ErrLogFilterInvalidTopicsString = errors.New("invalid log filter: invalid field topics: string")
	ErrLogFilterInvalidTopicsNumber = errors.New("invalid log filter: invalid field topics: number")
	ErrLogFilterInvalidHexCharacter = errors.New("invalid log filter: invalid hex string, invalid character")
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

	// check if user set up filters
	if len(requestRPC.Params) > 1 {
		jsonParams, err := json.Marshal(requestRPC.Params[1])
		if err != nil {
			responseRPC.Error = "failed to read log filters"
			return
		}

		var filterParams SubscribeLogsFilterParams
		if err := json.Unmarshal(jsonParams, &filterParams); err != nil {
			responseRPC.Error = "failed to read log filters, incorrect json format"
			return
		}

		if err := c.buildFilters(filterParams); err != nil {
			responseRPC.Error = err.Error()
			return
		}
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

func (c *Client) buildFilters(params SubscribeLogsFilterParams) error {
	c.logsFilters = &logsFilters{}
	c.logsFilters.Addresses = make([]string, 0)
	c.logsFilters.Topics = make([][]string, 0)

	// check if address is given in a signe field
	if utils.IsAddressValid(params.Address) {
		c.logsFilters.Addresses = append(c.logsFilters.Addresses, params.Address)
	}

	// parse filter addresses
	for _, addr := range params.Addresses {
		if utils.IsAddressValid(addr) {
			c.logsFilters.Addresses = append(c.logsFilters.Addresses, addr)
		}
	}

	// parsing topics
	for _, topics := range params.Topics {
		// if we have null at the position
		if topics == nil {
			c.logsFilters.Topics = append(c.logsFilters.Topics, nil)
		}

		// check if it's single string or array of strings
		switch topics.(type) {
		case string:
			if utils.IsTopicValid(topics.(string)) {
				t := make([]string, 0)
				t = append(t, topics.(string))
				c.logsFilters.Topics = append(c.logsFilters.Topics, t)
			} else {
				return errors.New("Invalid topic ")
			}
		case []interface{}:
			parsedTopics := make([]string, 0)
			for _, ta := range topics.([]interface{}) {
				switch ta.(type) {
				case string:
					parsedTopics = append(parsedTopics, ta.(string))
				}
			}
			c.logsFilters.Topics = append(c.logsFilters.Topics, parsedTopics)
		}
	}

	return nil
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

			logs, err := c.FilterLogs(newLogs.([]byte))
			if err != nil {
				c.log.Error().Err(err).Msg(fmt.Sprintf("filter logs output: %v", err))
				c.newLogsLocker.Unlock()
				continue
			}

			//
			if logs == nil {
				// after filtering we got no logs
				c.newLogsLocker.Unlock()
				continue
			}
			c.newLogsLocker.Unlock()

			// set parsed matched logs
			clientResponse.Params.Result = logs

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

// FilterLogs filters logs
// If no filter is set up, it returns log as it is
// Otherwise it checks if log contains AT LEAST 1 address from "address" filter OR AT LEAST 1 topic from "topic" filter
func (c *Client) FilterLogs(newLog []byte) ([]byte, error) {
	// no filters set up
	if len(c.logsFilters.Addresses) == 0 && len(c.logsFilters.Topics) == 0 {
		return newLog, nil
	}

	// unmarshall log
	rawLog := source.EthLog{}
	if err := json.Unmarshal(newLog, &rawLog); err != nil {
		return nil, err
	}

	// if address didn't match
	if len(c.logsFilters.Addresses) != 0 && !utils.Includes(c.logsFilters.Addresses, rawLog.Address) {
		return nil, nil
	}

	// expecting more topics then there are
	if len(c.logsFilters.Topics) > len(rawLog.Topics) {
		return nil, nil
	}

	// check topic match
	for ind, topic := range c.logsFilters.Topics {
		// if corresponding requirement was 'null' skip
		if topic == nil {
			continue
		}

		// if topic filter didn't match return empty
		if !utils.Includes(topic, rawLog.Topics[ind]) {
			return nil, nil
		}
	}

	return newLog, nil
}

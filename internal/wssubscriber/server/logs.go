package server

import (
	"encoding/json"
	"fmt"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/source"
	"regexp"

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

	// check if user set up filters
	if len(requestRPC.Params) > 1 {
		jsonParams, err := json.Marshal(requestRPC.Params[1])
		if err != nil {
			responseRPC.Error = "failed to read log filters"
			return
		}

		var filterParams SubscribeLogsFilterParams
		if err := json.Unmarshal(jsonParams, &filterParams); err != nil {
			responseRPC.Error = "failed to read log filters"
			return
		}

		if err := c.buildFilters(filterParams); err != nil {
			responseRPC.Error = fmt.Sprintf("failed to set up log filters: %v", err)
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

func isHexNumber(str string) bool {
	hexPattern := "^0[xX][0-9a-fA-F]*$"
	regex := regexp.MustCompile(hexPattern)
	return regex.MatchString(str)
}

func (c *Client) buildFilters(params SubscribeLogsFilterParams) error {
	c.logsFilters = &logsFilters{
		Addresses: make(map[string]struct{}),
		Topics:    make(map[string]struct{}),
	}

	for i := range params.Addresses {
		// it seems, if field "address" in log is empty, than we return "address":"0x", but user may not know it
		// and set up filter to get logs with empty "address" field as "" or "null"
		// if "address" isn't "" or "null", it should be presented at least as hex number
		if params.Addresses[i] != "null" && params.Addresses[i] != "" && !isHexNumber(params.Addresses[i]) {
			return fmt.Errorf("can't filter by %v addres, addres should be hex number", params.Addresses[i])
		}
		c.logsFilters.Addresses[params.Addresses[i]] = struct{}{}
	}

	for i := range params.Topics {
		c.logsFilters.Topics[params.Topics[i]] = struct{}{}
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
				return

			} else if logs == nil {
				// after filtering we got no logs
				c.newLogsLocker.Unlock()
				continue
			} else {
				clientResponse.Params.Result = logs
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

// FilterLogs filters logs
// If no filter is set up, it returns log as it is
// Otherwise it checks if log contains AT LEAST 1 address from "address" filter OR AT LEAST 1 topic from "topic" filter
func (c *Client) FilterLogs(newLog []byte) ([]byte, error) {
	// no filters set up
	if c.logsFilters == nil {
		return newLog, nil
	}

	rowLog := source.EthLog{}
	if err := json.Unmarshal(newLog, &rowLog); err != nil {
		return nil, err
	}

	filters := c.logsFilters

	// check if "empty address" filter is set and log contains "0x" address, which is representation for
	// empty address
	_, ok1 := filters.Addresses["null"]
	_, ok2 := filters.Addresses[""]
	if (ok1 || ok2) && rowLog.Address == "0x" {
		return newLog, nil
	}

	// check if log contains address, which is set up as filter
	_, ok := filters.Addresses[rowLog.Address]
	if ok {
		return newLog, nil
	}

	// check if "empty topic" filter is set and log contains no topics
	_, ok1 = filters.Topics["null"]
	_, ok2 = filters.Topics[""]
	if (ok1 || ok2) && rowLog.Topics == nil {
		return newLog, nil
	}

	// check if log contains topic, which is set up as filter
	if rowLog.Topics != nil {
		for _, topic := range rowLog.Topics {
			_, ok := filters.Topics[topic]
			if ok {
				return newLog, nil
			}
		}
	}
	return nil, nil
}

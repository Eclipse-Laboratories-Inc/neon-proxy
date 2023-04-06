package server

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"regexp"
	"strings"

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
			responseRPC.Error = "failed to read log filters"
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

	// addresses can be any type, but we set up as filter only if it's
	// not empty string
	switch addresses := params.Addresses.(type) {
	case string:
		if addresses != "" {
			c.logsFilters.Addresses = make(map[string]struct{})
			c.logsFilters.Addresses[addresses] = struct{}{}
		}
	case []interface{}:
		if len(addresses) > 0 {
			c.logsFilters.Addresses = make(map[string]struct{})
		}
		for _, addr := range addresses {
			switch a := addr.(type) {
			case string:
				if a != "" {
					c.logsFilters.Addresses[a] = struct{}{}
				}
			}
		}
	}

	// topics can be only array of strings, which have prefix "0x" and have to be
	// longer, than 64 symbols (32 bytes)
	topicsFormat := regexp.MustCompile(`^[0-9a-fA-F]*$`)
	switch topics := params.Topics.(type) {
	case string:
		return ErrLogFilterInvalidTopicsString
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
		return ErrLogFilterInvalidTopicsNumber
	case []interface{}:
		if len(topics) > 0 {
			c.logsFilters.Topics = make(map[string]struct{})
		}
		for _, topic := range topics {
			switch t := topic.(type) {
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
				return ErrLogFilterInvalidTopicsNumber
			case string:
				if !strings.HasPrefix(t, "0x") {
					return ErrLogFilterInvalidDataTypes
				}

				addr := strings.Trim(t, "0x")
				if len(addr) < 64 {
					return fmt.Errorf("invalid log filter: data type size mismatch, expected 32, got %v", len(addr)/2)
				}
				if !topicsFormat.MatchString(addr) {
					return ErrLogFilterInvalidHexCharacter
				}
				c.logsFilters.Topics[t] = struct{}{}
			}
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
	if c.logsFilters == nil || c.logsFilters.Addresses == nil && c.logsFilters.Topics == nil {
		return newLog, nil
	}

	rowLog := source.EthLog{}
	if err := json.Unmarshal(newLog, &rowLog); err != nil {
		return nil, err
	}

	filters := c.logsFilters

	// it seems, in Ethereum filter by topics has higher priority
	// if we send request { "id": 1, "method": "eth_subscribe", "params": [ "logs", { "addresses": ["0xc309c03e4d5065ea4a7d591fb9a4c418cb6e4812f0a768623ec9bd878fe2c829"], "topics": []}]}
	// for subscription to Ethereum, it returns all logs with all addresses without filtering
	if filters.Topics == nil {
		return newLog, nil
	}

	// check if any address filter is set up and log contains this topic
	if filters.Topics != nil {
		for _, topic := range rowLog.Topics {
			_, ok := filters.Topics[topic]
			if ok {
				return newLog, nil
			}
		}
	}

	// it seems, in Ethereum in case if we can't pass filter by topics, we return nothing,
	// so it doesn't make sense to check filter by address

	// check if any address filter is set up and log contains this address
	/*if filters.Addresses != nil {
		_, ok := filters.Addresses[rowLog.Address]
		if ok {
			return newLog, nil
		}
	}*/

	return nil, nil
}

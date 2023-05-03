package server

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	validAddress1 = "0x123"
	validAddress2 = "0x123AScmlsd"
	validAddress3 = "0x"
	validAddress4 = ""

	validTopic          = "0x602f1aeac0ca2e7a13e281a9ef0ad7838542712ce16780fa2ecffd351f05f8999"
	invalidTopicsNumber = 123
	invalidTopicHex     = "0x602f1aeac0ca2e7a13e281a9ef0ad7838542712ce16780fa2ecffd351f05f899z"
	invalidTopicString  = "602f1aeac0ca2e7a13e281a9ef0ad7838542712ce16780fa2ecffd351f05f899z"
)

func TestBuildFilters(t *testing.T) {
	c := Client{}

	t.Run("no params", func(t *testing.T) {
		err := c.buildFilters(SubscribeLogsFilterParams{})
		assert.NoError(t, err)
	})

	t.Run("empty params", func(t *testing.T) {
		err := c.buildFilters(SubscribeLogsFilterParams{Topics: nil, Addresses: nil})
		assert.NoError(t, err)
	})

	t.Run("addresses string", func(t *testing.T) {
		params := SubscribeLogsFilterParams{Addresses: []string{validAddress1}}
		err := c.buildFilters(params)
		assert.NoError(t, err)

		params = SubscribeLogsFilterParams{Addresses: []string{validAddress2}}
		err = c.buildFilters(params)
		assert.NoError(t, err)

		params = SubscribeLogsFilterParams{Addresses: []string{validAddress3}}
		err = c.buildFilters(params)
		assert.NoError(t, err)

		params = SubscribeLogsFilterParams{Addresses: []string{validAddress4}}
		err = c.buildFilters(params)
		assert.NoError(t, err)
	})

	t.Run("invalid topics string", func(t *testing.T) {
		topics := make([]interface{}, 1)
		topics[0] = validTopic
		params := SubscribeLogsFilterParams{Topics: topics}
		err := c.buildFilters(params)
		assert.Error(t, err)

		params = SubscribeLogsFilterParams{Topics: nil}
		err = c.buildFilters(params)
		assert.Equal(t, nil, err)
	})

	t.Run("invalid topics hex", func(t *testing.T) {
		params := SubscribeLogsFilterParams{Topics: []any{invalidTopicHex}}
		err := c.buildFilters(params)
		assert.Error(t, err)
		assert.Equal(t, ErrLogFilterInvalid.Error(), err.Error())
	})

	t.Run("invalid topics data type", func(t *testing.T) {
		params := SubscribeLogsFilterParams{Topics: []any{invalidTopicString}}
		err := c.buildFilters(params)
		assert.Error(t, err)
		assert.Equal(t, ErrLogFilterInvalid.Error(), err.Error())
	})

	t.Run("invalid array: contains empty string", func(t *testing.T) {
		params := SubscribeLogsFilterParams{Topics: []any{""}}
		err := c.buildFilters(params)
		assert.Error(t, err)
		assert.Equal(t, ErrLogFilterInvalid.Error(), err.Error())
	})

	t.Run("invalid topics array: contains number", func(t *testing.T) {
		params := SubscribeLogsFilterParams{
			Topics: []any{validTopic, invalidTopicsNumber},
		}
		err := c.buildFilters(params)
		assert.Error(t, err)
		assert.Equal(t, ErrLogFilterInvalid.Error(), err.Error())
	})

	t.Run("invalid topic size", func(t *testing.T) {
		topics := make([]interface{}, 1)
		topics[0] = "0x111"
		params := SubscribeLogsFilterParams{Topics: topics}
		err := c.buildFilters(params)
		assert.Error(t, err)
	})
}

func TestFilterLogs(t *testing.T) {
	c := Client{}
	log := []byte(`{
		"address": "0x5B38Da6a701c568545dCfcB03FcB875f56beddC4",
		"topics": [
		  "0x95b4472199b7a3877350ba99969c60f547e6f9ecae0f13e99b71d1aaff9e2612"
		],
		"data": "0x0000000000000000000000000000000000000000000000000000000000000001"
	  }`)

	t.Run("no filters set up", func(t *testing.T) {
		c.logsFilters = &logsFilters{}
		filteredLog, err := c.FilterLogs(log)
		assert.NoError(t, err)
		assert.Equal(t, log, filteredLog)
	})

	t.Run("filter by topic set up and topic found", func(t *testing.T) {
		topics := make([][]string, 1)
		topics[0] = []string{"0x95b4472199b7a3877350ba99969c60f547e6f9ecae0f13e99b71d1aaff9e2612"}
		c.logsFilters = &logsFilters{
			Topics: topics,
		}

		filteredLog, err := c.FilterLogs(log)
		assert.NoError(t, err)
		assert.Equal(t, log, filteredLog)
	})

	t.Run("filter by topic set up and topic not found", func(t *testing.T) {
		topics := make([][]string, 1)
		topics[0] = []string{"0x1234567890abcdef"}
		c.logsFilters = &logsFilters{
			Topics: topics,
		}

		filteredLog, err := c.FilterLogs(log)
		assert.NoError(t, err)
		assert.Nil(t, filteredLog)
	})

	t.Run("filter by address set up and address found", func(t *testing.T) {
		addresses := make([]string, 1)
		addresses[0] = "0x5B38Da6a701c568545dCfcB03FcB875f56beddC4"
		c.logsFilters = &logsFilters{
			Addresses: addresses,
		}

		filteredLog, err := c.FilterLogs(log)
		assert.NoError(t, err)
		assert.Equal(t, log, filteredLog)
	})

	t.Run("filter by address set up, but ignored", func(t *testing.T) {
		addresses := make([]string, 1)
		addresses[0] = "0x1234567890abcdef"
		c.logsFilters = &logsFilters{
			Addresses: addresses,
		}

		_, err := c.FilterLogs(log)
		assert.NoError(t, err)
	})
}

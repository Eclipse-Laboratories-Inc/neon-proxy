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
		err := c.buildFilters(SubscribeLogsFilterParams{Topics: []any{}, Addresses: []any{}})
		assert.NoError(t, err)
	})

	t.Run("addresses string", func(t *testing.T) {
		params := SubscribeLogsFilterParams{Addresses: validAddress1}
		err := c.buildFilters(params)
		assert.NoError(t, err)

		_, ok := c.logsFilters.Topics[validAddress1]
		assert.True(t, ok)

		params = SubscribeLogsFilterParams{Addresses: validAddress2}
		err = c.buildFilters(params)
		assert.NoError(t, err)

		_, ok = c.logsFilters.Topics[validAddress2]
		assert.True(t, ok)

		params = SubscribeLogsFilterParams{Addresses: validAddress3}
		err = c.buildFilters(params)
		assert.NoError(t, err)

		_, ok = c.logsFilters.Topics[validAddress3]
		assert.True(t, ok)

		params = SubscribeLogsFilterParams{Addresses: validAddress4}
		err = c.buildFilters(params)
		assert.NoError(t, err)

		_, ok = c.logsFilters.Topics[validAddress4]
		assert.False(t, ok)
	})

	t.Run("invalid topics string", func(t *testing.T) {
		params := SubscribeLogsFilterParams{Topics: validTopic}
		err := c.buildFilters(params)
		assert.Error(t, err)
		assert.Equal(t, ErrLogFilterInvalidTopicsString, err)

		params = SubscribeLogsFilterParams{Topics: ""}
		err = c.buildFilters(params)
		assert.Error(t, err)
		assert.Equal(t, ErrLogFilterInvalidTopicsString, err)
	})

	t.Run("invalid topics number", func(t *testing.T) {
		params := SubscribeLogsFilterParams{Topics: invalidTopicsNumber}
		err := c.buildFilters(params)
		assert.Error(t, err)
		assert.Equal(t, ErrLogFilterInvalidTopicsNumber, err)
	})

	t.Run("invalid topics hex", func(t *testing.T) {
		params := SubscribeLogsFilterParams{Topics: []any{invalidTopicHex}}
		err := c.buildFilters(params)
		assert.Error(t, err)
		assert.Equal(t, ErrLogFilterInvalidHexCharacter, err)
	})

	t.Run("invalid topics data type", func(t *testing.T) {
		params := SubscribeLogsFilterParams{Topics: []any{invalidTopicString}}
		err := c.buildFilters(params)
		assert.Error(t, err)
		assert.Equal(t, ErrLogFilterInvalidDataTypes, err)
	})

	t.Run("valid topics array", func(t *testing.T) {
		params := SubscribeLogsFilterParams{Topics: []any{validTopic}}
		err := c.buildFilters(params)
		assert.NoError(t, err)
		_, ok := c.logsFilters.Topics[validTopic]
		assert.True(t, ok)
	})

	t.Run("invalid array: contains empty string", func(t *testing.T) {
		params := SubscribeLogsFilterParams{Topics: []any{""}}
		err := c.buildFilters(params)
		assert.Error(t, err)
		assert.Equal(t, ErrLogFilterInvalidDataTypes, err)
	})

	t.Run("invalid topics array: contains number", func(t *testing.T) {
		params := SubscribeLogsFilterParams{
			Topics: []any{validTopic, invalidTopicsNumber},
		}
		err := c.buildFilters(params)
		assert.Error(t, err)
		assert.Equal(t, ErrLogFilterInvalidTopicsNumber, err)
	})

	t.Run("invalid topic size", func(t *testing.T) {
		params := SubscribeLogsFilterParams{Topics: "0x111"}
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
		c.logsFilters = &logsFilters{
			Topics: map[string]struct{}{
				"0x95b4472199b7a3877350ba99969c60f547e6f9ecae0f13e99b71d1aaff9e2612": {},
			},
		}

		filteredLog, err := c.FilterLogs(log)
		assert.NoError(t, err)
		assert.Equal(t, log, filteredLog)
	})

	t.Run("filter by topic set up and topic not found", func(t *testing.T) {
		c.logsFilters = &logsFilters{
			Topics: map[string]struct{}{
				"0x1234567890abcdef": {},
			},
		}

		filteredLog, err := c.FilterLogs(log)
		assert.NoError(t, err)
		assert.Nil(t, filteredLog)
	})

	t.Run("filter by address set up and address found", func(t *testing.T) {
		c.logsFilters = &logsFilters{
			Addresses: map[string]struct{}{
				"0x5B38Da6a701c568545dCfcB03FcB875f56beddC4": {},
			},
		}

		filteredLog, err := c.FilterLogs(log)
		assert.NoError(t, err)
		assert.Equal(t, log, filteredLog)
	})

	t.Run("filter by address set up, but ignored", func(t *testing.T) {
		c.logsFilters = &logsFilters{
			Addresses: map[string]struct{}{
				"0x1234567890abcdef": {},
			},
		}

		filteredLog, err := c.FilterLogs(log)
		assert.NoError(t, err)
		assert.Equal(t, log, filteredLog)
	})
}

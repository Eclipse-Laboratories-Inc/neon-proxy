package server

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/source"
)

func (c *Client) evmProxyMethods(req []byte, id uint64) (response SubscribeJsonResponseRCP) {
	response.Version = rpcVersion
	response.ID = id

	ret, err := source.EvmJsonRpc(req, c.evmRpcEndpoint)
	if err != nil {
		c.log.Error().Err(err).Msg(fmt.Sprintf("Error: %v %v", err, bytes.NewBuffer(req)))
		response.Error = &SubscriptionError{Code: EvmRpcFailed, Message: "Failed to call EVM rpc : " + err.Error()}
		return
	}

	if err = json.Unmarshal(ret, &response); err != nil {
		c.log.Error().Err(err).Msg(fmt.Sprintf("Error: %v %v", err, bytes.NewBuffer(req)))
		response.Error = &SubscriptionError{Code: EvmRpcFailed, Message: "Failed to unmarshal rpc: " + err.Error()}
		return
	}

	return response
}

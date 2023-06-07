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
		c.log.Error().Err(err).Msg(fmt.Sprintf("XXXxxxerror: %v %v", err, bytes.NewBuffer(req)))
	}

	if err = json.Unmarshal(ret, &response); err != nil {
		c.log.Error().Err(err).Msg(fmt.Sprintf("XXXxxxerror: %v %v", err, bytes.NewBuffer(req)))
	}

	return response
}

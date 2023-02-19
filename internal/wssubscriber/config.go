package wssubscriber

import (
  "os"
  "errors"
)

const (
    solanaRPCEndpoint = "SOLANA_RPC_ENDPOINT"
    wssubscriberPort = "NEON_WEBSOCKET_PORT"
)

type WSSubscriberConfig struct {
  solanaRPCEndpoint string
  wssubscriberPort string
}

func CreateConfigFromEnv() (cfg *WSSubscriberConfig, err error) {
  // check if endpoint is set in env
  endpoint := os.Getenv(solanaRPCEndpoint)
  if len(endpoint) == 0 {
    return nil, errors.New(solanaRPCEndpoint + " env variable not set")
  }

  // check if port is set in env
  port := os.Getenv(wssubscriberPort)
  if len(port) == 0 {
    port = "8080"
  }

	return &WSSubscriberConfig{
    solanaRPCEndpoint: endpoint,
    wssubscriberPort: port,
  }, nil
}

var headSubscribe = `{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "blockSubscribe",
  "params": [
    "all",
    {
      "commitment": "confirmed",
      "encoding": "jsonParsed",
      "showRewards": false,
      "transactionDetails": "signatures"
    }
  ]
}`

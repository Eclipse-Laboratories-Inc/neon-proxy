package wssubscriber

import (
  "os"
  "errors"
)

const (
    neonSolanaWebsocketEnv = "NEON_WEBSOCKET_ENDPOINT"
    wssubscriberPort = "NEON_WEBSOCKET_PORT"
)

type WSSubscriberConfig struct {
  solanaWebsocketEndpoint string
  wssubscriberPort string
}

func CreateConfigFromEnv() (cfg *WSSubscriberConfig, err error) {
  // check if endpoint is set in env
  solanaWssEndpoint := os.Getenv(neonSolanaWebsocketEnv)
  if len(solanaWssEndpoint) == 0 {
    return nil, errors.New(neonSolanaWebsocketEnv + " env variable not set")
  }

  // check if port is set in env
  port := os.Getenv(wssubscriberPort)
  if len(port) == 0 {
    port = "8080"
  }

	return &WSSubscriberConfig{
    solanaWebsocketEndpoint: solanaWssEndpoint,
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

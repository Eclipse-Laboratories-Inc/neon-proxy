package proxy

import (
  "os"
  "errors"
)

const (
    neonSolanaWebsocketEnv = "NEON_WEBSOCKET_ENDPOINT"
    websocketPort = "NEON_WEBSOCKET_PORT"
)

type TransactionProxyConfig struct {
  solanaWebsocketEndpoint string
  websocketPort string
}

func CreateConfigFromEnv() (cfg *TransactionProxyConfig, err error) {
  // check if endpoint is set in env
  solanaWssEndpoint := os.Getenv(neonSolanaWebsocketEnv)
  if len(solanaWssEndpoint) == 0 {
    return nil, errors.New(neonSolanaWebsocketEnv + " env variable not set")
  }

  // check if endpoint is set in env
  websocketPort := os.Getenv(websocketPort)
  if len(websocketPort) == 0 {
    websocketPort = "8080"
  }

	return &TransactionProxyConfig{
    solanaWebsocketEndpoint: solanaWssEndpoint,
    websocketPort: websocketPort,
  }, nil
}

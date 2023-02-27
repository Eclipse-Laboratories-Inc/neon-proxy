package config

import (
  "os"
  "errors"
)

const (
    solanaRPCEndpoint = "SOLANA_RPC_ENDPOINT"
    wssubscriberPort = "NEON_WEBSOCKET_PORT"
)

type WSSubscriberConfig struct {
  SolanaRPCEndpoint string
  WssubscriberPort string
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
    SolanaRPCEndpoint: endpoint,
    WssubscriberPort: port,
  }, nil
}

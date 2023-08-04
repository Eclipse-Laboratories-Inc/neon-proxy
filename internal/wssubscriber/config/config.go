package config

import (
	"errors"
	"os"
)

const (
	solanaRPCEndpoint = "SOLANA_RPC_ENDPOINT"
	wssubscriberPort  = "NEON_WEBSOCKET_PORT"
	EvmAddress        = "EVM_ADDRESS"
	EvmEndpoint       = "EVM_ENDPOINT"
)

// declare wssubscriber configuration parameters
type WSSubscriberConfig struct {
	SolanaRPCEndpoint string
	EvmEndpoint       string
	WssubscriberPort  string
	EvmAddress        string
}

// validate and return env variable
func CreateConfigFromEnv() (cfg *WSSubscriberConfig, err error) {
	// check if endpoint is set in env
	endpoint := os.Getenv(solanaRPCEndpoint)
	if len(endpoint) == 0 {
		return nil, errors.New(solanaRPCEndpoint + " env variable not set")
	}

	evmAddr := os.Getenv(EvmAddress)
	if len(evmAddr) == 0 {
		return nil, errors.New(EvmAddress + " env variable not set")
	}

	// check if port is set in env
	port := os.Getenv(wssubscriberPort)
	if len(port) == 0 {
		port = "8080"
	}

	evmEndpoint := os.Getenv(EvmEndpoint)
	if len(evmEndpoint) == 0 {
		return nil, errors.New(EvmEndpoint + " env variable not set")
	}

	return &WSSubscriberConfig{
		SolanaRPCEndpoint: endpoint,
		EvmEndpoint:       evmEndpoint,
		WssubscriberPort:  port,
		EvmAddress:        evmAddr,
	}, nil
}

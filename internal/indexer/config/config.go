package config

import (
	"errors"
	"github.com/gagliardetto/solana-go"
	"os"
	"strconv"
)

const (
	indexerPollCnt           = "INDEXER_POLL_COUNT"
	indexerPollCntMaxVal     = 1000
	indexerPollCntDefaultVal = 1

	evmLoaderId = "EVM_LOADER"

	solanaRPCEndpoint = "SOLANA_RPC_ENDPOINT"
)

type IndexerConfig struct {
	SolanaEndpoint string
	EvmLoaderID    solana.PublicKey
	IndexerPollCnt int
}

func CreateConfigFromEnv() (cfg *IndexerConfig, err error) {
	cfg = &IndexerConfig{}

	endpoint := os.Getenv(solanaRPCEndpoint)
	if len(endpoint) == 0 {
		return nil, errors.New(solanaRPCEndpoint + " env variable not set")
	}

	cfg.SolanaEndpoint = endpoint

	cntStr := os.Getenv(indexerPollCnt)
	if len(cntStr) == 0 {
		cfg.IndexerPollCnt = indexerPollCntDefaultVal
	} else {
		cnt, err := strconv.Atoi(cntStr)
		if err != nil {
			return nil, err
		}
		if cnt > indexerPollCntMaxVal {
			return nil, errors.New("maximal value for INDEXER_POLL_COUNT is 1000")
		}
		cfg.IndexerPollCnt = cnt
	}

	evmAddr := os.Getenv(evmLoaderId)
	if len(evmAddr) == 0 {
		return nil, errors.New(evmLoaderId + " env variable not set")
	}

	pubKey, err := solana.PublicKeyFromBase58(evmLoaderId)
	if err != nil {
		return nil, err
	}
	cfg.EvmLoaderID = pubKey

	return cfg, nil
}

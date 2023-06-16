package proxy

import (
	"context"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/neonlabsorg/neon-service-framework/pkg/logger"
)

type Proxy struct {
	ctx             context.Context
	solanaRpcClient *rpc.Client
	logger          logger.Logger
}

func NewProxy(
	ctx context.Context,
	solanaRpcClient *rpc.Client,
	log logger.Logger,
) *Proxy {
	return &Proxy{
		ctx:             ctx,
		solanaRpcClient: solanaRpcClient,
		logger:          log,
	}
}

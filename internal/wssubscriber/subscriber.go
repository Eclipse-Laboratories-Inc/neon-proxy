package wssubscriber

import (
	"context"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/config"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/server"
)

const (
	defaultChanLen = 0
)

type Transaction struct {
	signature string
}

type WSSubscriber struct {
	cfg             *config.WSSubscriberConfig
	ctx             context.Context
	logger          logger.Logger
	subscriberError chan error
}

func NewWSSubscriber(
	cfg *config.WSSubscriberConfig,
	ctx context.Context,
	log logger.Logger,
) *WSSubscriber {
	return &WSSubscriber{
		cfg:             cfg,
		ctx:             ctx,
		logger:          log,
		subscriberError: make(chan error, 0),
	}
}

func (s *WSSubscriber) Run() error {
	// create server
	server := server.NewServer(&s.ctx, s.logger)

	// creates a broadcaster already pulling new heads from solana and registers the broadcaster
	if err := server.StartNewHeadBroadcaster(s.cfg.SolanaRPCEndpoint); err != nil {
		return err
	}

	// creates a broadcaster already pulling pending transactions from mempool
	if err := server.StartPendingTransactionBroadcaster(); err != nil {
		return err
	}

	// start ws server for incoming subscriptions
	go server.StartServer(s.cfg.WssubscriberPort)

	return nil
}

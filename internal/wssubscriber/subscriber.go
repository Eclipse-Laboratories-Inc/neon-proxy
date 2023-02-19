package wssubscriber

import (
	"context"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
)

const (
	defaultChanLen = 0
)

type Transaction struct {
	signature string
}

type WSSubscriber struct {
	cfg             *WSSubscriberConfig
	ctx             context.Context
	logger          logger.Logger
	subscriberError chan error
}

func NewWSSubscriber(
	cfg *WSSubscriberConfig,
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
	// for checking registration errors
	var err error

	// create server
	server := NewServer(&s.ctx)

	// creates a broadcaster already pulling new heads from solana and registers the broadcaster
	if server.newHeadsBroadcaster, err = server.GetNewHeadBroadcaster(s.cfg.solanaRPCEndpoint); err != nil {
		return err
	}

	// create and register new pending transaction broadcaster and start pulling new transactions
	//go server.RegisterNewPendingTransactionBroadcaster(NewPendingTransactionBroadcaster(s.cfg.solanaWebsocketEndpoint)).Start()

	// start ws server for incoming subscriptions
	go server.StartServer(s.cfg.wssubscriberPort)

	return nil
}

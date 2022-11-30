package subscriber

import (
	"context"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
)

type TransactionSubscriber struct {
	cfg             *TransactionSubscriberConfig
	ctx             context.Context
	solanaRpcClient *rpc.Client
	logger          logger.Logger
}

func NewTransactionSubscriber(
	cfg *TransactionSubscriberConfig,
	ctx context.Context,
	solanaRpcClient *rpc.Client,
	log logger.Logger,
) *TransactionSubscriber {
	return &TransactionSubscriber{
		cfg:             cfg,
		ctx:             ctx,
		solanaRpcClient: solanaRpcClient,
		logger:          log,
	}
}

func (s *TransactionSubscriber) RunWithLongPoling() error {
	tick := time.NewTicker(time.Duration(s.cfg.Interval * int(time.Second)))
	defer tick.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return nil
		case <-tick.C:
			err := s.processTick()
			if err != nil {
				s.logger.Error().Err(err).Msg("Error on subscriber process")
			}
		}
	}
}

func (s *TransactionSubscriber) processTick() error {
	s.logger.Info().Msg("process tick")
	return nil
}

func (s *TransactionSubscriber) RunWithWebsocket() error {
	return nil
}

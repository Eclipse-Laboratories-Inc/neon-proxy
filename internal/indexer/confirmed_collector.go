package indexer

import (
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/neonlabsorg/neon-proxy/internal/indexer/config"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"github.com/neonlabsorg/neon-proxy/pkg/solana"
)

type ConfirmedCollector struct {
	*SolanaTxMetaCollector
}

func NewConfirmedCollector(cfg *config.IndexerConfig,
	logger logger.Logger,
	solanaClient *solana.Client,
	dict *SolTxMetaDict,
	isFinalized bool) *ConfirmedCollector {
	return &ConfirmedCollector{
		NewSolanaTxMetaCollector(cfg,
			logger,
			solanaClient,
			dict,
			rpc.CommitmentConfirmed,
			isFinalized),
	}
}

// IterTxMeta runs collector for given slots diapason.
// It gets new info about ~confirmed~ tx signatures from solana and then request full info about txs from solana by their signatures
func (c *ConfirmedCollector) IterTxMeta(startSlot, stopSlot uint64) ([]*SolTxMetaInfo, error) {
	if startSlot < stopSlot {
		panic("start slot must be greater than or equal to stop slot")
	}
	infos, err := c.iterSigSlot(nil, startSlot, stopSlot)
	if err != nil {
		return nil, err
	}
	return c.iterTxMeta(infos)
}

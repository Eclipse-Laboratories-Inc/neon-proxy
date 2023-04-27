package indexer

import (
	"context"
	"database/sql"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/neonlabsorg/neon-proxy/internal/indexer/config"
	"github.com/neonlabsorg/neon-proxy/pkg/solana"

	"github.com/neonlabsorg/neon-proxy/pkg/logger"
)

type IndexerApp struct {
	ctx                context.Context
	logger             logger.Logger
	db                 *IndexerDB
	finalizedCollector CollectorInterface
	confirmedCollector CollectorInterface

	gatherStatistics  bool
	skipCancelTimeout int
	holderTimeout     int
}

func NewIndexerApp(
	ctx context.Context, cfg *config.IndexerConfig, log logger.Logger,
	db *sql.DB, client *rpc.Client, gatherStatistics bool) *IndexerApp {
	iDB := &IndexerDB{}
	iDB.Init(db)

	solanaClient := solana.NewClient(log, client)

	return &IndexerApp{
		ctx:                ctx,
		logger:             log,
		db:                 iDB,
		finalizedCollector: NewFinalizedCollector(cfg, log, solanaClient, NewSolTxMetaDict(), iDB.solanaSignsDB, nil, 0, 0, true), // TODO read vals from env variables
		confirmedCollector: NewConfirmedCollector(cfg, log, solanaClient, NewSolTxMetaDict(), false),
		gatherStatistics:   gatherStatistics,
		skipCancelTimeout:  100, //todo read from Env variable
		holderTimeout:      100, //todo read from Env variable
	}
}

func (i *IndexerApp) Run() {

}

func (i *IndexerApp) cancelOldNeonTxs() {

}

func (i *IndexerApp) cancelNeonTxs() {

}

func (i *IndexerApp) completeNeonBlock() {

}

func (i *IndexerApp) commitTxStat() {

}

func (i *IndexerApp) commitBlockStat() {

}

func (i *IndexerApp) commitStatusStat() {

}

func (i *IndexerApp) commitStats() {

}

func (i *IndexerApp) getSolanaBlockDeque() {

}

func (i *IndexerApp) locateNeonBlock() {

}

// TODO implement to run collector
func (i *IndexerApp) runSolanaTxCollector(state SolNeonTxDecoderState, slotProcessingDelay int) {

}

func (i *IndexerApp) hasNewBlocks() {

}

func (i *IndexerApp) logStats() {

}

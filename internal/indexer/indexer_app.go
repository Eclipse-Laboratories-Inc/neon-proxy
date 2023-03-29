package indexer

import (
	"context"
	"database/sql"

	"github.com/neonlabsorg/neon-proxy/pkg/logger"
)

type IndexerApp struct {
	ctx    context.Context
	logger logger.Logger
	db     *IndexerDB
}

func NewIndexerApp(ctx context.Context, log logger.Logger, db *sql.DB) *IndexerApp {
	iDB := &IndexerDB{}
	iDB.Init(db)

	return &IndexerApp{
		ctx:    ctx,
		logger: log,
		db:     iDB,
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

func (i *IndexerApp) runSolanaTxCollector() {
	i.collector.RunSolanaTxs()
}

func (i *IndexerApp) hasNewBlocks() {

}

func (i *IndexerApp) logStats() {

}

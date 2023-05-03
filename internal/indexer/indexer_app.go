package indexer

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/neonlabsorg/neon-proxy/internal/indexer/config"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"github.com/neonlabsorg/neon-proxy/pkg/solana"
	"strconv"
	"time"
)

type IndexerApp struct {
	ctx                context.Context
	logger             logger.Logger
	cfg                *config.IndexerConfig
	db                 *IndexerDB
	finalizedCollector CollectorInterface
	confirmedCollector CollectorInterface

	lastOpAccountUpdateTime int64

	startBlockSlot         uint64
	lastConfirmedBlockSlot uint64
	lastFinalizedBlockSlot uint64

	gatherStatistics  bool
	skipCancelTimeout int
	holderTimeout     int
}

func NewIndexerApp(
	ctx context.Context, cfg *config.IndexerConfig, log logger.Logger,
	db *sql.DB, client *rpc.Client, lastKnownSlot uint64, gatherStatistics bool) (*IndexerApp, error) {
	iDB := &IndexerDB{}
	iDB.Init(db)

	solanaClient := solana.NewClient(log, client)
	startSlot, err := initStartSlot(ctx, log, cfg, solanaClient, "receipt", lastKnownSlot)
	if err != nil {
		return nil, err
	}

	return &IndexerApp{
		ctx:                ctx,
		logger:             log,
		db:                 iDB,
		cfg:                cfg,
		finalizedCollector: NewFinalizedCollector(cfg, log, solanaClient, NewSolTxMetaDict(), iDB.solanaSignsDB, nil, 0, 0, true), // TODO read vals from env variables
		confirmedCollector: NewConfirmedCollector(cfg, log, solanaClient, NewSolTxMetaDict(), false),
		gatherStatistics:   gatherStatistics,
		startBlockSlot:     startSlot,
		skipCancelTimeout:  100, //todo read from Env variable
		holderTimeout:      100, //todo read from Env variable
	}, nil
}

func (i *IndexerApp) Run() {
	ticker := time.NewTicker(i.cfg.IndexerCheckMsec)
	defer func() {
		if r := recover(); r != nil {
			i.logger.Error().Msg(fmt.Sprintf("Caught a panic on transaction processing: %v", r))
		}
	}()

	for range ticker.C {
		if err := i.processFunctions(); err != nil {
			i.logger.Error().Err(err).Msg("Error on transaction processing")
		}
	}
}

func (i *IndexerApp) processFunctions() error {
	return nil
}

func (i *IndexerApp) cancelOldNeonTxs(state *SolNeonTxDecoderState, solTxMeta *SolTxMetaInfo) {

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

func initStartSlot(
	ctx context.Context,
	log logger.Logger,
	config *config.IndexerConfig,
	solanaClient *solana.Client,
	name string, lastKnownSlot uint64) (uint64, error) {

	latestSlot, err := solanaClient.GetLatestBlockSlot(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return 0, err
	}

	var startIntSlot int64
	name = fmt.Sprintf("%v slot", name)

	startSlot := config.StartSlot
	log.Info().Msg(fmt.Sprintf("Starting %v with LATEST_KNOWN_LOST=%v and START_SLOT=%v", name, lastKnownSlot, startSlot))

	switch startSlot {
	case "CONTINUE":
		if lastKnownSlot > 0 {
			log.Info().Msg(fmt.Sprintf("START_SLOT=%v: started the %v from previous run %v", startSlot, name, lastKnownSlot))
			return lastKnownSlot, nil
		} else {
			log.Info().Msg(fmt.Sprintf("START_SLOT=%v: forced the %v from the latest Solana slot", startSlot, name))
			startSlot = "LATEST"
		}

	case "LATEST":
		log.Info().Msg(fmt.Sprintf("START_SLOT=%v: started the %v from the latest Solana slot %v", startSlot, name, latestSlot))
		return latestSlot, nil
	}

	startIntSlot, _ = strconv.ParseInt(startSlot, 10, 64)
	if latestSlot < uint64(startIntSlot) {
		startIntSlot = int64(latestSlot)
	}

	if uint64(startIntSlot) < lastKnownSlot {
		log.Info().Msg(fmt.Sprintf("START_SLOT=%v: started the %v from previous run, because %v < %v", startSlot, name, startIntSlot, lastKnownSlot))
		return lastKnownSlot, nil

	}

	log.Info().Msg(fmt.Sprintf("START_SLOT=%v: started the %v from %v", startSlot, name, startIntSlot))
	return uint64(startIntSlot), nil
}

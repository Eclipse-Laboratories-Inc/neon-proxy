package indexer

import (
	"errors"
	"fmt"
	solana2 "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/neonlabsorg/neon-proxy/internal/indexer/config"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"github.com/neonlabsorg/neon-proxy/pkg/solana"
)

type FinalizedCollector struct {
	*SolanaTxMetaCollector

	solanaSignsDb SolanaSignsDBInterface
	lastSignInfo  *SolTxSigSlotInfo
	stopSlot      uint64
	signCnt       int64
}

func NewFinalizedCollector(cfg *config.IndexerConfig,
	logger logger.Logger,
	solanaClient *solana.Client,
	dict *SolTxMetaDict,
	solanaSignsDb SolanaSignsDBInterface,
	lastSignInfo *SolTxSigSlotInfo,
	stopSlot uint64,
	signCnt int64,
	isFinalized bool) *FinalizedCollector {
	return &FinalizedCollector{
		SolanaTxMetaCollector: NewSolanaTxMetaCollector(cfg,
			logger,
			solanaClient,
			dict,
			rpc.CommitmentFinalized,
			isFinalized),
		solanaSignsDb: solanaSignsDb,
		lastSignInfo:  lastSignInfo,
		stopSlot:      stopSlot,
		signCnt:       signCnt,
	}
}

func (c *FinalizedCollector) LastBlockSlot() uint64 {
	return c.stopSlot
}

// buildCheckpointList builds and saves N checkpoints (info about signature + block_slot) in db
func (c *FinalizedCollector) buildCheckpointList(startSlot uint64) error {
	maxSign, err := c.solanaSignsDb.GetMaxSign()
	if err != nil && !errors.Is(err, ErrRecordNotFound) {
		return err
	}

	if maxSign != nil && maxSign.BlockSlot > c.stopSlot {
		c.stopSlot = maxSign.BlockSlot
	}

	// get info about signatures starting from blockchain top to last known slot
	infos, err := c.iterSigSlot(nil, startSlot, c.stopSlot)
	if err != nil {
		return err
	}

	// save all signatures according to c.cfg.IndexerPollCnt setting
	for i := range infos {
		c.saveCheckpoint(&infos[i], 1)
	}
	return nil
}

// saveCheckpoint saves signatures data every N signatures
// In case if process of gathering info about signatures wasn't finished because of some reason,
// we can continue from last saved data
func (c *FinalizedCollector) saveCheckpoint(info *SolTxSigSlotInfo, cnt int) {
	c.signCnt += int64(cnt)
	switch {
	case info == nil || c.signCnt < int64(c.cfg.IndexerPollCnt):
		return
	case c.lastSignInfo == nil || c.lastSignInfo.BlockSlot == info.BlockSlot:
		c.lastSignInfo = info
	case c.lastSignInfo.BlockSlot != info.BlockSlot:
		c.logger.Debug().Msg(fmt.Sprintf("save checkpoint: %v: %v", *c.lastSignInfo, c.signCnt))
		if err := c.solanaSignsDb.AddSign(*c.lastSignInfo); err == nil {
			c.resetCheckpointCache()
		}
	}
}

func (c *FinalizedCollector) resetCheckpointCache() {
	c.lastSignInfo = nil
	c.signCnt = 0
}

// iterSigSlotList gets info about signatures and their block_slots from solana using previous saved checkpoints (if exist) it set up
// slots diapason till it will be nothing to gather
func (c *FinalizedCollector) iterSigSlotList(startSlot uint64, longList bool) ([]SolTxSigSlotInfo, error) {
	var (
		nextInfo    *SolTxSigSlotInfo
		startSig    solana2.Signature
		startSigPtr *solana2.Signature
		err         error
	)
	startSigPtr = &startSig
	signSlotLists := make([]SolTxSigSlotInfo, 0)

	// try to get info till we have checkpoint or startSlot is more fresh than stopSlot
	for startSigPtr != nil && startSlot >= c.stopSlot {
		startSigPtr = nil
		if longList {
			// get the nearest checkpoint after stopSlot
			nextInfo, err = c.solanaSignsDb.GetNextSign(c.stopSlot)
			if err != nil && !errors.Is(err, ErrRecordNotFound) {
				return nil, err
			}
			if nextInfo != nil {
				// use signature of checkpoint to search other signatures from it backwards
				startSigPtr = &nextInfo.SolSign
			}
		}

		// get signatures list
		signSlotList, err := c.iterSigSlot(startSigPtr, startSlot, c.stopSlot)
		if err != nil {
			return nil, err
		}

		// nothing found in current slots diapason
		if len(signSlotList) == 0 {
			if nextInfo != nil {
				// move stopSlot nearer to startSlot and keep going
				// (in case if we have next checkpoint, we will use its signature to request more fresh data from solana)
				c.stopSlot = nextInfo.BlockSlot
				continue
			}
			return signSlotLists, nil
		}

		// move stopSlot nearer to startSlot and keep going
		if nextInfo == nil {
			c.stopSlot = signSlotList[0].BlockSlot + 1
		} else {
			c.stopSlot = nextInfo.BlockSlot
		}

		if !longList {
			c.saveCheckpoint(&signSlotList[0], len(signSlotList))
		}

		signSlotLists = append(signSlotLists, signSlotList...)

	}

	// remove duplicates but keeping signatures order
	seen := make(map[SolTxSigSlotInfo]bool)
	var result []SolTxSigSlotInfo
	for _, v := range signSlotLists {
		if _, ok := seen[v]; !ok {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result, nil
}

func (c *FinalizedCollector) pruneTxMetaDict() {
	for _, signSlot := range c.txMetaDict.Keys() {
		if signSlot.BlockSlot < c.stopSlot {
			c.txMetaDict.Delete(signSlot)
		}
	}
}

// IterTxMeta runs collector for given slots diapason.
// It builds checkpoints and save them to db for future run, gets new info about ~finalized~ tx signatures from solana and then request
// full info about txs from solana by their signatures
func (c *FinalizedCollector) IterTxMeta(startSlot, stopSlot uint64) ([]*SolTxMetaInfo, error) {
	if startSlot < stopSlot {
		return nil, errors.New("start slot must be greater than or equal to stop slot")
	}
	isLongList := (startSlot - stopSlot) > 10
	if isLongList {
		if err := c.buildCheckpointList(startSlot); err != nil {
			return nil, err
		}
	}
	c.stopSlot = stopSlot

	// get txs signatures from solana
	infos, err := c.iterSigSlotList(startSlot, isLongList)
	if err != nil {
		return nil, err
	}

	// get full info about txs by their signatures
	txMetas, err := c.iterTxMeta(infos)
	if err != nil {
		return nil, err
	}

	// return got info and delete it from dictionary
	for _, txMeta := range txMetas {
		c.txMetaDict.Pop(txMeta.ident)
	}
	return txMetas, nil
}

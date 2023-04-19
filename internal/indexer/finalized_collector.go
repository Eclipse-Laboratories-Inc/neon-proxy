package indexer

import (
	"errors"
	"fmt"
	solana2 "github.com/gagliardetto/solana-go"
)

type FinalizedCollector struct {
	*collector

	solanaSignsDb *SolanaSignsDB
	lastSignInfo  *SolTxSigSlotInfo
	stopSlot      uint64
	signCnt       int64
}

func (c *FinalizedCollector) LastBlockSlot() uint64 {
	return c.stopSlot
}

func (c *FinalizedCollector) buildCheckpointList(startSlot uint64) error {
	maxSign, err := c.solanaSignsDb.GetMaxSign()

	if maxSign != nil && maxSign.BlockSlot > c.stopSlot {
		c.stopSlot = maxSign.BlockSlot
	}

	infos, err := c.getSignInfoBySlot(nil, startSlot, c.stopSlot)
	if err != nil {
		return err
	}
	for i := range infos {
		c.saveCheckpoint(&infos[i], 1)
	}
	return nil
}

func (c *FinalizedCollector) saveCheckpoint(info *SolTxSigSlotInfo, cnt int) {
	c.signCnt += int64(cnt)
	switch {
	case info == nil || c.signCnt < c.cfg.IndexerPollCnt:
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

func (c *FinalizedCollector) getSignsInfoListBySlot(startSlot uint64, longList bool) ([]SolTxSigSlotInfo, error) {
	var (
		nextInfo  *SolTxSigSlotInfo
		startSign *solana2.Signature
		err       error
	)
	signSlotLists := make([]SolTxSigSlotInfo, 0)

	for startSign != nil || len(signSlotLists) == 0 {
		startSign = nil

		if longList {
			nextInfo, err = c.solanaSignsDb.GetNextSign(c.stopSlot)
			if err != nil && !errors.Is(err, ErrRecordNotFound) {
				return nil, err
			}
			if nextInfo != nil {
				startSign = &nextInfo.SolSign
			}
		}

		signSlotList, err := c.getSignInfoBySlot(startSign, startSlot, c.stopSlot)
		if err != nil {
			return nil, err
		}

		if len(signSlotList) == 0 {
			if nextInfo != nil {
				c.stopSlot = nextInfo.BlockSlot
				continue
			}
			break
		}

		if nextInfo != nil {
			c.stopSlot = nextInfo.BlockSlot
		} else {
			c.stopSlot = signSlotList[0].BlockSlot + 1
		}

		if !longList {
			c.saveCheckpoint(&signSlotList[0], len(signSlotList))
		}

		signSlotLists = append(signSlotLists, signSlotList...)
	}
	return signSlotLists, nil
}

func (c *FinalizedCollector) pruneTxMetaDict() {
	for _, signSlot := range c.txMetaDict.Keys() {
		if signSlot.BlockSlot < c.stopSlot {
			c.txMetaDict.Delete(signSlot)
		}
	}
}

func (c *FinalizedCollector) GetTxMeta(startSlot, stopSlot uint64) ([]*SolTxMetaInfo, error) {
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
	infos, err := c.getSignsInfoListBySlot(startSlot, isLongList)
	if err != nil {
		return nil, err
	}

	txMetas, err := c.getTxMeta(infos)
	if err != nil {
		return nil, err
	}

	for _, txMeta := range txMetas {
		c.txMetaDict.Pop(txMeta.ident)
	}
	return txMetas, nil
}

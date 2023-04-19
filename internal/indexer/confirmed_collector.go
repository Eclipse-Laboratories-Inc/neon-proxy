package indexer

import (
	"errors"
)

type ConfirmedCollector struct {
	*collector
}

func (c *ConfirmedCollector) GetTxMeta(startSlot, stopSlot uint64) ([]*SolTxMetaInfo, error) {
	if startSlot < stopSlot {
		return nil, errors.New("start slot must be greater than or equal to stop slot")
	}
	infos, err := c.getSignInfoBySlot(nil, startSlot, stopSlot)
	if err != nil {
		return nil, err
	}
	return c.getTxMeta(infos)
}

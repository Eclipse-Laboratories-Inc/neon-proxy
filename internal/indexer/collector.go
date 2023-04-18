package indexer

import (
	"context"
	solana2 "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"github.com/neonlabsorg/neon-proxy/pkg/solana"
	"sync"
)

// Config TODO implement
type Config struct {
}

type сollector struct {
	cfg    *Config
	logger logger.Logger

	solanaClient *solana.Client
	txMetaDict   *SolTxMetaDict
	commitment   rpc.CommitmentType
	isFinalized  bool
}

func NewCollector(cfg *Config,
	solanaClient *solana.Client,
	logger logger.Logger,
	dict *SolTxMetaDict,
	commitment rpc.CommitmentType,
	isFinalized bool) *сollector {
	return &сollector{
		cfg:          cfg,
		logger:       logger,
		solanaClient: solanaClient,
		txMetaDict:   dict,
		commitment:   commitment,
		isFinalized:  isFinalized,
	}
}

func (c *сollector) GetCommitment() rpc.CommitmentType {
	return c.commitment
}

func (c *сollector) IsFinalized() bool {
	return c.isFinalized
}

// requestTxMetaList gets solana signatures and requests for each signature tx receipt from solana
// and adds info to txMetaDict
func (c *сollector) requestTxMetaList(sigSlotList ...SolTxSigSlotInfo) error {
	var sigList []solana2.Signature
	for _, sigSlot := range sigSlotList {
		sigList = append(sigList, sigSlot.SolSign)
	}

	ctx := context.Background()

	metaList, err := c.solanaClient.GetTxReceiptList(ctx, sigList, solana2.EncodingJSON, c.GetCommitment())
	if err != nil {
		return err
	}
	for i, sigSlot := range sigSlotList {
		if err := c.txMetaDict.Add(sigSlot, (*SolTxReceipt)(metaList[i])); err != nil {
			return err
		}
	}
	return nil
}

// gatherInfoIntoTxMetaDict gets grouped solana signatures and calls requestTxMetaList to get tx receipts for
// each signature in each group
func (c *сollector) gatherInfoIntoTxMetaDict(groupedSigSlotList [][]SolTxSigSlotInfo) {
	if len(groupedSigSlotList) == 1 {
		if err := c.requestTxMetaList(groupedSigSlotList[0]...); err != nil {
			c.logger.Error().Err(err).Msg("error requesting tx receipt for signature")
		}
	} else if len(groupedSigSlotList) > 1 {
		var wg sync.WaitGroup

		for _, group := range groupedSigSlotList {
			wg.Add(1)

			go func(g []SolTxSigSlotInfo) {
				defer wg.Done()
				if err := c.requestTxMetaList(g...); err != nil {
					c.logger.Error().Err(err).Msg("error requesting tx receipt for signature")
				}
			}(group)
		}
		wg.Wait()
	}
}

func (c *сollector) getTxMeta(sigSlotList []SolTxSigSlotInfo) ([]*SolTxMetaInfo, error) {
	const groupLen = 20
	filteredSigSlotList := make([]SolTxSigSlotInfo, 0, len(sigSlotList))

	for _, sigSlot := range sigSlotList {
		// check if we already collected tx meta for this sigSlot
		if !c.txMetaDict.HasSig(sigSlot) {
			filteredSigSlotList = append(filteredSigSlotList, sigSlot)
		}
	}

	flatLen := len(filteredSigSlotList)
	groupedSigSlotList := make([][]SolTxSigSlotInfo, 0, (flatLen+groupLen-1)/groupLen)

	// group filtered sigSlot to batches
	for i := 0; i < flatLen; i += groupLen {
		end := i + groupLen
		if end > flatLen {
			end = flatLen
		}
		groupedSigSlotList = append(groupedSigSlotList, filteredSigSlotList[i:end])
	}

	// get info about each tx by its solana signature
	c.gatherInfoIntoTxMetaDict(groupedSigSlotList)

	resultSlice := make([]*SolTxMetaInfo, len(sigSlotList))
	for i := 0; i < len(sigSlotList); i++ {
		sig, err := c.txMetaDict.Get(sigSlotList[i])
		if err != nil {
			return nil, err
		}
		resultSlice[i] = sig
	}
	return resultSlice, nil
}

// getSignInfoBySlot gets solana signatures for EvmLoaderID address for slots, which re between
// startSlot and stopSlot till it reaches the stopSlot or will receive 0 signatures for slot
func (c *сollector) getSignInfoBySlot(startSign *solana2.Signature, startSlot int64, stopSlot int64) ([]SolTxSigSlotInfo, error) {
	var result []SolTxSigSlotInfo
	for {
		responseList, err := c.solanaClient.GetSignListForAddress(
			context.Background(),
			c.cfg.EvmLoaderID,
			c.cfg.IndexerPollCnt,
			startSign,
			nil,
			c.GetCommitment(),
			nil,
		)
		if err != nil {
			return nil, err
		}
		if len(responseList) == 0 {
			break
		}

		*startSign = responseList[len(responseList)-1].Signature

		for _, response := range responseList {
			blockSlot := response.Slot
			if blockSlot > uint64(startSlot) {
				continue
			} else if blockSlot < uint64(stopSlot) {
				return result, nil
			}
			result = append(result, SolTxSigSlotInfo{
				BlockSlot: int(blockSlot),
				SolSign:   response.Signature,
			})
		}
	}
	return result, nil
}

func (c *сollector) RunSolanaTxs() {

}

package indexer

import (
	"context"
	solana2 "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/neonlabsorg/neon-proxy/internal/indexer/config"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"github.com/neonlabsorg/neon-proxy/pkg/solana"
	"sync"
)

type SolanaTxMetaCollector struct {
	cfg    *config.IndexerConfig
	logger logger.Logger

	solanaClient *solana.Client
	txMetaDict   *SolTxMetaDict
	commitment   rpc.CommitmentType
	isFinalized  bool
}

type CollectorInterface interface {
	IterTxMeta(startSlot, stopSlot uint64) ([]*SolTxMetaInfo, error)
	GetCommitment() rpc.CommitmentType
	IsFinalized() bool
}

func NewSolanaTxMetaCollector(
	cfg *config.IndexerConfig,
	logger logger.Logger,
	solanaClient *solana.Client,
	dict *SolTxMetaDict,
	commitment rpc.CommitmentType,
	isFinalized bool) *SolanaTxMetaCollector {
	return &SolanaTxMetaCollector{
		cfg:          cfg,
		logger:       logger,
		solanaClient: solanaClient,
		txMetaDict:   dict,
		commitment:   commitment,
		isFinalized:  isFinalized,
	}
}

func (c *SolanaTxMetaCollector) GetCommitment() rpc.CommitmentType {
	return c.commitment
}

func (c *SolanaTxMetaCollector) IsFinalized() bool {
	return c.isFinalized
}

// requestTxMetaList gets solana signatures and requests for each signature tx receipt from solana
// and adds info to txMetaDict
func (c *SolanaTxMetaCollector) requestTxMetaList(sigSlotList ...SolTxSigSlotInfo) error {
	sigList := make([]solana2.Signature, 0, len(sigSlotList))
	for _, sigSlot := range sigSlotList {
		sigList = append(sigList, sigSlot.SolSign)
	}

	metaList, err := c.solanaClient.GetTxReceiptList(context.Background(), sigList, solana2.EncodingJSON, c.GetCommitment())
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
func (c *SolanaTxMetaCollector) gatherInfoIntoTxMetaDict(groupedSigSlotList [][]SolTxSigSlotInfo) {
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

// iterTxMeta returns meta (full) info about txs requested from solana by
// txs signatures
func (c *SolanaTxMetaCollector) iterTxMeta(sigSlotList []SolTxSigSlotInfo) ([]*SolTxMetaInfo, error) {
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

	// get info about each tx by its solana signature and save it to txMetaDict
	c.gatherInfoIntoTxMetaDict(groupedSigSlotList)

	resultSlice := make([]*SolTxMetaInfo, len(sigSlotList))
	for i := 0; i < len(sigSlotList); i++ {
		// check if we found info about each signature, which got from solana
		sig, err := c.txMetaDict.Get(sigSlotList[i])
		if err != nil {
			return nil, err
		}
		resultSlice[i] = sig
	}
	return resultSlice, nil
}

// iterSigSlot requests tx signatures by given address in set up slot diapason
func (c *SolanaTxMetaCollector) iterSigSlot(startSign *solana2.Signature, startSlot, stopSlot uint64) ([]SolTxSigSlotInfo, error) {
	var result []SolTxSigSlotInfo
	for {
		// get signatures from newest to oldest in amount of set up limit
		responseList, err := c.solanaClient.GetSignListForAddress(
			context.Background(),
			c.cfg.EvmLoaderID,
			&c.cfg.IndexerPollCnt,
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

		// set up new start signature, to look up new signatures backwards
		if startSign != nil {
			*startSign = responseList[len(responseList)-1].Signature
		} else {
			startSign = &responseList[len(responseList)-1].Signature
		}

		// filter signatures, which are not in slots diapason
		// signatures are ranged by slots from newest to oldest
		for _, response := range responseList {
			blockSlot := response.Slot
			if blockSlot > startSlot {
				// signature is too new, keep filtering
				continue
			} else if blockSlot < stopSlot {
				// signature is too old and next will be even older
				return result, nil
			}
			result = append(result, SolTxSigSlotInfo{
				BlockSlot: blockSlot,
				SolSign:   response.Signature,
			})
		}
	}
	return result, nil
}

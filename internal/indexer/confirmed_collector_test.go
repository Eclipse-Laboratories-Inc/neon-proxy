package indexer

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/golang/mock/gomock"
	"github.com/neonlabsorg/neon-proxy/internal/indexer/config"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	csolana "github.com/neonlabsorg/neon-proxy/pkg/solana"
	"github.com/neonlabsorg/neon-proxy/pkg/solana/mocks"
	"github.com/test-go/testify/assert"
	"testing"
)

func TestConfirmedCollector_IterTxMeta(t *testing.T) {
	const (
		indexerPollCnt = 10
		currentSlot    = 21
		startSlot      = 20
		stopSlot       = 19
		oldestSlot     = 18
	)
	var (
		alreadyKnownTxSignature, newSolanaSignature1, newSolanaSignature2, newSolanaSignature3 solana.Signature
		err                                                                                    error
		evmLoader                                                                              solana.PublicKey
	)

	evmLoader, err = solana.PublicKeyFromBase58("9ZNTfG4NyQgxy2SWjSiQoUyBPEvXT2xo7fKc5hPYYJ7b")
	assert.NoError(t, err)

	alreadyKnownTxSignature, err = solana.SignatureFromBase58("5h6xBEauJ3PK6SWCZ1PGjBvj8vDdWG3KpwATGy1ARAXFSDwt8GFXM7W5Ncn16wmqokgpiKRLuS83KUxyZyv2sUYv")
	assert.NoError(t, err)

	newSolanaSignature1, err = solana.SignatureFromBase58("2nBhEBYYvfaAe16UMNqRHre4YNSskvuYgx3M6E4JP1oDYvZEJHvoPzyUidNgNX5r9sTyN1J9UxtbCXy2rqYcuyuv")
	assert.NoError(t, err)

	newSolanaSignature2, err = solana.SignatureFromBase58("5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJLgp8uirBgmQpjKhoR4tjF3ZpRzrFmBV6UjKdiSZkQUW")
	assert.NoError(t, err)

	newSolanaSignature3, err = solana.SignatureFromBase58("3T7fk6qjfBchQ2miP7T22TqhPVxP4rAEVrJtLnEsxsFV7fmawzgZMb7TrMH4y5isuDCsu4YiSgSjG4pix9JZGm2j")
	assert.NoError(t, err)

	alreadyKnownTxSlotInfo := SolTxSigSlotInfo{
		SolSign:   alreadyKnownTxSignature,
		BlockSlot: stopSlot,
	}
	alreadyKnownTxReceipt := SolTxReceipt{
		Slot: stopSlot,
		Transaction: &csolana.TransactionResultEnvelope{
			AsParsedTransaction: &solana.Transaction{
				Signatures: []solana.Signature{alreadyKnownTxSignature},
			},
		},
		Meta: &rpc.TransactionMeta{},
	}
	txMetaDict := NewSolTxMetaDict()
	err = txMetaDict.Add(alreadyKnownTxSlotInfo, &alreadyKnownTxReceipt)
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	m := mocks.NewMockSolanaRpcConnection(ctrl)

	log, err := logger.NewLogger("test", logger.LogSettings{})
	assert.NoError(t, err)
	cl := csolana.NewClient(log, m)

	expectedMetaTxInfos := []*SolTxMetaInfo{
		{
			ident:     alreadyKnownTxSlotInfo,
			blockSlot: stopSlot,
		},
		{
			ident: SolTxSigSlotInfo{
				SolSign:   newSolanaSignature1,
				BlockSlot: stopSlot,
			},
			blockSlot: stopSlot,
		},
		{
			ident: SolTxSigSlotInfo{
				SolSign:   newSolanaSignature2,
				BlockSlot: startSlot,
			},
			blockSlot: startSlot,
		},
		{
			ident: SolTxSigSlotInfo{
				SolSign:   newSolanaSignature3,
				BlockSlot: startSlot,
			},
			blockSlot: startSlot,
		},
	}
	collector := NewConfirmedCollector(&config.IndexerConfig{
		EvmLoaderID:    evmLoader,
		IndexerPollCnt: indexerPollCnt,
	}, log, cl, txMetaDict, false)

	// Test scenario:
	// 1. Get list of solana tx signatures in given slots diapason
	// 2. Gather full information about txs by signatures got in step 1.
	t.Run("success: got info about 3 new transactions from start slot to stop slot", func(t *testing.T) {
		indexerPollCnt := indexerPollCnt
		gomock.InOrder(
			m.EXPECT().GetSignaturesForAddressWithOpts(gomock.Any(), evmLoader, &rpc.GetSignaturesForAddressOpts{
				Limit:      &indexerPollCnt,
				Commitment: collector.GetCommitment(),
			}).Return(
				[]*rpc.TransactionSignature{
					{
						Slot:               currentSlot,
						ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
					},
					{
						Signature:          newSolanaSignature3,
						Slot:               startSlot,
						ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
					},
					{
						Signature:          newSolanaSignature2,
						Slot:               startSlot,
						ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
					},
					{
						Signature:          newSolanaSignature1,
						Slot:               stopSlot,
						ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
					},
					{
						Signature:          alreadyKnownTxSignature,
						Slot:               stopSlot,
						ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
					},
					{
						Slot:               oldestSlot,
						ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
					},
				},
				nil,
			),
			m.EXPECT().GetTransaction(gomock.Any(), newSolanaSignature3, &rpc.GetTransactionOpts{
				Encoding:   solana.EncodingJSON,
				Commitment: rpc.CommitmentConfirmed,
			}).Return(&rpc.GetTransactionResult{
				Slot: startSlot, Transaction: &rpc.TransactionResultEnvelope{},
				Meta: &rpc.TransactionMeta{
					Fee: 10,
				},
			}, nil),

			m.EXPECT().GetTransaction(gomock.Any(), newSolanaSignature2, &rpc.GetTransactionOpts{
				Encoding:   solana.EncodingJSON,
				Commitment: rpc.CommitmentConfirmed,
			}).Return(&rpc.GetTransactionResult{
				Slot: startSlot, Transaction: &rpc.TransactionResultEnvelope{},
				Meta: &rpc.TransactionMeta{
					Fee: 13,
				},
			}, nil),

			m.EXPECT().GetTransaction(gomock.Any(), newSolanaSignature1, &rpc.GetTransactionOpts{
				Encoding:   solana.EncodingJSON,
				Commitment: rpc.CommitmentConfirmed,
			}).Return(&rpc.GetTransactionResult{
				Slot: stopSlot, Transaction: &rpc.TransactionResultEnvelope{},
				Meta: &rpc.TransactionMeta{
					Fee: 11,
				},
			}, nil),
		)

		txMetaInfos, err := collector.IterTxMeta(startSlot, stopSlot)
		assert.NoError(t, err)
		assert.Equal(t, len(expectedMetaTxInfos), len(txMetaInfos))
		for _, expectedInfo := range expectedMetaTxInfos {
			found := false
			for _, actualInfo := range txMetaInfos {
				if expectedInfo.ident.SolSign == actualInfo.ident.SolSign {
					found = true
					assert.Equal(t, expectedInfo.ident.BlockSlot, actualInfo.ident.BlockSlot)
				}
			}
			assert.True(t, found)
		}
	})

	// Test scenario:
	// 1. Get list of signatures in given slots diapason - no signatures found
	t.Run("success:  no signatures found", func(t *testing.T) {
		indexerPollCnt := indexerPollCnt

		m.EXPECT().GetSignaturesForAddressWithOpts(gomock.Any(), evmLoader, &rpc.GetSignaturesForAddressOpts{
			Limit:      &indexerPollCnt,
			Commitment: collector.GetCommitment(),
		}).Return(
			[]*rpc.TransactionSignature{
				{
					Slot:               currentSlot,
					ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
				},
				{
					Signature:          newSolanaSignature3,
					Slot:               startSlot,
					ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
				},
				{
					Signature:          newSolanaSignature2,
					Slot:               startSlot,
					ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
				},
				{
					Signature:          newSolanaSignature1,
					Slot:               stopSlot,
					ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
				},
				{
					Signature:          alreadyKnownTxSignature,
					Slot:               stopSlot,
					ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
				},
				{
					Slot:               oldestSlot,
					ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
				},
			}, nil)

		txMetaInfos, err := collector.IterTxMeta(40, 30)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(txMetaInfos))
	})
}

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

func TestFinalizedCollector_IterTxMeta(t *testing.T) {
	const (
		indexerPollCnt = 10
		currentSlot    = 21
		startSlot      = 20
		stopSlot       = 19
		oldestSlot     = 9
	)
	var (
		oldesSolanaSignature, newSolanaSignature1, newSolanaSignature2, newSolanaSignature3 solana.Signature
		err                                                                                 error
		evmLoader                                                                           solana.PublicKey
	)

	evmLoader, err = solana.PublicKeyFromBase58("9ZNTfG4NyQgxy2SWjSiQoUyBPEvXT2xo7fKc5hPYYJ7b")
	assert.NoError(t, err)

	newSolanaSignature1, err = solana.SignatureFromBase58("2nBhEBYYvfaAe16UMNqRHre4YNSskvuYgx3M6E4JP1oDYvZEJHvoPzyUidNgNX5r9sTyN1J9UxtbCXy2rqYcuyuv")
	assert.NoError(t, err)

	newSolanaSignature2, err = solana.SignatureFromBase58("5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJLgp8uirBgmQpjKhoR4tjF3ZpRzrFmBV6UjKdiSZkQUW")
	assert.NoError(t, err)

	newSolanaSignature3, err = solana.SignatureFromBase58("3T7fk6qjfBchQ2miP7T22TqhPVxP4rAEVrJtLnEsxsFV7fmawzgZMb7TrMH4y5isuDCsu4YiSgSjG4pix9JZGm2j")
	assert.NoError(t, err)

	oldesSolanaSignature, err = solana.SignatureFromBase58("3JXKEie3WbMfohy6wnVkfYnhbAVJcCQ7xKfq3yabqdseUvf9tpKyWZa4o1Xw2muFVSaf8jiweYWn7YXcCGMoNJ8m")
	assert.NoError(t, err)

	txMetaDict := NewSolTxMetaDict()

	ctrl := gomock.NewController(t)
	db := NewMockSolanaSignsDBInterface(ctrl)

	log, err := logger.NewLogger("test", logger.LogSettings{})
	assert.NoError(t, err)

	// Test scenario:
	// Gather short list of signatures from solana: gather info from less, than 10 slots.
	// We do not have any saved checkpoint in db: start getting info from top of blockchain.
	// 1. Try to take last saved checkpoint - no checkpoints saved.
	// 2. Take signatures from top of blockchain and filter them by slots (only signatures, which are in slots between start and stop slot will be in result slice)
	// 3. Try to save the freshest signature as checkpoint: checkpoint isn't saved because we collected to small amount of signatures in step 2.
	// 4. Gather full info about signatures, got in step 2 and save into txMetaDict
	t.Run("success with short list: 3 new signatures found, no checkpoints saved", func(t *testing.T) {
		indexerPollCnt := indexerPollCnt
		rpcConn := mocks.NewMockSolanaRpcConnection(ctrl)
		cl := csolana.NewClient(log, rpcConn)

		collector := NewFinalizedCollector(
			&config.IndexerConfig{
				EvmLoaderID:    evmLoader,
				IndexerPollCnt: indexerPollCnt,
			}, log, cl, txMetaDict, db, nil, stopSlot, 0, true,
		)

		expectedMetaTxInfos := []*SolTxMetaInfo{
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

		gomock.InOrder(
			rpcConn.EXPECT().GetSignaturesForAddressWithOpts(gomock.Any(), evmLoader, &rpc.GetSignaturesForAddressOpts{
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
						Slot:               oldestSlot,
						ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
					},
				},
				nil,
			),

			rpcConn.EXPECT().GetTransaction(gomock.Any(), newSolanaSignature3, &rpc.GetTransactionOpts{
				Encoding:   solana.EncodingJSON,
				Commitment: rpc.CommitmentFinalized,
			}).Return(&rpc.GetTransactionResult{
				Slot: startSlot, Transaction: &rpc.TransactionResultEnvelope{},
				Meta: &rpc.TransactionMeta{
					Fee: 10,
				},
			}, nil),

			rpcConn.EXPECT().GetTransaction(gomock.Any(), newSolanaSignature2, &rpc.GetTransactionOpts{
				Encoding:   solana.EncodingJSON,
				Commitment: rpc.CommitmentFinalized,
			}).Return(&rpc.GetTransactionResult{
				Slot: startSlot, Transaction: &rpc.TransactionResultEnvelope{},
				Meta: &rpc.TransactionMeta{
					Fee: 13,
				},
			}, nil),

			rpcConn.EXPECT().GetTransaction(gomock.Any(), newSolanaSignature1, &rpc.GetTransactionOpts{
				Encoding:   solana.EncodingJSON,
				Commitment: rpc.CommitmentFinalized,
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
	// Gather long list of signatures from solana: gather info from more, than 10 slots.
	// We do not have any saved checkpoint in db: we start getting info from top of blockchain.
	// 1. Take signatures from top of blockchain and filter them by slots (only signatures, which are in slots between start start and stop slot will be in result slice)
	// 2. Try to save N signatures into db as checkpoint: 0 checkpoint saved because of to small amount of signatures got in step 1.
	// 4. Look up for signatures from top of blockchain.
	// 5. Gather full info about signatures, got in step 4 and save into txMetaDict
	t.Run("success with long list: 4 new signatures found and no checkpoints saved", func(t *testing.T) {
		indexerPollCnt := 20
		rpcConn := mocks.NewMockSolanaRpcConnection(ctrl)
		cl := csolana.NewClient(log, rpcConn)

		collector := NewFinalizedCollector(
			&config.IndexerConfig{
				EvmLoaderID:    evmLoader,
				IndexerPollCnt: indexerPollCnt,
			}, log, cl, txMetaDict, db, nil, oldestSlot, 0, true,
		)
		expectedMetaTxInfos := []*SolTxMetaInfo{
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
			{
				ident: SolTxSigSlotInfo{
					SolSign:   oldesSolanaSignature,
					BlockSlot: oldestSlot,
				},
				blockSlot: oldestSlot,
			},
		}

		gomock.InOrder(
			db.EXPECT().GetMaxSign().Return(nil, ErrRecordNotFound),

			rpcConn.EXPECT().GetSignaturesForAddressWithOpts(gomock.Any(), evmLoader, &rpc.GetSignaturesForAddressOpts{
				Limit:      &indexerPollCnt,
				Commitment: collector.GetCommitment(),
			}).Return(
				[]*rpc.TransactionSignature{
					{
						Slot:               currentSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
					{
						Signature:          newSolanaSignature3,
						Slot:               startSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
					{
						Signature:          newSolanaSignature2,
						Slot:               startSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
					{
						Signature:          newSolanaSignature1,
						Slot:               stopSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
					{
						Signature:          oldesSolanaSignature,
						Slot:               oldestSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
				}, nil),

			rpcConn.EXPECT().GetSignaturesForAddressWithOpts(gomock.Any(), evmLoader, &rpc.GetSignaturesForAddressOpts{
				Before:     oldesSolanaSignature,
				Limit:      &indexerPollCnt,
				Commitment: collector.GetCommitment(),
			}).Return(nil, nil),

			db.EXPECT().GetNextSign(uint64(oldestSlot)).Return(nil, ErrRecordNotFound),

			rpcConn.EXPECT().GetSignaturesForAddressWithOpts(gomock.Any(), evmLoader, &rpc.GetSignaturesForAddressOpts{
				Limit:      &indexerPollCnt,
				Commitment: collector.GetCommitment(),
			}).Return(
				[]*rpc.TransactionSignature{
					{
						Slot:               currentSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
					{
						Signature:          newSolanaSignature3,
						Slot:               startSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
					{
						Signature:          newSolanaSignature2,
						Slot:               startSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
					{
						Signature:          newSolanaSignature1,
						Slot:               stopSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
					{
						Signature:          oldesSolanaSignature,
						Slot:               oldestSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
				}, nil),

			rpcConn.EXPECT().GetSignaturesForAddressWithOpts(gomock.Any(), evmLoader, &rpc.GetSignaturesForAddressOpts{
				Before:     oldesSolanaSignature,
				Limit:      &indexerPollCnt,
				Commitment: collector.GetCommitment(),
			}).Return(nil, nil),

			rpcConn.EXPECT().GetTransaction(gomock.Any(), newSolanaSignature3, &rpc.GetTransactionOpts{
				Encoding:   solana.EncodingJSON,
				Commitment: rpc.CommitmentFinalized,
			}).Return(&rpc.GetTransactionResult{
				Slot: startSlot, Transaction: &rpc.TransactionResultEnvelope{},
				Meta: &rpc.TransactionMeta{
					Fee: 10,
				},
			}, nil),

			rpcConn.EXPECT().GetTransaction(gomock.Any(), newSolanaSignature2, &rpc.GetTransactionOpts{
				Encoding:   solana.EncodingJSON,
				Commitment: rpc.CommitmentFinalized,
			}).Return(&rpc.GetTransactionResult{
				Slot: startSlot, Transaction: &rpc.TransactionResultEnvelope{},
				Meta: &rpc.TransactionMeta{
					Fee: 13,
				},
			}, nil),

			rpcConn.EXPECT().GetTransaction(gomock.Any(), newSolanaSignature1, &rpc.GetTransactionOpts{
				Encoding:   solana.EncodingJSON,
				Commitment: rpc.CommitmentFinalized,
			}).Return(&rpc.GetTransactionResult{
				Slot: stopSlot, Transaction: &rpc.TransactionResultEnvelope{},
				Meta: &rpc.TransactionMeta{
					Fee: 11,
				},
			}, nil),

			rpcConn.EXPECT().GetTransaction(gomock.Any(), oldesSolanaSignature, &rpc.GetTransactionOpts{
				Encoding:   solana.EncodingJSON,
				Commitment: rpc.CommitmentFinalized,
			}).Return(&rpc.GetTransactionResult{
				Slot: oldestSlot, Transaction: &rpc.TransactionResultEnvelope{},
				Meta: &rpc.TransactionMeta{
					Fee: 12,
				},
			}, nil),
		)

		txMetaInfos, err := collector.IterTxMeta(startSlot, oldestSlot)
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
	// Gather long list of signatures from solana: it means we gather info from more, than 10 slots.
	// We do not have any saved checkpoint in db - it means we start getting info from top of blockchain.
	// 1. Take signatures from top of blockchain and filter them by slots (only signatures, which are in slots between start start and stop slot will be in result slice)
	// 2. Save N signatures into db as checkpoint: in our case we will save only 1 checkpoint, because as result of step 1 we will get 4 signatures and indexerPollCnt setting = 2.
	// 3. Look up for new signatures using the nearest checkpoint to stopSlot.
	// 4. Look up for signatures starSlot to the latest checkpoint
	// 5. Gather full info about signatures, got in step 3-4 and save into txMetaDict
	t.Run("success with long list: 4 new signatures found and 1 checkpoint saved", func(t *testing.T) {
		indexerPollCnt := 2
		rpcConn := mocks.NewMockSolanaRpcConnection(ctrl)
		cl := csolana.NewClient(log, rpcConn)

		checkPoint := &SolTxSigSlotInfo{
			SolSign:   newSolanaSignature2,
			BlockSlot: startSlot,
		}

		collector := NewFinalizedCollector(
			&config.IndexerConfig{
				EvmLoaderID:    evmLoader,
				IndexerPollCnt: indexerPollCnt,
			}, log, cl, txMetaDict, db, nil, oldestSlot, int64(indexerPollCnt), true,
		)
		expectedMetaTxInfos := []*SolTxMetaInfo{
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
			{
				ident: SolTxSigSlotInfo{
					SolSign:   oldesSolanaSignature,
					BlockSlot: oldestSlot,
				},
				blockSlot: oldestSlot,
			},
		}

		gomock.InOrder(
			db.EXPECT().GetMaxSign().Return(nil, ErrRecordNotFound),

			// get signatures from top of blockchain
			rpcConn.EXPECT().GetSignaturesForAddressWithOpts(gomock.Any(), evmLoader, &rpc.GetSignaturesForAddressOpts{
				Limit:      &indexerPollCnt,
				Commitment: collector.GetCommitment(),
			}).Return(
				[]*rpc.TransactionSignature{
					{
						Slot:               currentSlot, // this signature is out of slots diapason
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
					{
						Signature:          newSolanaSignature3,
						Slot:               startSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
				}, nil),

			// get signatures, which are before newSolanaSignature3
			rpcConn.EXPECT().GetSignaturesForAddressWithOpts(gomock.Any(), evmLoader, &rpc.GetSignaturesForAddressOpts{
				Before:     newSolanaSignature3,
				Limit:      &indexerPollCnt,
				Commitment: collector.GetCommitment(),
			}).Return(
				[]*rpc.TransactionSignature{
					{
						Signature:          checkPoint.SolSign,
						Slot:               checkPoint.BlockSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
					{
						Signature:          newSolanaSignature1,
						Slot:               stopSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
				}, nil),

			// get signatures, which are before newSolanaSignature1
			rpcConn.EXPECT().GetSignaturesForAddressWithOpts(gomock.Any(), evmLoader, &rpc.GetSignaturesForAddressOpts{
				Before:     newSolanaSignature1,
				Limit:      &indexerPollCnt,
				Commitment: collector.GetCommitment(),
			}).Return(
				[]*rpc.TransactionSignature{
					{
						Signature:          oldesSolanaSignature, // cnt = 1
						Slot:               oldestSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
				}, nil),

			// get signatures, which are before oldesSolanaSignature - no signatures found
			rpcConn.EXPECT().GetSignaturesForAddressWithOpts(gomock.Any(), evmLoader, &rpc.GetSignaturesForAddressOpts{
				Before:     oldesSolanaSignature,
				Limit:      &indexerPollCnt,
				Commitment: collector.GetCommitment(),
			}).Return(nil, nil),

			// save checkpoint
			db.EXPECT().AddSign(*checkPoint).Return(nil),

			// get nearest to oldestSlot saved checkpoint
			db.EXPECT().GetNextSign(uint64(oldestSlot)).Return(checkPoint, nil),

			// get signatures from solana, which are before signature saved as checkpoint
			rpcConn.EXPECT().GetSignaturesForAddressWithOpts(gomock.Any(), evmLoader, &rpc.GetSignaturesForAddressOpts{
				Before:     checkPoint.SolSign,
				Limit:      &indexerPollCnt,
				Commitment: collector.GetCommitment(),
			}).Return(
				[]*rpc.TransactionSignature{
					{
						Signature:          newSolanaSignature1,
						Slot:               stopSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
					{
						Signature:          oldesSolanaSignature,
						Slot:               oldestSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
				}, nil),

			rpcConn.EXPECT().GetSignaturesForAddressWithOpts(gomock.Any(), evmLoader, &rpc.GetSignaturesForAddressOpts{
				Before:     oldesSolanaSignature,
				Limit:      &indexerPollCnt,
				Commitment: collector.GetCommitment(),
			}).Return(nil, nil),

			db.EXPECT().GetNextSign(uint64(startSlot)).Return(nil, ErrRecordNotFound),

			// get signatures from top of blockchain
			rpcConn.EXPECT().GetSignaturesForAddressWithOpts(gomock.Any(), evmLoader, &rpc.GetSignaturesForAddressOpts{
				Limit:      &indexerPollCnt,
				Commitment: collector.GetCommitment(),
			}).Return(
				[]*rpc.TransactionSignature{
					{
						Slot:               currentSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
					{
						Signature:          newSolanaSignature3,
						Slot:               startSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
				}, nil),

			rpcConn.EXPECT().GetSignaturesForAddressWithOpts(gomock.Any(), evmLoader, &rpc.GetSignaturesForAddressOpts{
				Before:     newSolanaSignature3,
				Limit:      &indexerPollCnt,
				Commitment: collector.GetCommitment(),
			}).Return(
				[]*rpc.TransactionSignature{
					{
						Signature:          checkPoint.SolSign,
						Slot:               checkPoint.BlockSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
					{
						Signature:          newSolanaSignature1,
						Slot:               stopSlot,
						ConfirmationStatus: rpc.ConfirmationStatusFinalized,
					},
				}, nil),

			// get full info about tx by its signature
			rpcConn.EXPECT().GetTransaction(gomock.Any(), newSolanaSignature1, &rpc.GetTransactionOpts{
				Encoding:   solana.EncodingJSON,
				Commitment: rpc.CommitmentFinalized,
			}).Return(&rpc.GetTransactionResult{
				Slot: stopSlot, Transaction: &rpc.TransactionResultEnvelope{},
				Meta: &rpc.TransactionMeta{
					Fee: 10,
				},
			}, nil),

			rpcConn.EXPECT().GetTransaction(gomock.Any(), oldesSolanaSignature, &rpc.GetTransactionOpts{
				Encoding:   solana.EncodingJSON,
				Commitment: rpc.CommitmentFinalized,
			}).Return(&rpc.GetTransactionResult{
				Slot: oldestSlot, Transaction: &rpc.TransactionResultEnvelope{},
				Meta: &rpc.TransactionMeta{
					Fee: 12,
				},
			}, nil),

			rpcConn.EXPECT().GetTransaction(gomock.Any(), newSolanaSignature3, &rpc.GetTransactionOpts{
				Encoding:   solana.EncodingJSON,
				Commitment: rpc.CommitmentFinalized,
			}).Return(&rpc.GetTransactionResult{
				Slot: startSlot, Transaction: &rpc.TransactionResultEnvelope{},
				Meta: &rpc.TransactionMeta{
					Fee: 11,
				},
			}, nil),

			rpcConn.EXPECT().GetTransaction(gomock.Any(), checkPoint.SolSign, &rpc.GetTransactionOpts{
				Encoding:   solana.EncodingJSON,
				Commitment: rpc.CommitmentFinalized,
			}).Return(&rpc.GetTransactionResult{
				Slot: checkPoint.BlockSlot, Transaction: &rpc.TransactionResultEnvelope{},
				Meta: &rpc.TransactionMeta{
					Fee: 13,
				},
			}, nil),
		)

		txMetaInfos, err := collector.IterTxMeta(startSlot, oldestSlot)
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
}

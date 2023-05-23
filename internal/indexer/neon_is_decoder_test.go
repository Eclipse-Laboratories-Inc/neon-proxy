package indexer

import (
	"encoding/hex"
	"fmt"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	csolana "github.com/neonlabsorg/neon-proxy/pkg/solana"
	"github.com/test-go/testify/assert"
	"strings"
	"testing"
)

const (
	testSolanaSign = "54686520717569636b2062726f776e20666f78206a756d7073206f7665722074"
	testHolderAcc  = "3UVYmECPPMZSCqWKfENfuoTv51fTDTWicX9xmBD2euKe"
)

var (
	testAccountKeys = []string{
		"3UVYmECPPMZSCqWKfENfuoTv51fTDTWicX9xmBD2euKe",
		"AjozzgE83A3x1sHNUR64hfH7zaEBWeMaFuAN9kQgujrc",
		"SysvarS1otHashes111111111111111111111111111",
		"SysvarC1ock11111111111111111111111111111111",
		"Vote111111111111111111111111111111111111111",
		"SysvarC1ock11111111111111111111111111111111",
		"Vote111111111111111111111111111111111111111"}
)

func initUnifiedTestSolState(t *testing.T) SolNeonTxDecoderState {
	data, err := hex.DecodeString("54686520717569636b2062726f776e20666f78206a756d7073206f76657220746865206c617a7920646f672e")
	assert.Empty(t, err)

	return SolNeonTxDecoderState{
		solNeonIxCnt: 1,
		solTx: &SolTxReceiptInfo{
			operator:       testHolderAcc,
			ixList:         make([]map[string]string, 0),
			innerIxList:    make([]map[string]string, 0),
			accountKeyList: testAccountKeys,
			ixLogMsgList:   []SolIxLogState{},
		},
		solTxMeta: &SolTxMetaInfo{
			tx: &SolTxReceipt{
				Slot:        430,
				Transaction: &csolana.TransactionResultEnvelope{},
			},
		},
		solNeonIx: &SolNeonIxReceiptInfo{
			metaInfo: SolIxMetaInfo{
				neonTxSig:        "0x" + testSolanaSign,
				neonTotalGasUsed: 32131890,
			},
			solSign: testSolanaSign,
			ixData:  append([]byte{0}, data...), // TODO switch to real data after impl NewNeonTxFromSigData
			ident: Ident{
				solSign: testSolanaSign,
			},
			accounts:    []int{0, 1, 2, 3, 4, 5, 6},
			accountKeys: testAccountKeys,
		},
		neonBlockDeque: []NeonIndexedBlockData{
			{
				neonIndexedBlockInfo: &NeonIndexedBlockInfo{
					solBlock: SolBlockInfo{
						BlockSlot: 430,
						Finalized: true,
					},
					neonTxs:     make(map[string]*NeonIndexedTxInfo),
					neonHolders: make(map[string]NeonIndexedHolderInfo),
				},
				finalized: true,
			},
		},
	}
}

func TestTxExecFromDataIxDecoder(t *testing.T) {
	log, err := logger.NewLogger("test", logger.LogSettings{})
	assert.NoError(t, err)

	t.Run("TxExecFromDataIxDecoder decoding single tx failed: empty ix.Data", func(t *testing.T) {
		decoder := InitTxExecFromDataIxDecoder(&IxDecoder{
			log:    log,
			name:   "TransactionExecuteFromInstruction",
			ixCode: 0x1f,
			state:  SolNeonTxDecoderState{solNeonIx: &SolNeonIxReceiptInfo{}}})

		result := decoder.Execute()
		assert.False(t, result)
	})

	t.Run("TxExecFromDataIxDecoder decoding single tx without log events: success",
		func(t *testing.T) {
			expectedDecodedTxKey := strings.ToLower(testSolanaSign)
			solState := initUnifiedTestSolState(t)

			decoder := InitTxExecFromDataIxDecoder(&IxDecoder{
				log:    log,
				name:   "TransactionExecuteFromInstruction",
				ixCode: 0x1f,
				state:  solState,
			})

			result := decoder.Execute()
			assert.True(t, result)

			tx, ok := solState.NeonBlock().neonTxs[expectedDecodedTxKey]
			assert.True(t, ok)
			assert.Equal(t, NeonIndexedTxTypeSingle, tx.txType)
			assert.Equal(t, "", tx.storageAccount)
			assert.Equal(t, NeonIndexedTxInfoStatusInProgress, tx.status)
		})
}

func TestTxExecFromAccountIxDecoder(t *testing.T) {
	log, err := logger.NewLogger("test", logger.LogSettings{})
	assert.NoError(t, err)

	t.Run("TxExecFromAccountIxDecoder decoding single tx failed: no accounts in SolIx", func(t *testing.T) {
		decoder := InitTxExecFromAccountIxDecoder(&IxDecoder{
			log:    log,
			name:   "TransactionExecFromAccount",
			ixCode: 0x2a,
			state:  SolNeonTxDecoderState{solNeonIx: &SolNeonIxReceiptInfo{}}})

		result := decoder.Execute()
		assert.False(t, result)
	})

	t.Run("TxExecFromAccountIxDecoder decoding single tx without receipt and log events: success",
		func(t *testing.T) {
			expectedDecodedTxKey := strings.ToLower(testSolanaSign)

			solState := initUnifiedTestSolState(t)
			decoder := InitTxExecFromAccountIxDecoder(&IxDecoder{
				log:    log,
				name:   "TransactionExecFromAccount",
				ixCode: 0x2a,
				state:  solState,
			})

			result := decoder.Execute()
			assert.True(t, result)

			tx, ok := solState.NeonBlock().neonTxs[expectedDecodedTxKey]
			assert.True(t, ok)
			assert.Equal(t, NeonIndexedTxTypeSingleFromAccount, tx.txType)
			assert.Equal(t, testHolderAcc, tx.storageAccount)
			assert.Equal(t, NeonIndexedTxInfoStatusInProgress, tx.status)
			assert.False(t, tx.canceled)
		})
}

func TestTxStepFromDataIxDecoder(t *testing.T) {
	log, err := logger.NewLogger("test", logger.LogSettings{})
	assert.NoError(t, err)

	t.Run("TxStepFromDataIxDecoder decoding iter tx failed: no ix.Data", func(t *testing.T) {
		decoder := InitTxStepFromDataIxDecoder(&IxDecoder{
			log:    log,
			name:   "TransactionStepFromInstruction",
			ixCode: 0x20,
			state:  SolNeonTxDecoderState{solNeonIx: &SolNeonIxReceiptInfo{}}})

		result := decoder.Execute()
		assert.False(t, result)
	})

	t.Run("TxStepFromDataIxDecoder decoding iter tx without receipt and log events: success",
		func(t *testing.T) {
			expectedDecodedTxKey := testSolanaSign
			solState := initUnifiedTestSolState(t)

			decoder := InitTxStepFromDataIxDecoder(&IxDecoder{
				log:    log,
				name:   "TransactionStepFromInstruction",
				ixCode: 0x20,
				state:  solState,
			})

			result := decoder.Execute()
			assert.True(t, result)

			tx, ok := solState.NeonBlock().neonTxs[expectedDecodedTxKey]
			assert.True(t, ok)
			assert.Equal(t, NeonIndexedTxTypeIterFromData, tx.txType)
			assert.Equal(t, testHolderAcc, tx.storageAccount)
			assert.Equal(t, NeonIndexedTxInfoStatusInProgress, tx.status)
		})
}

func TestTxStepFromAccountIxDecoder(t *testing.T) {
	log, err := logger.NewLogger("test", logger.LogSettings{})
	assert.NoError(t, err)

	t.Run("TestTxStepFromAccountIxDecoder decoding iter tx failed: no ix.Data", func(t *testing.T) {
		decoder := InitTxStepFromAccountIxDecoder(&IxDecoder{
			log:    log,
			name:   "TransactionStepFromAccount",
			ixCode: 0x21,
			state:  SolNeonTxDecoderState{solNeonIx: &SolNeonIxReceiptInfo{}}})

		result := decoder.Execute()
		assert.False(t, result)
	})

	t.Run("TxStepFromAccountIxDecoder decoding iter tx without receipt and log events: success",
		func(t *testing.T) {
			expectedDecodedTxKey := testSolanaSign
			solState := initUnifiedTestSolState(t)

			decoder := InitTxStepFromAccountIxDecoder(&IxDecoder{
				log:    log,
				name:   "TransactionStepFromAccount",
				ixCode: 0x21,
				state:  solState,
			})

			result := decoder.Execute()
			assert.True(t, result)

			tx, ok := solState.NeonBlock().neonTxs[expectedDecodedTxKey]
			assert.True(t, ok)
			assert.Equal(t, NeonIndexedTxTypeIterFromAccount, tx.txType)
			assert.Equal(t, testHolderAcc, tx.storageAccount)
			assert.Equal(t, NeonIndexedTxInfoStatusInProgress, tx.status)
		})
}

func TestTxStepFromAccountNoChainIdIxDecoder(t *testing.T) {
	log, err := logger.NewLogger("test", logger.LogSettings{})
	assert.NoError(t, err)

	t.Run("TestTxStepFromAccountNoChainIdIxDecoder decoding iter tx failed: no ix.Data", func(t *testing.T) {
		decoder := InitTxStepFromAccountNoChainIdIxDecoder(&IxDecoder{
			log:    log,
			name:   "TransactionStepFromAccountNoChainId",
			ixCode: 0x22,
			state:  SolNeonTxDecoderState{solNeonIx: &SolNeonIxReceiptInfo{}}})

		result := decoder.Execute()
		assert.False(t, result)
	})

	t.Run("TxStepFromAccountNoChainIdIxDecoder decoding iter tx without receipt and log events: success",
		func(t *testing.T) {
			expectedDecodedTxKey := testSolanaSign
			solState := initUnifiedTestSolState(t)

			decoder := InitTxStepFromAccountNoChainIdIxDecoder(&IxDecoder{
				log:    log,
				name:   "TransactionStepFromAccountNoChainId",
				ixCode: 0x22,
				state:  solState,
			})

			result := decoder.Execute()
			assert.True(t, result)

			tx, ok := solState.NeonBlock().neonTxs[expectedDecodedTxKey]
			assert.True(t, ok)
			assert.Equal(t, NeonIndexedTxTypeIterFromAccountWoChainId, tx.txType)
			assert.Equal(t, testHolderAcc, tx.storageAccount)
			assert.Equal(t, NeonIndexedTxInfoStatusInProgress, tx.status)
		})
}

func TestWriteHolderAccountIx(t *testing.T) {
	log, err := logger.NewLogger("test", logger.LogSettings{})
	assert.NoError(t, err)

	t.Run("WriteHolderAccountIx decoding iter tx failed: not enough accounts", func(t *testing.T) {
		decoder := InitWriteHolderAccountIx(&IxDecoder{
			log:    log,
			name:   "WriteHolderAccount",
			ixCode: 0x26,
			state:  SolNeonTxDecoderState{solNeonIx: &SolNeonIxReceiptInfo{}}})

		result := decoder.Execute()
		assert.False(t, result)
	})

	t.Run("WriteHolderAccountIx decoding iter tx without receipt and log events: success",
		func(t *testing.T) {
			expectedDecodedTxKey := strings.ToLower(testSolanaSign)
			expectedHolderTxKey := fmt.Sprintf("%v:%v", testHolderAcc, testSolanaSign)
			solState := initUnifiedTestSolState(t)

			decoder := InitWriteHolderAccountIx(&IxDecoder{
				log:    log,
				name:   "WriteHolderAccount",
				ixCode: 0x26,
				state:  solState,
			})

			result := decoder.Execute()
			assert.True(t, result)

			_, ok := solState.NeonBlock().neonTxs[expectedDecodedTxKey]
			assert.False(t, ok)

			_, ok = solState.NeonBlock().neonHolders[expectedHolderTxKey]
			assert.True(t, ok)
		})
}

func TestCancelWithHashIxDecoder(t *testing.T) {
	log, err := logger.NewLogger("test", logger.LogSettings{})
	assert.NoError(t, err)

	t.Run("CancelWithHashIxDecoder decoding iter tx failed: not enough accounts", func(t *testing.T) {
		decoder := InitCancelWithHashIxDecoder(&IxDecoder{
			log:    log,
			name:   "CancelWithHash",
			ixCode: 0x23,
			state:  SolNeonTxDecoderState{solNeonIx: &SolNeonIxReceiptInfo{}}})

		result := decoder.Execute()
		assert.False(t, result)
	})

	t.Run("CancelWithHashIxDecoder decoding iter tx without receipt and log events: success",
		func(t *testing.T) {
			expectedDecodedTxKey := strings.ToLower(testSolanaSign)
			solState := initUnifiedTestSolState(t)

			decoder := InitCancelWithHashIxDecoder(&IxDecoder{
				log:    log,
				name:   "CancelWithHash",
				ixCode: 0x23, // TODO какой тип тогда будет у tx ?
				state:  solState,
			})

			result := decoder.Execute()
			assert.True(t, result)

			tx, ok := solState.NeonBlock().neonTxs[expectedDecodedTxKey]
			assert.True(t, ok)
			assert.Equal(t, testHolderAcc, tx.storageAccount)
			assert.Equal(t, true, tx.neonReceipt.neonTxRes.canceled)
			assert.Equal(t, true, tx.neonReceipt.neonTxRes.completed)
			assert.Equal(t, 4, len(tx.blockedAccounts))
			assert.Equal(t, 1, len(tx.neonEvents))
			assert.Equal(t, testSolanaSign, tx.neonEvents[0].solSig)
		})
}

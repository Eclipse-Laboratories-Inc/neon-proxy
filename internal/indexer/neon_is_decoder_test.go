package indexer

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/utils"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"github.com/test-go/testify/assert"
	"strconv"
	"strings"
	"testing"
)

func initUnifiedTestSolState(operator string, receipt *SolNeonIxReceiptInfo) SolNeonTxDecoderState {
	return SolNeonTxDecoderState{
		solNeonIxCnt: 1,
		solTx: &SolTxReceiptInfo{
			operator:       operator,
			ixList:         make([]map[string]string, 0),
			innerIxList:    make([]map[string]string, 0),
			accountKeyList: receipt.accountKeys,
			ixLogMsgList:   []SolIxLogState{},
		},
		solTxMeta: &SolTxMetaInfo{
			tx: &SolTxReceipt{
				Slot: uint64(receipt.blockSlot),
			},
		},
		solNeonIx: receipt,
		neonBlockDeque: []NeonIndexedBlockData{
			{
				neonIndexedBlockInfo: &NeonIndexedBlockInfo{
					solBlock: SolBlockInfo{
						BlockSlot: receipt.blockSlot,
					},
					neonTxs:     make(map[string]*NeonIndexedTxInfo),
					neonHolders: make(map[string]*NeonIndexedHolderInfo),
				},
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
			const (
				solSig    = "2P5gUGvMziec8onTZYjnxKTAUSBn3qqDKjswzSvV9vbGVpEZBW11n21rMP5K5KsBmhnZtEAFtXT6hciL6mRjM758"
				operator  = "82YcsM5eN83trdhdShGUF4crAC4CGgFJ7EWd2vnGiSsb"
				blockSlot = 6397
			)

			neonTxSig := utils.Base64stringToHex("ctpBMrOwjDxlZyQk2Ii71+E5uMFyWpLMp4uf5fItkYA=")
			gasUsed := int(binary.LittleEndian.Uint32(utils.Base64stringToBytes("ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")))
			totalGasUsed := int(binary.LittleEndian.Uint32(utils.Base64stringToBytes("ECcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")))
			ixData := base58.Decode("B9gky482U43WuVuRnnCM33ZhMWgev1YwHx2ogNpd2kUVWLJM9aAmmztTdHxJ8jcS6nrD8Ztta2G1fmqmgDVzbKvoh8yTLr5mEmEn3SsYw92AF3K7Wjj5Vz4TuZTincozrDbM31htuZUEnYyhNeD3pEA2qxwKP5QwsreB17JAafqWMK6n44pgwfJGbDPvU7dE2cHUB2YchE3urGAcsoftG5xh8in2BxPZ5wgr2yNvfBqhWKafWs9")

			expectedDecodedTxKey := strings.ToLower(neonTxSig[2:])
			neonTxAddr := "0x0614d72ee03a49a956e108d6c2bc553d7272346b"
			expectedNeonTx := NeonTxInfo{
				addr:     &neonTxAddr,
				sig:      neonTxSig,
				nonce:    "03",
				gasPrice: "14ad8bfe11",
				gasLimit: "d6d8",
				toAddr:   "0x3ba699a292c08640982e361d6ee7c186ae948387",
				value:    "",
				callData: "0xb9176795000000000000000000000000d1686024ee7176e06f88e702d00734628af7452f000000000000000000000000906b15c6f1c10fb243aff1750c44236d98ae8e0a",
				v:        "0x102",
				r:        "0xa0f2a91e85f32db798a1369f36b0dcfcf50036dac0aab06ce4caa8bb5a71e73f",
				s:        "0x1e12989c69311c416fbfc0a3e62ac581e0e7a9d96e8308fed1a380d390e75bf0",
			}

			solTx := &SolNeonIxReceiptInfo{
				metaInfo: SolIxMetaInfo{
					status:           1, // success
					neonGasUsed:      int64(gasUsed),
					neonTotalGasUsed: int64(totalGasUsed),
					neonTxSig:        neonTxSig,
					neonTxEvents:     []NeonLogTxEvent{},
				},
				solSign:   solSig,
				blockSlot: blockSlot,
				programIx: 0x1f,
				ixData:    ixData,
				solTxCost: SolTxCostInfo{
					solSign:   solSig,
					blockSlot: blockSlot,
					operator:  operator,
					solSpent:  4997902349000 - 4997902339000,
				},
				ident: Ident{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				accounts: []int{0, 1, 6, 7, 9, 2, 3, 4, 5},
				accountKeys: []string{
					operator,
					"5Vnsi99cwbuTiPWhpV4qyXE3M6Ltn6MAFY1H8U3iYCgR",
					"AXCVmCoRWNfTvSdm1G9S11SZJh7v7VNvSUVWxofW4Shw",
					"ExMThAwH8ZTrwvYxSbyQJEftGD64w83gCfuDyTqXD1vA",
					"Ffdwpann7sFWFAL9GJi1y5HpP8KXEwQJf8M9zSgwwB2f",
					"FzRaNhxAh268jkf2oqa5Cz2KguTQG3DdQdRCac1saQkY",
					"JDcZX2w1iFvQo2LYHBqMfd3APQAHxsA2JtbnXy3y9WZS",
					"11111111111111111111111111111111",
					"ComputeBudget111111111111111111111111111111",
					"53DfF883gyixYNXnM7s5xhdeyV8mVk9T4i2hGV9vG9io",
				},
			}

			solState := initUnifiedTestSolState(operator, solTx)
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
			assert.Equal(t, NeonIndexedTxInfoStatusDone, tx.status)
			assert.False(t, tx.canceled)

			neonTx := tx.neonReceipt.neonTx
			assert.Empty(t, neonTx.err)
			assert.Equal(t, expectedNeonTx.sig, neonTx.sig)
			assert.Equal(t, *expectedNeonTx.addr, *neonTx.addr)
			assert.Equal(t, expectedNeonTx.nonce, neonTx.nonce)
			assert.Equal(t, expectedNeonTx.value, neonTx.value)
			assert.Equal(t, expectedNeonTx.callData, neonTx.callData)
			assert.Equal(t, expectedNeonTx.gasLimit, neonTx.gasLimit)
			assert.Equal(t, expectedNeonTx.gasPrice, neonTx.gasPrice)
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
			const (
				solSig         = "4f5BGKsFXKyy7wpqSFiFyCCbiVckb9GPVAesZPPRWCLSWBVTreaaLYwXY8vXJMaHuPfeHbaCTyBe5T3kdWmcY7Qr"
				operator       = "7C6iuRYzEJEwe878X2TeMDoCHPEw85ZhaxapNEBuqwL9"
				expectedHolder = "EXfDGv289atyLdCZkkdpN7Jw8fi9qfoW2TjAndVbnAUd"
				blockSlot      = 1713
			)

			neonTxSig := utils.Base64stringToHex("xi4Ek5sG8ACFZbjT+KZg4gQKQV1IYrsIu6w5uRsR938=")
			ixData := base58.Decode("886TCTvoks5mpH9CmZ4YepiKKSaCCiM9ze4YSywkeydUPQXeAcA2A8JSW9ZFJvPuE8ik1ENbQvvq9eHVMeANwYwnqfsgbS2jHnsaBnZkZXGrDweCGFBM4K1AwAAzPfRgyw18pTQH33icBTYcLSXFV2E24ohJKcGPLSCnZJX262frw4XQsnXj43HX9mSkYKbEcYsfCv1MXcwtdj87VihH7UW8Sgf1eobbDAwJJQGsSaXrSmQspUzJVcwaCSq6NcVL6hECQKeFS3MFnPLAycByxF7G8f4FnvrKhACsdREdqqMcpzM94Sm6WdxyeWPx3dxtwxhoQsJr5ra2A5QwMT2fC8oPioUmmK4dRabVt1Edb4uSbJNA44BdkxTyfz7hgsmZwS1Mkk7mxGP991hqk76gvRZLeVUfaFvXZBEksekHoCzMJ2iYnGWQespYomkesZmNHKSoGbkP7ujJRznjp7CJbPscHPynu5trubqQtf7h6ZTGckXixvWtskCiKkvSESsTrxW7U3mZeVKsKmsCh8k2dsiUxvttsZwYABKhvMor8sZfWJV4y1uVc353DZLA7eMCsmNxbskeEjYH7b4trcfHkrp7T88EKqpdGSFGpJLxJV6uc43YdsU5nGjkdMPq7o8aqCQs9ZoG3t8sQjLGgCtqpDxvAjna9MWU5aw3iSzGirr9meoMQr4MeSEeiU6YuM")
			expectedDecodedTxKey := strings.ToLower(neonTxSig[2:])

			solTx := &SolNeonIxReceiptInfo{
				metaInfo: SolIxMetaInfo{
					status:    1, // success
					neonTxSig: neonTxSig,
				},
				solSign:   solSig,
				blockSlot: blockSlot,
				programIx: 0x2a,
				ixData:    ixData,
				solTxCost: SolTxCostInfo{
					solSign:   solSig,
					blockSlot: 0,
					operator:  operator,
					solSpent:  4997984299280 - 4997984294280,
				},
				ident: Ident{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				accounts: []int{1, 0},
				accountKeys: []string{
					operator,
					"EXfDGv289atyLdCZkkdpN7Jw8fi9qfoW2TjAndVbnAUd",
					"53DfF883gyixYNXnM7s5xhdeyV8mVk9T4i2hGV9vG9io",
				},
			}

			solState := initUnifiedTestSolState(operator, solTx)
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
			assert.Equal(t, expectedHolder, tx.storageAccount)
			assert.Equal(t, NeonIndexedTxInfoStatusDone, tx.status)
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

	t.Run("TxStepFromDataIxDecoder decoding iter tx failed: signatures are not same", func(t *testing.T) {
		ixData := base58.Decode("CUrU71iEchwyUYbwpusc7ciwTfJs1mnxngF5eQA5zrkL5CJDxtedKXbUhb4FuYXLk9RVxXAjrNbaZuytdf4nqNPW2oBmFvZxUZi8QcGMjSEZ5fxwCWBbXQ5E6kyLk3MdEAHvFqChx4e7qbYdXsCmNoeWg78Ddqne7VP5R17quCcex2wnj5eJ51oWkfpTkiWr1F5kgTBRH3mmetuW8CX529VbtK6gKvfZqJg9KV2aZr6SyKmAGtVrmVdRnDwv2bvaXg")

		decoder := InitTxStepFromDataIxDecoder(&IxDecoder{
			log:    log,
			name:   "TransactionStepFromInstruction",
			ixCode: 0x20,
			state: initUnifiedTestSolState("", &SolNeonIxReceiptInfo{
				metaInfo: SolIxMetaInfo{
					neonTxSig: "some_non_valid_signature",
				},
				ixData: ixData,
			}),
		})

		result := decoder.Execute()
		assert.False(t, result)
	})

	t.Run("TxStepFromDataIxDecoder decoding iter tx without receipt and log events: success",
		func(t *testing.T) {
			const (
				solSig         = "5P2DCMNFpXwoY1qFQkRK6xAXwPgzyG88oXP7xFPHL5aGGCC85eLJfWjDvcqz8RktPyaivSsJAMmdVkwmWsTtDjyZ"
				operator       = "7C6iuRYzEJEwe878X2TeMDoCHPEw85ZhaxapNEBuqwL9"
				expectedHolder = "GUoWJTagZaFV23H7cT3tkkmNYk7Hy12Swsz29ZxUrssW"
				blockSlot      = 1164
			)

			neonTxSig := utils.Base64stringToHex("/0uSkeSM8X6KskVMeT6KtZe2ZAZNZmQJ2yIQ8YXtsUQ=")
			gasUsed := int(binary.LittleEndian.Uint32(utils.Base64stringToBytes("mDoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")))
			totalGasUsed := int(binary.LittleEndian.Uint32(utils.Base64stringToBytes("mDoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")))
			ixData := base58.Decode("CUrU71iEchwyUYbwpusc7ciwTfJs1mnxngF5eQA5zrkL5CJDxtedKXbUhb4FuYXLk9RVxXAjrNbaZuytdf4nqNPW2oBmFvZxUZi8QcGMjSEZ5fxwCWBbXQ5E6kyLk3MdEAHvFqChx4e7qbYdXsCmNoeWg78Ddqne7VP5R17quCcex2wnj5eJ51oWkfpTkiWr1F5kgTBRH3mmetuW8CX529VbtK6gKvfZqJg9KV2aZr6SyKmAGtVrmVdRnDwv2bvaXg")
			expectedDecodedTxKey := strings.ToLower(neonTxSig[2:])

			solTx := &SolNeonIxReceiptInfo{
				metaInfo: SolIxMetaInfo{
					status:           1, // success
					neonTxSig:        neonTxSig,
					neonGasUsed:      int64(gasUsed),
					neonTotalGasUsed: int64(totalGasUsed),
				},
				solSign:   solSig,
				blockSlot: blockSlot,
				programIx: 0x20,
				ixData:    ixData,
				solTxCost: SolTxCostInfo{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				ident: Ident{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				accountKeys: []string{
					"7r5GAh4SDhBwxg98vT86Q8sA8c9zEgJduSWWCV1y48V",
					"HAaFkyDXGrrnRBzrheZebTQw5m1U9SKe2P6E4puWY3t",
					"2Lf3ek6ayv7FWKe8NYydxyv9165jBU9m8gTfgn2zsbUZ",
					"2Zv3J7yBRzdbX2XyEmDxKvHtNPWNU8jgRXK6neASvSSK",
					"3nhuEyPNxF6QJVvaZWFcKYVL6ruLUrs7cA665xvn1jWN",
					"72sKYGHs8FaztMfbqqJRS9H8WAKHsjyxMe4SXFjHFhk5",
					"ADAfCJq8ytXHHpWqX9Nmr3K2oSs6b2NNBz6RJQpD5RBH",
					"DqPxusfWAP52z4urZYpNcMZQ6A9FdoRkUayk9e5myZUm",
					"FqnfGMByCC7nWVxvWDtXqDUNj3pV7UEEXT4WtxBeRayX",
					expectedHolder,
					"11111111111111111111111111111111",
					"ComputeBudget111111111111111111111111111111",
					"53DfF883gyixYNXnM7s5xhdeyV8mVk9T4i2hGV9vG9io",
				},
				accounts: []int{
					9,
					0,
					4,
					3,
					10,
					12,
					7,
					5,
					8,
					6,
					2,
					1,
				},
			}

			neonTxAddr := "0xaa4d6f4ff831181a2bbfd4d62260dabdea964ff1"
			expectedNeonTx := NeonTxInfo{
				addr:     &neonTxAddr,
				sig:      neonTxSig,
				nonce:    "bf",
				gasPrice: "14b1b0cf32",
				gasLimit: "3b9aca00",
				toAddr:   "0xca12f8c0ca275bd38937c1bf354da40cc116bb68",
				value:    "",
				callData: "0xc9c6539600000000000000000000000058b2145cfa2406097be00c0057d24a3f3f90361100000000000000000000000010acfd050938dfdaf3d3d9831c05fc6ed9e4194b",
				v:        "0x102",
				r:        "0x9fdc62d9d340b345cf08ffe373d7302d64800bfdedac2d7567ee90ed0e5648d0",
				s:        "0x2e36e6bac35772258a4797afa267bfbd6ad1a6f29fd6a282ef311c097ae414df",
			}

			solState := initUnifiedTestSolState(operator, solTx)
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
			assert.Equal(t, expectedHolder, tx.storageAccount)
			assert.Equal(t, NeonIndexedTxInfoStatusInProgress, tx.status)
			assert.False(t, tx.canceled)

			neonTx := tx.neonReceipt.neonTx
			assert.Empty(t, neonTx.err)
			assert.Equal(t, expectedNeonTx.sig, neonTx.sig)
			assert.Equal(t, *expectedNeonTx.addr, *neonTx.addr)
			assert.Equal(t, expectedNeonTx.nonce, neonTx.nonce)
			assert.Equal(t, expectedNeonTx.value, neonTx.value)
			assert.Equal(t, expectedNeonTx.callData, neonTx.callData)
			assert.Equal(t, expectedNeonTx.gasLimit, neonTx.gasLimit)
			assert.Equal(t, expectedNeonTx.gasPrice, neonTx.gasPrice)
		})

	t.Run("TxStepFromDataIxDecoder decoding iter tx with log events and return event: success",
		func(t *testing.T) {
			const (
				solSig         = "5P2DCMNFpXwoY1qFQkRK6xAXwPgzyG88oXP7xFPHL5aGGCC85eLJfWjDvcqz8RktPyaivSsJAMmdVkwmWsTtDjyZ"
				operator       = "7C6iuRYzEJEwe878X2TeMDoCHPEw85ZhaxapNEBuqwL9"
				expectedHolder = "GUoWJTagZaFV23H7cT3tkkmNYk7Hy12Swsz29ZxUrssW"
				blockSlot      = 1164
			)

			neonTxSig := utils.Base64stringToHex("/0uSkeSM8X6KskVMeT6KtZe2ZAZNZmQJ2yIQ8YXtsUQ=")
			gasUsed := int(binary.LittleEndian.Uint32(utils.Base64stringToBytes("mDoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")))
			totalGasUsed := int(binary.LittleEndian.Uint32(utils.Base64stringToBytes("mDoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")))
			ixData := base58.Decode("CUrU71iEchwyUYbwpusc7ciwTfJs1mnxngF5eQA5zrkL5CJDxtedKXbUhb4FuYXLk9RVxXAjrNbaZuytdf4nqNPW2oBmFvZxUZi8QcGMjSEZ5fxwCWBbXQ5E6kyLk3MdEAHvFqChx4e7qbYdXsCmNoeWg78Ddqne7VP5R17quCcex2wnj5eJ51oWkfpTkiWr1F5kgTBRH3mmetuW8CX529VbtK6gKvfZqJg9KV2aZr6SyKmAGtVrmVdRnDwv2bvaXg")
			expectedDecodedTxKey := strings.ToLower(neonTxSig[2:])

			returnEventTotalGasUsed, err := strconv.ParseInt(fmt.Sprintf("%x", totalGasUsed), 16, 64)
			assert.Empty(t, err)
			expectedReturnEvent := NeonLogTxEvent{
				eventType:    Return,
				hidden:       true,
				address:      "",
				topics:       nil,
				data:         convertHexStringToLittleEndianByte("0x1"),
				solSig:       solSig,
				totalGasUsed: returnEventTotalGasUsed + 5000,
				reverted:     false,
			}

			solTx := &SolNeonIxReceiptInfo{
				metaInfo: SolIxMetaInfo{
					status:           1, // success
					err:              nil,
					neonTxSig:        neonTxSig,
					neonGasUsed:      int64(gasUsed),
					neonTotalGasUsed: int64(totalGasUsed),
					neonTxReturn: &NeonLogTxReturn{
						GasUsed:  int64(gasUsed),
						Status:   1,
						Canceled: false,
					},
					neonTxEvents: []NeonLogTxEvent{
						{
							eventType:    EnterCall,
							hidden:       true,
							address:      "0x5e152aa201dd2d48739c84ee77ad71dc43d1d747",
							topics:       []string{},
							data:         []byte{},
							solSig:       solSig,
							totalGasUsed: int64(totalGasUsed),
							reverted:     false,
							eventLevel:   2,
							eventOrder:   2,
						},
						{
							eventType:    ExitReturn,
							hidden:       true,
							address:      "",
							topics:       []string{},
							data:         []byte{},
							solSig:       solSig,
							totalGasUsed: int64(totalGasUsed),
							reverted:     false,
							eventLevel:   2,
							eventOrder:   3,
						},
						{
							eventType:    EnterCall,
							hidden:       true,
							address:      "0x5e152aa201dd2d48739c84ee77ad71dc43d1d747",
							topics:       []string{},
							data:         []byte{},
							solSig:       solSig,
							totalGasUsed: int64(totalGasUsed),
							reverted:     false,
							eventLevel:   3,
							eventOrder:   4,
						},
					},
					isLogTruncated:     false,
					isAlreadyFinalized: false,
				},
				solSign:   solSig,
				blockSlot: blockSlot,
				programIx: 0x20,
				ixData:    ixData,
				solTxCost: SolTxCostInfo{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				ident: Ident{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				accounts: []int{
					9,
					0,
					4,
					3,
					10,
					12,
					7,
					5,
					8,
					6,
					2,
					1,
				},
				accountKeys: []string{
					"7r5GAh4SDhBwxg98vT86Q8sA8c9zEgJduSWWCV1y48V",
					"HAaFkyDXGrrnRBzrheZebTQw5m1U9SKe2P6E4puWY3t",
					"2Lf3ek6ayv7FWKe8NYydxyv9165jBU9m8gTfgn2zsbUZ",
					"2Zv3J7yBRzdbX2XyEmDxKvHtNPWNU8jgRXK6neASvSSK",
					"3nhuEyPNxF6QJVvaZWFcKYVL6ruLUrs7cA665xvn1jWN",
					"72sKYGHs8FaztMfbqqJRS9H8WAKHsjyxMe4SXFjHFhk5",
					"ADAfCJq8ytXHHpWqX9Nmr3K2oSs6b2NNBz6RJQpD5RBH",
					"DqPxusfWAP52z4urZYpNcMZQ6A9FdoRkUayk9e5myZUm",
					"FqnfGMByCC7nWVxvWDtXqDUNj3pV7UEEXT4WtxBeRayX",
					expectedHolder,
					"11111111111111111111111111111111",
					"ComputeBudget111111111111111111111111111111",
					"53DfF883gyixYNXnM7s5xhdeyV8mVk9T4i2hGV9vG9io",
				},
			}

			neonTxAddr := "0xaa4d6f4ff831181a2bbfd4d62260dabdea964ff1"
			expectedNeonTx := NeonTxInfo{
				addr:     &neonTxAddr,
				sig:      neonTxSig,
				nonce:    "bf",
				gasPrice: "14b1b0cf32",
				gasLimit: "3b9aca00",
				toAddr:   "0xca12f8c0ca275bd38937c1bf354da40cc116bb68",
				value:    "",
				callData: "0xc9c6539600000000000000000000000058b2145cfa2406097be00c0057d24a3f3f90361100000000000000000000000010acfd050938dfdaf3d3d9831c05fc6ed9e4194b",
				v:        "0x102",
				r:        "0x9fdc62d9d340b345cf08ffe373d7302d64800bfdedac2d7567ee90ed0e5648d0",
				s:        "0x2e36e6bac35772258a4797afa267bfbd6ad1a6f29fd6a282ef311c097ae414df",
			}

			solState := initUnifiedTestSolState(operator, solTx)
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
			assert.Equal(t, expectedHolder, tx.storageAccount)
			assert.Equal(t, NeonIndexedTxInfoStatusDone, tx.status)
			assert.False(t, tx.canceled)

			neonTx := tx.neonReceipt.neonTx
			assert.Empty(t, neonTx.err)
			assert.Equal(t, expectedNeonTx.sig, neonTx.sig)
			assert.Equal(t, *expectedNeonTx.addr, *neonTx.addr)
			assert.Equal(t, expectedNeonTx.nonce, neonTx.nonce)
			assert.Equal(t, expectedNeonTx.value, neonTx.value)
			assert.Equal(t, expectedNeonTx.callData, neonTx.callData)
			assert.Equal(t, expectedNeonTx.gasLimit, neonTx.gasLimit)
			assert.Equal(t, expectedNeonTx.gasPrice, neonTx.gasPrice)

			expectedEvents := solTx.metaInfo.neonTxEvents
			actualEvents := tx.neonEvents
			assert.Equal(t, 4, len(actualEvents))

			for i := range expectedEvents {
				assert.Equal(t, expectedEvents[i].totalGasUsed, actualEvents[i].totalGasUsed)
				assert.Equal(t, expectedEvents[i].topics, actualEvents[i].topics)
				assert.Equal(t, expectedEvents[i].solSig, actualEvents[i].solSig)
				assert.Equal(t, expectedEvents[i].eventType, actualEvents[i].eventType)
				assert.Equal(t, expectedEvents[i].eventOrder, actualEvents[i].eventOrder)
				assert.Equal(t, expectedEvents[i].eventLevel, actualEvents[i].eventLevel)
				assert.Equal(t, expectedEvents[i].hidden, actualEvents[i].hidden)
				assert.Equal(t, expectedEvents[i].reverted, actualEvents[i].reverted)
				assert.Equal(t, expectedEvents[i].address, actualEvents[i].address)
			}

			actualReturnEvent := actualEvents[3]
			assert.Equal(t, expectedReturnEvent.totalGasUsed, actualReturnEvent.totalGasUsed)
			assert.Equal(t, expectedReturnEvent.topics, actualReturnEvent.topics)
			assert.Equal(t, expectedReturnEvent.solSig, actualReturnEvent.solSig)
			assert.Equal(t, expectedReturnEvent.eventType, actualReturnEvent.eventType)
			assert.Equal(t, expectedReturnEvent.eventOrder, actualReturnEvent.eventOrder)
			assert.Equal(t, expectedReturnEvent.eventLevel, actualReturnEvent.eventLevel)
			assert.Equal(t, expectedReturnEvent.hidden, actualReturnEvent.hidden)
			assert.Equal(t, expectedReturnEvent.reverted, actualReturnEvent.reverted)
			assert.Equal(t, expectedReturnEvent.address, actualReturnEvent.address)
		},
	)
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
			const (
				solSig         = "5P2DCMNFpXwoY1qFQkRK6xAXwPgzyG88oXP7xFPHL5aGGCC85eLJfWjDvcqz8RktPyaivSsJAMmdVkwmWsTtDjyZ"
				operator       = "7C6iuRYzEJEwe878X2TeMDoCHPEw85ZhaxapNEBuqwL9"
				expectedHolder = "GUoWJTagZaFV23H7cT3tkkmNYk7Hy12Swsz29ZxUrssW"
				blockSlot      = 1164
			)

			neonTxSig := utils.Base64stringToHex("/0uSkeSM8X6KskVMeT6KtZe2ZAZNZmQJ2yIQ8YXtsUQ=")
			gasUsed := int(binary.LittleEndian.Uint32(utils.Base64stringToBytes("mDoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")))
			totalGasUsed := int(binary.LittleEndian.Uint32(utils.Base64stringToBytes("mDoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")))
			ixData := base58.Decode("CUrU71iEchwyUYbwpusc7ciwTfJs1mnxngF5eQA5zrkL5CJDxtedKXbUhb4FuYXLk9RVxXAjrNbaZuytdf4nqNPW2oBmFvZxUZi8QcGMjSEZ5fxwCWBbXQ5E6kyLk3MdEAHvFqChx4e7qbYdXsCmNoeWg78Ddqne7VP5R17quCcex2wnj5eJ51oWkfpTkiWr1F5kgTBRH3mmetuW8CX529VbtK6gKvfZqJg9KV2aZr6SyKmAGtVrmVdRnDwv2bvaXg")
			expectedDecodedTxKey := strings.ToLower(neonTxSig[2:])

			solTx := &SolNeonIxReceiptInfo{
				metaInfo: SolIxMetaInfo{
					status:           1, // success
					neonTxSig:        neonTxSig,
					neonGasUsed:      int64(gasUsed),
					neonTotalGasUsed: int64(totalGasUsed),
				},
				solSign:   solSig,
				blockSlot: blockSlot,
				programIx: 0x20,
				ixData:    ixData,
				solTxCost: SolTxCostInfo{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				ident: Ident{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				accountKeys: []string{
					"7r5GAh4SDhBwxg98vT86Q8sA8c9zEgJduSWWCV1y48V",
					"HAaFkyDXGrrnRBzrheZebTQw5m1U9SKe2P6E4puWY3t",
					"2Lf3ek6ayv7FWKe8NYydxyv9165jBU9m8gTfgn2zsbUZ",
					"2Zv3J7yBRzdbX2XyEmDxKvHtNPWNU8jgRXK6neASvSSK",
					"3nhuEyPNxF6QJVvaZWFcKYVL6ruLUrs7cA665xvn1jWN",
					"72sKYGHs8FaztMfbqqJRS9H8WAKHsjyxMe4SXFjHFhk5",
					"ADAfCJq8ytXHHpWqX9Nmr3K2oSs6b2NNBz6RJQpD5RBH",
					"DqPxusfWAP52z4urZYpNcMZQ6A9FdoRkUayk9e5myZUm",
					"FqnfGMByCC7nWVxvWDtXqDUNj3pV7UEEXT4WtxBeRayX",
					expectedHolder,
					"11111111111111111111111111111111",
					"ComputeBudget111111111111111111111111111111",
					"53DfF883gyixYNXnM7s5xhdeyV8mVk9T4i2hGV9vG9io",
				},
				accounts: []int{
					9,
					0,
					4,
					3,
					10,
					12,
					7,
					5,
					8,
					6,
					2,
					1,
				},
			}

			expectedNeonTx := NeonTxInfo{
				sig: neonTxSig,
			}

			solState := initUnifiedTestSolState(operator, solTx)

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
			assert.Equal(t, expectedHolder, tx.storageAccount)
			assert.Equal(t, NeonIndexedTxInfoStatusInProgress, tx.status)
			assert.False(t, tx.canceled)

			neonTx := tx.neonReceipt.neonTx
			assert.Empty(t, neonTx.err)
			assert.Equal(t, expectedNeonTx.sig, neonTx.sig)
			assert.Equal(t, expectedNeonTx.addr, neonTx.addr)
			assert.Equal(t, expectedNeonTx.nonce, neonTx.nonce)
			assert.Equal(t, expectedNeonTx.value, neonTx.value)
			assert.Equal(t, expectedNeonTx.callData, neonTx.callData)
			assert.Equal(t, expectedNeonTx.gasLimit, neonTx.gasLimit)
			assert.Equal(t, expectedNeonTx.gasPrice, neonTx.gasPrice)
		})

	t.Run("TxStepFromAccountIxDecoder decoding iter tx with return neon event: success",
		func(t *testing.T) {
			const (
				solSig         = "5P2DCMNFpXwoY1qFQkRK6xAXwPgzyG88oXP7xFPHL5aGGCC85eLJfWjDvcqz8RktPyaivSsJAMmdVkwmWsTtDjyZ"
				operator       = "7C6iuRYzEJEwe878X2TeMDoCHPEw85ZhaxapNEBuqwL9"
				expectedHolder = "GUoWJTagZaFV23H7cT3tkkmNYk7Hy12Swsz29ZxUrssW"
				blockSlot      = 1164
			)

			neonTxSig := utils.Base64stringToHex("/0uSkeSM8X6KskVMeT6KtZe2ZAZNZmQJ2yIQ8YXtsUQ=")
			gasUsed := int(binary.LittleEndian.Uint32(utils.Base64stringToBytes("mDoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")))
			totalGasUsed := int(binary.LittleEndian.Uint32(utils.Base64stringToBytes("mDoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")))
			ixData := base58.Decode("CUrU71iEchwyUYbwpusc7ciwTfJs1mnxngF5eQA5zrkL5CJDxtedKXbUhb4FuYXLk9RVxXAjrNbaZuytdf4nqNPW2oBmFvZxUZi8QcGMjSEZ5fxwCWBbXQ5E6kyLk3MdEAHvFqChx4e7qbYdXsCmNoeWg78Ddqne7VP5R17quCcex2wnj5eJ51oWkfpTkiWr1F5kgTBRH3mmetuW8CX529VbtK6gKvfZqJg9KV2aZr6SyKmAGtVrmVdRnDwv2bvaXg")
			expectedDecodedTxKey := strings.ToLower(neonTxSig[2:])

			returnEventTotalGasUsed, err := strconv.ParseInt(fmt.Sprintf("%x", totalGasUsed), 16, 64)
			assert.Empty(t, err)
			expectedReturnEvent := NeonLogTxEvent{
				eventType:    Return,
				hidden:       true,
				address:      "",
				topics:       nil,
				data:         convertHexStringToLittleEndianByte("0x1"),
				solSig:       solSig,
				totalGasUsed: returnEventTotalGasUsed + 5000,
				reverted:     false,
			}

			solTx := &SolNeonIxReceiptInfo{
				metaInfo: SolIxMetaInfo{
					status:           1, // success
					neonTxSig:        neonTxSig,
					neonGasUsed:      int64(gasUsed),
					neonTotalGasUsed: int64(totalGasUsed),
					neonTxReturn: &NeonLogTxReturn{
						GasUsed:  int64(totalGasUsed),
						Status:   1,
						Canceled: false,
					},
					neonTxEvents:       nil,
					isLogTruncated:     false,
					isAlreadyFinalized: false,
				},
				solSign:   solSig,
				blockSlot: blockSlot,
				programIx: 0x20,
				ixData:    ixData,
				solTxCost: SolTxCostInfo{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				ident: Ident{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				accountKeys: []string{
					"7r5GAh4SDhBwxg98vT86Q8sA8c9zEgJduSWWCV1y48V",
					"HAaFkyDXGrrnRBzrheZebTQw5m1U9SKe2P6E4puWY3t",
					"2Lf3ek6ayv7FWKe8NYydxyv9165jBU9m8gTfgn2zsbUZ",
					"2Zv3J7yBRzdbX2XyEmDxKvHtNPWNU8jgRXK6neASvSSK",
					"3nhuEyPNxF6QJVvaZWFcKYVL6ruLUrs7cA665xvn1jWN",
					"72sKYGHs8FaztMfbqqJRS9H8WAKHsjyxMe4SXFjHFhk5",
					"ADAfCJq8ytXHHpWqX9Nmr3K2oSs6b2NNBz6RJQpD5RBH",
					"DqPxusfWAP52z4urZYpNcMZQ6A9FdoRkUayk9e5myZUm",
					"FqnfGMByCC7nWVxvWDtXqDUNj3pV7UEEXT4WtxBeRayX",
					expectedHolder,
					"11111111111111111111111111111111",
					"ComputeBudget111111111111111111111111111111",
					"53DfF883gyixYNXnM7s5xhdeyV8mVk9T4i2hGV9vG9io",
				},
				accounts: []int{
					9,
					0,
					4,
					3,
					10,
					12,
					7,
					5,
					8,
					6,
					2,
					1,
				},
			}

			expectedNeonTx := NeonTxInfo{
				sig: neonTxSig,
			}

			solState := initUnifiedTestSolState(operator, solTx)

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
			assert.Equal(t, expectedHolder, tx.storageAccount)
			assert.Equal(t, NeonIndexedTxInfoStatusDone, tx.status)
			assert.False(t, tx.canceled)

			neonTx := tx.neonReceipt.neonTx
			assert.Empty(t, neonTx.err)
			assert.Equal(t, expectedNeonTx.sig, neonTx.sig)
			assert.Equal(t, expectedNeonTx.addr, neonTx.addr)
			assert.Equal(t, expectedNeonTx.nonce, neonTx.nonce)
			assert.Equal(t, expectedNeonTx.value, neonTx.value)
			assert.Equal(t, expectedNeonTx.callData, neonTx.callData)
			assert.Equal(t, expectedNeonTx.gasLimit, neonTx.gasLimit)
			assert.Equal(t, expectedNeonTx.gasPrice, neonTx.gasPrice)

			actualReturnEvent := tx.neonEvents[0]
			assert.Equal(t, expectedReturnEvent.totalGasUsed, actualReturnEvent.totalGasUsed)
			assert.Equal(t, expectedReturnEvent.topics, actualReturnEvent.topics)
			assert.Equal(t, expectedReturnEvent.solSig, actualReturnEvent.solSig)
			assert.Equal(t, expectedReturnEvent.eventType, actualReturnEvent.eventType)
			assert.Equal(t, expectedReturnEvent.eventOrder, actualReturnEvent.eventOrder)
			assert.Equal(t, expectedReturnEvent.eventLevel, actualReturnEvent.eventLevel)
			assert.Equal(t, expectedReturnEvent.hidden, actualReturnEvent.hidden)
			assert.Equal(t, expectedReturnEvent.reverted, actualReturnEvent.reverted)
			assert.Equal(t, expectedReturnEvent.address, actualReturnEvent.address)
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
			const (
				solSig         = "3ZVbKg66KnKQL51m5vhzBiaQM3rfrACgnYMeLrqzuazWPsKFf7Wuw81swdeM8CFZpytThWzG9xpZjZqnUDoEFPSC"
				operator       = "896VtgKpVopcBY9KFxvmXCzriV7gmfyynuXsbTzPhf4t"
				expectedHolder = "896VtgKpVopcBY9KFxvmXCzriV7gmfyynuXsbTzPhf4t"
				blockSlot      = 11074
			)

			neonTxSig := utils.Base64stringToHex("/vstpTXpJ7hf5o64HLLkpYJ8kF94OBoB7yMiqpsK7o4=")
			ixData := base58.Decode("gG26gyJTR8s1JuUH8R9yPSaFsSx8itzokDUKnvBTekV42pco95FovLZ4SEXewtCKNgZAb9qt1qX1wymD6Wg1ZVmtRw6FWvWmdhhXV3tdBTB67EFKaDT7CJyR5NoAD3dFeE8qzFhR57HuKXw5vcjqQSxKDp2i1PY4pbcMw8eHCwXDRTCheZGiT7HD3ASN83p1PSeYnSJZzgD2g1MLBSMCsvek7zAjrkaptfvZvJbektV2ajeSrqekrouj5w4Uwtm2Ab465ZgBfFnXm7NK5xcFdjc3DR7HTgoxrFDxzoprHjc8o5GBC5sanuWJth9dQQ3QNA8vT2Q3M3gGCEbp8zSqFS5fJzLoEZRaB67nvX3WH3DMuSQzkkLjWy4zSjcMSkrtkmSb7gGzpg26zwNgns1nAaWh66te1vzhbQzUtEWzzDvgSyXuFbXNZK8bc57oNAkvVPAW9wd9TBCmkzkBNuNLZTgc9VMXE2Vf3aFB7SmSJEtJxANFtDkSmdLVvTxsnPDm1kJfp1XBQ4yik5PpL1LfK1r5afjWfx9USHNBeJ3ogiLgv1HDymkMmTvExUfUFA14ZQNHzqfGaUjL3SXyAVGJRVesNGHzN3GQFvdHtY4TgRTCFwCV974L7BzK6WPtAoP3mCnQnqio4JMvavR3aTGuhjatDkMmJnPFUVp9d78z7shJkqHhMyp1mvBbRYJpkg7ziLw8Mzhyf8WgRCyv2EZp9XK8Vt634C5Xxn9LVHJd4mrK6CMBc3UFSF1oigCwr28EWwZYZqXGETMkb3zNgU6afjU9EmbH2uU2TtjxNuwJ8Hz5RNLRRAuC7kLAEt2srMG12Lvhn281qGTVeUu47Z6LQdHvtmjNMeR5QWkxD73t4qmizw91qtA1zSinUq8DBwfJ9Heu7aBBywDBBY1UP9Ff3fQuUJzS87qiMb6zFTshDGv2v5qHVs1mL9tuExtADfBmA5ebHZM76QL1BHgkDRt7eworNfyqSaMAo5dANgrKMogbPhu9gSuANjFpvs6R12dqUJH5F5QCoHpNd6tueKKqLEanvTYxpMhrUeY2cVVLGnYqJbNAcyJWStvpAyHbv8XUNQy96uMp9EAPHppqqTiRrmwU6pMqkc8TosrnfoZ8aro7YLih58kKQsBSWTSeAvCQ7EceJsX6Hv5WJPHjuJmmq9NyAb1Qbap3pRUgzeuqpQTwcGCsYwZ9FZs53647WxC5xnRjeENpzGAyWx8Jru1cp6zqpA6SdrFXeuNGwz9Kv8Q3U1JzqUwY19YFvu7DThfcCYXxAQGUdUBSk5k2czhtMtefPgamsyorDgExQUvCL9Apw72ThCKW4ZnXxKXhUxWqyjJc6aMxb")
			expectedDecodedTxKey := strings.ToLower(neonTxSig[2:])

			solTx := &SolNeonIxReceiptInfo{
				metaInfo: SolIxMetaInfo{
					status:    1, // success
					neonTxSig: neonTxSig,
				},
				solSign:   solSig,
				blockSlot: blockSlot,
				programIx: 0x22,
				ixData:    ixData,
				solTxCost: SolTxCostInfo{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				ident: Ident{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				accountKeys: []string{
					"3NqgsSRfjpmDfzRH4PLKrzBvMc8MgFXgU58Yy8n41KF5",
					"3MTrvrjHerq8JAnNvr3QvfMK4VFs1493JAwQj4FMf8bt",
					expectedHolder,
					"F9bMPQ3xBnfkRrhFPDSpjnhXb5yxkzfLScYirtR87tUD",
					"GS6DZNvbv7dmjZgquokzyX3sFccgjjd5MXeK46PH1F7N",
					"GvRqcwRyKgWfGA4AyrNZaRyDe8HqiUDT5sRhPSS6vpt2",
					"11111111111111111111111111111111",
					"ComputeBudget111111111111111111111111111111",
					"53DfF883gyixYNXnM7s5xhdeyV8mVk9T4i2hGV9vG9io",
				},
				accounts: []int{
					2,
					0,
					3,
					4,
					6,
					8,
					1,
					5},
			}

			expectedNeonTx := NeonTxInfo{
				sig: neonTxSig,
			}

			solState := initUnifiedTestSolState(operator, solTx)

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
			assert.Equal(t, expectedHolder, tx.storageAccount)
			assert.Equal(t, NeonIndexedTxInfoStatusInProgress, tx.status)

			neonTx := tx.neonReceipt.neonTx
			assert.Empty(t, neonTx.err)
			assert.Equal(t, expectedNeonTx.sig, neonTx.sig)
			assert.Equal(t, expectedNeonTx.addr, neonTx.addr)
			assert.Equal(t, expectedNeonTx.nonce, neonTx.nonce)
			assert.Equal(t, expectedNeonTx.value, neonTx.value)
			assert.Equal(t, expectedNeonTx.callData, neonTx.callData)
			assert.Equal(t, expectedNeonTx.gasLimit, neonTx.gasLimit)
			assert.Equal(t, expectedNeonTx.gasPrice, neonTx.gasPrice)
		})

	t.Run("TxStepFromAccountNoChainIdIxDecoder decoding iter tx wit return neon event: success",
		func(t *testing.T) {
			const (
				solSig         = "3ZVbKg66KnKQL51m5vhzBiaQM3rfrACgnYMeLrqzuazWPsKFf7Wuw81swdeM8CFZpytThWzG9xpZjZqnUDoEFPSC"
				operator       = "896VtgKpVopcBY9KFxvmXCzriV7gmfyynuXsbTzPhf4t"
				expectedHolder = "896VtgKpVopcBY9KFxvmXCzriV7gmfyynuXsbTzPhf4t"
				blockSlot      = 11074
			)

			neonTxSig := utils.Base64stringToHex("/vstpTXpJ7hf5o64HLLkpYJ8kF94OBoB7yMiqpsK7o4=")
			ixData := base58.Decode("gG26gyJTR8s1JuUH8R9yPSaFsSx8itzokDUKnvBTekV42pco95FovLZ4SEXewtCKNgZAb9qt1qX1wymD6Wg1ZVmtRw6FWvWmdhhXV3tdBTB67EFKaDT7CJyR5NoAD3dFeE8qzFhR57HuKXw5vcjqQSxKDp2i1PY4pbcMw8eHCwXDRTCheZGiT7HD3ASN83p1PSeYnSJZzgD2g1MLBSMCsvek7zAjrkaptfvZvJbektV2ajeSrqekrouj5w4Uwtm2Ab465ZgBfFnXm7NK5xcFdjc3DR7HTgoxrFDxzoprHjc8o5GBC5sanuWJth9dQQ3QNA8vT2Q3M3gGCEbp8zSqFS5fJzLoEZRaB67nvX3WH3DMuSQzkkLjWy4zSjcMSkrtkmSb7gGzpg26zwNgns1nAaWh66te1vzhbQzUtEWzzDvgSyXuFbXNZK8bc57oNAkvVPAW9wd9TBCmkzkBNuNLZTgc9VMXE2Vf3aFB7SmSJEtJxANFtDkSmdLVvTxsnPDm1kJfp1XBQ4yik5PpL1LfK1r5afjWfx9USHNBeJ3ogiLgv1HDymkMmTvExUfUFA14ZQNHzqfGaUjL3SXyAVGJRVesNGHzN3GQFvdHtY4TgRTCFwCV974L7BzK6WPtAoP3mCnQnqio4JMvavR3aTGuhjatDkMmJnPFUVp9d78z7shJkqHhMyp1mvBbRYJpkg7ziLw8Mzhyf8WgRCyv2EZp9XK8Vt634C5Xxn9LVHJd4mrK6CMBc3UFSF1oigCwr28EWwZYZqXGETMkb3zNgU6afjU9EmbH2uU2TtjxNuwJ8Hz5RNLRRAuC7kLAEt2srMG12Lvhn281qGTVeUu47Z6LQdHvtmjNMeR5QWkxD73t4qmizw91qtA1zSinUq8DBwfJ9Heu7aBBywDBBY1UP9Ff3fQuUJzS87qiMb6zFTshDGv2v5qHVs1mL9tuExtADfBmA5ebHZM76QL1BHgkDRt7eworNfyqSaMAo5dANgrKMogbPhu9gSuANjFpvs6R12dqUJH5F5QCoHpNd6tueKKqLEanvTYxpMhrUeY2cVVLGnYqJbNAcyJWStvpAyHbv8XUNQy96uMp9EAPHppqqTiRrmwU6pMqkc8TosrnfoZ8aro7YLih58kKQsBSWTSeAvCQ7EceJsX6Hv5WJPHjuJmmq9NyAb1Qbap3pRUgzeuqpQTwcGCsYwZ9FZs53647WxC5xnRjeENpzGAyWx8Jru1cp6zqpA6SdrFXeuNGwz9Kv8Q3U1JzqUwY19YFvu7DThfcCYXxAQGUdUBSk5k2czhtMtefPgamsyorDgExQUvCL9Apw72ThCKW4ZnXxKXhUxWqyjJc6aMxb")
			expectedDecodedTxKey := strings.ToLower(neonTxSig[2:])

			returnEventTotalGasUsed, err := strconv.ParseInt(fmt.Sprintf("%x", 0), 16, 64)
			assert.Empty(t, err)
			expectedReturnEvent := NeonLogTxEvent{
				eventType:    Return,
				hidden:       true,
				address:      "",
				topics:       nil,
				data:         convertHexStringToLittleEndianByte("0x1"),
				solSig:       solSig,
				totalGasUsed: returnEventTotalGasUsed + 5000,
				reverted:     false,
			}

			solTx := &SolNeonIxReceiptInfo{
				metaInfo: SolIxMetaInfo{
					status:    1, // success
					neonTxSig: neonTxSig,
					neonTxReturn: &NeonLogTxReturn{
						GasUsed:  0,
						Status:   1,
						Canceled: false,
					},
				},
				solSign:   solSig,
				blockSlot: blockSlot,
				programIx: 0x22,
				ixData:    ixData,
				solTxCost: SolTxCostInfo{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				ident: Ident{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				accountKeys: []string{
					"3NqgsSRfjpmDfzRH4PLKrzBvMc8MgFXgU58Yy8n41KF5",
					"3MTrvrjHerq8JAnNvr3QvfMK4VFs1493JAwQj4FMf8bt",
					expectedHolder,
					"F9bMPQ3xBnfkRrhFPDSpjnhXb5yxkzfLScYirtR87tUD",
					"GS6DZNvbv7dmjZgquokzyX3sFccgjjd5MXeK46PH1F7N",
					"GvRqcwRyKgWfGA4AyrNZaRyDe8HqiUDT5sRhPSS6vpt2",
					"11111111111111111111111111111111",
					"ComputeBudget111111111111111111111111111111",
					"53DfF883gyixYNXnM7s5xhdeyV8mVk9T4i2hGV9vG9io",
				},
				accounts: []int{
					2,
					0,
					3,
					4,
					6,
					8,
					1,
					5},
			}

			expectedNeonTx := NeonTxInfo{
				sig: neonTxSig,
			}

			solState := initUnifiedTestSolState(operator, solTx)

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
			assert.Equal(t, expectedHolder, tx.storageAccount)
			assert.Equal(t, NeonIndexedTxInfoStatusDone, tx.status)
			assert.Equal(t, 1, len(tx.neonEvents))

			neonTx := tx.neonReceipt.neonTx
			assert.Empty(t, neonTx.err)
			assert.Equal(t, expectedNeonTx.sig, neonTx.sig)
			assert.Equal(t, expectedNeonTx.addr, neonTx.addr)
			assert.Equal(t, expectedNeonTx.nonce, neonTx.nonce)
			assert.Equal(t, expectedNeonTx.value, neonTx.value)
			assert.Equal(t, expectedNeonTx.callData, neonTx.callData)
			assert.Equal(t, expectedNeonTx.gasLimit, neonTx.gasLimit)
			assert.Equal(t, expectedNeonTx.gasPrice, neonTx.gasPrice)

			actualReturnEvent := tx.neonEvents[0]
			assert.Equal(t, expectedReturnEvent.totalGasUsed, actualReturnEvent.totalGasUsed)
			assert.Equal(t, expectedReturnEvent.topics, actualReturnEvent.topics)
			assert.Equal(t, expectedReturnEvent.solSig, actualReturnEvent.solSig)
			assert.Equal(t, expectedReturnEvent.eventType, actualReturnEvent.eventType)
			assert.Equal(t, expectedReturnEvent.eventOrder, actualReturnEvent.eventOrder)
			assert.Equal(t, expectedReturnEvent.eventLevel, actualReturnEvent.eventLevel)
			assert.Equal(t, expectedReturnEvent.hidden, actualReturnEvent.hidden)
			assert.Equal(t, expectedReturnEvent.reverted, actualReturnEvent.reverted)
			assert.Equal(t, expectedReturnEvent.address, actualReturnEvent.address)
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
			const (
				solSig         = "3ZVbKg66KnKQL51m5vhzBiaQM3rfrACgnYMeLrqzuazWPsKFf7Wuw81swdeM8CFZpytThWzG9xpZjZqnUDoEFPSC"
				operator       = "896VtgKpVopcBY9KFxvmXCzriV7gmfyynuXsbTzPhf4t"
				expectedHolder = "896VtgKpVopcBY9KFxvmXCzriV7gmfyynuXsbTzPhf4t"
				blockSlot      = 1111
			)

			neonTxSig := utils.Base64stringToHex("/vstpTXpJ7hf5o64HLLkpYJ8kF94OBoB7yMiqpsK7o4=")

			// generate data for test
			decodedNeonTx, err := hex.DecodeString(neonTxSig[2:])
			assert.Empty(t, err)
			dataChunk := bytes.Repeat([]byte{0x1}, 20)
			ixData := append([]byte{0}, decodedNeonTx...)
			ixData = append(ixData, dataChunk...)

			expectedDecodedTxKey := strings.ToLower(solSig)
			expectedHolderTxKey := fmt.Sprintf("%v:%v", expectedHolder, strings.ToLower(neonTxSig[2:]))

			solTx := &SolNeonIxReceiptInfo{
				metaInfo: SolIxMetaInfo{
					status:    1, // success
					neonTxSig: neonTxSig,
					neonTxReturn: &NeonLogTxReturn{
						GasUsed:  0,
						Status:   1,
						Canceled: false,
					},
				},
				solSign:   solSig,
				blockSlot: blockSlot,
				programIx: 0x22,
				ixData:    ixData,
				solTxCost: SolTxCostInfo{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				ident: Ident{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				accountKeys: []string{
					"3NqgsSRfjpmDfzRH4PLKrzBvMc8MgFXgU58Yy8n41KF5",
					"3MTrvrjHerq8JAnNvr3QvfMK4VFs1493JAwQj4FMf8bt",
					expectedHolder,
					"F9bMPQ3xBnfkRrhFPDSpjnhXb5yxkzfLScYirtR87tUD",
					"GS6DZNvbv7dmjZgquokzyX3sFccgjjd5MXeK46PH1F7N",
					"GvRqcwRyKgWfGA4AyrNZaRyDe8HqiUDT5sRhPSS6vpt2",
					"11111111111111111111111111111111",
					"ComputeBudget111111111111111111111111111111",
					"53DfF883gyixYNXnM7s5xhdeyV8mVk9T4i2hGV9vG9io",
				},
				accounts: []int{
					2,
					0,
					3,
					4,
					6,
					8,
					1,
					5},
			}

			solState := initUnifiedTestSolState(operator, solTx)
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

			holder, ok := solState.NeonBlock().neonHolders[expectedHolderTxKey]
			assert.True(t, ok)
			assert.Equal(t, expectedHolder, holder.Account())
			assert.Equal(t, neonTxSig[2:], holder.NeonTxSig())
			assert.NotEmpty(t, holder.data)
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

	t.Run("CancelWithHashIxDecoder decoding iter tx wit return event: success",
		func(t *testing.T) {
			const (
				solSig         = "4y1mhzcL6w64bm4EGvTKu5m6yM5jDufgSjxAnZq1WDkdMPoFTNWvcicFoSNFFDcyarF2wUtzMN9ej2rrxgBJkGou"
				operator       = "896VtgKpVopcBY9KFxvmXCzriV7gmfyynuXsbTzPhf4t"
				expectedHolder = "9gZjcTuq3iyxc2v1mFMTBPy4ajAhmkDeEUEUxHLtvRja"
				blockSlot      = 1642
			)

			neonTxSig := utils.Base64stringToHex("mm+dAi50RbJaxJOOkRXQrtIyCH0jVRIHyNRd3F0/v8A=")

			// generate data for test
			decodedNeonTx, err := hex.DecodeString(neonTxSig[2:])
			assert.Empty(t, err)
			ixData := append([]byte{0}, decodedNeonTx...)

			gasUsed := int(binary.LittleEndian.Uint32(utils.Base64stringToBytes("mDoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")))
			totalGasUsed := int(binary.LittleEndian.Uint32(utils.Base64stringToBytes("mDoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")))

			expectedDecodedTxKey := strings.ToLower(neonTxSig[2:])
			returnEventTotalGasUsed, err := strconv.ParseInt(fmt.Sprintf("%x", totalGasUsed), 16, 64)
			assert.Empty(t, err)

			expectedReturnEvent := NeonLogTxEvent{
				eventType:    Cancel,
				hidden:       true,
				address:      "",
				topics:       nil,
				data:         convertHexStringToLittleEndianByte("0x1"),
				solSig:       solSig,
				totalGasUsed: returnEventTotalGasUsed + 5000,
				reverted:     false,
			}

			solTx := &SolNeonIxReceiptInfo{
				metaInfo: SolIxMetaInfo{
					status:           1, // success
					neonTxSig:        neonTxSig,
					neonGasUsed:      int64(gasUsed),
					neonTotalGasUsed: int64(totalGasUsed),
				},
				solSign:   solSig,
				blockSlot: blockSlot,
				programIx: 0x23,
				ixData:    ixData,
				solTxCost: SolTxCostInfo{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				ident: Ident{
					solSign:   solSig,
					blockSlot: blockSlot,
				},
				accountKeys: []string{
					"BMp6gEnveANdvSvspESJUrNczuHz1GF5UQKjVLCkAZih",
					"2Y4cZc5invCZLZVaHUEss3S8DwobUFPSzsppGGPskkrw",
					"3gDyAzr9P8t1489MD86rpBETQ1zpQWd84J6YiFH1mEXy",
					expectedHolder,
					"BQf6ke1FB9oeDqq9To7NGdJpxXhQxaAdTsJNapmRwjeL",
					"D7aa3D2MNjTZ1xqJ8pAS6Km45dg5MyjCQ3bFWSHwr5jx",
					"11111111111111111111111111111111",
					"ComputeBudget111111111111111111111111111111",
					"53DfF883gyixYNXnM7s5xhdeyV8mVk9T4i2hGV9vG9io",
				},
				accounts: []int{
					3,
					0,
					4,
					1,
					6,
					8,
					5,
					2,
				},
			}

			expectedNeonTx := NeonTxInfo{
				sig: neonTxSig,
			}

			solState := initUnifiedTestSolState(operator, solTx)

			decoder := InitCancelWithHashIxDecoder(&IxDecoder{
				log:    log,
				name:   "CancelWithHash",
				ixCode: 0x23,
				state:  solState,
			})

			result := decoder.Execute()
			assert.True(t, result)

			tx, ok := solState.NeonBlock().neonTxs[expectedDecodedTxKey]
			assert.True(t, ok)
			assert.Equal(t, expectedHolder, tx.storageAccount)
			assert.Equal(t, true, tx.neonReceipt.neonTxRes.canceled)
			assert.Equal(t, true, tx.neonReceipt.neonTxRes.completed)

			neonTx := tx.neonReceipt.neonTx
			assert.Empty(t, neonTx.err)
			assert.Equal(t, expectedNeonTx.sig, neonTx.sig)
			assert.Equal(t, expectedNeonTx.addr, neonTx.addr)
			assert.Equal(t, expectedNeonTx.nonce, neonTx.nonce)
			assert.Equal(t, expectedNeonTx.value, neonTx.value)
			assert.Equal(t, expectedNeonTx.callData, neonTx.callData)
			assert.Equal(t, expectedNeonTx.gasLimit, neonTx.gasLimit)
			assert.Equal(t, expectedNeonTx.gasPrice, neonTx.gasPrice)

			actualReturnEvent := tx.neonEvents[0]
			assert.Equal(t, expectedReturnEvent.totalGasUsed, actualReturnEvent.totalGasUsed)
			assert.Equal(t, expectedReturnEvent.topics, actualReturnEvent.topics)
			assert.Equal(t, expectedReturnEvent.solSig, actualReturnEvent.solSig)
			assert.Equal(t, expectedReturnEvent.eventType, actualReturnEvent.eventType)
			assert.Equal(t, expectedReturnEvent.eventOrder, actualReturnEvent.eventOrder)
			assert.Equal(t, expectedReturnEvent.eventLevel, actualReturnEvent.eventLevel)
			assert.Equal(t, expectedReturnEvent.hidden, actualReturnEvent.hidden)
			assert.Equal(t, expectedReturnEvent.reverted, actualReturnEvent.reverted)
			assert.Equal(t, expectedReturnEvent.address, actualReturnEvent.address)
		})
}

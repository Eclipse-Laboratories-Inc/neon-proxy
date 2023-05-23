package utils

import (
	"github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/test-go/testify/assert"
	"math/big"
	"testing"
)

func TestRlpDecoding(t *testing.T) {
	const (
		ixData    = "CUrU71iEchwyUYbwpusc7ciwTfJs1mnxngF5eQA5zrkL5CJDxtedKXbUhb4FuYXLk9RVxXAjrNbaZuytdf4nqNPW2oBmFvZxUZi8QcGMjSEZ5fxwCWBbXQ5E6kyLk3MdEAHvFqChx4e7qbYdXsCmNoeWg78Ddqne7VP5R17quCcex2wnj5eJ51oWkfpTkiWr1F5kgTBRH3mmetuW8CX529VbtK6gKvfZqJg9KV2aZr6SyKmAGtVrmVdRnDwv2bvaXg"
		expectedR = "9fdc62d9d340b345cf08ffe373d7302d64800bfdedac2d7567ee90ed0e5648d0"
		expectedS = "2e36e6bac35772258a4797afa267bfbd6ad1a6f29fd6a282ef311c097ae414df"
	)

	bigR, ok := big.NewInt(0).SetString(expectedR, 16)
	assert.True(t, ok)

	bigS, ok := big.NewInt(0).SetString(expectedS, 16)
	assert.True(t, ok)

	expectedTx := NeonTx{
		Nonce:     big.NewInt(0xbf),
		GasPrice:  big.NewInt(0x14b1b0cf32),
		GasLimit:  big.NewInt(0x3b9aca00),
		ToAddress: []byte("ca12f8c0ca275bd38937c1bf354da40cc116bb68"),
		Value:     big.NewInt(0),
		CallData:  []byte("0xc9c6539600000000000000000000000058b2145cfa2406097be00c0057d24a3f3f90361100000000000000000000000010acfd050938dfdaf3d3d9831c05fc6ed9e4194b"),
		V:         big.NewInt(0x102),
		R:         bigR,
		S:         bigS,
	}

	txData := base58.Decode(ixData)
	var tx NeonTx

	// check if ethereum lib works fine
	err := rlp.DecodeBytes(txData[13:], &tx)
	assert.Empty(t, err)
	assert.Equal(t, 0, tx.V.Cmp(expectedTx.V))
	assert.Equal(t, 0, tx.R.Cmp(expectedTx.R))
	assert.Equal(t, 0, tx.S.Cmp(expectedTx.S))
	assert.Equal(t, expectedTx.Nonce, tx.Nonce)
	assert.Equal(t, expectedTx.GasPrice, tx.GasPrice)
	assert.Equal(t, expectedTx.GasLimit, tx.GasLimit)

	t.Run("get tx sign: success", func(t *testing.T) {
		sign := tx.HexTxSig()
		assert.Equal(t, "0xff4b9291e48cf17e8ab2454c793e8ab597b664064d666409db2210f185edb144", sign)
	})

	t.Run("get sender address: success", func(t *testing.T) {
		const expectedSenderAddr = "0xaa4d6f4ff831181a2bbfd4d62260dabdea964ff1"
		senderAddr := tx.HexSender()
		assert.Equal(t, expectedSenderAddr, senderAddr)
	})

	t.Run("get call data hex: success", func(t *testing.T) {
		callData := tx.HexCallData()
		assert.Equal(t, string(expectedTx.CallData), callData)
	})
}

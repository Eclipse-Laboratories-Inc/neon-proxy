package solana

import (
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"github.com/stretchr/testify/assert"
	"testing"
)

// used testnet endpoint from https://docs.solana.com/cluster/rpc-endpoints
func initClient() (*Client, error) {
	log, err := logger.NewLogger("logger", logger.LogSettings{})
	if err != nil {
		return nil, err
	}
	return NewClient(log, "https://api.testnet.solana.com"), nil
}

func TestGetBlockBySlot(t *testing.T) {
	expectedBase58Hash := "8ofhVKwbvo78FxfZPq8fpHTpVZXaCA38T4uHvJejqePN"
	expectedBytesHash, err := base58.Decode(expectedBase58Hash)
	assert.NoError(t, err)

	expectedBase58ParentHash := "6mVEmyp6jwEBJPK8FyHdMvZFtu3AkLR6WpxrhcgGtoib"
	expectedBytesParentHash, err := base58.Decode(expectedBase58ParentHash)
	assert.NoError(t, err)

	expectedBlock := &rpc.GetBlockResult{
		Blockhash:         solana.HashFromBytes(expectedBytesHash),
		PreviousBlockhash: solana.HashFromBytes(expectedBytesParentHash),
		ParentSlot:        190780808,
	}

	client, err := initClient()
	assert.NoError(t, err)

	block, err := client.GetBlockInfoBySlot(context.Background(), 190780809, rpc.CommitmentFinalized, nil)
	assert.NoError(t, err)

	assert.Equal(t, expectedBlock.Blockhash, block.Blockhash)
	assert.Equal(t, expectedBlock.PreviousBlockhash, block.PreviousBlockhash)
	assert.Equal(t, expectedBlock.ParentSlot, block.ParentSlot)
}

func TestGetLatestBlockSlot(t *testing.T) {
	client, err := initClient()
	assert.NoError(t, err)

	slot, err := client.GetLatestBlockSlot(context.Background(), rpc.CommitmentFinalized)
	assert.NoError(t, err)

	fmt.Printf("finalised slot is %v\n", slot)

	slot, err = client.GetLatestBlockSlot(context.Background(), rpc.CommitmentRecent)
	assert.NoError(t, err)

	fmt.Printf("recent slot is %v\n", slot)
}

func TestGetLatestBlock(t *testing.T) {
	client, err := initClient()
	assert.NoError(t, err)

	block, err := client.GetLatestBlock(context.Background(), rpc.CommitmentFinalized)
	assert.NoError(t, err)

	fmt.Printf("block parent slot is %v\n", block.ParentSlot)
}

func TestGetBlockHeight(t *testing.T) {
	client, err := initClient()
	assert.NoError(t, err)

	t.Run("get latest known block height", func(t *testing.T) {
		_, err := client.GetBlockHeight(context.Background(), nil, rpc.CommitmentFinalized)
		assert.NoError(t, err)
	})

	t.Run("get block height by block slot", func(t *testing.T) {
		slot := uint64(190780809)
		_, err := client.GetBlockHeight(context.Background(), &slot, rpc.CommitmentFinalized)
		assert.NoError(t, err)
	})
}

func TestCheckConfirmOfTxSignsList(t *testing.T) {
	cl := Client{
		conn: &MockSolanaClient{},
	}

	firstSignature := solana.SignatureFromBytes([]byte("0x111"))
	secondSignature := solana.SignatureFromBytes([]byte("0x222"))
	thirdSignature := solana.SignatureFromBytes([]byte("0x333"))

	t.Run("success: all txs have 'confirmed' status", func(t *testing.T) {
		// tell mocked method, that it is 'success' case
		ctx := context.WithValue(context.Background(), "success", "bool")
		ok, err := cl.CheckConfirmOfTxSignsList(ctx, []solana.Signature{firstSignature, secondSignature, thirdSignature}, 3)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("success: 0 txs got as parameters -> return true", func(t *testing.T) {
		ok, err := cl.CheckConfirmOfTxSignsList(context.Background(), []solana.Signature{}, 3)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("fail: first tx doesn't 'confirmed' status", func(t *testing.T) {
		ok, err := cl.CheckConfirmOfTxSignsList(context.Background(), []solana.Signature{firstSignature, secondSignature, thirdSignature}, 3)
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestGetBlockStatus(t *testing.T) {
	cl := Client{
		conn: &MockSolanaClient{},
	}

	t.Run("block status is 'finalized'", func(t *testing.T) {
		status, err := cl.GetBlockStatus(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, string(rpc.CommitmentFinalized), status.Commitment)
	})

	t.Run("block status is 'safe'", func(t *testing.T) {
		status, err := cl.GetBlockStatus(context.Background(), 2)
		assert.NoError(t, err)
		assert.Equal(t, "safe", status.Commitment)
	})

	t.Run("block status is 'confirmed'", func(t *testing.T) {
		status, err := cl.GetBlockStatus(context.Background(), 3)
		assert.NoError(t, err)
		assert.Equal(t, string(rpc.CommitmentConfirmed), status.Commitment)
	})
}

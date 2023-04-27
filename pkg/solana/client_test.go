package solana

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/golang/mock/gomock"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"github.com/neonlabsorg/neon-proxy/pkg/solana/mocks"
	"github.com/test-go/testify/assert"
	"testing"
)

var (
	client *Client
	conn   *mocks.MockSolanaRpcConnection
)

func TestInitClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	conn = mocks.NewMockSolanaRpcConnection(ctrl)

	log, err := logger.NewLogger("logger", logger.LogSettings{})
	assert.NoError(t, err)
	client = NewClient(log, conn)
}

func TestCheckConfirmOfTxSignsList(t *testing.T) {
	if client == nil {
		TestInitClient(t)
	}

	const (
		blockHeight      = 1234987
		validBlockHeight = blockHeight + 2
	)
	firstSignature := solana.SignatureFromBytes([]byte("0x111"))
	secondSignature := solana.SignatureFromBytes([]byte("0x222"))
	thirdSignature := solana.SignatureFromBytes([]byte("0x333"))

	ctx := context.Background()
	t.Run("success: all txs have 'confirmed' status", func(t *testing.T) {
		gomock.InOrder(
			conn.EXPECT().GetBlockHeight(gomock.Any(), rpc.CommitmentConfirmed).Return(uint64(blockHeight), nil),
			conn.EXPECT().GetSignatureStatuses(gomock.Any(), false, []solana.Signature{firstSignature, secondSignature, thirdSignature}).
				Return(&rpc.GetSignatureStatusesResult{
					RPCContext: rpc.RPCContext{},
					Value: []*rpc.SignatureStatusesResult{
						{
							Slot:               uint64(blockHeight),
							ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
						},
						{
							Slot:               uint64(blockHeight),
							ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
						},
						{
							Slot:               uint64(blockHeight),
							ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
						}},
				}, nil),
		)

		ok, err := client.CheckConfirmOfTxSignsList(ctx, []solana.Signature{firstSignature, secondSignature, thirdSignature}, validBlockHeight)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("success: 0 txs got as parameters -> return true", func(t *testing.T) {
		ok, err := client.CheckConfirmOfTxSignsList(context.Background(), []solana.Signature{}, validBlockHeight)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("fail: first tx doesn't 'confirmed' status", func(t *testing.T) {
		gomock.InOrder(
			conn.EXPECT().GetBlockHeight(gomock.Any(), rpc.CommitmentConfirmed).Return(uint64(blockHeight), nil),
			conn.EXPECT().GetSignatureStatuses(gomock.Any(), false, []solana.Signature{firstSignature, secondSignature, thirdSignature}).
				Return(&rpc.GetSignatureStatusesResult{
					RPCContext: rpc.RPCContext{},
					Value: []*rpc.SignatureStatusesResult{
						{
							Slot:               uint64(blockHeight),
							ConfirmationStatus: rpc.ConfirmationStatusFinalized,
						},
						{
							Slot:               uint64(blockHeight),
							ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
						},
						{
							Slot:               uint64(blockHeight),
							ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
						}},
				}, nil),
		)

		ok, err := client.CheckConfirmOfTxSignsList(ctx, []solana.Signature{firstSignature, secondSignature, thirdSignature}, validBlockHeight)
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestGetBlockStatus(t *testing.T) {
	if client == nil {
		TestInitClient(t)
	}

	const blockSlot = uint64(1234)

	t.Run("block status is 'finalized'", func(t *testing.T) {
		gomock.InOrder(
			conn.EXPECT().GetBlockWithOpts(gomock.Any(), blockSlot, &rpc.GetBlockOpts{
				Commitment: rpc.CommitmentFinalized,
			}).Return(&rpc.GetBlockResult{}, nil),
		)

		block, err := client.GetBlockStatus(context.Background(), blockSlot)
		assert.NoError(t, err)
		assert.Equal(t, blockSlot, block.BlockSlot)
		assert.Equal(t, string(rpc.CommitmentFinalized), block.Commitment)
	})

	t.Run("block status is 'safe'", func(t *testing.T) {
		gomock.InOrder(
			conn.EXPECT().GetBlockWithOpts(gomock.Any(), blockSlot, &rpc.GetBlockOpts{
				Commitment: rpc.CommitmentFinalized,
			}).Return(nil, nil),

			conn.EXPECT().GetBlockCommitment(gomock.Any(), blockSlot).
				Return(&rpc.GetBlockCommitmentResult{
					Commitment: []uint64{1, 3, 2, 9, 2},
					TotalStake: 3,
				}, nil),
		)

		status, err := client.GetBlockStatus(context.Background(), blockSlot)
		assert.NoError(t, err)
		assert.Equal(t, "safe", status.Commitment)
	})

	t.Run("block status is 'confirmed'", func(t *testing.T) {
		gomock.InOrder(
			conn.EXPECT().GetBlockWithOpts(gomock.Any(), blockSlot, &rpc.GetBlockOpts{
				Commitment: rpc.CommitmentFinalized,
			}).Return(nil, nil),

			conn.EXPECT().GetBlockCommitment(gomock.Any(), blockSlot).
				Return(&rpc.GetBlockCommitmentResult{
					Commitment: []uint64{3},
					TotalStake: 89,
				}, nil),
		)

		status, err := client.GetBlockStatus(context.Background(), blockSlot)
		assert.NoError(t, err)
		assert.Equal(t, string(rpc.CommitmentConfirmed), status.Commitment)
	})
}

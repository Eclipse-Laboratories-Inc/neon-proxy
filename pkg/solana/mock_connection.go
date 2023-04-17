package solana

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type SolanaRpcConnection interface {
	Close() error
	GetHealth(ctx context.Context) (out string, err error)
	GetSignaturesForAddressWithOpts(ctx context.Context, account solana.PublicKey, opts *rpc.GetSignaturesForAddressOpts) (out []*rpc.TransactionSignature, err error)
	GetBlockWithOpts(ctx context.Context, slot uint64, opts *rpc.GetBlockOpts) (out *rpc.GetBlockResult, err error)
	GetBlockHeight(ctx context.Context, commitment rpc.CommitmentType) (out uint64, err error)
	GetRecentBlockhash(ctx context.Context, commitment rpc.CommitmentType) (out *rpc.GetRecentBlockhashResult, err error)
	GetAccountInfoWithOpts(ctx context.Context, account solana.PublicKey, opts *rpc.GetAccountInfoOpts) (*rpc.GetAccountInfoResult, error)
	GetProgramAccountsWithOpts(ctx context.Context, publicKey solana.PublicKey, opts *rpc.GetProgramAccountsOpts) (out rpc.GetProgramAccountsResult, err error)
	GetBalance(ctx context.Context, publicKey solana.PublicKey, commitment rpc.CommitmentType) (out *rpc.GetBalanceResult, err error)
	GetSignatureStatuses(ctx context.Context, searchTransactionHistory bool, transactionSignatures ...solana.Signature) (out *rpc.GetSignatureStatusesResult, err error)
	GetBlockCommitment(ctx context.Context, block uint64) (out *rpc.GetBlockCommitmentResult, err error)
	GetMinimumBalanceForRentExemption(ctx context.Context, dataSize uint64, commitment rpc.CommitmentType) (lamport uint64, err error)
	SendTransaction(ctx context.Context, transaction *solana.Transaction) (signature solana.Signature, err error)
	GetTransaction(ctx context.Context, txSig solana.Signature, opts *rpc.GetTransactionOpts) (out *rpc.GetTransactionResult, err error)
	GetClusterNodes(ctx context.Context) (out []*rpc.GetClusterNodesResult, err error)
	GetSlot(ctx context.Context, commitment rpc.CommitmentType) (out uint64, err error)
}

type MockSolanaClient struct{}

func (c *MockSolanaClient) Close() error                                          { return nil }
func (c *MockSolanaClient) GetHealth(ctx context.Context) (out string, err error) { return "ok", nil }
func (c *MockSolanaClient) GetSignaturesForAddressWithOpts(ctx context.Context, account solana.PublicKey, opts *rpc.GetSignaturesForAddressOpts) (out []*rpc.TransactionSignature, err error) {
	return nil, nil
}
func (c *MockSolanaClient) GetBlockWithOpts(ctx context.Context, slot uint64, opts *rpc.GetBlockOpts) (out *rpc.GetBlockResult, err error) {
	switch {
	case opts != nil && opts.Commitment == rpc.CommitmentFinalized && slot == 1:
		return &rpc.GetBlockResult{
			Blockhash:  solana.HashFromBytes([]byte("0x111")),
			ParentSlot: slot,
		}, nil
	case opts != nil && opts.Commitment == rpc.CommitmentConfirmed && slot == 2:
		return &rpc.GetBlockResult{
			Blockhash:  solana.HashFromBytes([]byte("0x222")),
			ParentSlot: slot,
		}, nil
	default:
		return nil, nil
	}
}

func (c *MockSolanaClient) GetRecentBlockhash(ctx context.Context, commitment rpc.CommitmentType) (out *rpc.GetRecentBlockhashResult, err error) {
	return nil, nil
}
func (c *MockSolanaClient) GetAccountInfoWithOpts(ctx context.Context, account solana.PublicKey, opts *rpc.GetAccountInfoOpts) (*rpc.GetAccountInfoResult, error) {
	return nil, nil
}
func (c *MockSolanaClient) GetProgramAccountsWithOpts(ctx context.Context, publicKey solana.PublicKey, opts *rpc.GetProgramAccountsOpts) (out rpc.GetProgramAccountsResult, err error) {
	return nil, nil
}
func (c *MockSolanaClient) GetBalance(ctx context.Context, publicKey solana.PublicKey, commitment rpc.CommitmentType) (out *rpc.GetBalanceResult, err error) {
	return nil, nil
}
func (c *MockSolanaClient) GetSignatureStatuses(ctx context.Context, searchTransactionHistory bool, transactionSignatures ...solana.Signature) (out *rpc.GetSignatureStatusesResult, err error) {
	result := make([]*rpc.SignatureStatusesResult, 0)
	confirmations1 := uint64(10)
	confirmations2 := uint64(30)
	confirmations3 := uint64(50)

	firstConfirmTxStatus := rpc.ConfirmationStatusProcessed
	if ctx.Value("success") != nil {
		// if we are testing success case, set up status for all tx signatures 'confirmed'
		firstConfirmTxStatus = rpc.ConfirmationStatusConfirmed
	}

	for _, sign := range transactionSignatures {
		if sign == solana.SignatureFromBytes([]byte("0x111")) {
			result = append(result, &rpc.SignatureStatusesResult{
				Slot:               1,
				Confirmations:      &confirmations1,
				ConfirmationStatus: firstConfirmTxStatus,
			})
		}
		if sign == solana.SignatureFromBytes([]byte("0x222")) {
			result = append(result, &rpc.SignatureStatusesResult{
				Slot:               1,
				Confirmations:      &confirmations2,
				ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
			})
		}
		if sign == solana.SignatureFromBytes([]byte("0x333")) {
			result = append(result, &rpc.SignatureStatusesResult{
				Slot:               1,
				Confirmations:      &confirmations3,
				ConfirmationStatus: rpc.ConfirmationStatusConfirmed,
			})
		}
	}
	return &rpc.GetSignatureStatusesResult{
		Value: result,
	}, nil
}
func (c *MockSolanaClient) GetBlockCommitment(ctx context.Context, block uint64) (out *rpc.GetBlockCommitmentResult, err error) {
	if block == 3 {
		return &rpc.GetBlockCommitmentResult{
			Commitment: []uint64{0, 0, 0, 0, 0, 1},
			TotalStake: 100,
		}, nil
	} else {
		return &rpc.GetBlockCommitmentResult{
			Commitment: []uint64{0, 3, 10, 45, 1, 10},
			TotalStake: 100,
		}, nil
	}
}

func (c *MockSolanaClient) GetMinimumBalanceForRentExemption(ctx context.Context, dataSize uint64, commitment rpc.CommitmentType) (lamport uint64, err error) {
	return 0, nil
}
func (c *MockSolanaClient) SendTransaction(ctx context.Context, transaction *solana.Transaction) (signature solana.Signature, err error) {
	return [64]byte{}, err
}
func (c *MockSolanaClient) GetTransaction(ctx context.Context, txSig solana.Signature, opts *rpc.GetTransactionOpts) (out *rpc.GetTransactionResult, err error) {
	return nil, nil
}
func (c *MockSolanaClient) GetClusterNodes(ctx context.Context) (out []*rpc.GetClusterNodesResult, err error) {
	return nil, nil
}
func (c *MockSolanaClient) GetSlot(ctx context.Context, commitment rpc.CommitmentType) (out uint64, err error) {
	return 0, nil
}
func (receiver *MockSolanaClient) GetBlockHeight(ctx context.Context, commitment rpc.CommitmentType) (out uint64, err error) {
	return 11111, nil
}

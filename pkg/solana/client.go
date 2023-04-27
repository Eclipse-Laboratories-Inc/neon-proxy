package solana

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
)

type Client struct {
	log  logger.Logger
	conn SolanaRpcConnection
}

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

func NewClient(log logger.Logger, rpcConn SolanaRpcConnection) *Client {
	return &Client{
		log:  log,
		conn: rpcConn,
	}
}

func (c *Client) IsHealth(ctx context.Context) (bool, error) {
	resp, err := c.conn.GetHealth(ctx)
	if err != nil {
		return false, err
	}

	if resp == rpc.HealthOk {
		return true, nil
	}
	return false, nil
}

// GetSignListForAddress gets signatures from newest to oldest
// https://docs.solana.com/api/http#getsignaturesforaddress
func (c *Client) GetSignListForAddress(
	ctx context.Context, addr solana.PublicKey,
	limit *int, before, until *solana.Signature, commitmentType rpc.CommitmentType,
	minContextSlot *uint64) ([]*rpc.TransactionSignature, error) {
	opts := &rpc.GetSignaturesForAddressOpts{
		Limit:          limit,
		Commitment:     commitmentType,
		MinContextSlot: minContextSlot,
	}
	if before != nil {
		opts.Before = *before
	}

	if until != nil {
		opts.Until = *until
	}

	return c.conn.GetSignaturesForAddressWithOpts(ctx, addr, opts)
}

func (c *Client) GetBlockInfoBySlot(ctx context.Context, slot uint64,
	commitmentType rpc.CommitmentType, encodingType *solana.EncodingType) (*rpc.GetBlockResult, error) {
	if encodingType != nil {
		return c.conn.GetBlockWithOpts(ctx, slot, &rpc.GetBlockOpts{
			Encoding:   *encodingType,
			Commitment: commitmentType,
		})
	}
	return c.conn.GetBlockWithOpts(ctx, slot, &rpc.GetBlockOpts{
		Commitment: commitmentType,
	})
}

func (c *Client) GetBlockInfoListBySlot(ctx context.Context, slots []uint64,
	commitmentType rpc.CommitmentType, encodingType solana.EncodingType) ([]*rpc.GetBlockResult, error) {
	result := make([]*rpc.GetBlockResult, 0)
	for _, slot := range slots {
		resp, err := c.conn.GetBlockWithOpts(ctx, slot, &rpc.GetBlockOpts{
			Encoding:   encodingType,
			Commitment: commitmentType,
		})
		if err != nil {
			return nil, err
		}
		result = append(result, resp)
	}
	return result, nil
}

func (c *Client) GetBlockHash(ctx context.Context, slot uint64, commitmentType rpc.CommitmentType) (solana.Hash, error) {
	blockInfo, err := c.GetBlockInfoBySlot(ctx, slot, commitmentType, nil)
	if err != nil {
		return solana.Hash{}, err
	}

	return blockInfo.Blockhash, nil
}

func (c *Client) GetBlockHeight(ctx context.Context, slot *uint64, commitmentType rpc.CommitmentType) (*uint64, error) {
	if slot == nil {
		height, err := c.conn.GetBlockHeight(ctx, commitmentType)
		if err != nil {
			return nil, err
		}
		return &height, nil
	}

	blockInfo, err := c.GetBlockInfoBySlot(ctx, *slot, commitmentType, nil)
	if err != nil {
		return nil, err
	}

	return blockInfo.BlockHeight, nil
}

func (c *Client) GetRecentBlockHash(ctx context.Context, commitmentType rpc.CommitmentType) (*rpc.GetRecentBlockhashResult, error) {
	return c.conn.GetRecentBlockhash(ctx, commitmentType)
}

func (c *Client) GetAccountInfo(ctx context.Context, account solana.PublicKey,
	encoding solana.EncodingType, commitmentType rpc.CommitmentType,
	dataSlice *rpc.DataSlice, minContextSlot *uint64) (*rpc.GetAccountInfoResult, error) {
	return c.conn.GetAccountInfoWithOpts(ctx, account, &rpc.GetAccountInfoOpts{
		Encoding:       encoding,
		Commitment:     commitmentType,
		DataSlice:      dataSlice,
		MinContextSlot: minContextSlot,
	})
}

func (c *Client) GetAccountInfoList(ctx context.Context, accounts []solana.PublicKey,
	encoding solana.EncodingType, commitmentType rpc.CommitmentType,
	dataSlice *rpc.DataSlice, minContextSlot *uint64) ([]*rpc.GetAccountInfoResult, error) {

	result := make([]*rpc.GetAccountInfoResult, 0)
	for _, account := range accounts {
		resp, err := c.conn.GetAccountInfoWithOpts(ctx, account, &rpc.GetAccountInfoOpts{
			Encoding:       encoding,
			Commitment:     commitmentType,
			DataSlice:      dataSlice,
			MinContextSlot: minContextSlot,
		})
		if err != nil {
			return nil, err
		}
		result = append(result, resp)
	}
	return result, nil
}

func (c *Client) GetProgramAccountInfoList(ctx context.Context, account solana.PublicKey,
	encoding solana.EncodingType, commitmentType rpc.CommitmentType,
	dataSlice *rpc.DataSlice, filters []rpc.RPCFilter,
) (rpc.GetProgramAccountsResult, error) {
	// TODO implement decode_account_info
	return c.conn.GetProgramAccountsWithOpts(ctx, account, &rpc.GetProgramAccountsOpts{
		Commitment: commitmentType,
		Encoding:   encoding,
		DataSlice:  dataSlice,
		Filters:    filters,
	})
}

func (c *Client) GetSolBalance(ctx context.Context, account solana.PublicKey, commitmentType rpc.CommitmentType) (*rpc.GetBalanceResult, error) {
	return c.conn.GetBalance(ctx, account, commitmentType)
}

func (c *Client) GetSolBalanceList(ctx context.Context, accounts []solana.PublicKey, commitmentType rpc.CommitmentType) ([]*rpc.GetBalanceResult, error) {
	result := make([]*rpc.GetBalanceResult, 0)
	for _, acc := range accounts {
		resp, err := c.conn.GetBalance(ctx, acc, commitmentType)
		if err != nil {
			return nil, err
		}
		result = append(result, resp)
	}
	return result, nil
}

// CheckConfirmOfTxSignsList checks if all transactions with given signatures are confirmed
func (c *Client) CheckConfirmOfTxSignsList(ctx context.Context, txsSigns []solana.Signature, validBlockHeight uint64) (bool, error) {
	const limit = 100
	if len(txsSigns) == 0 {
		return true, nil
	}
	var (
		searchInHistory bool
		partTxsSigns    []solana.Signature
	)

	blockHeight, err := c.GetBlockHeight(ctx, nil, rpc.CommitmentConfirmed)
	if err != nil {
		return false, err
	}
	if blockHeight != nil && *blockHeight >= validBlockHeight {
		searchInHistory = true
	}

	for len(txsSigns) > 0 {
		if len(txsSigns) > limit {
			// divide signatures to 'batch' and 'rest'
			partTxsSigns, txsSigns = txsSigns[:limit], txsSigns[limit:]
		} else {
			partTxsSigns = txsSigns
			txsSigns = txsSigns[:0]
		}
		signsStatuses, err := c.conn.GetSignatureStatuses(ctx, searchInHistory, partTxsSigns...)
		if err != nil {
			return false, err
		} else if signsStatuses != nil && len(signsStatuses.Value) == 0 {
			return false, nil
		}
		for _, signStatus := range signsStatuses.Value {
			if signStatus.ConfirmationStatus != rpc.ConfirmationStatusConfirmed {
				return false, err
			}
		}
	}
	return true, nil
}

func (c *Client) GetBlockStatus(ctx context.Context, blockSlot uint64) (*BlockStatus, error) {
	blockInfo, err := c.GetBlockInfoBySlot(ctx, blockSlot, rpc.CommitmentFinalized, nil)
	if err != nil {
		return nil, err
	} else if blockInfo != nil {
		return &BlockStatus{
			BlockSlot:  blockSlot,
			Commitment: string(rpc.CommitmentFinalized),
		}, nil
	}

	resp, err := c.conn.GetBlockCommitment(ctx, blockSlot)
	if err != nil {
		return nil, err
	}

	var (
		votedStake uint64
		commitment string
	)
	for _, commitmentItem := range resp.Commitment {
		votedStake += commitmentItem
	}

	stake := float64(votedStake*100) / float64(resp.TotalStake)
	if stake > 66.67 {
		// optimistic-finalized => 2/3 of validators
		commitment = "safe"
	} else {
		commitment = string(rpc.CommitmentConfirmed)
	}

	return &BlockStatus{
		BlockSlot:  blockSlot,
		Commitment: commitment,
	}, nil
}

func (c *Client) GetMultipleRentExemptBalancesForSize(ctx context.Context, commitmentType rpc.CommitmentType, dataSize uint64) (uint64, error) {
	return c.conn.GetMinimumBalanceForRentExemption(ctx, dataSize, commitmentType)
}

func (c *Client) SendTxList(ctx context.Context, txs []*solana.Transaction) ([]solana.Signature, error) {
	result := make([]solana.Signature, 0)
	for _, tx := range txs {
		sign, err := c.conn.SendTransaction(ctx, tx)
		if err != nil {
			return nil, err
		}
		result = append(result, sign)
	}
	return result, nil
}

func (c *Client) GetTxReceiptList(ctx context.Context, signatures []solana.Signature,
	encodingType solana.EncodingType, commitmentType rpc.CommitmentType) ([]*TransactionResult, error) {
	result := make([]*TransactionResult, 0)
	for _, sign := range signatures {
		resp, err := c.conn.GetTransaction(ctx, sign, &rpc.GetTransactionOpts{
			Encoding:   encodingType,
			Commitment: commitmentType,
		})
		if err != nil {
			return nil, err
		}

		tx, _ := resp.Transaction.GetTransaction()
		result = append(result, &TransactionResult{
			Slot:      resp.Slot,
			BlockTime: resp.BlockTime,
			Transaction: &TransactionResultEnvelope{
				AsDecodedBinary:     resp.Transaction.GetData(),
				AsParsedTransaction: tx,
			},
			Meta:    resp.Meta,
			Version: resp.Version,
		})
	}
	return result, nil
}

func (c *Client) GetClusterNodes(ctx context.Context) ([]*rpc.GetClusterNodesResult, error) {
	return c.conn.GetClusterNodes(ctx)
}

func (c *Client) GetLatestBlockSlot(ctx context.Context, commitmentType rpc.CommitmentType) (uint64, error) {
	return c.conn.GetSlot(ctx, commitmentType)
}

func (c *Client) GetLatestBlock(ctx context.Context, commitmentType rpc.CommitmentType) (*rpc.GetBlockResult, error) {
	slotNum, err := c.conn.GetSlot(ctx, commitmentType)
	if err != nil {
		return nil, err
	}
	return c.GetBlockInfoBySlot(ctx, slotNum, commitmentType, nil)
}

func (c *Client) GetNeonAccountInfo() {
	// TODO implement after functionality for NeonAddress
	// look  proxy-model.py/proxy/common_neon/address.py
	panic("GetNeonAccountInfo: implement me")
}

func (c *Client) GetNeonAccountInfoList() {
	c.GetNeonAccountInfo()
}

func (c *Client) GetHolderAccountInfo() {
	// TODO implement after GetNeonAccountInfo
	panic("GetHolderAccountInfo: implement me")
}

func (c *Client) GetAccountLookupTableInfo() {
	// TODO implement after GetNeonAccountInfo
	panic("GetAccountLookupTableInfo: implement me")
}

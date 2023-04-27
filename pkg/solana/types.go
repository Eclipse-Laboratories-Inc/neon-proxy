package solana

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type BlockStatus struct {
	BlockSlot  uint64
	Commitment string
}

type TransactionResultEnvelope struct {
	AsDecodedBinary     solana.Data
	AsParsedTransaction *solana.Transaction
}

func (e *TransactionResultEnvelope) GetTransaction() (*solana.Transaction, error) {
	// TODO remove 'if' after tests
	if e.AsParsedTransaction == nil {
		return &solana.Transaction{}, nil
	}
	return e.AsParsedTransaction, nil
}

type TransactionResult struct {
	Slot uint64 `json:"slot"`

	// Estimated production time, as Unix timestamp (seconds since the Unix epoch)
	// of when the transaction was processed.
	// Nil if not available.
	BlockTime *solana.UnixTimeSeconds `json:"blockTime" bin:"optional"`

	Transaction *TransactionResultEnvelope `json:"transaction" bin:"optional"`
	Meta        *rpc.TransactionMeta       `json:"meta,omitempty" bin:"optional"`
	Version     rpc.TransactionVersion     `json:"version"`
}

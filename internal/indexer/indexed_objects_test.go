package indexer

import (
	"testing"
)

func Test_NewTxInfoKey(t *testing.T) {
	solNeonIx := SolNeonIxReceiptInfo{
		metaInfo: metaInfo{
			neonTxSig: "tx:123",
		},
	}
	//assert.Equal(t, "tx:123", NewTxInfoKey(123))
}

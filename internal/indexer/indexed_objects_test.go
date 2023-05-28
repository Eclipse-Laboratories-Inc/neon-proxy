package indexer

import (
	"testing"

	"github.com/test-go/testify/assert"
)

func Test_NewTxInfoKey(t *testing.T) {
	solNeonIx := SolNeonIxReceiptInfo{
		metaInfo: SolIxMetaInfo{
			neonTxSig: "0x:123",
		},
	}
	assert.Equal(t, ":123", NewNeonIndexedTxInfoKey(solNeonIx).value)
}

func Test_CompleteEventListInvalid(t *testing.T) {
	neonTxInfo := NeonIndexedTxInfo{
		neonReceipt: &NeonTxReceiptInfo{
			neonTxRes: NeonTxResultInfo{
				status: "0x1",
			},
		},
		neonEvents: []NeonLogTxEvent{
			{
				eventType: ExitRevert,
			},
			{
				eventType: ExitStop,
			},
			{
				eventType: EnterCreate,
			},
			{
				eventType: Log,
			},
			{
				eventType: EnterStaticCall,
			},
			{
				eventType: EnterCreate,
			},
			{
				eventType: ExitSelfDestruct,
			},
		},
	}

	neonTxInfo.CompleteEventList()
	assert.Empty(t, neonTxInfo.neonReceipt.neonTxRes.logs)
}

func Test_CompleteEventListValid(t *testing.T) {
	neonTxInfo := NeonIndexedTxInfo{
		neonReceipt: &NeonTxReceiptInfo{
			neonTxRes: NeonTxResultInfo{
				status:    "0x1",
				gasUsed:   "0x1",
				blockHash: "BlockHash",
			},
		},
		neonEvents: []NeonLogTxEvent{
			{
				eventType:    ExitRevert,
				address:      "0x123",
				data:         []byte("0x456"),
				hidden:       true,
				solSig:       "0x:789",
				idx:          4,
				innerIdx:     2,
				totalGasUsed: 1,
				reverted:     false,
			},
			{
				eventType:    ExitStop,
				address:      "0x10123",
				data:         []byte("0x10456"),
				hidden:       true,
				solSig:       "0x:10789",
				idx:          7,
				innerIdx:     5,
				totalGasUsed: 10,
				reverted:     true,
			},
		},
	}

	neonTxInfo.CompleteEventList()
	assert.Equal(t, 2, len(neonTxInfo.neonReceipt.neonTxRes.logs))
	// test first log
	log0 := neonTxInfo.neonReceipt.neonTxRes.logs[0]
	assert.Equal(t, "0x3078313233", log0["address"])
	assert.Equal(t, "0x3078343536", log0["data"])
	assert.Equal(t, "0x0", log0["neonEventLevel"])
	assert.Equal(t, "0x1", log0["neonEventOrder"])
	assert.Equal(t, "204", log0["neonEventType"])
	assert.Equal(t, "0x2", log0["neonInnerIxIdx"])
	assert.Equal(t, true, log0["neonIsHidden"])
	assert.Equal(t, false, log0["neonIsReverted"])
	assert.Equal(t, "0x4", log0["neonIxIdx"])
	assert.Equal(t, "0x:789", log0["neonSolHash"])
	assert.Empty(t, log0["topics"])

	// test second log
	log1 := neonTxInfo.neonReceipt.neonTxRes.logs[1]
	assert.Equal(t, "0x30783130313233", log1["address"])
	assert.Equal(t, "0x30783130343536", log1["data"])
	assert.Equal(t, "0x1", log1["neonEventLevel"])
	assert.Equal(t, "0x2", log1["neonEventOrder"])
	assert.Equal(t, "201", log1["neonEventType"])
	assert.Equal(t, "0x5", log1["neonInnerIxIdx"])
	assert.Equal(t, true, log1["neonIsHidden"])
	assert.Equal(t, true, log1["neonIsReverted"])
	assert.Equal(t, "0x7", log1["neonIxIdx"])
	assert.Equal(t, "0x:10789", log1["neonSolHash"])
	assert.Empty(t, log0["topics"])
}

func Test_NeonIndexedBlockInfoAddTxHolder(t *testing.T) {
	neonBlockInfo := NeonIndexedBlockInfo{
		neonHolders: make(map[string]*NeonIndexedHolderInfo),
	}

	solNeonIx := SolNeonIxReceiptInfo{
		metaInfo: SolIxMetaInfo{
			neonTxSig: "0x:123",
		},
	}

	neonBlockInfo.AddNeonTxHolder("123", solNeonIx)
	assert.Equal(t, 1, len(neonBlockInfo.neonHolders))
	assert.Equal(t, ":123", neonBlockInfo.neonHolders[":123"].key.value)
}

func Test_NeonIndexedBlockInfoAddNeonTx(t *testing.T) {
	neonBlockInfo := NeonIndexedBlockInfo{
		neonTxs: make(map[string]*NeonIndexedTxInfo),
	}

	solNeonIx := SolNeonIxReceiptInfo{
		metaInfo: SolIxMetaInfo{
			neonTxSig: "0x:123",
		},
	}

	neonBlockInfo.AddNeonTx(NeonIndexedTxTypeSingle, NeonTxInfo{}, "", []string{}, solNeonIx)
	assert.Equal(t, 1, len(neonBlockInfo.neonTxs))
	assert.Equal(t, ":123", neonBlockInfo.neonTxs[":123"].key.value)
}

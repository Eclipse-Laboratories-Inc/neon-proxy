package indexer

import (
	"strings"
	"time"
)

type IndexedObjectInfo interface {
	StartBlockSlot() int
	LastBlockSlot() int
}

type NeonIndexedHolderInfo struct {
	startBlockSlot int
	lastBlockSlot  int
}

func (n NeonIndexedHolderInfo) StartBlockSlot() int {
	return n.startBlockSlot
}

func (n NeonIndexedHolderInfo) LastBlockSlot() int {
	return n.lastBlockSlot
}

func (n *NeonIndexedHolderInfo) AddSolanaNeonIx(solanaNeonIx SolNeonIxReceiptInfo) {
	n.SetStartBlockSlot(solanaNeonIx.blockSlot)
	n.SetLastBlockSlot(solanaNeonIx.blockSlot)
}

func (n *NeonIndexedHolderInfo) SetStartBlockSlot(blockSlot int) {
	if n.startBlockSlot == 0 || blockSlot < n.startBlockSlot {
		n.startBlockSlot = blockSlot
	}
}

func (n *NeonIndexedHolderInfo) SetLastBlockSlot(blockSlot int) {
	if blockSlot > n.lastBlockSlot {
		n.lastBlockSlot = blockSlot
	}
}

type TxStatus int

const (
	InProgress TxStatus = iota
	Canceled
	Done
)

type TxType int

const (
	UnknownType TxType = iota
	Single
	SingleFromAccount
	IterFromData
	IterFromAccount
	IterFromAccountWoChainId
)

type TxInfoKey struct {
	value string
}

func NewTxInfoKey(solNeonIx SolNeonIxReceiptInfo) TxInfoKey {
	sign := solNeonIx.metaInfo.neonTxSig
	if sign[:2] == "0x" {
		sign = sign[2:]
	}
	return TxInfoKey{value: strings.ToLower(sign)}
}

func (ti *TxInfoKey) String() string {
	return ti.value
}

func (ti *TxInfoKey) IsEmpty() bool {
	return ti.value == ""
}

type NeonIndexedTxInfo struct {
	startBlockSlot int
	lastBlockSlot  int

	key             TxInfoKey
	neonReceipt     NeonTxReceiptInfo
	txType          TxType
	storageAccount  string
	blockedAccounts []string
	status          TxStatus
	canceled        bool
	neonEvents      []NeonLogTxEvent
}

func (n NeonIndexedTxInfo) StartBlockSlot() int {
	return n.startBlockSlot
}

func (n NeonIndexedTxInfo) LastBlockSlot() int {
	return n.lastBlockSlot
}

type NeonTxReceiptInfo struct {
	neonTx    NeonTxInfo
	neonTxRes NeonTxResultInfo
}

type NeonTxInfo struct {
	addr     string
	sig      string
	nonce    string
	gasPrice string
	gasLimit string
	toAddr   string
	contract string
	value    string
	callData string
	v        string
	r        string
	s        string
	err      error
}

type NeonTxResultInfo struct {
	blockSlot int
	blockHash string
	txIdx     int

	solSig        string
	solIxIdx      int
	solIxInnerIdx int

	neonSig string
	gasUsed string
	status  string

	logs []map[string]string

	canceledStatus int
	lostStatus     int
}

type NeonAccountInfo struct {
	neonAddress string
	pdaAddress  string
	blockSlot   int
	code        string
	solSig      string
}

type LogTxEventType int

const (
	Log LogTxEventType = 1

	EnterCall         LogTxEventType = 101
	EnterCallCode     LogTxEventType = 102
	EnterStaticCall   LogTxEventType = 103
	EnterDelegateCall LogTxEventType = 104
	EnterCreate       LogTxEventType = 105
	EnterCreate2      LogTxEventType = 106

	ExitStop         LogTxEventType = 201
	ExitReturn       LogTxEventType = 202
	ExitSelfDestruct LogTxEventType = 203
	ExitRevert       LogTxEventType = 204

	Return LogTxEventType = 300
	Cancel LogTxEventType = 301
)

type NeonLogTxEvent struct {
	eventType LogTxEventType
	Hidden    bool

	address string
	topics  []string
	data    string

	solSig       string
	idx          int
	innerIdx     int
	totalGasUsed int
	reverted     bool
	eventLevel   int
	eventOrder   int
}

type NeonIndexedBlockInfo struct {
	solBlocks         []SolBlockInfo
	historyBlockDeque []SolBlockInfo
	completed         bool

	neonHolders map[string]NeonIndexedHolderInfo
	neonTxs     map[string]NeonIndexedTxInfo

	doneNeonTxs []NeonIndexedTxInfo

	solNeonIxs []SolNeonIxReceiptInfo
	solTxCosts []SolTxCostInfo

	StatNeonTxs map[TxType]NeonTxStatData
}

type SolBlockInfo struct{}       // todo implemented
type NeonTxStatData struct{}     // todo implemented
type SolTxMetaCollector struct{} // todo implemented

type IndexedBlockStat struct {
	neonBlockCnt    int
	neonHolderCnt   int
	neonTxCnt       int
	historyBlockCnt int
	solNeonIxCnt    int
	minBlockSlot    int
}

type NeonIndexedBlockMap struct {
	neonBlockMap       map[int]NeonIndexedBlockInfo
	finalizedNeonBlock NeonIndexedBlockInfo
	stat               IndexedBlockStat
}

type NeonIndexedBlockData struct {
	neonIndexedBlockInfo NeonIndexedBlockInfo
	finalized            bool
}

type SolNeonTxDecoderState struct {
	startTime          time.Time
	initBlockSlot      int
	startBlockSlot     int
	stopBlockSlot      int
	solTxMetaCnt       int
	solNeonIxCnt       int
	solTxMetaCollector SolTxMetaCollector

	solTx     SolTxReceiptInfo
	solTxMeta SolTxMetaInfo
	solNeonIx SolNeonIxReceiptInfo

	neonBlockDeque []NeonIndexedBlockData
}

package indexer

import (
	"encoding/hex"
	"fmt"
	utils2 "github.com/neonlabsorg/neon-proxy/pkg/utils"
	"log"
	"math"
	"sort"
	"strconv"
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
	key            TxInfoKey
	dataSize       int
	data           []byte
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

func (n *NeonIndexedHolderInfo) NeonTxSig() string {
	return n.key.neonTxSig
}

func (n *NeonIndexedHolderInfo) Account() string {
	return n.key.account
}

func (n *NeonIndexedHolderInfo) AddDataChank(chunk TxInfoDataChunk) {
	end := chunk.offset + chunk.lenght
	dataLen := len(n.data)
	if end > dataLen {
		n.data = append(n.data, make([]byte, end-dataLen)...)
	}

	n.data = append(n.data[:chunk.offset], chunk.data...)
	n.data = append(n.data[:chunk.offset+chunk.lenght], n.data[end:]...)
	n.dataSize += chunk.lenght
}

type TxStatus int

const (
	NeonIndexedTxInfoStatusInProgress TxStatus = iota
	NeonIndexedTxInfoStatusCanceled
	NeonIndexedTxInfoStatusDone
)

type NeonIndexedTxType int

const (
	NeonIndexedTxTypeUnknown                  NeonIndexedTxType = 0
	NeonIndexedTxTypeSingle                   NeonIndexedTxType = 0x1f
	NeonIndexedTxTypeSingleFromAccount        NeonIndexedTxType = 0x2a
	NeonIndexedTxTypeIterFromData             NeonIndexedTxType = 0x20
	NeonIndexedTxTypeIterFromAccount          NeonIndexedTxType = 0x21
	NeonIndexedTxTypeIterFromAccountWoChainId NeonIndexedTxType = 0x22
)

type TxInfoKey struct {
	value     string
	neonTxSig string
	account   string
}

func NewNeonIndexedTxInfoKey(solNeonIx SolNeonIxReceiptInfo) TxInfoKey {
	sign := solNeonIx.metaInfo.neonTxSig
	if sign[:2] == "0x" {
		sign = sign[2:]
	}
	return TxInfoKey{value: strings.ToLower(sign)}
}

func NewNeonIndexedHolderInfoKey(account, neonSig string) TxInfoKey {
	sign := neonSig
	if sign[:2] == "0x" {
		sign = sign[2:]
	}
	sign = strings.ToLower(sign)
	return TxInfoKey{
		value:     fmt.Sprintf("%v:%v", account, sign),
		neonTxSig: sign,
		account:   account,
	}
}

func (ti *TxInfoKey) String() string {
	return ti.value
}

func (ti *TxInfoKey) IsEmpty() bool {
	return ti.value == ""
}

type TxInfoDataChunk struct {
	offset int
	lenght int
	data   []byte
}

// TODO implement
func (t *TxInfoDataChunk) String() string {
	return ""
}

func (t *TxInfoDataChunk) IsValid() bool {
	return t.lenght > 0 && len(t.data) == t.lenght
}

type NeonIndexedTxInfo struct {
	startBlockSlot int
	lastBlockSlot  int

	key             TxInfoKey
	neonReceipt     *NeonTxReceiptInfo
	txType          NeonIndexedTxType
	storageAccount  string
	blockedAccounts []string
	status          TxStatus
	canceled        bool
	neonEvents      []NeonLogTxEvent
}

func NewNeonIndexedTxInfo(txType NeonIndexedTxType, key TxInfoKey, neonTx NeonTxInfo, holderAccount string, blockedAccounts []string) *NeonIndexedTxInfo {
	if key.IsEmpty() {
		panic("TxInfoKey info key can't be empty")
	}
	return &NeonIndexedTxInfo{
		key:             key,
		neonReceipt:     NewNeonTxReceiptInfo(neonTx, NeonTxResultInfo{}),
		txType:          txType,
		storageAccount:  holderAccount,
		blockedAccounts: blockedAccounts,
		status:          NeonIndexedTxInfoStatusInProgress,
	}
}

func (n *NeonIndexedTxInfo) StartBlockSlot() int {
	return n.startBlockSlot
}

func (n *NeonIndexedTxInfo) LastBlockSlot() int {
	return n.lastBlockSlot
}

func (n *NeonIndexedTxInfo) IsCanceled() bool {
	return n.canceled
}

func (n *NeonIndexedTxInfo) LenNeonEventList() int {
	return len(n.neonEvents)
}

func (n *NeonIndexedTxInfo) SetStatus(value TxStatus, blockSlot int) {
	n.status = value
	n.SetLastBlockSlot(blockSlot)
}

func (n *NeonIndexedTxInfo) SetStartBlockSlot(blockSlot int) {
	if n.startBlockSlot == 0 || blockSlot < n.startBlockSlot {
		n.startBlockSlot = blockSlot
	}
}

func (n *NeonIndexedTxInfo) SetLastBlockSlot(blockSlot int) {
	if blockSlot > n.lastBlockSlot {
		n.lastBlockSlot = blockSlot
	}
}

func (n *NeonIndexedTxInfo) AddSolanaNeonIx(solanaNeonIx SolNeonIxReceiptInfo) {
	n.SetStartBlockSlot(solanaNeonIx.blockSlot)
	n.SetLastBlockSlot(solanaNeonIx.blockSlot)
}

func (n *NeonIndexedTxInfo) SetNeonTx(neonTx NeonTxInfo, holder NeonIndexedHolderInfo) {
	n.neonReceipt.neonTx = neonTx
	n.SetStartBlockSlot(holder.startBlockSlot)
	n.SetLastBlockSlot(holder.lastBlockSlot)
}

func (n *NeonIndexedTxInfo) AddNeonEvent(event NeonLogTxEvent) {
	n.neonEvents = append(n.neonEvents, event)
}

func (n *NeonIndexedTxInfo) SortNeonIndexedList() {
	if len(n.neonEvents) > 0 {
		sort.Sort(SortNeonEventList(n.neonEvents))
	}
}

func (n *NeonIndexedTxInfo) CompleteEventList() {
	eventsLen := len(n.neonEvents)
	if !n.neonReceipt.neonTxRes.IsValid() || len(n.neonReceipt.neonTxRes.logs) > 0 || eventsLen == 0 {
		return
	}
	neonEventList := make([]NeonLogTxEvent, 0, eventsLen)
	curLevel := 1
	revertedLevel := -1
	curOrder := eventsLen
	isFailed := (n.neonReceipt.neonTxRes.status == "0x0")
	var isReverted, isHidden bool

	n.SortNeonIndexedList()
	for _, event := range n.neonEvents {
		if event.reverted {
			isReverted = true
			isHidden = true
		} else {
			if event.isStartEventType() {
				curLevel--
				if (revertedLevel != -1) && (curLevel < revertedLevel) {
					revertedLevel = -1
				}
			} else if event.isExitEventType() {
				curLevel += 1
				if (event.eventType == ExitRevert) && (revertedLevel == -1) {
					revertedLevel = curLevel
				}
			}
			isReverted = (revertedLevel != -1) || isFailed
			isHidden = event.Hidden || isReverted
		}
		neonLogEvent := event.DeepCopy()
		neonLogEvent.Hidden = isHidden
		neonLogEvent.reverted = isReverted
		neonLogEvent.eventLevel = curLevel
		neonLogEvent.eventOrder = curOrder
		neonEventList = append(neonEventList, neonLogEvent)
		curOrder--
	}

	for i := len(neonEventList) - 1; i >= 0; i-- {
		n.neonReceipt.neonTxRes.AddEvent(neonEventList[i])
	}
}

type NeonTxReceiptInfo struct {
	neonTx    NeonTxInfo
	neonTxRes NeonTxResultInfo
}

func NewNeonTxReceiptInfo(neonTx NeonTxInfo, neonTxRes NeonTxResultInfo) *NeonTxReceiptInfo {
	return &NeonTxReceiptInfo{
		neonTx:    neonTx,
		neonTxRes: neonTxRes,
	}
}

type NeonTxInfo struct {
	addr     *string
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

func NewNeonTxInfoFromNeonTx(tx *utils2.NeonTx) *NeonTxInfo {
	addr := tx.HexSender()

	return &NeonTxInfo{
		addr:     &addr,
		sig:      tx.HexTxSig(),
		nonce:    hex.EncodeToString(tx.Nonce.Bytes()),
		gasPrice: hex.EncodeToString(tx.GasPrice.Bytes()),
		gasLimit: hex.EncodeToString(tx.GasLimit.Bytes()),
		toAddr:   tx.HexToAddress(),
		contract: tx.HexContract(),
		value:    hex.EncodeToString(tx.Value.Bytes()),
		callData: tx.HexCallData(),
		v:        hex.EncodeToString(tx.V.Bytes()),
		r:        hex.EncodeToString(tx.R.Bytes()),
		s:        hex.EncodeToString(tx.S.Bytes()),
	}
}

func (ntx *NeonTxInfo) IsValid() bool {
	return ntx.addr != nil && ntx.err == nil
}

// TODO remove stab
func NewNeonTxFromSigData(rlpSigData []byte) *NeonTxInfo {
	tx, err := utils2.NewNeonTxFromString(rlpSigData) // TODO uncomment after tests
	if err != nil {
		return &NeonTxInfo{err: err}
	}

	return NewNeonTxInfoFromNeonTx(tx)

	/*	addr := ""
		return &NeonTxInfo{sig: "54686520717569636b2062726f776e20666f78206a756d7073206f7665722074", addr: &addr}*/
}

func NewNeonTxFromNeonSig(sig string) *NeonTxInfo {
	return &NeonTxInfo{sig: sig}
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

	logs []map[string][]string

	completed bool
	canceled  bool
}

func (n *NeonTxResultInfo) IsValid() bool {
	return n.gasUsed != ""
}

func (n *NeonTxResultInfo) IsCompleted() bool {
	return n.completed
}

func (n *NeonTxResultInfo) IsCanceled() bool {
	return n.canceled
}

func (n *NeonTxResultInfo) AddEvent(event NeonLogTxEvent) {
	if n.blockSlot != 0 {
		return
	}

	curLog := make(map[string][]string)

	var addr string
	if event.address != "" {
		addr = "0x" + hex.EncodeToString([]byte(event.address))
	}
	curLog["address"] = []string{addr}

	var topics []string
	for _, topic := range event.topics {
		topics = append(topics, "0x"+hex.EncodeToString([]byte(topic)))
	}
	curLog["topics"] = topics

	var data string
	if event.data != nil {
		data = "0x" + hex.EncodeToString(event.data)
	}
	curLog["data"] = []string{data}

	curLog["neonSolHash"] = []string{event.solSig}
	curLog["neonIxIdx"] = []string{fmt.Sprintf("%x", event.idx)}
	curLog["neonInnerIxIdx"] = []string{fmt.Sprintf("%x", event.innerIdx)}
	curLog["neonEventType"] = []string{fmt.Sprintf("%v", event.eventType)}
	curLog["neonEventLevel"] = []string{fmt.Sprintf("%x", event.eventLevel)}
	curLog["neonEventOrder"] = []string{fmt.Sprintf("%x", event.eventOrder)}

	hidden := "false"
	if event.Hidden {
		hidden = "true"
	}
	curLog["neonIsHidden"] = []string{hidden}

	reverted := "false"
	if event.reverted {
		reverted = "true"
	}
	curLog["neonIsReverted"] = []string{reverted}

	n.logs = append(n.logs, curLog)
}

func (n *NeonTxResultInfo) SetResult(status, gasUsed int64) {
	n.status = fmt.Sprintf("%x", status)
	n.gasUsed = fmt.Sprintf("%x", gasUsed)
	n.completed = true
}

func (n *NeonTxResultInfo) SetLostResult(gasUsed int64) {
	n.status = fmt.Sprintf("%x", 0)
	n.gasUsed = fmt.Sprintf("%x", gasUsed)
	n.completed = false
}

func (n *NeonTxResultInfo) SetSolSigInfo(solSig string, idx, innerIdx int) {
	n.solSig = solSig
	n.solIxIdx = idx
	n.solIxInnerIdx = innerIdx
}

func (n *NeonTxResultInfo) SetCanceledResult(neonTotalGasUsed int64) {
	n.SetResult(0, neonTotalGasUsed)
	n.canceled = true
}

func (n *NeonTxResultInfo) SetBlockInfo(block SolBlockInfo, neonSig string, txIdx int, logIdx int) int {
	n.blockSlot = block.BlockSlot
	n.blockHash = block.BlockHash
	n.solSig = neonSig
	n.txIdx = txIdx

	hexBlockSlot := fmt.Sprintf("0x%x", n.blockSlot)
	hexTxIdx := fmt.Sprintf("0x%x", n.txIdx)
	txLogIdx := 0

	for _, rec := range n.logs {
		rec["transactionHash"] = []string{n.solSig}
		rec["blockHash"] = []string{n.blockHash}
		rec["blockNumber"] = []string{hexBlockSlot}
		rec["transactionIndex"] = []string{hexTxIdx}
		if _, ok := rec["neonIsHidden"]; !ok {
			rec["logIndex"] = []string{fmt.Sprintf("0x%x", logIdx)}
			rec["transactionLogIndex"] = []string{fmt.Sprintf("0x%x", txLogIdx)}
			logIdx += 1
			txLogIdx += 1
		}
	}

	return logIdx
}

type NeonAccountInfo struct {
	neonAddress string
	pdaAddress  string
	blockSlot   int
	code        string
	solSig      string
}

type NeonIndexedBlockInfo struct {
	solBlock          SolBlockInfo
	historyBlockDeque []SolBlockInfo
	completed         bool

	neonHolders map[string]NeonIndexedHolderInfo
	neonTxs     map[string]*NeonIndexedTxInfo

	doneNeonTxs []*NeonIndexedTxInfo

	solNeonIxs []SolNeonIxReceiptInfo
	solTxCosts []SolTxCostInfo

	StatNeonTxs map[NeonIndexedTxType]NeonTxStatData
}

func NewNeonIndexedBlockInfo(historyBlockDeque []SolBlockInfo) *NeonIndexedBlockInfo {
	return &NeonIndexedBlockInfo{
		solBlock:          historyBlockDeque[len(historyBlockDeque)-1],
		historyBlockDeque: historyBlockDeque,
		neonHolders:       make(map[string]NeonIndexedHolderInfo),
		neonTxs:           make(map[string]*NeonIndexedTxInfo),
		StatNeonTxs:       make(map[NeonIndexedTxType]NeonTxStatData),
	}
}

func (n *NeonIndexedBlockInfo) Clone(historyBlockDeque []SolBlockInfo) *NeonIndexedBlockInfo {
	solBlock := historyBlockDeque[len(historyBlockDeque)-1]
	if solBlock.BlockSlot <= n.solBlock.BlockSlot {
		panic("Clone:NeonIndexedBlockInfo: solBlock.BlockSlot <= n.solBlock.BlockSlot")
	}

	clone := NewNeonIndexedBlockInfo(n.historyBlockDeque)
	// deep copy neonHolders
	for k, v := range n.neonHolders {
		clone.neonHolders[k] = v
	}

	// deep copy neonTxs
	for k, v := range n.neonTxs {
		clone.neonTxs[k] = v
	}

	return clone
}

func (n *NeonIndexedBlockInfo) SetFinalized(value bool) {
	for _, block := range n.historyBlockDeque {
		block.SetFinalized(value)
	}
}

func (n *NeonIndexedBlockInfo) FinalizeHistoryList(finalizedBlockSlot int) int {
	removedCount := 0
	for len(n.historyBlockDeque) > 0 && finalizedBlockSlot >= n.historyBlockDeque[0].BlockSlot {
		n.historyBlockDeque = n.historyBlockDeque[1:]
		removedCount++
	}
	if len(n.historyBlockDeque) == 0 {
		panic("FinalizeHistoryList: len(n.historyBlockDeque) == 0")
	}
	return removedCount
}

func (n *NeonIndexedBlockInfo) AddSolNeonIx(solNeonIx SolNeonIxReceiptInfo) {
	n.solNeonIxs = append(n.solNeonIxs, solNeonIx)
}

func (n *NeonIndexedBlockInfo) AssSolTxCost(txCost SolTxCostInfo) {
	n.solTxCosts = append(n.solTxCosts, txCost)
}

func (n *NeonIndexedBlockInfo) FindNeonTxHolder(account string, solNeonIx SolNeonIxReceiptInfo) *NeonIndexedHolderInfo {
	key := NewNeonIndexedHolderInfoKey(account, solNeonIx.solSign)
	holder, ok := n.neonHolders[key.value]
	if ok {
		holder.AddSolanaNeonIx(solNeonIx)
		return &holder
	}
	return nil
}

func (n *NeonIndexedBlockInfo) AddNeonTxHolder(account string, solNeonIx SolNeonIxReceiptInfo) *NeonIndexedHolderInfo {
	key := NewNeonIndexedHolderInfoKey(account, solNeonIx.solSign)
	_, ok := n.neonHolders[key.value]
	if ok {
		panic("AddNeonTxHolder: holder already in use!")
	}

	holder := NeonIndexedHolderInfo{
		key: key,
	}
	holder.AddSolanaNeonIx(solNeonIx)
	n.neonHolders[key.value] = holder
	return &holder
}

func (n *NeonIndexedBlockInfo) AddNeonAccount(info NeonAccountInfo, ix SolNeonIxReceiptInfo) {}

func (n *NeonIndexedBlockInfo) DeleteNeonHolder(holder NeonIndexedHolderInfo) {
	delete(n.neonHolders, holder.key.value)
}

func (n *NeonIndexedBlockInfo) FailNeonHolder(holder NeonIndexedHolderInfo) {
	n.DeleteNeonHolder(holder)
}

func (n *NeonIndexedBlockInfo) DoneNeonHolder(holder NeonIndexedHolderInfo) {
	n.DeleteNeonHolder(holder)
}

func (n *NeonIndexedBlockInfo) FindNeonTx(solNeonIx SolNeonIxReceiptInfo) *NeonIndexedTxInfo {
	key := NewNeonIndexedTxInfoKey(solNeonIx)
	tx, ok := n.neonTxs[key.value]
	if ok {
		tx.AddSolanaNeonIx(solNeonIx)
		return tx
	}
	return nil
}

func (n *NeonIndexedBlockInfo) AddNeonTx(
	txType NeonIndexedTxType,
	neonTx NeonTxInfo,
	holderAccount string, blockedAccounts []string,
	solNeonIx SolNeonIxReceiptInfo) *NeonIndexedTxInfo {
	key := NewNeonIndexedTxInfoKey(solNeonIx)

	_, ok := n.neonTxs[key.value]
	if ok {
		panicMsg := fmt.Sprintf("AddNeonTx: %s tx already in use!", key.value)
		panic(panicMsg)
	}

	tx := NewNeonIndexedTxInfo(txType, key, neonTx, holderAccount, blockedAccounts)
	tx.AddSolanaNeonIx(solNeonIx)
	n.neonTxs[key.value] = tx
	return tx
}

func (n *NeonIndexedBlockInfo) DeleteNeonTx(tx *NeonIndexedTxInfo) {
	delete(n.neonTxs, tx.key.value)
}

func (n *NeonIndexedBlockInfo) FailNeonTx(tx *NeonIndexedTxInfo) {
	if tx.status != NeonIndexedTxInfoStatusInProgress && tx.status != NeonIndexedTxInfoStatusCanceled {
		panic("FailNeonTx: attempt to fail the completed tx") // change warning ?
	}
	n.DeleteNeonTx(tx)
}

func (n *NeonIndexedBlockInfo) DoneNeonTx(tx *NeonIndexedTxInfo, solNeonIx *SolNeonIxReceiptInfo) {
	if tx.status != NeonIndexedTxInfoStatusInProgress && tx.status != NeonIndexedTxInfoStatusCanceled {
		panic("DoneNeonTx: attempt to done the completed tx") // change warning ?
	}

	tx.SetStatus(NeonIndexedTxInfoStatusDone, solNeonIx.blockSlot)
	n.doneNeonTxs = append(n.doneNeonTxs, tx)
}

func (n *NeonIndexedBlockInfo) HistoryBlockCount() int {
	return len(n.historyBlockDeque)
}

func (n *NeonIndexedBlockInfo) GetNeonTxCount() int {
	return len(n.neonTxs)
}

func (n *NeonIndexedBlockInfo) GetNeonHolderCount() int {
	return len(n.neonHolders)
}

func (n *NeonIndexedBlockInfo) GetSolNeonIxCount() int {
	return len(n.solNeonIxs)
}

func (n *NeonIndexedBlockInfo) GetSolTxCostCount() int {
	return len(n.solTxCosts)
}

func (n *NeonIndexedBlockInfo) GetNeonTxCountByType(txType NeonIndexedTxType) int {
	count := 0
	for _, tx := range n.neonTxs {
		if tx.txType == txType {
			count++
		}
	}
	return count
}

func (n *NeonIndexedBlockInfo) CalculateStat(gatherStatistics bool, opAccountSet map[string]bool) {
	// todo(argishti) remove gatherStatistics param after calling this function from Indexer
	if !gatherStatistics {
		return
	}

	for _, solNeonIx := range n.solNeonIxs {
		txType := NeonIndexedTxTypeUnknown
		tx, trTypeFound := n.neonTxs[NewNeonIndexedTxInfoKey(solNeonIx).value]
		if trTypeFound {
			txType = tx.txType
		}
		isOpSolNeonIx := opAccountSet[solNeonIx.solTxCost.operator]

		stat, statFound := n.StatNeonTxs[txType]
		if !statFound {
			stat = *NewNeonTxStatData(txType)
		}

		neonIncome := int64(0)
		if trTypeFound && strings.HasPrefix(tx.neonReceipt.neonTx.gasPrice, "0x") {
			decimalNum, err := strconv.ParseInt(tx.neonReceipt.neonTx.gasPrice[2:], 16, 64)
			if err != nil {
				log.Fatal(err)
			}
			neonIncome = solNeonIx.metaInfo.neonGasUsed * decimalNum
		}

		solSpent := 0
		if !solNeonIx.solTxCost.calculatedStat {
			solSpent = solNeonIx.solTxCost.solSpent
			solNeonIx.solTxCost.calculatedStat = true
			stat.solSpent += solSpent
			stat.solTxCnt++
		}

		stat.neonIncome += neonIncome
		stat.neonStepCnt += solNeonIx.neonStepCnt
		stat.bpfCycleCnt += solNeonIx.metaInfo.usedBpfCycleCnt

		if isOpSolNeonIx {
			stat.opSolSpent += solSpent
			stat.opNeonIncome += neonIncome
		}

		if solNeonIx.metaInfo.neonTxReturn != nil {
			if solNeonIx.metaInfo.neonTxReturn.Canceled {
				stat.canceledNeonTxCnt++
				if isOpSolNeonIx {
					stat.opCanceledNeonTxCnt++
				}
			} else {
				stat.completedNeonTxCnt++
				if isOpSolNeonIx {
					stat.opCompletedNeonTxCnt++
				}
			}
		}
		// update stats by tx type
		n.StatNeonTxs[txType] = stat
	}
}

func (n *NeonIndexedBlockInfo) FillLogInfo() {
	logIdx := 0
	txIdx := 0
	for _, tx := range n.doneNeonTxs {
		tx.CompleteEventList()
		logIdx = tx.neonReceipt.neonTxRes.SetBlockInfo(n.solBlock, tx.neonReceipt.neonTx.sig, txIdx, logIdx)
		txIdx++
	}
}

func (n *NeonIndexedBlockInfo) CompleteBlock(skipCancelTimeout int, holdertimeout int) {
	for _, tx := range n.doneNeonTxs {
		n.DeleteNeonTx(tx)
	}

	n.completed = true
	// clear slices keeping alocated memory
	n.doneNeonTxs = n.doneNeonTxs[:0]
	n.solTxCosts = n.solTxCosts[:0]
	n.solNeonIxs = n.solNeonIxs[:0]

	for _, tx := range n.neonTxs {
		if math.Abs(float64(n.solBlock.BlockSlot-tx.lastBlockSlot)) > float64(skipCancelTimeout) {
			//log.Debug(fmt.Sprintf("skip to cancel %s", &tx.key))
			n.FailNeonTx(tx)
		}
	}

	for _, holder := range n.neonHolders {
		if math.Abs(float64(n.solBlock.BlockSlot-holder.lastBlockSlot)) > float64(holdertimeout) {
			//log.Debug(fmt.Sprintf("skip the neon holder %s", holder))
			n.FailNeonHolder(holder)
		}
	}
}

type SolBlockInfo struct {
	BlockSlot   int
	BlockHash   string
	BlockTime   int64
	BlockHeight int64

	ParentBlockSlot int
	ParentBlockHash string
	Finalized       bool
}

func (s *SolBlockInfo) SetFinalized(value bool) {
	s.Finalized = value
}

func (s *SolBlockInfo) SetBlockHash(hash string) {
	s.BlockHash = hash
}

type NeonTxStatData struct {
	txType             string
	completedNeonTxCnt int
	canceledNeonTxCnt  int
	solTxCnt           int
	solSpent           int
	neonIncome         int64
	neonStepCnt        int
	bpfCycleCnt        int

	opSolSpent           int
	opNeonIncome         int64
	opCompletedNeonTxCnt int
	opCanceledNeonTxCnt  int
}

func NewNeonTxStatData(txType NeonIndexedTxType) *NeonTxStatData {
	var typeName string
	switch txType {
	case NeonIndexedTxTypeSingle:
		typeName = "single"
	case NeonIndexedTxTypeSingleFromAccount:
		typeName = "single-holder"
	case NeonIndexedTxTypeIterFromData:
		typeName = "iterative"
	case NeonIndexedTxTypeIterFromAccount:
		typeName = "holder"
	case NeonIndexedTxTypeIterFromAccountWoChainId:
		typeName = "wochainid"
	default:
		typeName = "other"
	}
	return &NeonTxStatData{
		txType: typeName,
	}
}

type IndexedBlockStat struct {
	neonBlockCnt    int
	neonHolderCnt   int
	neonTxCnt       int
	historyBlockCnt int
	solNeonIxCnt    int
	minBlockSlot    int
}

func NewIndexedBlockStat(neonHolderCnt int, neonTxCnt int, historyBlockCnt int, solNeonIxCnt int) *IndexedBlockStat {
	return &IndexedBlockStat{
		neonBlockCnt:    1,
		neonHolderCnt:   neonHolderCnt,
		neonTxCnt:       neonTxCnt,
		historyBlockCnt: historyBlockCnt,
		solNeonIxCnt:    solNeonIxCnt,
		minBlockSlot:    0,
	}
}

func NewEmptyIndexedBlockStat() *IndexedBlockStat {
	return NewIndexedBlockStat(0, 0, 0, 0)
}

func NewIndexedBlockStatFromNeonBlock(neonBlock NeonIndexedBlockInfo) *IndexedBlockStat {
	return NewIndexedBlockStat(neonBlock.GetNeonHolderCount(),
		neonBlock.GetNeonTxCount(),
		neonBlock.HistoryBlockCount(),
		neonBlock.GetSolNeonIxCount())
}

func (n *IndexedBlockStat) DecreaseHistoryBlockCount(removedBlockCount int) {
	n.historyBlockCnt -= removedBlockCount
}

func (n *IndexedBlockStat) AddStat(src *IndexedBlockStat) {
	n.neonBlockCnt += src.neonBlockCnt
	n.neonHolderCnt += src.neonHolderCnt
	n.neonTxCnt += src.neonTxCnt
	n.historyBlockCnt += src.historyBlockCnt
	n.solNeonIxCnt += src.solNeonIxCnt
}

func (n *IndexedBlockStat) DelStat(src *IndexedBlockStat) {
	n.neonBlockCnt -= src.neonBlockCnt
	n.neonHolderCnt -= src.neonHolderCnt
	n.neonTxCnt -= src.neonTxCnt
	n.historyBlockCnt -= src.historyBlockCnt
	n.solNeonIxCnt -= src.solNeonIxCnt
}

type NeonIndexedBlockMap struct {
	neonBlockMap       map[int]NeonIndexedBlockInfo
	finalizedNeonBlock *NeonIndexedBlockInfo
	stat               IndexedBlockStat
}

func NewNeonIndexedBlockMap() *NeonIndexedBlockMap {
	return &NeonIndexedBlockMap{
		neonBlockMap:       make(map[int]NeonIndexedBlockInfo),
		finalizedNeonBlock: nil,
		stat:               *NewEmptyIndexedBlockStat(),
	}
}

func (n *NeonIndexedBlockMap) GetNeonBlock(blockSlot int) (NeonIndexedBlockInfo, bool) {
	neonBlock, ok := n.neonBlockMap[blockSlot]
	if ok && n.finalizedNeonBlock != nil {
		removedBlockCnt := neonBlock.FinalizeHistoryList(n.finalizedNeonBlock.solBlock.BlockSlot)
		n.stat.DecreaseHistoryBlockCount(removedBlockCnt)
	}
	return neonBlock, ok
}

func FindMinBlockSlot(neonBlock NeonIndexedBlockInfo) int {
	minBlockSlot := neonBlock.solBlock.BlockSlot
	for _, holder := range neonBlock.neonHolders {
		if holder.StartBlockSlot() < minBlockSlot {
			minBlockSlot = holder.StartBlockSlot()
		}
	}
	return minBlockSlot
}

func (n *NeonIndexedBlockMap) AddNeonBlock(neonBlock NeonIndexedBlockInfo) {
	if _, ok := n.neonBlockMap[neonBlock.solBlock.BlockSlot]; ok {
		return
	}

	stat := NewIndexedBlockStatFromNeonBlock(neonBlock)
	n.stat.AddStat(stat)
	n.neonBlockMap[neonBlock.solBlock.BlockSlot] = neonBlock
}

func (n *NeonIndexedBlockMap) FinalizeNeonBlock(neonBlock NeonIndexedBlockInfo) {
	if _, ok := n.neonBlockMap[neonBlock.solBlock.BlockSlot]; !ok {
		return
	}

	if n.finalizedNeonBlock != nil {
		for blockSlot := n.finalizedNeonBlock.solBlock.BlockSlot; blockSlot < neonBlock.solBlock.BlockSlot; blockSlot++ {
			oldNeonBlock, ok := n.neonBlockMap[blockSlot]
			if ok {
				stat := NewIndexedBlockStatFromNeonBlock(oldNeonBlock)
				n.stat.DelStat(stat)
				delete(n.neonBlockMap, blockSlot)
			}
		}
	}

	n.finalizedNeonBlock = &neonBlock
	n.stat.minBlockSlot = FindMinBlockSlot(neonBlock)
}

type NeonIndexedBlockData struct {
	neonIndexedBlockInfo *NeonIndexedBlockInfo
	finalized            bool
}

type SolNeonTxDecoderState struct {
	startTime          time.Time
	initBlockSlot      uint64
	startBlockSlot     uint64
	stopBlockSlot      uint64
	solTxMetaCnt       int
	solNeonIxCnt       int
	solTxMetaCollector CollectorInterface // todo(argishti) possible to use pointer

	solTx     *SolTxReceiptInfo
	solTxMeta *SolTxMetaInfo
	solNeonIx *SolNeonIxReceiptInfo

	neonBlockDeque []NeonIndexedBlockData
}

func NewSolNeonTxDecoderState(solTxMetaCollector CollectorInterface, startBlockSlot uint64, neonBlock *NeonIndexedBlockInfo) *SolNeonTxDecoderState {
	state := SolNeonTxDecoderState{
		startTime:          time.Now(),
		initBlockSlot:      startBlockSlot,
		startBlockSlot:     startBlockSlot,
		stopBlockSlot:      startBlockSlot,
		solTxMetaCollector: solTxMetaCollector,
		neonBlockDeque:     make([]NeonIndexedBlockData, 0),
	}
	if neonBlock != nil {
		state.SetNeonBlock(neonBlock)
	}
	return &state
}

func (s *SolNeonTxDecoderState) SetNeonBlock(neonBlock *NeonIndexedBlockInfo) {
	if len(s.neonBlockDeque) > 0 && s.neonBlockDeque[0].finalized {
		s.neonBlockDeque = s.neonBlockDeque[1:]
	}
	isFinalized := s.solTxMetaCollector.IsFinalized()
	s.neonBlockDeque = append(s.neonBlockDeque, NeonIndexedBlockData{neonBlock, isFinalized})
}

func (s *SolNeonTxDecoderState) ShifttoCollector(collector CollectorInterface) {
	s.startBlockSlot = s.stopBlockSlot + 1
	s.stopBlockSlot = s.startBlockSlot
	s.solTxMetaCollector = collector
}

func (s *SolNeonTxDecoderState) ProcessTimeMs() float64 {
	return time.Since(s.startTime).Seconds() * 1000
}

func (s *SolNeonTxDecoderState) NeonBlockCount() int {
	return len(s.neonBlockDeque)
}

func (s *SolNeonTxDecoderState) HasNeonBlock() bool {
	return s.NeonBlockCount() > 0
}

func (s *SolNeonTxDecoderState) NeonBlock() *NeonIndexedBlockInfo {
	if !s.HasNeonBlock() {
		panic("SolNeonTxDecoderState: No Neon Block")
	}
	return s.neonBlockDeque[len(s.neonBlockDeque)-1].neonIndexedBlockInfo
}

func (s *SolNeonTxDecoderState) IsNeonBlockFinalized() bool {
	if !s.HasNeonBlock() {
		panic("SolNeonTxDecoderState: No Neon Block")
	}
	return s.neonBlockDeque[len(s.neonBlockDeque)-1].finalized
}

func (s *SolNeonTxDecoderState) HasSolTx() bool {
	return s.solTx != nil
}

func (s *SolNeonTxDecoderState) SolTx() *SolTxReceiptInfo {
	if !s.HasSolTx() {
		panic("SolNeonTxDecoderState: No Sol Tx")
	}
	return s.solTx
}

func (s *SolNeonTxDecoderState) HasSolTxMeta() bool {
	return s.solTxMeta != nil
}

func (s *SolNeonTxDecoderState) SolTxMeta() *SolTxMetaInfo {
	if !s.HasSolTxMeta() {
		panic("SolNeonTxDecoderState: No Sol Tx Meta")
	}
	return s.solTxMeta
}

func (s *SolNeonTxDecoderState) HasSolNeonIx() bool {
	return s.solNeonIx != nil
}

func (s *SolNeonTxDecoderState) SolNeonIx() *SolNeonIxReceiptInfo {
	if !s.HasSolNeonIx() {
		panic("SolNeonTxDecoderState: No Sol Neon Ix")
	}
	return s.solNeonIx
}

func (s *SolNeonTxDecoderState) EndRange() *SolTxMetaInfo {
	return SolTxMetaInfoFromEndRange(s.stopBlockSlot, s.solTxMetaCollector.GetCommitment())
}

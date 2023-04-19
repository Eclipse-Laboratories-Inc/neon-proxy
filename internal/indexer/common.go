package indexer

import (
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/labstack/gommon/log"

	"github.com/prometheus/client_golang/prometheus"
)

type SolTxMetaDict struct {
	txMetaDict map[SolTxSigSlotInfo]*SolTxMetaInfo
}

func NewSolTxMetaDict() *SolTxMetaDict {
	return &SolTxMetaDict{
		txMetaDict: make(map[SolTxSigSlotInfo]*SolTxMetaInfo),
	}
}

func (s *SolTxMetaDict) HasSig(sigSlot SolTxSigSlotInfo) bool {
	_, ok := s.txMetaDict[sigSlot]
	return ok
}

func (s *SolTxMetaDict) Add(sigSlot SolTxSigSlotInfo, txMeta *SolTxReceipt) error {
	if txMeta == nil {
		return fmt.Errorf("solana receipt %v not found", sigSlot.SolSign)
	}

	tx, err := txMeta.Transaction.GetTransaction()
	if err != nil {
		return err
	}

	var signature solana.Signature
	if tx.Signatures != nil {
		signature = tx.Signatures[0]
	}

	if int(txMeta.Slot) != sigSlot.BlockSlot {
		return fmt.Errorf("solana receipt %v on another history branch: sugnature %v, block slot %v", sigSlot, signature, txMeta.Slot)

	}

	s.txMetaDict[sigSlot] = SolTxMetaInfoFromResponse(sigSlot, txMeta)
	return nil
}

func (s *SolTxMetaDict) Get(sigSlot SolTxSigSlotInfo) (*SolTxMetaInfo, error) {
	meta, ok := s.txMetaDict[sigSlot]
	if !ok {
		return nil, fmt.Errorf("no solana receipt for the signature %v", sigSlot.SolSign)
	}
	return meta, nil
}

func (s *SolTxMetaDict) Delete(sigSlot SolTxSigSlotInfo) {
	delete(s.txMetaDict, sigSlot)
}

func (s *SolTxMetaDict) Keys() []SolTxSigSlotInfo {
	keys := make([]SolTxSigSlotInfo, 0)
	for key := range s.txMetaDict {
		keys = append(keys, key)
	}
	return keys
}

type SolTxSigSlotInfo struct {
	SolSign   string
	BlockSlot int

	hash int
	str  string
}

func (st SolTxSigSlotInfo) String() string {
	if st.str == "" {
		st.str = fmt.Sprintf("%v:%v", st.BlockSlot, st.SolSign)
	}
	return st.str
}

func (st SolTxSigSlotInfo) Hash() (int, error) {
	h := fnv.New32a()
	_, err := h.Write([]byte(st.String()))
	if err != nil {
		return 0, err
	}
	return int(h.Sum32()), nil
}

type SolTxReceipt rpc.GetTransactionResult

type SolTxMetaInfo struct {
	ident SolTxSigSlotInfo

	blockSlot int
	solSign   string
	tx        *SolTxReceipt

	str   string
	reqID string
}

func SolTxMetaInfoFromResponse(slotInfo SolTxSigSlotInfo, resp *SolTxReceipt) *SolTxMetaInfo {
	return &SolTxMetaInfo{
		ident:     slotInfo,
		blockSlot: slotInfo.BlockSlot,
		tx:        resp,
	}
}

func NewSolTxMetaInfoFromEndRange(blockSlot int, commitment string) *SolTxMetaInfo {
	ident := SolTxSigSlotInfo{
		BlockSlot: blockSlot,
		SolSign:   fmt.Sprintf("end-%v", commitment),
	}
	return &SolTxMetaInfo{
		ident:     ident,
		blockSlot: blockSlot,
		solSign:   ident.SolSign,
	}

}

func (stm *SolTxMetaInfo) GetReqID() string {
	if stm.reqID == "" {
		stm.reqID = fmt.Sprintf("%s%d", stm.ident.SolSign[:7], stm.blockSlot)
	}
	return stm.reqID
}

func (stm *SolTxMetaInfo) String() string {
	// TODO implement after implementing str_fmt_object (task ...)
	/*	if stm.str == "" {
		stm.str = str_fmt_object(stm.ident)
	}*/
	return stm.str
}

type Status int

const (
	UnknownStatus Status = iota
	Success
	Failed
)

type SolIxLogState struct {
	program string
	level   int

	maxBpfCycleCnt  int
	usedBpfCycleCnt int
	usedHeapSize    int

	logs      []SolIxLogState
	innerLogs []SolIxLogState
}

type SolIxMetaInfo struct {
	ix       map[string]string
	idx      int
	innerIdx int

	program string
	level   int
	status  Status
	err     error

	usedHeapSize    int
	maxBpfCycleCnt  int
	usedBpfCycleCnt int

	neonTxSig        string
	neonGasUsed      int
	neonTotalGasUsed int

	neonTxReturn *NeonLogTxReturn
	neonTxEvents []NeonLogTxEvent
}

type SolTxCostInfo struct {
	solSign   string
	blockSlot int
	operator  string
	solSpent  int

	str            string
	calculatedStat bool
}

type SolTxLogDecoder struct{} //todo move to decoder?
type NeonLogTxReturn struct {
	Cancled bool
	GasUsed int
	Status  int
} // todo move to decoder?

type NeonLogTxCancel struct {
	GasUsed int
}

type Ident struct {
	solSign   string
	blockSlot int
	idx       int
	innerIdx  int
}

type SolNeonIxReceiptInfo struct {
	metaInfo SolIxMetaInfo

	solSign   string
	blockSlot int

	programIx int
	ixData    []byte

	neonStepCnt int

	solTxCost SolTxCostInfo

	ident Ident

	str         string
	accounts    []int
	accountKeys []string
}

type SolTxReceiptInfo struct {
	solCost  SolTxCostInfo
	operator string

	ixList         []map[string]string
	innerIxList    []map[string]string
	accountKeyList []string
	ixLogMsgList   []SolIxLogState
}

type SolBlockInfo struct {
	BlockSlot int
	BlockTime *int
	BlockHash string
	finalized bool
}

func (s *SolBlockInfo) SetFinalized(value bool) {
	s.finalized = value
}

func (s *SolBlockInfo) SetBlockHash(value string) {
	s.BlockHash = value
}

func (s *SolBlockInfo) IsEmpty() bool {
	return s.BlockTime == nil
}

func InsertBatchImpl(indexerDB DBInterface, pCounter prometheus.Counter, data []map[string]string) (int64, error) {
	colums := indexerDB.GetColums()
	sqlStr := fmt.Sprintf("INSERT INTO %s(%s) VALUES ", indexerDB.GetTableName(), strings.Join(colums, ", "))
	vals := []interface{}{}

	for _, row := range data {
		//sqlStr += "(?, ?, ?, ?, ?, ?),"
		rp := strings.Repeat("?, ", len(colums))
		sqlStr += "(" + rp[:len(rp)-2] + "),"
		rows := make([]interface{}, 0, len(colums))
		for _, column := range colums {
			rows = append(rows, row[column])
		}
		vals = append(vals, rows...)
	}
	// trim the last `,`
	sqlStr = sqlStr[0 : len(sqlStr)-1]
	stmt, err := indexerDB.GetDB().Prepare(sqlStr)
	if err != nil {
		return 0, err
	}

	// format all values at once
	res, err := stmt.Exec(vals...)
	if err != nil {
		return 0, err
	}
	pCounter.Add(float64(len(data)))
	return res.LastInsertId()
}

type NeonTxResultInfo struct {
	blockSlot *int
	blockHash string
	txIdx     int

	solSig        string
	solIxIdx      int
	solIxInnerIdx int

	neonSig string
	gasUsed string
	status  string

	logs []map[string]interface{}

	canceledStatus int
	lostStatus     int

	completed bool
	canceld   bool
}

func (n *NeonTxResultInfo) IsValid() bool {
	return n.gasUsed != ""
}

func (n *NeonTxResultInfo) AddEvent(event NeonLogTxEvent) {
	if n.blockSlot != nil {
		log.Warnf("Neon tx %s has completed event logs", n.neonSig)
		return
	}

	topics := make([]string, 0, len(event.topics))
	for i, topic := range event.topics {
		topics[i] = "0x" + hex.EncodeToString([]byte(topic))
	}

	rec := map[string]interface{}{
		"address":        "0x" + hex.EncodeToString(event.address),
		"topics":         topics,
		"data":           "0x" + hex.EncodeToString(event.data),
		"neonSolHash":    event.solSig,
		"neonIxIdx":      fmt.Sprintf("0x%x", event.idx),
		"neonInnerIxIdx": fmt.Sprintf("0x%x", event.innerIdx),
		"neonEventType":  fmt.Sprintf("%d", event.eventType),
		"neonEventLevel": fmt.Sprintf("0x%x", event.eventLevel),
		"neonEventOrder": fmt.Sprintf("0x%x", event.eventOrder),
		"neonIsHidden":   event.hidden,
		"neonIsReverted": event.reverted,
	}

	n.logs = append(n.logs, rec)
}

func (n *NeonTxResultInfo) SetRes(status int, gasUsed int) {
	n.status = fmt.Sprintf("0x%x", status)
	n.gasUsed = fmt.Sprintf("0x%x", gasUsed)
	n.completed = true
}

func (n *NeonTxResultInfo) SetBlockInfo(block SolBlockInfo, neonSig string, txIdx int, logIdx int) int {
	n.blockSlot = new(int)
	*n.blockSlot = block.BlockSlot
	n.blockHash = block.BlockHash
	n.solSig = neonSig
	n.txIdx = txIdx

	hexBlockSlot := fmt.Sprintf("0x%x", n.blockSlot)
	hexTxIdx := fmt.Sprintf("0x%x", n.txIdx)
	txLogIdx := 0

	for _, rec := range n.logs {
		rec["transactionHash"] = n.solSig
		rec["blockHash"] = n.blockHash
		rec["blockNumber"] = hexBlockSlot
		rec["transactionIndex"] = hexTxIdx
		if _, ok := rec["neonIsHidden"]; !ok {
			rec["logIndex"] = fmt.Sprintf("0x%x", logIdx)
			rec["transactionLogIndex"] = fmt.Sprintf("0x%x", txLogIdx)
			logIdx += 1
			txLogIdx += 1
		}
	}

	return logIdx
}

func (n *NeonTxResultInfo) SetCanceledRes(gasUsed int) {
	n.SetRes(0, gasUsed)
	n.canceld = true
}

func (n *NeonTxResultInfo) SetLostRes(gasUsed int) {
	n.SetRes(0, gasUsed)
	n.completed = false
}

func (n *NeonTxResultInfo) SetSolSigInfo(solSig string, solIxIdx int, solIxInnerIdx int) {
	n.solSig = solSig
	n.solIxIdx = solIxIdx
	n.solIxInnerIdx = solIxInnerIdx
}

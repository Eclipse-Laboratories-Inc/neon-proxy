package indexer

import (
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	solana2 "github.com/neonlabsorg/neon-proxy/pkg/solana"
	"hash/fnv"
	"strings"

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

	if txMeta.Slot != sigSlot.BlockSlot {
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

func (s *SolTxMetaDict) Pop(sigSlot SolTxSigSlotInfo) (*SolTxMetaInfo, error) {
	info, err := s.Get(sigSlot)
	if err != nil {
		return nil, err
	}
	s.Delete(sigSlot)
	return info, nil
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
	SolSign   solana.Signature
	BlockSlot uint64

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

type SolTxReceipt solana2.TransactionResult

type SolTxMetaInfo struct {
	ident SolTxSigSlotInfo

	blockSlot uint64
	tx        *SolTxReceipt

	str   string
	reqID string
}

func SolTxMetaInfoFromEndRange(slot uint64, commitmentType rpc.CommitmentType) *SolTxMetaInfo {
	// TODO implement me properly
	return &SolTxMetaInfo{}
}

func SolTxMetaInfoFromResponse(slotInfo SolTxSigSlotInfo, resp *SolTxReceipt) *SolTxMetaInfo {
	return &SolTxMetaInfo{
		ident:     slotInfo,
		blockSlot: slotInfo.BlockSlot,
		tx:        resp,
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
	neonGasUsed      int64 // TODO may by decimal.Decimal (?)
	neonTotalGasUsed int64

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
	GasUsed  string
	Status   string
	Canceled bool
} // todo move to decoder?

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
	solTxCost   SolTxCostInfo

	ident Ident

	str         string
	accounts    []int
	accountKeys []string
}

// TODO implement
func (sni *SolNeonIxReceiptInfo) GetAccount(accountIdx int) string {
	return ""
}

// TODO implement
func (sni *SolNeonIxReceiptInfo) IterAccount(accountIdx int) []string {
	return nil
}

func (sni *SolNeonIxReceiptInfo) AccountCnt() int {
	return len(sni.accounts)
}

func (sni *SolNeonIxReceiptInfo) SetNeonStepCnt(cnt int) {
	if cnt == 0 {
		panic("neon step cnt can't be 0")
	}
	sni.neonStepCnt = cnt
}

type SolTxReceiptInfo struct {
	solCost  SolTxCostInfo
	operator string

	ixList         []map[string]string
	innerIxList    []map[string]string
	accountKeyList []string
	ixLogMsgList   []SolIxLogState
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

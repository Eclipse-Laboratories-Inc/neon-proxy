package indexer

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type SolTxSigSlotInfo struct {
	SolSign   string
	BlockSlot int

	hash int
	str  string
}

type SolTxMetaInfo struct {
	ident SolTxSigSlotInfo

	blockSlot int
	tx        map[string]string

	str   string
	reqID string
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

	neonTxReturn NeonLogTxReturn
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
type NeonLogTxReturn struct{} // todo move to decoder?

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

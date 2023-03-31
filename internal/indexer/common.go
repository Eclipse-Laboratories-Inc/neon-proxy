package indexer

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

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

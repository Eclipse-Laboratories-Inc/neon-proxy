package indexer

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	neonTxLogsInsertedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "neon-proxy",
		Subsystem: "indexer",
		Name:      "inserted_neon_tx_logs_total",
		Help:      "The total number of inserted neon transaction logs",
	})
)

type NeonTxLogsDB struct {
	db *sql.DB
}

func (n NeonTxLogsDB) GetColums() []string {
	return []string{
		"log_topic1", "log_topic2", "log_topic3", "log_topic4",
		"log_topic_cnt", "log_data",
		"block_slot", "tx_hash", "tx_idx", "tx_log_idx", "log_idx", "address",
		"event_order", "event_level", "sol_sig", "idx", "inner_idx",
	}
}

func (n NeonTxLogsDB) GetTableName() string {
	return "neon_transaction_logs"
}

func (s NeonTxLogsDB) InsertBatch(data []map[string]string) (int64, error) {
	return InsertBatchImpl(s, neonTxLogsInsertedCounter, data)
}

func (s NeonTxLogsDB) GetDB() *sql.DB {
	return s.db
}

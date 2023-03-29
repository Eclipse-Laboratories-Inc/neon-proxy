package indexer

import (
	"database/sql"
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
	return InsertBatchImpl(s, s.db, data)
}

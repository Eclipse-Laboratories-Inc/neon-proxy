package indexer

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	neonTxInsertedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "neon-proxy",
		Subsystem: "indexer",
		Name:      "inserted_neon_tx_total",
		Help:      "The total number of inserted neon transactions",
	})
)

type NeonTxsDB struct {
	db *sql.DB
}

func (n NeonTxsDB) GetColums() []string {
	return []string{
		"neon_sig", "from_addr", "sol_sig", "sol_ix_idx", "sol_ix_inner_idx", "block_slot",
		"tx_idx", "nonce", "gas_price", "gas_limit", "to_addr", "contract", "value",
		"calldata", "v", "r", "s", "status", "gas_used", "logs",
	}
}

func (n NeonTxsDB) GetTableName() string {
	return "neon_transactions"
}

func (s NeonTxsDB) InsertBatch(data []map[string]string) (int64, error) {
	return InsertBatchImpl(s, neonTxInsertedCounter, data)
}

func (s NeonTxsDB) GetDB() *sql.DB {
	return s.db
}

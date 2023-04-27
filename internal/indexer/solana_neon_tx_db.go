package indexer

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	solanaNeonTxInsertedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "neon_proxy",
		Subsystem: "indexer",
		Name:      "inserted_solana_neon_tx_total",
		Help:      "The total number of inserted solana neon transactions",
	})
)

type SolanaNeonTxDB struct {
	db *sql.DB
}

func (s SolanaNeonTxDB) GetColums() []string {
	return []string{
		"sol_sig", "block_slot", "idx", "inner_idx", "neon_sig", "neon_step_cnt", "neon_income",
		"heap_size", "max_bpf_cycle_cnt", "used_bpf_cycle_cnt",
	}
}

func (s SolanaNeonTxDB) GetTableName() string {
	return "solana_neon_transactions"
}

func (s SolanaNeonTxDB) InsertBatch(data []map[string]string) (int64, error) {
	return InsertBatchImpl(s, solanaNeonTxInsertedCounter, data)
}

func (s SolanaNeonTxDB) GetDB() *sql.DB {
	return s.db
}

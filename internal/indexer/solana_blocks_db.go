package indexer

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	solanaBlocksInsertedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "neon-proxy",
		Subsystem: "indexer",
		Name:      "inserted_solana_blocks_total",
		Help:      "The total number of inserted solana blocks",
	})
)

type SolanaBlocksDB struct {
	db *sql.DB
}

func (s SolanaBlocksDB) GetColums() []string {
	return []string{"block_slot", "block_hash", "block_time", "parent_block_slot", "is_finalized", "is_active"}
}

func (s SolanaBlocksDB) GetTableName() string {
	return "solana_blocks"
}

func (s SolanaBlocksDB) InsertBatch(data []map[string]string) (int64, error) {
	return InsertBatchImpl(s, solanaBlocksInsertedCounter, data)
}

func (s SolanaBlocksDB) GetDB() *sql.DB {
	return s.db
}

package indexer

import (
	"database/sql"
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
	return InsertBatchImpl(s, s.db, data)
}

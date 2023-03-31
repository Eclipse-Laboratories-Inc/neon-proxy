package indexer

import (
	"database/sql"
)

type SolanaSignsDB struct {
	db *sql.DB
}

func (s SolanaSignsDB) GetColums() []string {
	return []string{"block_slot", "signature"}
}

func (s SolanaSignsDB) GetTableName() string {
	return "solana_transaction_signatures"
}

func (s SolanaSignsDB) InsertBatch(_ []map[string]string) (int64, error) {
	return 0, nil
}

func (s SolanaSignsDB) GetDB() *sql.DB {
	return s.db
}

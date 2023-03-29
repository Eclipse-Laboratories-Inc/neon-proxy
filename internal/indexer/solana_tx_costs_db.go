package indexer

import "database/sql"

type SolanaTxCostsDB struct {
	db *sql.DB
}

func (s SolanaTxCostsDB) GetColums() []string {
	return []string{"sol_sig", "block_slot", "operator", "sol_spent"}
}

func (s SolanaTxCostsDB) GetTableName() string {
	return "solana_transaction_costs"
}

func (s SolanaTxCostsDB) InsertBatch(data []map[string]string) (int64, error) {
	return InsertBatchImpl(s, s.db, data)
}

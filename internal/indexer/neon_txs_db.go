package indexer

import "database/sql"

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
	return InsertBatchImpl(s, s.db, data)
}

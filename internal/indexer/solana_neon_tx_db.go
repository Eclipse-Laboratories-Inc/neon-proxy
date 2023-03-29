package indexer

import "database/sql"

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
	return InsertBatchImpl(s, s.db, data)
}

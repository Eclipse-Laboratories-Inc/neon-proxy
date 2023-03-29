package indexer

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	solanaTxCostsInsertedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "neon-proxy",
		Subsystem: "indexer",
		Name:      "inserted_solana_tx_costs_total",
		Help:      "The total number of inserted solana transaction costs",
	})
)

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
	return InsertBatchImpl(s, s.db, solanaTxCostsInsertedCounter, data)
}

package indexer

import (
	"database/sql"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/exp/slices"
)

var (
	constantsInsertedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "neon-proxy",
		Subsystem: "indexer",
		Name:      "inserted_constants_total",
		Help:      "The total number of inserted constants",
	})
)

type ConstantsDB struct {
	db *sql.DB
}

func (c ConstantsDB) Items() []string {
	return []string{"min_receipt_block_slot", "latest_block_slot", "starting_block_slot", "finalized_block_slot"}
}

func (c ConstantsDB) GetColums() []string {
	return []string{"key", "value"}
}

func (c ConstantsDB) GetTableName() string {
	return "constants"
}

func (c ConstantsDB) InsertBatch(data []map[string]string) (int64, error) {
	return InsertBatchImpl(c, constantsInsertedCounter, data)
}

func (c ConstantsDB) GetDB() *sql.DB {
	return c.db
}

func (c ConstantsDB) GetItem(item string) error {
	if slices.Index(c.Items(), item) == -1 {
		return fmt.Errorf("constants DB: unknown item %s", item)
	}

	err := c.db.QueryRow("SELECT value FROM ? WHERE key = ?", c.GetTableName(), item).Scan(item)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("constants DB: unknown item %s", item)
		}
		return err
	}
	return nil
}

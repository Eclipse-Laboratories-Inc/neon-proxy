package indexer

import "database/sql"

// Db interface for indexer DBs
type DBInterface interface {
	// returns table colum names
	GetColums() []string
	// returns table name
	GetTableName() string
	// inserts elemts into table
	InsertBatch([]map[string]string) (int64, error)
	// returns sql DB connector
	GetDB() *sql.DB
}

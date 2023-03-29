package indexer

// Db interface for indexer DBs
type DBInterface interface {
	InsertBatch()
	Connect()
	IsConnected()
	Close()
}

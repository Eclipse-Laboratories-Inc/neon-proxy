package indexer

import (
	"database/sql"
	"errors"
)

var ErrRecordNotFound = errors.New("no record found")

type SolanaSignsDB struct {
	db *sql.DB
}

func (s *SolanaSignsDB) GetColums() []string {
	return []string{"block_slot", "signature"}
}

func (s *SolanaSignsDB) GetTableName() string {
	return "solana_transaction_signatures"
}

func (s *SolanaSignsDB) InsertBatch(_ []map[string]string) (int64, error) {
	return 0, nil
}

func (s *SolanaSignsDB) GetDB() *sql.DB {
	return s.db
}

func (s *SolanaSignsDB) AddSign(info SolTxSigSlotInfo) error {
	panic("AddSign: implement me!")
}

func (s *SolanaSignsDB) GetMaxSign() (*SolTxSigSlotInfo, error) {
	panic("GetMaxSign: implement me!")
}

func (s *SolanaSignsDB) GetNextSign(slot uint64) (*SolTxSigSlotInfo, error) {
	panic("GetNextSign: implement me!")
}

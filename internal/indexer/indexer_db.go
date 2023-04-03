package indexer

import "database/sql"

// Indexer database
type IndexerDB struct {
	solanaBlocksDB  *SolanaBlocksDB
	solanaNeonTxDB  *SolanaNeonTxDB
	solanaSignsDB   *SolanaSignsDB
	solanaTxCostsDB *SolanaTxCostsDB
	neonTxLogsDB    *NeonTxLogsDB
	neonTxsDB       *NeonTxsDB
	constantsDB     *ConstantsDB
}

func (db *IndexerDB) Init(sdb *sql.DB) {
	db.solanaBlocksDB = &SolanaBlocksDB{db: sdb}
	db.solanaNeonTxDB = &SolanaNeonTxDB{db: sdb}
	db.solanaSignsDB = &SolanaSignsDB{db: sdb}
	db.solanaTxCostsDB = &SolanaTxCostsDB{db: sdb}
	db.neonTxLogsDB = &NeonTxLogsDB{db: sdb}
	db.neonTxsDB = &NeonTxsDB{db: sdb}
	db.constantsDB = &ConstantsDB{db: sdb}
}

func (db *IndexerDB) IsHealthy() bool              { return true }
func (db *IndexerDB) SubmitBlock()                 {}
func (db *IndexerDB) FinalizeBlock()               {}
func (db *IndexerDB) ActivateBlockList()           {}
func (db *IndexerDB) GetBlockbySlot()              {}
func (db *IndexerDB) GetBlockByHash()              {}
func (db *IndexerDB) GetLatestBlock()              {}
func (db *IndexerDB) GetLatestBlockSlot()          {}
func (db *IndexerDB) GetFinalizedBlockSlot()       {}
func (db *IndexerDB) GetFinalizedBlock()           {}
func (db *IndexerDB) GetStartingBlock()            {}
func (db *IndexerDB) GetStartingBlockSlot()        {}
func (db *IndexerDB) GetMinReceiptBlockSlot()      {}
func (db *IndexerDB) SetMinReceiptBlockSlot()      {}
func (db *IndexerDB) GetLogList()                  {}
func (db *IndexerDB) GetTxlistByBlockSlot()        {}
func (db *IndexerDB) GetTxbyNeonSig()              {}
func (db *IndexerDB) GetTxbyBlockSlotTxIdx()       {}
func (db *IndexerDB) GetSolanaSignListByNeonSign() {}

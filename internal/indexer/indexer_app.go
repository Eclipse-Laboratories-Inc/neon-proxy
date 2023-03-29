package indexer

type IndexerApp struct {
	collector *Collector
}

func NewIndexerApp() *IndexerApp {
	return &IndexerApp{}
}

func (i *IndexerApp) Run() {

}

func (i *IndexerApp) cancelOldNeonTxs() {

}

func (i *IndexerApp) cancelNeonTxs() {

}

func (i *IndexerApp) completeNeonBlock() {

}

func (i *IndexerApp) commitTxStat() {

}

func (i *IndexerApp) commitBlockStat() {

}

func (i *IndexerApp) commitStatusStat() {

}

func (i *IndexerApp) commitStats() {

}

func (i *IndexerApp) getSolanaBlockDeque() {

}

func (i *IndexerApp) locateNeonBlock() {

}

func (i *IndexerApp) runSolanaTxCollector() {
	i.collector.RunSolanaTxs()
}

func (i *IndexerApp) hasNewBlocks() {

}

func (i *IndexerApp) logStats() {

}

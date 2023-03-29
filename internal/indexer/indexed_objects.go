package indexer

type IndexedObject interface {
	StartBlockSlot()
	LastBlockSlot()
}

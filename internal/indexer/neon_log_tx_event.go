package indexer

type LogTxEventType int

const (
	Log LogTxEventType = 1

	EnterCall         LogTxEventType = 101
	EnterCallCode     LogTxEventType = 102
	EnterStaticCall   LogTxEventType = 103
	EnterDelegateCall LogTxEventType = 104
	EnterCreate       LogTxEventType = 105
	EnterCreate2      LogTxEventType = 106

	ExitStop         LogTxEventType = 201
	ExitReturn       LogTxEventType = 202
	ExitSelfDestruct LogTxEventType = 203
	ExitRevert       LogTxEventType = 204

	Return LogTxEventType = 300
	Cancel LogTxEventType = 301
)

type NeonLogTxEvent struct {
	eventType LogTxEventType
	Hidden    bool

	address string
	topics  []string
	data    []byte

	solSig       string
	idx          int
	innerIdx     int
	totalGasUsed int64
	reverted     bool
	eventLevel   int
	eventOrder   int
}

func (n *NeonLogTxEvent) DeepCopy() NeonLogTxEvent {
	logEvent := *n
	logEvent.topics = make([]string, len(n.topics))
	copy(logEvent.topics, n.topics)
	return logEvent
}

func (n *NeonLogTxEvent) isStartEventType() bool {
	//todo implement
	return true
}

func (n *NeonLogTxEvent) isExitEventType() bool {
	//todo implement
	return true
}

type SortNeonEventList []NeonLogTxEvent

func (s SortNeonEventList) Len() int {
	return len(s)
}

func (s SortNeonEventList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortNeonEventList) Less(i, j int) bool {
	return s[i].totalGasUsed > s[j].totalGasUsed // reverse order by Gas Price
}

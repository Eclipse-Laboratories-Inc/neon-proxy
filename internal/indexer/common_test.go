package indexer

import (
	"testing"

	"github.com/test-go/testify/assert"
)

/*
func (n *NeonTxResultInfo) AddEvent(event NeonLogTxEvent) {
	if n.blockSlot != nil {
		log.Warnf("Neon tx %s has completed event logs", n.neonSig)
		return
	}

	topics := make([]string, 0, len(event.topics))
	for i, topic := range event.topics {
		topics[i] = "0x" + hex.EncodeToString([]byte(topic))
	}

	rec := map[string]interface{}{
		"address":        "0x" + hex.EncodeToString(event.address),
		"topics":         topics,
		"data":           "0x" + hex.EncodeToString(event.data),
		"neonSolHash":    event.solSig,
		"neonIxIdx":      fmt.Sprintf("0x%x", event.idx),
		"neonInnerIxIdx": fmt.Sprintf("0x%x", event.innerIdx),
		"neonEventType":  fmt.Sprintf("%d", event.eventType),
		"neonEventLevel": fmt.Sprintf("0x%x", event.eventLevel),
		"neonEventOrder": fmt.Sprintf("0x%x", event.eventOrder),
		"neonIsHidden":   event.hidden,
		"neonIsReverted": event.reverted,
	}

	n.logs = append(n.logs, rec)
}
*/

func Test_NeonTxResultInfoAddEvent(t *testing.T) {
	n := NeonTxResultInfo{}
	event := NeonLogTxEvent{
		eventType:    ExitRevert,
		address:      "0x001",
		data:         []byte("0x002"),
		Hidden:       true,
		solSig:       "0x:003",
		idx:          4,
		innerIdx:     2,
		totalGasUsed: 1,
		reverted:     false,
	}

	assert.Equal(t, 0, len(n.logs))

	n.AddEvent(event)

	assert.Equal(t, 1, len(n.logs))
	assert.Equal(t, "0x3078303031", n.logs[0]["address"])
	assert.Equal(t, "0x3078303032", n.logs[0]["data"])
	assert.Equal(t, "0x0", n.logs[0]["neonEventLevel"])
	assert.Equal(t, "0x0", n.logs[0]["neonEventOrder"])
	assert.Equal(t, "204", n.logs[0]["neonEventType"])
	assert.Equal(t, "0x2", n.logs[0]["neonInnerIxIdx"])
	assert.Equal(t, true, n.logs[0]["neonIsHidden"])
	assert.Equal(t, false, n.logs[0]["neonIsReverted"])
	assert.Equal(t, "0x4", n.logs[0]["neonIxIdx"])
	assert.Equal(t, "0x:003", n.logs[0]["neonSolHash"])
	assert.Empty(t, n.logs[0]["topics"])
}

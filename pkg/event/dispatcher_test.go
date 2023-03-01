package event

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestEvent struct {
	name string
}

func (e TestEvent) Name() string {
	return e.name
}

func (e TestEvent) IsAsynchronous() bool {
	return false
}

var processedTestHandlers []string

type TestHandler struct {
	name string
}

func (h TestHandler) Handle(event Event) {
	processedTestHandlers = append(processedTestHandlers, h.name)
}

func TestDispatcherOneHandler(t *testing.T) {
	processedTestHandlers = nil
	d := NewDispatcher()
	ev := TestEvent{
		name: "ABCD-event",
	}
	h := TestHandler{
		name: "ABCD-Handler",
	}
	d.Register(h, ev, High)
	d.Notify(ev)
	assert.Equal(t, "ABCD-event", ev.Name())
	assert.Len(t, processedTestHandlers, 1)
	assert.Equal(t, processedTestHandlers[0], "ABCD-Handler")
}

func TestDispatcherMultyHandlersWithPriority(t *testing.T) {
	processedTestHandlers = nil
	d := NewDispatcher()
	ev := TestEvent{
		name: "ABCD-event",
	}
	priorities := []int{15, High, Normal, 80, Low, -20, Low, 30, High}
	for _, priority := range priorities {
		h := TestHandler{
			name: fmt.Sprintf("ABCD-handler %d", priority),
		}
		d.Register(h, ev, priority)
	}

	d.Notify(ev)
	assert.Equal(t, "ABCD-event", ev.Name())
	assert.Equal(t, len(processedTestHandlers), len(priorities))

	// check if Handlers are called by priority
	handlerNames := []string{
		"ABCD-handler 100", "ABCD-handler 100", "ABCD-handler 80", "ABCD-handler 30",
		"ABCD-handler 15", "ABCD-handler 0", "ABCD-handler -20", "ABCD-handler -100",
		"ABCD-handler -100",
	}
	assert.Equal(t, handlerNames, processedTestHandlers)
}

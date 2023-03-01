package event

import (
	"errors"
	"fmt"
	"testing"
	"time"

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

func (h TestHandler) Handle(event Event) error {
	processedTestHandlers = append(processedTestHandlers, h.name)
	return nil
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

func TestDispatcherOneHandlerMustTriggerWithoutError(t *testing.T) {
	processedTestHandlers = nil
	d := NewDispatcher()
	ev := TestEvent{
		name: "ABCD-event",
	}
	h := TestHandler{
		name: "ABCD-Handler",
	}
	d.Register(h, ev, High)

	d.MustTrigger(ev)
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

// handler function
type TestHandlerFunc func(e Event) error

// Handle event. implements the Handler interface
func (fn TestHandlerFunc) Handle(e Event) error {
	return fn(e)
}

func testHandlerFunc1(e Event) error {
	processedTestHandlers = append(processedTestHandlers, "testHandlerFunc1")
	return nil
}
func testHandlerFunc2(e Event) error {
	processedTestHandlers = append(processedTestHandlers, "testHandlerFunc2")
	return nil
}

func TestDispatcherHandlerFuncWithPriority(t *testing.T) {
	processedTestHandlers = nil
	d := NewDispatcher()
	ev := TestEvent{
		name: "ABCD-event",
	}
	priorities := []int{Normal, Low, 20, High}
	for _, priority := range priorities {
		h := TestHandler{
			name: fmt.Sprintf("ABCD-handler %d", priority),
		}
		d.Register(h, ev, priority)
	}
	f1 := TestHandlerFunc(testHandlerFunc1)
	f2 := TestHandlerFunc(testHandlerFunc2)
	d.Register(f1, ev, 10)
	d.Register(f2, ev, 15)

	d.Notify(ev)
	assert.Equal(t, "ABCD-event", ev.Name())
	assert.Equal(t, len(processedTestHandlers), 2+len(priorities))

	// check if Handlers are called by priority
	handlerNames := []string{
		"ABCD-handler 100", "ABCD-handler 20", "testHandlerFunc2", "testHandlerFunc1",
		"ABCD-handler 0", "ABCD-handler -100",
	}
	assert.Equal(t, handlerNames, processedTestHandlers)
}

type TestEventAsync struct {
	name string
}

func (e TestEventAsync) Name() string {
	return e.name
}

func (e TestEventAsync) IsAsynchronous() bool {
	return true
}

func TestDispatcherAsyncEvent(t *testing.T) {
	processedTestHandlers = nil
	d := NewDispatcher()
	ev := TestEventAsync{
		name: "ABCD-event-async",
	}
	priorities := []int{High, Low, Low, Normal}
	for _, priority := range priorities {
		h := TestHandler{
			name: fmt.Sprintf("ABCD-handler-async %d", priority),
		}
		d.Register(h, ev, priority)
	}

	d.Notify(ev)

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, "ABCD-event-async", ev.Name())
	assert.Equal(t, len(processedTestHandlers), len(priorities))

	// check if Handlers are called by priority for async event
	handlerNames := []string{
		"ABCD-handler-async 100", "ABCD-handler-async 0",
		"ABCD-handler-async -100", "ABCD-handler-async -100",
	}
	assert.Equal(t, handlerNames, processedTestHandlers)
}

func testHandlerFuncError(e Event) error {
	processedTestHandlers = append(processedTestHandlers, "testHandlerFuncError")
	return errors.New("Something Wrong")
}

func TestDispatcherMustTriggerWithError(t *testing.T) {
	processedTestHandlers = nil
	d := NewDispatcher()
	ev := TestEvent{
		name: "ABCD-event",
	}

	ferr := TestHandlerFunc(testHandlerFuncError)
	d.Register(ferr, ev, High)

	d.Notify(ev)
	assert.Equal(t, "ABCD-event", ev.Name())
	assert.Len(t, processedTestHandlers, 1)
	assert.Equal(t, processedTestHandlers[0], "testHandlerFuncError")

	// check MustTrigger
	assert.PanicsWithError(t, "Something Wrong", func() { d.MustTrigger(ev) })
}

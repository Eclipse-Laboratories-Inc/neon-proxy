package event

import (
	"sort"
	"sync"
)

type DispatcherInterface interface {
	Register(handler Handler, event Event, priority int)
	UnRegister(event Event)
	Notify(event Event)
	MustTrigger(event Event) // panic on error
}

// some priority constants
const (
	Low    = -100
	Normal = 0
	High   = 100
)

// Event Handler with Priority
type HandlerData struct {
	handler  Handler
	priority int
}

// Array of Event Handlers ordered by Priority
type SortedHandlers []HandlerData

func (s SortedHandlers) Insert(handler Handler, priority int) SortedHandlers {
	// find the handlers position in sorted Handlers
	i := sort.Search(len(s), func(i int) bool { return s[i].priority <= priority })
	if i == len(s) {
		return append(s, HandlerData{handler, priority})
	}
	s = append(s[:i+1], s[i:]...)
	s[i] = HandlerData{handler, priority}
	return s
}

type Dispatcher struct {
	mu       sync.Mutex
	handlers map[string]SortedHandlers
}

func (d *Dispatcher) Register(handler Handler, event Event, priority int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	name := event.Name()
	d.handlers[name] = d.handlers[name].Insert(handler, priority)
}

// Unsubscribe from all Handler
func (d *Dispatcher) UnRegister(event Event) {
	d.mu.Lock()
	defer d.mu.Unlock()

	name := event.Name()
	delete(d.handlers, name)
}

func (d *Dispatcher) Notify(event Event) {
	panicOnError := false
	d.notify(event, panicOnError)
}

func (d *Dispatcher) MustTrigger(event Event) {
	panicOnError := true
	d.notify(event, panicOnError)
}

func (d *Dispatcher) notify(event Event, panicOnError bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if event.IsAsynchronous() {
		go d.notifyEvent(event, panicOnError)
	} else {
		d.notifyEvent(event, panicOnError)
	}
}

func (d *Dispatcher) notifyEvent(event Event, panicOnError bool) {
	name := event.Name()
	for _, hd := range d.handlers[name] {
		err := hd.handler.Handle(event)
		if panicOnError && err != nil {
			panic(err)
		}
	}
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string]SortedHandlers),
	}
}

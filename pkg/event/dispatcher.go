package event

import "sort"

type DispatcherInterface interface {
	Register(handler Handler, event Event, priority int)
	UnRegister(event Event)
	Notify(event Event)
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
	i := sort.Search(len(s), func(i int) bool { return s[i].priority >= priority })
	if i == len(s) {
		return append(s, HandlerData{handler, priority})
	}
	s = append(s[:i+1], s[i:]...)
	s[i] = HandlerData{handler, priority}
	return s
}

type Dispatcher struct {
	handlers map[string]SortedHandlers
}

func (d *Dispatcher) Register(handler Handler, event Event, priority int) {
	name := event.Name()
	d.handlers[name] = d.handlers[name].Insert(handler, priority)
}

// Unsubscribe from all Handler
func (d *Dispatcher) UnRegister(event Event) {
	name := event.Name()
	delete(d.handlers, name)
}

func (d *Dispatcher) Notify(event Event) {
	if event.IsAsynchronous() {
		go d.notifyEvent(event)
	} else {
		d.notifyEvent(event)
	}
}

func (d *Dispatcher) notifyEvent(event Event) {
	name := event.Name()
	for _, hd := range d.handlers[name] {
		hd.handler.Handle(event)
	}
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string]SortedHandlers),
	}
}

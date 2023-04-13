package event

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

var (
	once               sync.Once
	dispatcherInstance *Dispatcher
)

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

func DispatcherInstance() *Dispatcher {
	once.Do(func() { //atomic
		dispatcherInstance = &Dispatcher{
			handlers: make(map[string]SortedHandlers),
		}

	})
	return dispatcherInstance
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

func (d *Dispatcher) Notify(event Event) error {
	panicOnError := false
	return d.notify(event, panicOnError)
}

func (d *Dispatcher) MustTrigger(event Event) {
	panicOnError := true
	d.notify(event, panicOnError)
}

func (d *Dispatcher) notify(event Event, panicOnError bool) error {
	if event.IsAsynchronous() {
		go d.notifyEvent(event, panicOnError)
		return nil // todo: ignoring errors in async mode
	}
	return d.notifyEvent(event, panicOnError)
}

func (d *Dispatcher) notifyEvent(event Event, panicOnError bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	name := event.Name()
	var errs []error
	for _, hd := range d.handlers[name] {
		err := hd.handler.Handle(event)
		if err != nil {
			errs = append(errs, err)
		}
	}

	err := formatErrors(errs)
	if panicOnError && err != nil {
		panic(err)
	}
	return err
}

func formatErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	errMsg := ""
	for _, err := range errs {
		errMsg += fmt.Sprintf("EventDispatcher: error on notify event %s \n", err.Error())
	}
	return errors.New(errMsg)
}

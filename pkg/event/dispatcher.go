package event

type DispatcherInterface interface {
	Register(handler Handler, event Event, priority int)
	UnRegister(event Event)
	Notify(event Event) error
	MustTrigger(event Event) // panic on error
}

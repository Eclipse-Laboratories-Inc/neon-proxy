package event

// Handler interface
type Handler interface {
	Handle(ev Event) error
}

// handler function
type HandlerFunc func(e Event) error

// Handle event. implements the Handler interface
func (fn HandlerFunc) Handle(e Event) error {
	return fn(e)
}

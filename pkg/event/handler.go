package event

// Handler interface
type Handler interface {
	Handle(ev Event)
}

// handler function
type HandlerFunc func(e Event)

// Handle event. implements the Handler interface
func (fn HandlerFunc) Handle(e Event) {
	fn(e)
}

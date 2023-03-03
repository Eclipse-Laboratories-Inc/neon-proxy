package event

// abstract event
type Event interface {
	Name() string
	IsAsynchronous() bool
}

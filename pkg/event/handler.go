package event

type Handler interface {
	Handle(ev Event)
}

package broadcaster

import (
	"context"

	"github.com/neonlabsorg/neon-service-framework/pkg/logger"
)

/*
Broadcaster system has one source that receives data and distributes
to the subscribers. By using subscribe method it creates a new channel and pushes copied
incoming data into each subscriber's channel. All the broadcaster functions are thread safe.
*/

type listener struct {
	Params      interface{}
	DataChannel chan interface{}
}

type Broadcaster struct {
	ctx            *context.Context
	source         chan interface{}
	sourceError    chan error
	listeners      []chan interface{}
	addListener    chan chan interface{}
	removeListener chan (<-chan interface{})
	start          func()
	log            logger.Logger
}

// create initial server structure with source nil
func NewBroadcaster(ctx *context.Context, log logger.Logger) *Broadcaster {
	return &Broadcaster{
		ctx:            ctx,
		log:            log,
		source:         make(chan interface{}),
		sourceError:    make(chan error),
		listeners:      make([]chan interface{}, 0),
		addListener:    make(chan (chan interface{})),
		removeListener: make(chan (<-chan interface{})),
	}
}

// sets external sources for incoming data
func (broadcaster *Broadcaster) SetSources(source chan interface{}, sourceErr chan error) {
	broadcaster.source = source
	broadcaster.sourceError = sourceErr
}

// subscribing to the broadcaster
func (broadcaster *Broadcaster) Subscribe() chan interface{} {
	newListener := make(chan interface{}, 10)
	broadcaster.addListener <- newListener
	return newListener
}

// unsubscribes
func (broadcaster *Broadcaster) CancelSubscription(channel <-chan interface{}) {
	broadcaster.removeListener <- channel
}

// closes all the listeners
func (broadcaster *Broadcaster) closeListeners() {
	for _, listener := range broadcaster.listeners {
		if listener != nil {
			close(listener)
		}
	}
}

/*
Start is the main routine for subscriber that listens to new request for adding listener, removing
listener or receiving incoming source data and distributing among registered/subscribed listeners in thread safe mannet
*/
func (broadcaster *Broadcaster) Start() {
	// defer closing listeners
	defer broadcaster.closeListeners()

	// listen to incoming actions
	for {
		select {
		// check if server is shut down
		case <-(*broadcaster.ctx).Done():
			return

		// if new listener is added, add as a new channel
		case newListener := <-broadcaster.addListener:
			broadcaster.listeners = append(broadcaster.listeners, newListener)

		// when unsubscribing rpemove the listener
		case listenerToRemove := <-broadcaster.removeListener:
			for i, ch := range broadcaster.listeners {
				if ch == listenerToRemove {
					broadcaster.listeners[i] = broadcaster.listeners[len(broadcaster.listeners)-1]
					broadcaster.listeners = broadcaster.listeners[:len(broadcaster.listeners)-1]
					close(ch)
					break
				}
			}

		// when receiving a new transaction from the source broadcast it to the listeners
		case event := <-broadcaster.source:
			for _, listener := range broadcaster.listeners {
				if listener != nil {
					select {
					case listener <- event:
					case <-(*broadcaster.ctx).Done():
						return
					}
				}
			}

		// in case we catch error
		case err := <-broadcaster.sourceError:
			broadcaster.log.Error().Err(err).Msg("Error on subscriber process")
			continue
		}
	}
}

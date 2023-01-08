package proxy

import (
  "net/http"
  "context"
  "fmt"

  "github.com/gorilla/websocket"
)

type Server struct {
  ctx context.Context
  source chan Transaction
  sourceError chan error
  listeners []chan Transaction
  addListener chan chan Transaction
  removeListener chan (<-chan Transaction)
}

// create initial server structure with source nil
func NewServer(ctx context.Context) *Server {
  return &Server{
		ctx:            ctx,
		source:         make(chan Transaction, 0),
    sourceError:    make(chan error),
    listeners:      make([]chan Transaction, 0),
    addListener:    make(chan (chan Transaction)),
    removeListener: make(chan (<-chan Transaction)),
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// subscribing to the serverc
func (server *Server) Subscribe() <-chan Transaction {
  newListener := make(chan Transaction, 10)
  server.addListener <- newListener
  return newListener
}

// unsubscribes
func (server *Server) CancelSubscription(channel <-chan Transaction) {
  server.removeListener <- channel
}

// closes all the listeners
func (server *Server) closeListeners() {
  for _, listener := range server.listeners {
    if listener != nil {
        close(listener)
    }
  }
}

func (server *Server) StartBroadcaster() {
  // defer closing listeners
  defer server.closeListeners()

  // listen to incoming actions
  for {
    select {
      // check if server is shut down
    case <-server.ctx.Done():
        return

      // if new listener is added, add as a new channel
    case newListener := <- server.addListener:
        server.listeners = append(server.listeners, newListener)

      // when unsubscribing remove the listener
    case listenerToRemove := <- server.removeListener:
        for i, ch := range server.listeners {
          if ch == listenerToRemove {
              server.listeners[i] = server.listeners[len(server.listeners)-1]
              server.listeners = server.listeners[:len(server.listeners)-1]
              close(ch)
              break
          }
        }

      // when receiving a new transaction from the source broadcast it to the listeners
    case transaction := <-server.source:
        for _, listener := range server.listeners {
          if listener != nil {
            select {
             case listener <- transaction:
             case <-server.ctx.Done():
              return
            }
          }
        }
    }
  }
}

func (server * Server) wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}

  // subscribe to transaction
  transactions := server.Subscribe()

  // defer unsubscribe
  defer server.CancelSubscription(transactions)

  // for each incoming transaction send it to the connected user
  for {
    select {
    case transaction := <- transactions:
      if err := ws.WriteMessage(1, []byte(transaction.signature)); err != nil {
  			fmt.Println(err)
  			return
  		}
    case <- server.ctx.Done():
      return
    }
	}
}

func (server * Server) StartServer(port string) {
  http.HandleFunc("/ws", server.wsEndpoint)
  fmt.Println(http.ListenAndServe(":" + port, nil))
}

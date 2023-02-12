package wssubscriber

import (
  "context"
  "net/http"
  "fmt"

  "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
  ReadBufferSize:  8192,
  WriteBufferSize: 8192,
  CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewServer(ctx *context.Context) *Server{
  return &Server{ctx: ctx}
}

type Server struct {
  ctx *context.Context

  // head broadcaster instance
  newHeadsBroadcaster *Broadcaster

  //pending transaction broadcaster instance
  pendingTransactionBroadcaster *Broadcaster
}

// upgrade connection to websocket and register the client
func (server * Server) wsEndpoint(w http.ResponseWriter, r *http.Request) {
  // upgrade this connection to a WebSocket
  conn, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
  	fmt.Println(err)
  }

  // create a new client associated with the connection
  client := NewClient(conn, server.newHeadsBroadcaster, server.pendingTransactionBroadcaster)

	// listen for incoming subscriptions.
	go client.ReadPump()

  // send subscription data back to user
	go client.WritePump()
}

func (server *Server) GetNewHeadBroadcaster(solanaWSEndpoint string) (*Broadcaster, error) {
  // create a new broadcaster
  broadcaster := NewBroadcaster(server.ctx)

  // register source and sourceError for broadcaster that will we solana endpoint pulling new heads
	err := RegisterNewHeadBroadcasterSources(server.ctx, solanaWSEndpoint, broadcaster.source, broadcaster.sourceError)
	if err != nil {
		return nil, err
	}

  // start broadcasting incoming new heads to subscribers
	go broadcaster.Start()

  return broadcaster, nil
}

// start listening to incoming subscription connections
func (server * Server) StartServer(port string) {
  http.HandleFunc("/", server.wsEndpoint)
  fmt.Println(http.ListenAndServe(":" + port, nil))
}

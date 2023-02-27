package wssubscriber

import (
  "context"
  "net/http"
  "time"
  "fmt"
  "log"
  "runtime"

  "github.com/gorilla/websocket"
  "github.com/neonlabsorg/neon-proxy/pkg/logger"
)

var upgrader = websocket.Upgrader{
  ReadBufferSize:  8192,
  WriteBufferSize: 8192,
  CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewServer(ctx *context.Context, log logger.Logger) *Server{
  return &Server{ctx: ctx, log: log}
}

type Server struct {
  ctx *context.Context

  // head broadcaster instance
  newHeadsBroadcaster *Broadcaster

  //pending transaction broadcaster instance
  pendingTransactionBroadcaster *Broadcaster

  // logger instance
  log logger.Logger
}

// upgrade connection to websocket and register the client
func (server * Server) wsEndpoint(w http.ResponseWriter, r *http.Request) {
  server.log.Info().Msg("new client connection is establishing ... ")

  // upgrade this connection to a WebSocket
  conn, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
  	server.log.Error().Err(err).Msg("Error on upgrading client connection to ws")
    return
  }

  // create a new client associated with the connection
  client := NewClient(conn, server.log, server.newHeadsBroadcaster, server.pendingTransactionBroadcaster)

	// listen for incoming subscriptions.
	go client.ReadPump()

	// send subscription data back to user
	go client.WritePump()
}

func (server *Server) GetNewHeadBroadcaster(solanaWSEndpoint string) (*Broadcaster, error) {
  // create a new broadcaster
  broadcaster := NewBroadcaster(server.ctx, server.log)

  // start broadcasting incoming new heads to subscribers
  go broadcaster.Start()

  // register source and sourceError for broadcaster that will we solana endpoint pulling new heads
  err := RegisterNewHeadBroadcasterSources(server.ctx, server.log, solanaWSEndpoint, broadcaster.source, broadcaster.sourceError)
  if err != nil {
    return nil, err
  }

  server.log.Info().Msg("NewHeads broadcaster sources registered")
  return broadcaster, nil
}

func (server *Server) GetPendingTransactionBroadcaster(solanaWSEndpoint string) (*Broadcaster, error) {
  // create a new broadcaster
  broadcaster := NewBroadcaster(server.ctx, server.log)

  // start broadcasting incoming new heads to subscribers
  go broadcaster.Start()

  // register source and sourceError for broadcaster that will we solana endpoint pulling new heads
  err := RegisterPendingTransactionBroadcasterSources(server.ctx, server.log, solanaWSEndpoint, broadcaster.source, broadcaster.sourceError)
  if err != nil {
    return nil, err
  }

  server.log.Info().Msg("NewHeads broadcaster sources registered")
  return broadcaster, nil
}

// start listening to incoming subscription connections
func (server * Server) StartServer(port string) {
  go func() {
    for {
    time.Sleep(time.Second * 2)
    fmt.Println(runtime.NumGoroutine())
  }
  }()
  http.HandleFunc("/", server.wsEndpoint)
  log.Println("starting on ", port)
  fmt.Println(http.ListenAndServe(":" + port, nil))
}

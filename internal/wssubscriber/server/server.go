package server

import (
  "context"
  "net/http"
  "fmt"

  "github.com/gorilla/websocket"
  "github.com/neonlabsorg/neon-proxy/pkg/logger"
  "github.com/neonlabsorg/neon-proxy/internal/wssubscriber/source"
  "github.com/neonlabsorg/neon-proxy/internal/wssubscriber/broadcaster"
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
  newHeadsBroadcaster *broadcaster.Broadcaster

  //pending transaction broadcaster instance
  pendingTransactionBroadcaster *broadcaster.Broadcaster

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

func (server *Server) StartNewHeadBroadcaster(solanaWSEndpoint string) error {
  // create a new broadcaster
  broadcaster := broadcaster.NewBroadcaster(server.ctx, server.log)

  // start broadcasting incoming new heads to subscribers
  go broadcaster.Start()

  // register source and sourceError for broadcaster that will we solana endpoint pulling new heads
  err := source.RegisterNewHeadBroadcasterSources(server.ctx, server.log, solanaWSEndpoint, broadcaster)
  if err != nil {
    return err
  }

  // register head broadcaster in server
  server.newHeadsBroadcaster = broadcaster

  server.log.Info().Msg("newHeads broadcaster sources registered")
  return nil
}

func (server *Server) StartPendingTransactionBroadcaster() error {
  // create a new broadcaster
  broadcaster := broadcaster.NewBroadcaster(server.ctx, server.log)

  // start broadcasting incoming new heads to subscribers
  go broadcaster.Start()

  // register source and sourceError for broadcaster that will we solana endpoint pulling new heads
  err := source.RegisterPendingTransactionBroadcasterSources(server.ctx, server.log, broadcaster)
  if err != nil {
    return err
  }

  // register pendingTransaction broadcaster
  server.pendingTransactionBroadcaster = broadcaster
  server.log.Info().Msg("pendingTransaction broadcaster sources registered")
  return nil
}

// start listening to incoming subscription connections
func (server * Server) StartServer(port string) {
  http.HandleFunc("/", server.wsEndpoint)
  fmt.Println(http.ListenAndServe(":" + port, nil))
}

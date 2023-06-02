package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/broadcaster"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/source"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewServer(ctx *context.Context, log logger.Logger, endpoint string) *Server {
	return &Server{ctx: ctx, log: log, evmRpcEndpoint: endpoint}
}

type Server struct {
	ctx *context.Context

	// head broadcaster instance
	newHeadsBroadcaster *broadcaster.Broadcaster

	//pending transaction broadcaster instance
	pendingTransactionBroadcaster *broadcaster.Broadcaster

	logsBroadcaster *broadcaster.Broadcaster

	// logger instance
	log logger.Logger

	evmRpcEndpoint string
}

// upgrade connection to websocket and register the client
func (server *Server) wsEndpoint(w http.ResponseWriter, r *http.Request) {
	server.log.Info().Msg("new client connection is establishing ... ")

	// upgrade this connection to a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		server.log.Error().Err(err).Msg("Error on upgrading client connection to ws")
		return
	}

	// create a new client associated with the connection
	client := NewClient(conn, server.log, server.newHeadsBroadcaster, server.pendingTransactionBroadcaster, server.logsBroadcaster, server.evmRpcEndpoint)

	// listen for incoming subscriptions.
	go client.ReadPump()

	// send subscription data back to user
	go client.WritePump()
}

// Creates broadcaster instance that receives new heads from the routine that is processing solana blockchain with rpc and distributes them to subscribed users
func (server *Server) StartNewHeadBroadcaster(solanaWSEndpoint string) error {
	// create a new broadcaster
	broadcaster := broadcaster.NewBroadcaster(server.ctx, server.log)

	/*
	   start broadcasting incoming new heads to subscribers. We need to activate broadcaster
	   before pushing new data into it's source as pushing will block if broadcaster isn't active and receiving the data
	*/
	go broadcaster.Start()

	/*
	   register source and sourceError for newHeads broadcaster.
	   After registering sources a separate routine will process new finalized block headers and push them into the source of the broadcaster
	   which on it's own distribute those block heads (headers) to subscribed users
	*/
	if err := source.RegisterNewHeadBroadcasterSources(server.ctx, server.log, solanaWSEndpoint, broadcaster); err != nil {
		return err
	}

	// register newHeads broadcaster to use with connected clients later. At this point the broadcaster is active and waiting for new subscribers to join.
	server.newHeadsBroadcaster = broadcaster
	server.log.Info().Msg("newHeads broadcaster sources registered")
	return nil
}

// Creates broadcaster instance that receives new pending transactions from source (mempool) and distributes them to subscribed users
func (server *Server) StartPendingTransactionBroadcaster() error {
	// create a new broadcaster
	broadcaster := broadcaster.NewBroadcaster(server.ctx, server.log)

	/*
	   start broadcasting incoming pending transactions to subscribers. We need to activate broadcaster
	   before pushing new data into it's source as pushing will block if broadcaster isn't active and receiving the data
	*/
	go broadcaster.Start()

	/*
	   register source and sourceError for pending transaction broadcaster.
	   After registering sources a separate routine will pull all the pending transactions from
	   mempool and push it into the source of broadcaster which on it's own distribute those transactions
	   to subscribed users
	*/
	if err := source.RegisterPendingTransactionBroadcasterSources(server.ctx, server.log, broadcaster); err != nil {
		return err
	}

	// register pendingTransaction broadcaster to use with connected clients later. At this point the broadcaster is active and waiting for new subscribers to join.
	server.pendingTransactionBroadcaster = broadcaster
	server.log.Info().Msg("pendingTransaction broadcaster sources registered")
	return nil
}

// creates broadcaster for receiving transaction logs from evm
func (server *Server) StartLogsBroadcaster(solanaWSEndpoint, evmAddr string) error {
	// create a new broadcaster
	broadcaster := broadcaster.NewBroadcaster(server.ctx, server.log)

	/*
	   start broadcasting incoming new logs to subscribers. We need to activate broadcaster
	   before pushing new data into it's source as pushing will block if broadcaster isn't active and receiving the data
	*/
	go broadcaster.Start()

	/*
	   register source and sourceError for newLogs broadcaster.
	   After registering sources a separate routine will process new transaction logs and push them into the source of the broadcaster
	   which on it's own distribute those logs to subscribed users, using user's filters
	*/
	if err := source.RegisterLogsBroadcasterSources(server.ctx, server.log, solanaWSEndpoint, evmAddr, broadcaster); err != nil {
		return err
	}

	// register newHeads broadcaster to use with connected clients later. At this point the broadcaster is active and waiting for new subscribers to join.
	server.logsBroadcaster = broadcaster
	server.log.Info().Msg("newLogs broadcaster sources registered")
	return nil
}

// start listening to incoming subscription connections
func (server *Server) StartServer(port string) {
	http.HandleFunc("/", server.wsEndpoint)
	fmt.Println(http.ListenAndServe(":"+port, nil))
}

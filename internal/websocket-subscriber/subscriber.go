package proxy

import (
	"fmt"
	"context"

	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
)

const (
	defaultChanLen = 0
)

type Transaction struct {
	signature string
}

type TransactionProxy struct {
	cfg             *TransactionProxyConfig
	ctx             context.Context
	logger          logger.Logger
	subscriberError chan error
}

func NewTransactionProxy(
	cfg *TransactionProxyConfig,
	ctx context.Context,
	log logger.Logger,
) *TransactionProxy {
	return &TransactionProxy{
		cfg:             cfg,
		ctx:             ctx,
		logger:          log,
		subscriberError: make(chan error, 0),
	}
}

func (s *TransactionProxy) Run() error {
	// start socket server for enabling users to subscribe to transactions
	webServer := NewServer(s.ctx)

	// subscribe to transactions from node
	go s.subscribeToTransactions(webServer)

	// start broadcasting
	go webServer.StartBroadcaster()

	// start server
	go webServer.StartServer(s.cfg.websocketPort)

	return nil
}

func (s *TransactionProxy) subscribeToTransactions(server *Server) {
	// connect to running solana websocket and create client
	client, err := ws.Connect(s.ctx, s.cfg.solanaWebsocketEndpoint)
	if err != nil {
		fmt.Println(err)
		return
	}

	// subscribe to "all"  transactions that are "finalized"
	subscription, err := client.LogsSubscribe("all", "finalized")
	if err != nil {
		fmt.Println(err)
		return
	}

	// subscribe to every result coming into the channel
	for {
		result, err := subscription.Recv()
		if err != nil {
			server.sourceError <- err
			fmt.Println(err)
			return
		} else {
			server.source <- Transaction{signature: result.Value.Signature.String()}
		}
	}
}

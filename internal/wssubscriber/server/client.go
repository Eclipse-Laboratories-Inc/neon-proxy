package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber/broadcaster"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
)

const (
	// Time allowed to write a message to the peer.
	deadline = 3 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 4096

	// rpc version
	rpcVersion = "2.0"

	// subscription method
	methodSubscription     = "eth_subscribe"
	methodSubscriptionName = "eth_subscription"
	methodUnsubscription   = "eth_unsubscribe"

	subscriptionNewHeads               = "newHeads"
	subscriptionLogs                   = "logs"
	subscriptionNewPendingTransactions = "newPendingTransactions"

	//error codes
	UnmarshalingErrorCode = -1
	MethodNotFoundErrorCode = -32601
	RequestIDErrorCode = -2
	IncorrectNumberOfParametersErrorCode = -3
	IncorrectParameterTypeErrorCode = -4
	SubscriptionAlreadyActiveErrorCode = -5
	IncorrectFilterErrorMessage = -6
)

// defining each connection parameters
type Client struct {
	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	clientResponseBuffer chan []byte

	// logger
	log logger.Logger

	// client closer once
	closeOnlyOnce sync.Once

	// head broadcaster instance
	newHeadsBroadcaster   *broadcaster.Broadcaster
	newHeadsSource        chan interface{}
	newHeadsLocker        sync.Mutex
	newHeadsIsActive      bool
	newHeadSubscriptionID string

	//pending transaction broadcaster instance
	pendingTransactionsBroadcaster    *broadcaster.Broadcaster
	pendingTransactionsSource         chan interface{}
	pendingTransactionsLocker         sync.Mutex
	pendingTransactionsIsActive       bool
	pendingTransactionsSubscriptionID string

	//logs broadcaster instance
	newLogsBroadcaster    *broadcaster.Broadcaster
	newLogsSource         chan interface{}
	newLogsLocker         sync.Mutex
	newLogsIsActive       bool
	newLogsSubscriptionID string
	logsFilters           *logsFilters
}

// json object sent back to the client
type ClientResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		Subscription string          `json:"subscription"`
		Result       json.RawMessage `json:"result"`
	} `json:"params"`
}

// subscription request
type SubscribeJsonRPC struct {
	Method string        `json:"method"`
	ID     uint64        `json:"id"`
	Params []interface{} `json:"params"`
}

type logsFilters struct {
	Addresses []string
	Topics    [][]string
}

type SubscribeLogsFilterParams struct {
	Addresses []string      `json:"addresses"`
	Address   string        `json:"address"`
	Topics    []interface{} `json:"topics"`
}


// defines error returned to user
type SubscriptionError struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

// subscription response from websocket
type SubscribeJsonResponseRCP struct {
	Version string `json:"jsonrpc"`
	ID      uint64 `json:"id"`
	Result  string `json:"result,omitempty"`
	Error   *SubscriptionError `json:"error,omitempty"`
}

// event type defines the data sent to the subscriber each time new event is caught
type Event struct {
	Version string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		Subscription string      `json:"subscription"`
		Result       interface{} `json:"result"`
	} `json:"params"`
}

// create new client when connecting
func NewClient(conn *websocket.Conn, log logger.Logger,
	headBroadcaster *broadcaster.Broadcaster, pendingTxBroadcaster *broadcaster.Broadcaster,
	newLogsBroadcaster *broadcaster.Broadcaster) *Client {
	return &Client{
		conn:                           conn,
		log:                            log,
		clientResponseBuffer:           make(chan []byte, 256),
		newHeadsBroadcaster:            headBroadcaster,
		pendingTransactionsBroadcaster: pendingTxBroadcaster,
		newLogsBroadcaster:             newLogsBroadcaster}
}

// readPump pumps messages from the websocket connection.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() {
	// close connection upon error
	defer c.conn.Close()

	c.conn.SetReadLimit(maxMessageSize)
	for {
		// read next request
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log.Error().Err(err).Msg(fmt.Sprintf("error: %v", err))
			}
			c.Close()
			break
		}

		// process request
		response := c.ProcessRequest(bytes.TrimSpace(bytes.Replace(message, []byte{'\n'}, []byte{' '}, -1)))
		res, err := json.Marshal(response)

		// check json marshaling error
		if err != nil {
			c.log.Error().Err(err).Msg(fmt.Sprintf("marshalling response output: %v", err))
			continue
		}

		c.clientResponseBuffer <- res
	}
}

// based on request data we determine what kind of subscription it is and make specific subscription for client (or unsubscribe)
func (c *Client) ProcessRequest(request []byte) (responseRPC SubscribeJsonResponseRCP) {
	// prepare response
	responseRPC.Version = rpcVersion

	// unmarshal request
	var requestRPC SubscribeJsonRPC
	if err := json.Unmarshal(request, &requestRPC); err != nil {
		responseRPC.Error = &SubscriptionError{Code: UnmarshalingErrorCode, Message: "Input should be in JSON format"}
		return
	}

	// set corresponding response id
	responseRPC.ID = requestRPC.ID

	// check rpc version
	if requestRPC.Method != methodSubscription && requestRPC.Method != methodUnsubscription {
		responseRPC.Error = &SubscriptionError{Code: MethodNotFoundErrorCode, Message: "The method " + requestRPC.Method + " does not exist/is not available"}
		return
	}

	// check request id to be valid
	if requestRPC.ID == 0 {
		responseRPC.Error = &SubscriptionError{Code: RequestIDErrorCode, Message: "id must be greater than 0"}
		return
	}

	// check params
	if len(requestRPC.Params) < 1 {
		responseRPC.Error = &SubscriptionError{Code: IncorrectNumberOfParametersErrorCode, Message: "Incorrect subscription parameters"}
		return
	}

	// check subscription type is correct
	if reflect.TypeOf(requestRPC.Params[0]).Name() != "string" {
		responseRPC.Error = &SubscriptionError{Code: IncorrectParameterTypeErrorCode, Message: "Incorrect first parameter type"}
		return
	}

	// activate subscription based on type
	switch {
	case requestRPC.Method == methodUnsubscription:
		c.unsubscribe(requestRPC, &responseRPC)
	case requestRPC.Params[0].(string) == subscriptionNewHeads:
		c.subscribeToNewHeads(requestRPC, &responseRPC)
	case requestRPC.Params[0].(string) == subscriptionLogs:
		c.subscribeToNewLogs(requestRPC, &responseRPC)
	case requestRPC.Params[0].(string) == subscriptionNewPendingTransactions:
		c.subscribeToNewPendingTransactions(requestRPC, &responseRPC)
	default:
		responseRPC.Error = &SubscriptionError{Code: MethodNotFoundErrorCode, Message: "The method " + requestRPC.Method + " does not exist/is not available"}
		return
	}

	return responseRPC
}

// writePump pumps messages from the client.
//
// A goroutine running writePump is started for each connection.
func (c *Client) WritePump() {
	defer c.Close()
	for {
		select {
		case message, ok := <-c.clientResponseBuffer:
			c.conn.SetWriteDeadline(time.Now().Add(deadline))
			if !ok {
				// channel closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// create new writer
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// send next data chunk
			w.Write(message)

			// close current writer
			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func (c *Client) unsubscribe(requestRPC SubscribeJsonRPC, responseRPC *SubscribeJsonResponseRCP) {
	// get subscription id to cancel subscription
	subscriptionID := requestRPC.Params[0].(string)

	// protect client vars
	c.newHeadsLocker.Lock()
	defer c.newHeadsLocker.Unlock()

	// check for empty subscription id
	if subscriptionID == "" {
		responseRPC.Error = &SubscriptionError{Code: MethodNotFoundErrorCode, Message: "Subscription not found"}
		return
	}

	// unsubscribe
	if c.newHeadSubscriptionID == subscriptionID {
		c.newHeadsBroadcaster.CancelSubscription(c.newHeadsSource)
		responseRPC.Result = "true"
		responseRPC.ID = requestRPC.ID
		c.newHeadSubscriptionID = ""
		c.newHeadsIsActive = false
		return
	}

	// protect client vars
	c.pendingTransactionsLocker.Lock()
	defer c.pendingTransactionsLocker.Unlock()

	// unsubscribe
	if c.pendingTransactionsSubscriptionID == subscriptionID {
		c.pendingTransactionsBroadcaster.CancelSubscription(c.pendingTransactionsSource)
		responseRPC.Result = "true"
		responseRPC.ID = requestRPC.ID
		c.pendingTransactionsSubscriptionID = ""
		c.pendingTransactionsIsActive = false
		return
	}

	responseRPC.Error = &SubscriptionError{Code: MethodNotFoundErrorCode, Message: "Subscription not found"}
	return
}

// closing client connection unsubscribes everything and closes connection, cancelling subscription is safe even if we hadn't subscribed
func (c *Client) Close() {
	c.closeOnlyOnce.Do(func() {
		c.conn.Close()
		c.newHeadsBroadcaster.CancelSubscription(c.newHeadsSource)
		c.pendingTransactionsBroadcaster.CancelSubscription(c.pendingTransactionsSource)
		c.newLogsBroadcaster.CancelSubscription(c.newLogsSource)
		close(c.clientResponseBuffer)
	})
}

package wssubscriber

import (
  "log"
  "time"
  "fmt"
  "sync"
  "bytes"
  "reflect"
  "encoding/json"

  "github.com/gorilla/websocket"
  "github.com/neonlabsorg/neon-proxy/pkg/logger"
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
  newHeadsBroadcaster *Broadcaster
  newHeadsSource chan interface{}
  newHeadsLocker sync.Mutex
  newHeadsIsActive bool
  newHeadSubscriptionID string

  //pending transaction broadcaster instance
  pendingTransactionsBroadcaster *Broadcaster
  pendingTransactionsSource chan interface{}
  pendingTransactionsLocker sync.Mutex
  pendingTransactionsIsActive bool
  pendingTransactionsSubscriptionID string
}

// json object sent back to the client
type ClientResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		Subscription string `json:"subscription"`
		Result       json.RawMessage `json:"result"`
	} `json:"params"`
}

const (
  // Time allowed to write a message to the peer.
  deadline = 3 * time.Second

  // Maximum message size allowed from peer.
  maxMessageSize = 4096

  // rpc version
  rpcVersion = "2.0"

  // subscription method
  methodSubscription = "eth_subscribe"
  methodSubscriptionName = "eth_subscription"
  methodUnsubscription = "eth_unsubscribe"

  subscriptionNewHeads = "newHeads"
  subscriptionLogs = "logs"
  subscriptionNewPendingTransactions = "newPendingTransactions"
)

// subscription request
type SubscribeJsonRPC struct {
  Method  string        `json:"method"`
  ID      uint64        `json:"id"`
  Params  []interface{} `json:"params"`
}

// subscription response from websocket
type SubscribeJsonResponseRCP struct {
  Version string `json:"json_rpc"`
  ID      uint64 `json:"id"`
  Result  string `json:"result,omitempty"`
  Error   string `json:"error,omitempty"`
}

// event type defines the data sent to the subscriber each time new event is caught
type Event struct {
  Version string `json:"json_rpc"`
  Method  string `json:"method"`
  Params struct {
    Subscription string      `json:"subscription"`
    Result       interface{} `json:"result"`
  } `json:"params"`
}

// create new client when connecting
func NewClient(conn *websocket.Conn, log logger.Logger, headBroadcaster *Broadcaster, pendingTxBroadcaster *Broadcaster) *Client {
  return &Client{conn: conn, log: log, clientResponseBuffer: make(chan []byte, 256), newHeadsBroadcaster: headBroadcaster, pendingTransactionsBroadcaster: pendingTxBroadcaster}
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
				log.Printf("error: %v", err)
			}
      c.Close()
			break
		}

    // process request
		response := c.ProcessRequest(bytes.TrimSpace(bytes.Replace(message, []byte{'\n'}, []byte{' '}, -1)));
    res, _ := json.Marshal(response)
    c.sendMessage(res)
	}
}

// based on request data we determine what kind of subscription it is and make specific subscription for client (or unsubscribe)
func (c *Client) ProcessRequest(request []byte) (responseRPC SubscribeJsonResponseRCP) {
  // prepare response
  responseRPC.Version = rpcVersion

  // unmarshal request
  var requestRPC SubscribeJsonRPC
  if err := json.Unmarshal(request, &requestRPC); err != nil {
    responseRPC.Error = err.Error()
    return
  }

  // set corresponding response id
  responseRPC.ID = requestRPC.ID

  // check rpc version
  if requestRPC.Method != methodSubscription &&  requestRPC.Method != methodUnsubscription {
    responseRPC.Error = "method incorrect"
    return
  }

  // check request id to be valid
  if requestRPC.ID == 0 {
    responseRPC.Error = "id must be greater than 0"
    return
  }

  // check params
  if len(requestRPC.Params) < 1 {
    responseRPC.Error = "Incorrect subscription parameters"
    return
  }

  // check subscription type is correct
  if reflect.TypeOf(requestRPC.Params[0]).Name() != "string" {
    responseRPC.Error = "Incorrect parameter 0"
    return
  }

  // activate subscription based on type
  switch {
  case requestRPC.Method == methodUnsubscription:
      c.unsubscribe(requestRPC, &responseRPC)
  case requestRPC.Params[0].(string) == subscriptionNewHeads:
      c.subscribeToNewHeads(requestRPC, &responseRPC)
  case requestRPC.Params[0].(string) == subscriptionLogs:
      c.subscribeNewLogs(requestRPC, &responseRPC)
  case requestRPC.Params[0].(string) == subscriptionNewPendingTransactions:
      c.subscribeNewPendingTransactions(requestRPC, &responseRPC)
  default:
      responseRPC.Error = "subscription type not found"
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

// send message back to subscriber
func (c *Client) sendMessage(message []byte) {
  c.conn.SetWriteDeadline(time.Now().Add(deadline))

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


func (c *Client) unsubscribe(requestRPC SubscribeJsonRPC, responseRPC *SubscribeJsonResponseRCP) {
  // get subscription id to cancel subscription
  subscriptionID := requestRPC.Params[0].(string)

  // protect client vars
  c.newHeadsLocker.Lock()
  defer c.newHeadsLocker.Unlock()

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

  responseRPC.Error = "Subscription not found"
  return
}

func (c *Client) subscribeNewLogs(requestRPC SubscribeJsonRPC, responseRPC *SubscribeJsonResponseRCP) {

}

func (c *Client) subscribeNewPendingTransactions(requestRPC SubscribeJsonRPC, responseRPC *SubscribeJsonResponseRCP) {
  // if new head subscription is active skip another subscription
  c.pendingTransactionsLocker.Lock()
  defer c.pendingTransactionsLocker.Unlock()

  // check if subscription type for the client is active
  if c.pendingTransactionsIsActive {
    responseRPC.Error = "pendingTransactions subscription already active. Subscription ID: " + c.pendingTransactionsSubscriptionID
    return
  }

  // if not subscribe to broadcaster
  c.pendingTransactionsSource = c.pendingTransactionsBroadcaster.Subscribe()

  // generate subscription id
  responseRPC.Result = NewID()
  responseRPC.ID = requestRPC.ID

  // register subscription id for client
  c.pendingTransactionsSubscriptionID = responseRPC.Result
  c.pendingTransactionsIsActive = true
  go c.CollectPendingTransactions()
}

func (c *Client) CollectPendingTransactions() {
  // listen for incoming pending transactions and send to user
  for  {
      select {
      case tx, ok := <- c.pendingTransactionsSource:
        //channel has been closed
        if ok == false {
          return
        }
        // case when response to subscribe request isn't sent yet
        c.pendingTransactionsLocker.Lock()
        if c.pendingTransactionsIsActive == false {
          c.pendingTransactionsLocker.Unlock()
          continue
        }

        // construct response object for new event
        var clientResponse ClientResponse
        clientResponse.Jsonrpc = rpcVersion
        clientResponse.Method = methodSubscriptionName
        clientResponse.Params.Subscription = c.pendingTransactionsSubscriptionID
        clientResponse.Params.Result = []byte("\"" + tx.(string) + "\"")
        c.pendingTransactionsLocker.Unlock()

        // marshal to send it as a json
        response, _ := json.Marshal(clientResponse)
        c.clientResponseBuffer <- response
      }
  }
}


func (c *Client) subscribeToNewHeads(requestRPC SubscribeJsonRPC, responseRPC *SubscribeJsonResponseRCP) {
  // if new head subscription is active skip another subscription
  c.newHeadsLocker.Lock()
  defer c.newHeadsLocker.Unlock()

  // check if subscription type for the client is active
  if c.newHeadsIsActive {
    responseRPC.Error = "newHeads subscription already active. Subscription ID: " + c.newHeadSubscriptionID
    return
  }

  // if not subscribe to broadcaster
  c.newHeadsSource = c.newHeadsBroadcaster.Subscribe()
  fmt.Println(c.newHeadsSource)
  // generate subscription id
  responseRPC.Result = NewID()
  responseRPC.ID = requestRPC.ID

  // register subscription id for client
  c.newHeadSubscriptionID = responseRPC.Result
  c.newHeadsIsActive = true
  c.log.Info().Msg("NewHeads subscription succeeded with ID: " + responseRPC.Result)
  go c.CollectNewHeads()
}

func (c *Client) CollectNewHeads() {
  // listen for incoming heads and send to user
  for  {
      select {
      case newHead, ok := <- c.newHeadsSource:
        c.log.Info().Msg("new head to be subscribed")
        //channel has been closed
        if ok == false {
          return
        }
        // case when subscription response isn't sent yet
        c.newHeadsLocker.Lock()
        if c.newHeadsIsActive == false {
          c.newHeadsLocker.Unlock()
          continue
        }

        // construct response object for new event
        var clientResponse ClientResponse
        clientResponse.Jsonrpc = rpcVersion
        clientResponse.Method = methodSubscriptionName
        clientResponse.Params.Subscription = c.newHeadSubscriptionID
        clientResponse.Params.Result = newHead.([]byte)
        c.newHeadsLocker.Unlock()

        // marshal to send it as a json
        response, _ := json.Marshal(clientResponse)
        c.clientResponseBuffer <- response
      }
  }
}

func (c *Client) Close() {
  c.closeOnlyOnce.Do(func() {
    c.conn.Close()
    c.newHeadsBroadcaster.CancelSubscription(c.newHeadsSource)
    c.pendingTransactionsBroadcaster.CancelSubscription(c.pendingTransactionsSource)
    close(c.clientResponseBuffer)
 	})
}

package wssubscriber

import (
  "log"
  "time"
  "sync"
  "bytes"
  "reflect"
  "encoding/json"
  "github.com/gorilla/websocket"
)

type Client struct {
  // The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	clientResponseBuffer chan []byte

  // client locker
  cMu sync.Mutex

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

const (
	// Time allowed to write a message to the peer.
	deadline = 3 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 4096

  // rpc version
  rpcVersion = "2.0"

  // subscription method
  methodSubscription = "eth_subscription"
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
func NewClient(conn *websocket.Conn, headBroadcaster *Broadcaster, pendingTxBroadcaster *Broadcaster) *Client {
  return &Client{conn: conn, clientResponseBuffer: make(chan []byte, 256), newHeadsBroadcaster: headBroadcaster, pendingTransactionsBroadcaster: pendingTxBroadcaster}
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
			break
		}

    // process request
		response := c.ProcessSubscription(bytes.TrimSpace(bytes.Replace(message, []byte{'\n'}, []byte{' '}, -1)));
    res, _ := json.Marshal(response)
    c.sendMessage(res)
	}
}

// unmarshall subscription request
func (c *Client) ProcessSubscription(request []byte) (responseRPC SubscribeJsonResponseRCP) {
  // prepare response
  responseRPC.Version = rpcVersion

  // unmarshall client request
  var requestRPC SubscribeJsonRPC
  if err := json.Unmarshal(request, &requestRPC); err != nil {
    responseRPC.Error = err.Error()
  }

  // set corresponding response id
  responseRPC.ID = requestRPC.ID

  // check rpc version
  if requestRPC.Method != methodSubscription &&  requestRPC.Method != methodUnsubscription {
    responseRPC.Error = "method incorrect"
  }

  // check request id to be valid
  if requestRPC.ID == 0 {
    responseRPC.Error = "id must be greater than 0"
  }

  // check params
  if len(requestRPC.Params) < 1 {
    responseRPC.Error = "Incorrect subscription parameters"
  }

  // check subscription type is correct
  if reflect.TypeOf(requestRPC.Params[0]).Name() != "string" {
    responseRPC.Error = "Incorrect parameter 0"
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
  }

  return responseRPC
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() {
	defer c.conn.Close()

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

}


func (c *Client) subscribeToNewHeads(requestRPC SubscribeJsonRPC, responseRPC *SubscribeJsonResponseRCP) {
  // if new head subscription is active skip another subscription
  c.newHeadsLocker.Lock()
  defer c.newHeadsLocker.Unlock()

  // check if subscription type for the client is active
  if c.newHeadsIsActive {
    responseRPC.Error = "newHeads subscription already active"
    return
  }

  // if not subscribe to broadcaster
  c.newHeadsSource = c.newHeadsBroadcaster.Subscribe()

  // generate subscription id
  responseRPC.Result = NewID()
  responseRPC.ID = requestRPC.ID

  // register subscription id for client
  c.newHeadSubscriptionID = responseRPC.Result
  go c.CollectNewHeads()
}

func (c *Client) CollectNewHeads() {
  // listen for incoming heads and send to user
  for  {
      select {
      case newHead, ok := <- c.newHeadsSource:
        //channel has been closed
        if ok == false {
          return
        }
        // case when subscription response isn't sent yet
        c.cMu.Lock()
        if c.newHeadsIsActive == false {
          c.cMu.Unlock()
          continue
        }
        c.cMu.Unlock()

        // encode new head and push it into send buffer
        jsonNewHead, _ := json.Marshal(newHead)
        c.clientResponseBuffer <- jsonNewHead
      }
  }
}

func (c *Client) Close() {

}

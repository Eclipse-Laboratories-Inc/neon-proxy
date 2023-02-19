package wssubscriber

import (
  "time"
  "bytes"
  "errors"
  "context"
  "strconv"
  "net/http"
  "io/ioutil"
  "encoding/json"
)

const (
  slotWasSkippedErrorCode = -32007
)

// define block header that we broadcast to users
type BlockHeader struct {
	Result  struct {
		BlockHeight       int    `json:"blockHeight"`
		BlockTime         int    `json:"blockTime"`
		Blockhash         string `json:"blockhash"`
		ParentSlot        int    `json:"parentSlot"`
		PreviousBlockhash string `json:"previousBlockhash"`
	} `json:"result"`
  Error   *struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
  } `json:"error,omitempty"`
}

func RegisterNewHeadBroadcasterSources(ctx *context.Context, solanaWebsocketEndpoint string, broadcaster chan interface{}, broadcasterErr chan error) (error){
  // subscribe to every result coming into the channel
  go func() {
    // store latest processed block slot
    var latestProcessedBlockSlot uint64

    // subscribe to every result coming into the channel
  	for {
      // Calling Sleep method
      time.Sleep(1 * time.Second)

      // get latest block number
      slotResponse, err := jsonRPC([]byte(`{"jsonrpc":"2.0","id":1, "method":"getSlot", "params":[{"commitment":"finalized"}]}`), solanaWebsocketEndpoint, "POST")
      if err != nil {
        broadcasterErr <- err
        continue
      }

      // declare response json
      var blockSlot  struct {
      	Jsonrpc string `json:"jsonrpc"`
      	Result  uint64    `json:"result"`
      	ID      int    `json:"id"`
      	Error   struct {
      		Code    int    `json:"code"`
      		Message string `json:"message"`
      	} `json:"error"`
      }


      // unrmarshal latest block slot
      err = json.Unmarshal([]byte(slotResponse), &blockSlot)
    	if err != nil {
        broadcasterErr <- err
        continue
    	}

      // error from rpc
      if blockSlot.Error.Message != "" {
        broadcasterErr <- err
        continue
      }

      // check unmarshaling error
      if err != nil {
        broadcasterErr <- err
        continue
      }

      // check the first iteration
      if latestProcessedBlockSlot == 0 {
        latestProcessedBlockSlot = blockSlot.Result
        continue
      }

      err = processBlocks(ctx, solanaWebsocketEndpoint, broadcaster, broadcasterErr, &latestProcessedBlockSlot, blockSlot.Result)
      // check unmarshaling error
      if err != nil {
        broadcasterErr <- err
        continue
      }
    }
  }()

  return nil
}

// request each block from "from" to "to" from the rpc endpoint and broadcast to user
func processBlocks(ctx *context.Context, solanaWebsocketEndpoint string, broadcaster chan interface{}, broadcasterErr chan error, from *uint64, to uint64) error {
  // process each block sequentially, broadcasting to peers
  for  *from < to {
    // get block with given slot
    response, err := jsonRPC([]byte(`{
      "jsonrpc": "2.0","id":1,
      "method":"getBlock",
      "params": [
        ` + strconv.FormatUint(*from, 10) + `,
        {
          "encoding": "json",
          "maxSupportedTransactionVersion":0,
          "transactionDetails":"none", "commitment" :"finalized",
          "rewards":false
        }
      ]
    }`), solanaWebsocketEndpoint, "POST")

    // check err
    if err != nil {
      return err
    }

    var blockHeader BlockHeader

    // unrmarshal latest block slot
    err = json.Unmarshal([]byte(response), &blockHeader)
    if err != nil {
      return err
    }

    // check rpc error
    if blockHeader.Error != nil {
      // check if the given slot was skipped if so we jump over it and continue processing blocks from the next slot
      if blockHeader.Error.Code == slotWasSkippedErrorCode {
        *from++
        continue
      }
      return errors.New(blockHeader.Error.Message)
    } else {
      // insert new block to be broadcasted
      clientResponse, _ := json.Marshal(blockHeader.Result)
      broadcaster <- clientResponse
      *from++
    }
  }

  return nil
}

// execute rpc call to the solana rpc endpoint
func jsonRPC(jsonStr []byte, url string, requestType string) ([]byte, error) {
	req, err := http.NewRequest(requestType, url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return body, nil
}

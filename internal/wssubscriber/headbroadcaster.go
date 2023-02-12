package wssubscriber

import (
  "context"
  "github.com/gagliardetto/solana-go/rpc/ws"
)

func RegisterNewHeadBroadcasterSources(ctx *context.Context, solanaWebsocketEndpoint string, broadcaster chan interface{}, broadcasterErr chan error) (error){
	// connect to running solana websocket and create client
	client, err := ws.Connect(*ctx, solanaWebsocketEndpoint)
	if err != nil {
		return err
	}

	// subscribe to "all"  transactions that are "finalized"
	subscription, err := client.BlockSubscribe(ws.NewBlockSubscribeFilterAll(), &ws.BlockSubscribeOpts{})
	if err != nil {
		return err
	}

  // subscribe to every result coming into the channel
  go func() {
    // subscribe to every result coming into the channel
  	for {
  		result, err := subscription.Recv()
  		if err != nil {
  			broadcasterErr <- err
  			return
  		} else {
  			broadcaster <- result
  		}
  	}
  }()

  return nil
}

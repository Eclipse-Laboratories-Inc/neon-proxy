package wssubscriber

import (
  "time"
  "context"

  "github.com/neonlabsorg/neon-proxy/pkg/logger"
)

// RegisterNewHeadBroadcasterSources passes data and error channels where new incoming data (block heads) will be pushed and redirected to broadcaster
func RegisterPendingTransactionBroadcasterSources(ctx *context.Context, log logger.Logger, solanaWebsocketEndpoint string, broadcaster chan interface{}, broadcasterErr chan error) (error){
  log.Info().Msg("pending transaction pulling from mempool started ... ")

  var fakemempool = [...]string {"3vCdSdwwzRgasgrj17Wc4PzMUwzujcen1W432Qe9wm2fFnqz1jQbm92cVgLnsaE7vtgSb1CCiCPWyhac5sJgkyNY",
    "3asEtuKPUKKCpvHWxatGxXzWGyJZauQu1Gb2Z8goMvem8UE62phiTs7P3sRxfvCWiijmCtcsytayFeNbAmaHDDrz",
    "5iKYFy3DriePzNWrETZKTAhF7x4zhHNvrMAZbsvHJGysUm2K8m25DV5eToe4kbPjEpNrS74WL2DezwTEU7hTPXb7",
    "2GbYuCAahhFBZiSAn9vNhE3ALZcudfyUKKYusUQ6DhRQQxZFAmSB3ptGZPyoMUu9GuAQ4afCrEBvcKqvXzjXBZCR"}

  // subscribe to every result coming into the channel
  go func() {
    // subscribe to every result coming into the channel
  	for k := 0; k <= 1000000; k++ {
      // Calling Sleep method
      time.Sleep(1 * time.Second)

      broadcaster <- fakemempool[k%4]
    }
  }()

  return nil
}

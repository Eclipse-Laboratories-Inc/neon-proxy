package main

import (
	"fmt"

	"github.com/neonlabsorg/neon-proxy/pkg/service"
	websocket "github.com/neonlabsorg/neon-proxy/internal/websocket-subscriber"
	"github.com/neonlabsorg/neon-proxy/pkg/service/configuration"
)

func main() {
	s := service.CreateService(&configuration.Config{
		Name: "proxy",
		// Storage: &configuration.ConfigStorageList{},
	})

	s.AddHandler(runTransactionProxy)

	s.Run()
}

func runTransactionProxy(s *service.Service) {
	cfg, err := websocket.CreateConfigFromEnv()
	if err != nil {
		panic(err)
	}

	transactionSubscriber := websocket.NewTransactionProxy(
		cfg,
		s.GetContext(),
		s.GetLogger(),
	)

	err = transactionSubscriber.Run()
	if err != nil {
		fmt.Println(err)
	}
}

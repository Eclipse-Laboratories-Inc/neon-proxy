package main

import (
	"fmt"

	"github.com/neonlabsorg/neon-proxy/pkg/service"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber"
	"github.com/neonlabsorg/neon-proxy/pkg/service/configuration"
)

func main() {
	s := service.CreateService(&configuration.Config{
		Name: "wssubscriber",
		// Storage: &configuration.ConfigStorageList{},
	})

	s.AddHandler(runTransactionProxy)

	s.Run()
}

func runTransactionProxy(s *service.Service) {
	cfg, err := wssubscriber.CreateConfigFromEnv()
	if err != nil {
		panic(err)
	}

	transactionSubscriber := wssubscriber.NewTransactionSubscriber(
		cfg,
		s.GetContext(),
		s.GetLogger(),
	)

	err = transactionSubscriber.Run()
	if err != nil {
		fmt.Println(err)
	}
}

package main

import (
	"github.com/neonlabsorg/neon-proxy/internal/subscriber"
	"github.com/neonlabsorg/neon-proxy/pkg/service"
	"github.com/neonlabsorg/neon-proxy/pkg/service/configuration"
)

func main() {
	s := service.CreateService(&configuration.Config{
		Name: "subscriber",
		// Storage: &configuration.ConfigStorageList{},
	})

	s.AddHandler(runTransactionSubscriber)

	s.Run()
}

func runTransactionSubscriber(s *service.Service) {
	cfg, err := subscriber.CreateFromEnv()
	if err != nil {
		panic(err)
	}

	transactionSubscriber := subscriber.NewTransactionSubscriber(
		cfg,
		s.GetContext(),
		s.GetSolanaRpcClient(),
		s.GetLogger(),
	)

	transactionSubscriber.RunWithLongPoling()
}

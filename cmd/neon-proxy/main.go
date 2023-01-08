package main

import (
	"fmt"
	
	"github.com/neonlabsorg/neon-proxy/pkg/service"
	"github.com/neonlabsorg/neon-proxy/internal/proxy"
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
	cfg, err := proxy.CreateConfigFromEnv()
	if err != nil {
		panic(err)
	}

	transactionProxy := proxy.NewTransactionProxy(
		cfg,
		s.GetContext(),
		s.GetLogger(),
	)

	err = transactionProxy.Run()
	if err != nil {
		fmt.Println(err)
	}
}

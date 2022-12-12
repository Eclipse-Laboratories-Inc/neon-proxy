package main

import (
	"github.com/neonlabsorg/neon-proxy/pkg/service"
	"github.com/neonlabsorg/neon-proxy/pkg/service/configuration"
)

func main() {
	s := service.CreateService(&configuration.Config{
		Name: "mempool",
	})

	s.AddHandler(runMempool)

	s.Run()
}

func runMempool(s *service.Service) {

}

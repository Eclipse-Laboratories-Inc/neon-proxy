package main

import (
	"github.com/neonlabsorg/neon-proxy/internal/mempool"
	"github.com/neonlabsorg/neon-service-framework/pkg/service"
	"github.com/neonlabsorg/neon-service-framework/pkg/service/configuration"
)

func main() {
	s := service.CreateService(&configuration.Config{
		Name: "mempool",
	})

	s.AddHandler(mempool.ServiceHandler)

	s.Run()
}

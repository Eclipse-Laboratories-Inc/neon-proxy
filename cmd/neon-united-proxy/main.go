package main

import (
	"github.com/neonlabsorg/neon-proxy/internal/indexer"
	"github.com/neonlabsorg/neon-proxy/internal/mempool"
	"github.com/neonlabsorg/neon-proxy/internal/proxy"
	"github.com/neonlabsorg/neon-proxy/internal/wssubscriber"
	"github.com/neonlabsorg/neon-service-framework/pkg/service"
	"github.com/neonlabsorg/neon-service-framework/pkg/service/configuration"
)

func main() {
	s := service.CreateService(&configuration.Config{
		Name: "united_proxy",
		Storage: &configuration.ConfigStorageList{
			Postgres: []string{
				"indexer",
			},
		},
	})

	s.AddHandler(indexer.ServiceHandler)
	s.AddHandler(mempool.ServiceHandler)
	s.AddHandler(proxy.ServiceHandler)
	s.AddHandler(wssubscriber.ServiceHandler)

	s.Run()
}

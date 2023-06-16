package main

import (
	"github.com/neonlabsorg/neon-proxy/internal/proxy"
	"github.com/neonlabsorg/neon-service-framework/pkg/service"
	"github.com/neonlabsorg/neon-service-framework/pkg/service/configuration"
)

func main() {
	s := service.CreateService(&configuration.Config{
		Name: "proxy",
		Storage: &configuration.ConfigStorageList{
			Postgres: []string{
				"indexer",
			},
		},
	})

	s.AddHandler(proxy.ServiceHandler)

	s.Run()
}

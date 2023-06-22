package main

import (
	"github.com/neonlabsorg/neon-proxy/internal/indexer"
	"github.com/neonlabsorg/neon-service-framework/pkg/service"
	"github.com/neonlabsorg/neon-service-framework/pkg/service/configuration"
)

func main() {
	s := service.CreateService(&configuration.Config{
		Name: "indexer",
		Storage: &configuration.ConfigStorageList{
			Postgres: []string{
				"indexer",
			},
		},
	})

	s.AddHandler(indexer.ServiceHandler)

	s.Run()
}

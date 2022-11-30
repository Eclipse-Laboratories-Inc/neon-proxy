package main

import (
	"github.com/neonlabsorg/neon-proxy/pkg/service"
	"github.com/neonlabsorg/neon-proxy/pkg/service/configuration"
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

	s.AddHandler(runIndexer)

	s.Run()
}

func runIndexer(s *service.Service) {

}

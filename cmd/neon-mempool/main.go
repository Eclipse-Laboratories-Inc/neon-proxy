package main

import (
	"github.com/neonlabsorg/neon-proxy/pkg/service"
	"github.com/neonlabsorg/neon-proxy/pkg/service/configuration"
)

func main() {
	s := service.CreateService(&configuration.Config{
		Name: "mempool",
		// Storage: &configuration.ConfigStorageList{
		// 	Postgres: []string{
		// 		"indexer",
		// 	},
		// },
	})

	s.AddHandler(runMempool)

	s.Run()
}

func runMempool(s *service.Service) {

}

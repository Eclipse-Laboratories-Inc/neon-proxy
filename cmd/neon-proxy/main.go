package main

import (
	"github.com/neonlabsorg/neon-proxy/pkg/service"
	"github.com/neonlabsorg/neon-proxy/pkg/service/configuration"
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

	s.AddHandler(runProxy)

	s.Run()
}

func runProxy(s *service.Service) {

}

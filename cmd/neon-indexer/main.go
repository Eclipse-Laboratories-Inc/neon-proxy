package main

import (
	"github.com/neonlabsorg/neon-proxy/internal/indexer"
	"github.com/neonlabsorg/neon-proxy/pkg/service"
	"github.com/neonlabsorg/neon-proxy/pkg/service/configuration"
)

const (
	indexerServiceName = "indexer"
)

func main() {
	s := service.CreateService(&configuration.Config{
		Name: indexerServiceName,
		Storage: &configuration.ConfigStorageList{
			Postgres: []string{
				indexerServiceName,
			},
		},
	})

	s.AddHandler(runIndexer)

	s.Run()
}

func runIndexer(s *service.Service) {
	indexerDB, err := s.GetDB(indexerServiceName)
	if err != nil {
		panic(err)
	}
	db := indexerDB.GetRawDB()
	app := indexer.NewIndexerApp(s.GetContext(), s.GetLogger(), db)
	app.Run()
}

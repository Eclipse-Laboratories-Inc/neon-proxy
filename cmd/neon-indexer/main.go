package main

import (
	"github.com/neonlabsorg/neon-proxy/internal/indexer"
	indconfig "github.com/neonlabsorg/neon-proxy/internal/indexer/config"
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
	/*
		// add PoolConnector to the indexer
		indexerPool, err := s.GetPool(indexerServiceName)
		if err != nil {
			panic(err)
		}
	*/

	cfg, err := indconfig.CreateConfigFromEnv("", "")
	if err != nil {
		panic(err)
	}

	app, err := indexer.NewIndexerApp(s.GetContext(), cfg, s.GetLogger(), indexerDB.GetRawDB(), s.GetSolanaRpcClient(), 0, s.GatherStatistics())
	if err != nil {
		panic(err)
	}
	app.Run()
}

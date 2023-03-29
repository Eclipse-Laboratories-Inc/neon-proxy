package indexer

import "github.com/neonlabsorg/neon-proxy/pkg/postgres"

type SolanaSignsDB struct {
	conn *postgres.Connector
}

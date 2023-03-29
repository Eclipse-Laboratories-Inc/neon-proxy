package indexer

import "github.com/neonlabsorg/neon-proxy/pkg/postgres"

type SolanaBlocksDB struct {
	conn *postgres.Connector
}

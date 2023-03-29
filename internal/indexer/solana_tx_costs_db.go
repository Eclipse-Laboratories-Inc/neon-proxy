package indexer

import "github.com/neonlabsorg/neon-proxy/pkg/postgres"

type SolanaTxCostsDB struct {
	conn *postgres.Connector
}

package indexer

import "github.com/neonlabsorg/neon-proxy/pkg/postgres"

type SolanaNeonTxDB struct {
	conn *postgres.Connector
}

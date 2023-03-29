package indexer

import "github.com/neonlabsorg/neon-proxy/pkg/postgres"

type NeonTxsDB struct {
	conn *postgres.Connector
}

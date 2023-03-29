package indexer

import "github.com/neonlabsorg/neon-proxy/pkg/postgres"

type NeonTxLogsDB struct {
	conn *postgres.Connector
}

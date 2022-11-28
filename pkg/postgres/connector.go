package postgres

import (
	"context"
	"database/sql"

	"github.com/neonlabsorg/neon-proxy/pkg/service/configuration"
)

type Connector struct {
	DB     *sql.DB
	ctx    context.Context
	config *configuration.PostgresConfiguration
	name   string
}

func NewPostgresConnector(name string, ctx context.Context, connect *sql.DB, config *configuration.PostgresConfiguration) *Connector {
	return &Connector{
		DB:     connect,
		ctx:    ctx,
		name:   name,
		config: config,
	}
}

func (c *Connector) GetRawDB() *sql.DB {
	return c.DB
}

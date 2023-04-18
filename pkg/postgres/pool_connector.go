package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neonlabsorg/neon-proxy/pkg/service/configuration"
)

type PoolConnector struct {
	Pool   *pgxpool.Pool
	ctx    context.Context
	config *configuration.PostgresConfiguration
	name   string
}

func NewPoolConnector(name string, ctx context.Context, pool *pgxpool.Pool, config *configuration.PostgresConfiguration) *PoolConnector {
	return &PoolConnector{
		Pool:   pool,
		ctx:    ctx,
		name:   name,
		config: config,
	}
}

func (c *PoolConnector) GetPool() *pgxpool.Pool {
	return c.Pool
}

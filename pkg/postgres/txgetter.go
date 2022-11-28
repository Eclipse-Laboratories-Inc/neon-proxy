package postgres

import (
	"context"
	"database/sql"
)

type PgTxGetter struct {
	db *sql.DB
}

func NewPgTxGetter(db *sql.DB) TxGetter {
	return &PgTxGetter{
		db: db,
	}
}

func (g *PgTxGetter) GetTX() (*sql.Tx, error) {
	return g.db.Begin()
}

func (g *PgTxGetter) GetTXWithOptions(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return g.db.BeginTx(ctx, opts)
}

type TxGetter interface {
	GetTX() (*sql.Tx, error)
	GetTXWithOptions(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

package store

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PG reúne os stores em cima de um pool pgx — a aplicação injeta um único PG.
type PG struct {
	Pool *pgxpool.Pool
}

// New cria o agregado de stores Postgres.
func New(pool *pgxpool.Pool) *PG { return &PG{Pool: pool} }

func isNoRows(err error) bool { return errors.Is(err, pgx.ErrNoRows) }

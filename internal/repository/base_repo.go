package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shashisharma307703/vedantam/db/dbgen"
)

type Repository struct {
	Pool *pgxpool.Pool
	*dbgen.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		Pool:    pool,
		Queries: dbgen.New(pool),
	}
}

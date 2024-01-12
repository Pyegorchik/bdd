package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionsRepo struct {
	pool *pgxpool.Pool
}

func NewTransactionsRepo(pool *pgxpool.Pool) Transactions {
	return &TransactionsRepo{pool: pool}
}

func (r *TransactionsRepo) BeginTransaction(ctx context.Context) (Transaction, error) {
	return r.pool.Begin(ctx)
}
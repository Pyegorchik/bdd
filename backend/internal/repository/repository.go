package repository

import (
	"context"

	"github.com/Pyegorchik/bdd/backend/internal/config"
	"github.com/Pyegorchik/bdd/backend/internal/domain"
	"github.com/Pyegorchik/bdd/backend/pkg/jwtoken"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)


var ErrNoRows = pgx.ErrNoRows

type Users interface {
	GetUserById(ctx context.Context, transaction Transaction, id int64,) (*domain.User, error)
	GetUserByAddress(ctx context.Context, transaction Transaction, address string,) (*domain.User, error)
	InsertUser(ctx context.Context, transaction Transaction, user *domain.User,) (int64, error)


	InsertAuthMessage(ctx context.Context, transaction Transaction, authMsg *domain.AuthMessage) error
	GetAuthMessageByAddress(ctx context.Context, transaction Transaction, address string) (*domain.AuthMessage, error)
	DeleteAuthMessage(ctx context.Context, transaction Transaction, address string) error
}


type JWTokens interface {
	InsertJWToken(ctx context.Context, transaction Transaction, tokenData jwtoken.JWTokenData) error
	GetJWTokenNumber(ctx context.Context, transaction Transaction, id int64, role domain.Role, purpose jwtoken.Purpose) (int, error)
	GetJWTokenSecret(ctx context.Context, transaction Transaction, id int64, role domain.Role, number int, purpose jwtoken.Purpose) (string, error)
	DropJWTokens(ctx context.Context, transaction Transaction, id int64, role domain.Role, number int) error
	DropAllJWTokens(ctx context.Context, transaction Transaction, id int64, role domain.Role) error
}


type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
type Transactions interface {
	BeginTransaction(ctx context.Context) (Transaction, error)
}

type Repository struct {
	Users
	JWTokens


	Transactions
}

func NewRepository(cfg *config.Config, pool *pgxpool.Pool) (*Repository, error) {

	return &Repository{
		Users:        NewUsersRepo(),
		JWTokens:         NewJWTokensRepo(),
		Transactions:     NewTransactionsRepo(pool),
	}, nil
}
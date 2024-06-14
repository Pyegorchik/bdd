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
	GetUserById(ctx context.Context, transaction Transaction, id int64) (*domain.UserChain, error)
	GetUserByAddress(ctx context.Context, transaction Transaction, address string) (*domain.UserChain, error)
	InsertUser(ctx context.Context, transaction Transaction, user *domain.UserChain) (int64, error)

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

type Dialogs interface {
	DialogExists(ctx context.Context, transaction Transaction, userOneID int64, userTwoID int64) (bool, error)
	GetDialogByUsers(ctx context.Context, transaction Transaction, userOneID int64, userTwoID int64) (int64, error)
	CreateDialogBetweenUsers(ctx context.Context, transaction Transaction, userOneID int64, userTwoID int64, dialodID int64) error
	CreateMessageInDialog(ctx context.Context, transaction Transaction, msg *domain.Message) error
	CreateDialog(ctx context.Context, transaction Transaction, userOneID int64, userTwoID int64) (int64, error)

	GetAllDialogsByUser(ctx context.Context, transaction Transaction, userID int64) ([]*domain.DialogParticipant, error)

	GetAllMessagesWithinDialogById(ctx context.Context, transaction Transaction, dialogID int64) ([]*domain.Message, error)
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
	Dialogs

	Transactions
}

func NewRepository(cfg *config.Config, pool *pgxpool.Pool) (*Repository, error) {
	return &Repository{
		Users:        NewUsersRepo(),
		Dialogs:      NewDialogsRepo(),
		JWTokens:     NewJWTokensRepo(),
		Transactions: NewTransactionsRepo(pool),
	}, nil
}

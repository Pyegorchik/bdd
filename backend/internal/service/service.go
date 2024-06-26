package service

import (
	"context"
	"time"

	"github.com/Pyegorchik/bdd/backend/internal/config"
	"github.com/Pyegorchik/bdd/backend/internal/domain"
	"github.com/Pyegorchik/bdd/backend/internal/repository"
	"github.com/Pyegorchik/bdd/backend/models"
	"github.com/Pyegorchik/bdd/backend/pkg/hash"
	"github.com/Pyegorchik/bdd/backend/pkg/jwtoken"
	"github.com/Pyegorchik/bdd/backend/pkg/logger"
)

type Auth interface {
	GetUserById(ctx context.Context, id int64) (*domain.UserChain, error)
	GetUserByJWToken(ctx context.Context, purpose jwtoken.Purpose, token string) (*domain.UserWithTokenNumber, error)
	RefreshJWTokens(ctx context.Context, id, number int64, role domain.Role) (*models.AuthResponse, *jwtoken.JWTokenData, *jwtoken.JWTokenData, error)
	Logout(ctx context.Context, id, number int64, role domain.Role) error
	FullLogout(ctx context.Context, id int64, role domain.Role) error
	GetAuthMessage(ctx context.Context, req *models.AuthMessageRequest) (*models.AuthMessageResponse, error)
	AuthByMessage(ctx context.Context, req *models.AuthBySignatureRequest) (*models.AuthResponse, *jwtoken.JWTokenData, *jwtoken.JWTokenData, error)
}

type Dialogs interface {
	SendMessage(ctx context.Context, req *models.SendMessageRequest, userID int64) error
	GetDialogs(ctx context.Context, userID int64) ([]*models.DialogsResponseItems0, error)
	GetMessages(ctx context.Context, dialogID int64) ([]*models.MessagesResponseItems0, error)
}

type Service interface {
	Auth
	Dialogs
	Shutdown()
}

type service struct {
	Auth
	Dialogs
	stopCh chan struct{}

	cfg     *config.ServiceConfig
	logging logger.Logger
}

func NewService(
	repo *repository.Repository,
	jwttokenManager jwtoken.JWTokenManager,
	hashManager hash.HashManager,
	cfg *config.ServiceConfig,
	logging logger.Logger,
) (Service, error) {
	var (
		stopCh = make(chan struct{})

		Auth = NewAuthService(cfg, repo.Users, repo.JWTokens, repo.Transactions, jwttokenManager,
			hashManager, logging)
		Dialogs = NewDialogsService(cfg, repo.Users, repo.Dialogs, repo.JWTokens, repo.Transactions, logging)
	)

	res := &service{
		Auth:    Auth,
		Dialogs: Dialogs,

		cfg:     cfg,
		logging: logging,
		stopCh:  stopCh,
	}

	return res, nil
}

func (s *service) Shutdown() {
	time.Sleep(1 * time.Second)
	for i := 0; i < 0; i++ {
		s.stopCh <- struct{}{}
	}

	close(s.stopCh)
}

package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Pyegorchik/bdd/backend/internal/config"
	"github.com/Pyegorchik/bdd/backend/internal/domain"
	"github.com/Pyegorchik/bdd/backend/internal/repository"
	"github.com/Pyegorchik/bdd/backend/models"
	"github.com/Pyegorchik/bdd/backend/pkg/logger"
	"github.com/jackc/pgx/v5"
)

type DialogsService struct {
	cfg              *config.ServiceConfig
	repoUsers        repository.Users
	repoDialogs      repository.Dialogs
	repoJWTokens     repository.JWTokens
	repoTransactions repository.Transactions

	logging logger.Logger
}

func NewDialogsService(
	cfg *config.ServiceConfig,
	repoUsers repository.Users,
	repoDialogs repository.Dialogs,
	repoJWTokens repository.JWTokens,
	repoTransactions repository.Transactions,

	logging logger.Logger) Dialogs {

	return &DialogsService{
		cfg:              cfg,
		repoDialogs:      repoDialogs,
		repoUsers:        repoUsers,
		repoJWTokens:     repoJWTokens,
		repoTransactions: repoTransactions,

		logging: logging,
	}
}

func (d *DialogsService) SendMessage(ctx context.Context, req *models.SendMessageRequest, userID int64) error {
	tx, err := d.repoTransactions.BeginTransaction(ctx)
	if err != nil {
		return newServiceError(code500, fmt.Errorf("SendMessage/BeginTransaction: %w", err), InternalError, "")
	}
	defer tx.Rollback(ctx)

	recepeint, err := d.repoUsers.GetUserByAddress(ctx, tx, *req.RecipientID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return newServiceError(code400, fmt.Errorf("SendMessage/GetUserByAddress: %w", err),
				BadRequest, fmt.Sprintf("user with address %v is not registered", *req.RecipientID))
		} else {
			return newServiceError(code500, fmt.Errorf("SendMessage/GetUserByAddress: %w", err), InternalError, "")
		}

	}

	sender, err := d.repoUsers.GetUserById(ctx, tx, userID)
	if err != nil {
		return newServiceError(code500, fmt.Errorf("SendMessage/GetUserById: %w", err), InternalError, "")
	}

	dialogExists, err := d.repoDialogs.DialogExists(ctx, tx, recepeint.ID, userID)
	if err != nil {
		return newServiceError(code500, fmt.Errorf("SendMessage/DialogExists: %w", err), InternalError, "")
	}

	var dialogId int64
	if dialogExists {
		dialogId, err = d.repoDialogs.GetDialogByUsers(ctx, tx, recepeint.ID, userID)
		if err != nil {
			return newServiceError(code500, fmt.Errorf("SendMessage/GetDialogByUsers: %w", err), InternalError, "")
		}
	} else {
		dialogId, err = d.repoDialogs.CreateDialog(ctx, tx, recepeint.ID, userID)
		if err != nil {
			return newServiceError(code500, fmt.Errorf("SendMessage/CreateDialog: %w", err), InternalError, "")
		}

		err = d.repoDialogs.CreateDialogBetweenUsers(ctx, tx, recepeint.ID, userID, dialogId)
		if err != nil {
			return newServiceError(code500, fmt.Errorf("SendMessage/CreateDialogBetweenUsers: %w", err), InternalError, "")
		}
	}

	msg := &domain.Message{
		DialogID:      dialogId,
		SenderAddress: sender.Address.String(),
		SenderID:      userID,
		Content:       *req.Content,
	}

	err = d.repoDialogs.CreateMessageInDialog(ctx, tx, msg)
	if err != nil {
		return newServiceError(code500, fmt.Errorf("SendMessage/CreateMessageInDialog: %w", err), InternalError, "")
	}

	if err := tx.Commit(ctx); err != nil {
		return newServiceError(code500,
			fmt.Errorf("SendMessage/Commit: %w", err), InternalError, "")
	}

	return nil
}

func (d *DialogsService) GetDialogs(ctx context.Context, userID int64) ([]*models.DialogsResponseItems0, error) {
	tx, err := d.repoTransactions.BeginTransaction(ctx)
	if err != nil {
		return nil, newServiceError(code500, fmt.Errorf("GetDialogs/BeginTransaction: %w", err), InternalError, "")
	}
	defer tx.Rollback(ctx)

	recepients, err := d.repoDialogs.GetAllDialogsByUser(ctx, tx, userID)
	if err != nil {
		return nil, newServiceError(code500, fmt.Errorf("GetDialogs/CreateMessageInDialog: %w", err), InternalError, "")
	}

	res := domain.RecepientsToRecepinetsResponce(recepients)

	return res, nil
}

func (d *DialogsService) GetMessages(ctx context.Context, dialogID int64) ([]*models.MessagesResponseItems0, error) {
	tx, err := d.repoTransactions.BeginTransaction(ctx)
	if err != nil {
		return nil, newServiceError(code500, fmt.Errorf("GetMessages/BeginTransaction: %w", err), InternalError, "")
	}
	defer tx.Rollback(ctx)

	msgs, err := d.repoDialogs.GetAllMessagesWithinDialogById(ctx, tx, dialogID)
	if err != nil {
		return nil, newServiceError(code500, fmt.Errorf("GetMessages/CreateMessageInDialog: %w", err), InternalError, "")
	}

	res := domain.MessageToMessageResponse(msgs)

	return res, nil
}

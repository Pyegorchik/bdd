package repository

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Pyegorchik/bdd/backend/internal/domain"
	"github.com/jackc/pgx/v5"
)

type DialogsRepo struct {
}

func NewDialogsRepo() Dialogs {
	return &DialogsRepo{}
}

func (repo *DialogsRepo) DialogExists(ctx context.Context, transaction Transaction, userOneID int64, userTwoID int64) (bool, error) {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return false, errors.New("DialogExists: error: type assertion failed on interface Transaction")
	}

	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM dialog_participants dp1
			JOIN dialog_participants dp2 ON dp1.dialog_id = dp2.dialog_id
			WHERE dp1.user_id = $1 AND dp2.user_id = $2
		)
	`

	row := tx.QueryRow(ctx, query, userOneID, userTwoID)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("DialogExists/Scan: %w", err)
	}

	return exists, nil
}

func (repo *DialogsRepo) GetDialogByUsers(ctx context.Context, transaction Transaction, userOneID int64, userTwoID int64) (int64, error) {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return 0, errors.New("GetDialogByUsers: error: type assertion failed on interface Transaction")
	}

	query := `
		SELECT dp1.dialog_id
		FROM dialog_participants dp1
		JOIN dialog_participants dp2 ON dp1.dialog_id = dp2.dialog_id
		WHERE dp1.user_id = $1 AND dp2.user_id = $2
	`

	row := tx.QueryRow(ctx, query, userOneID, userTwoID)
	var dialogID int64
	if err := row.Scan(&dialogID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNoRows
		}
		return 0, fmt.Errorf("GetDialogByUsers/Scan: %w", err)
	}

	return dialogID, nil
}

func (repo *DialogsRepo) CreateDialogBetweenUsers(ctx context.Context, transaction Transaction, userOneID int64, userTwoID int64, dialogID int64) error {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return errors.New("CreateDialogBetweenUsers: error: type assertion failed on interface Transaction")
	}

	query := `INSERT INTO dialog_participants (dialog_id, user_id) VALUES ($1, $2), ($1, $3)`
	_, err := tx.Exec(ctx, query, dialogID, userOneID, userTwoID)
	if err != nil {
		return fmt.Errorf("CreateDialogBetweenUsers/Exec: %w", err)
	}

	return nil
}

func (repo *DialogsRepo) CreateDialog(ctx context.Context, transaction Transaction, userOneID int64, userTwoID int64) (int64, error) {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return 0, errors.New("CreateDialog: error: type assertion failed on interface Transaction")
	}

	query := `INSERT INTO dialogs DEFAULT VALUES RETURNING id`
	row := tx.QueryRow(ctx, query)
	var dialogID int64
	if err := row.Scan(&dialogID); err != nil {
		return 0, fmt.Errorf("CreateDialog/Scan: %w", err)
	}

	return dialogID, nil
}

func (repo *DialogsRepo) CreateMessageInDialog(ctx context.Context, transaction Transaction, msg *domain.Message) error {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return errors.New("CreateMessageInDialog: error: type assertion failed on interface Transaction")
	}

	log.Printf("d  %v s %v c %v", msg.DialogID, msg.SenderID, msg.Content)
	query := `INSERT INTO messages (dialog_id, sender_id, content) VALUES ($1, $2, $3)`
	_, err := tx.Exec(ctx, query, msg.DialogID, msg.SenderID, msg.Content)
	if err != nil {
		return fmt.Errorf("CreateMessageInDialog/Exec: %w", err)
	}

	return nil
}

func (repo *DialogsRepo) GetAllDialogsByUser(ctx context.Context, transaction Transaction, userID int64) ([]*domain.DialogParticipant, error) {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return nil, errors.New("GetAllDialogsByUser: error: type assertion failed on interface Transaction")
	}

	query := `
		SELECT dp1.dialog_id, uc.address
		FROM dialog_participants dp1
		JOIN dialog_participants dp2 ON dp1.dialog_id = dp2.dialog_id
		JOIN users_chain uc ON dp2.user_id = uc.id
		WHERE dp1.user_id = $1 AND dp2.user_id != $1
	`

	rows, err := tx.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("GetAllDialogsByUser/Query: %w", err)
	}
	defer rows.Close()

	var participants []*domain.DialogParticipant
	for rows.Next() {
		var participant domain.DialogParticipant
		if err := rows.Scan(&participant.DialogID, &participant.UserAdress); err != nil {
			return nil, fmt.Errorf("GetAllDialogsByUser/Scan: %w", err)
		}
		participants = append(participants, &participant)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("GetAllDialogsByUser/Rows: %w", rows.Err())
	}

	return participants, nil
}

func (repo *DialogsRepo) GetAllMessagesWithinDialogById(ctx context.Context, transaction Transaction, dialogID int64) ([]*domain.Message, error) {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return nil, errors.New("GetAllMessagesWithinDialogById: error: type assertion failed on interface Transaction")
	}

	query := `
		SELECT m.id, m.dialog_id, m.sender_id, m.content, u_c.address
		FROM messages AS m
		JOIN users_chain as u_c ON m.sender_id = u_c.id
		WHERE dialog_id = $1
	`

	rows, err := tx.Query(ctx, query, dialogID)
	if err != nil {
		return nil, fmt.Errorf("GetAllMessagesWithinDialogById/Query: %w", err)
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		var message domain.Message
		if err := rows.Scan(&message.ID, &message.DialogID, &message.SenderID, &message.Content, &message.SenderAddress); err != nil {
			return nil, fmt.Errorf("GetAllMessagesWithinDialogById/Scan: %w", err)
		}
		messages = append(messages, &message)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("GetAllMessagesWithinDialogById/Rows: %w", rows.Err())
	}

	return messages, nil
}

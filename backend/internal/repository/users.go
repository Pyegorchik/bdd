package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Pyegorchik/bdd/backend/internal/domain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v5"
)

type UsersRepo struct {
}

func NewUsersRepo() Users {
	return &UsersRepo{}
}

func (r *UsersRepo) InsertUser(ctx context.Context, transaction Transaction, user *domain.UserChain) (int64, error) {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return 0, errors.New("InsertUser: error: type assertion failed on interface Transaction")
	}
	row := tx.QueryRow(ctx, `INSERT INTO users_chain (id, role, address) VALUES (DEFAULT, $1,$2) RETURNING id`,
		user.Role, strings.ToLower(user.Address.String()))

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("InsertUser/Scan: %w", err)
	}

	return id, nil
}

func (r *UsersRepo) GetUserById(ctx context.Context, transaction Transaction, id int64,
) (*domain.UserChain, error) {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return nil, errors.New("GetUserById: error: type assertion failed on interface Transaction")
	}

	row := tx.QueryRow(ctx, `SELECT u.id, u.role, u.address
		FROM users_chain AS u
		WHERE u.id=$1`, id)

	var (
		u    = &domain.UserChain{}
		addr string
	)
	err := row.Scan(&u.ID, &u.Role, &addr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRows
		}
		return nil, fmt.Errorf("GetUserById/Scan: %w", err)
	}
	u.Address = common.HexToAddress(addr)
	return u, nil
}

func (r *UsersRepo) GetUserByAddress(
	ctx context.Context,
	transaction Transaction,
	address string,
) (*domain.UserChain, error) {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return nil, errors.New("GetUserByAddress: error: type assertion failed on interface Transaction")
	}

	row := tx.QueryRow(ctx, `SELECT u.id, u.role, u.address
		FROM users_chain AS u
		WHERE u.address=$1`, strings.ToLower(address))

	var (
		u    = &domain.UserChain{}
		addr string
	)
	err := row.Scan(&u.ID, &u.Role, &addr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRows
		}
		return nil, fmt.Errorf("GetUserByAddress/Scan: %w", err)
	}

	u.Address = common.HexToAddress(addr)
	return u, nil
}

func (r *UsersRepo) GetAuthMessageByAddress(
	ctx context.Context,
	transaction Transaction,
	address string,
) (*domain.AuthMessage, error) {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return nil, errors.New("GetAuthMessageByAddress: error: type assertion failed on interface Transaction")
	}
	row := tx.QueryRow(ctx, `SELECT address, created_at, code FROM auth_messages_chain WHERE address = $1`, strings.ToLower(address))
	res := &domain.AuthMessage{}
	if err := row.Scan(&res.Address, &res.CreatedAt, &res.Message); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRows
		}
		return nil, fmt.Errorf("GetAuthMessageByAddress/Scan: %w", err)
	}
	return res, nil
}

func (r *UsersRepo) InsertAuthMessage(
	ctx context.Context,
	transaction Transaction,
	msg *domain.AuthMessage,
) error {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return errors.New("InsertAuthMessage: error: type assertion failed on interface Transaction")
	}
	if _, err := tx.Exec(ctx, `INSERT INTO auth_messages_chain (address, code, created_at) VALUES ($1,$2,$3)`,
		strings.ToLower(msg.Address), msg.Message, msg.CreatedAt); err != nil {
		return fmt.Errorf("InsertAuthMessage/Exec: %w", err)
	}
	return nil
}

func (r *UsersRepo) DeleteAuthMessage(
	ctx context.Context,
	transaction Transaction,
	address string,
) error {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return errors.New("DeleteAuthMessage: error: type assertion failed on interface Transaction")
	}
	if _, err := tx.Exec(ctx, `DELETE FROM auth_messages_chain WHERE address=$1`, strings.ToLower(address)); err != nil {
		return fmt.Errorf("DeleteAuthMessage/Exec: %w", err)
	}
	return nil
}

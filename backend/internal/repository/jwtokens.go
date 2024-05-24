package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Pyegorchik/bdd/backend/internal/domain"
	"github.com/Pyegorchik/bdd/backend/pkg/jwtoken"
	"github.com/jackc/pgx/v5"
)

type JWTokensRepo struct {
}

func NewJWTokensRepo() JWTokens {
	return &JWTokensRepo{}
}

func (r *JWTokensRepo) InsertJWToken(
	ctx context.Context,
	transaction Transaction,
	tokenData jwtoken.JWTokenData,
) error {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return errors.New("InsertJWToken: error: type assertion failed on interface Transaction")
	}
	if _, err := tx.Exec(ctx, `INSERT INTO jwtokens_chain (id,purpose,role,number,expires_at,secret) VALUES($1, $2, $3, $4, $5,$6)`,
		tokenData.ID, tokenData.Purpose, tokenData.Role,
		tokenData.Number, tokenData.ExpiresAt, tokenData.Secret); err != nil {
		return err
	}

	return nil
}

func (r *JWTokensRepo) GetJWTokenNumber(
	ctx context.Context,
	transaction Transaction,
	id int64,
	role domain.Role,
	purpose jwtoken.Purpose,
) (int, error) {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return 0, errors.New("GetJWTokenNumber: error: type assertion failed on interface Transaction")
	}
	rows, err := tx.Query(ctx, `SELECT number FROM jwtokens_chain WHERE id=$1 AND role=$2 AND purpose=$3 ORDER BY number`,
		id, role, int(purpose))

	if err != nil {
		return 0, fmt.Errorf("GetJWTokenNumber/Query: %w", err)
	}
	defer rows.Close()

	var number int
	if rowExist := rows.Next(); !rowExist {
		return number, nil
	}
	if err := rows.Scan(&number); err != nil {
		return 0, fmt.Errorf("GetJWTokenNumber/Scan: %w", err)
	}

	nextNum := number + 1
	for rows.Next() {
		if err := rows.Scan(&number); err != nil {
			return 0, fmt.Errorf("GetJWTokenNumber/Next/Scan: %w", err)
		}

		if number != nextNum {
			return nextNum, nil
		}

		nextNum++
	}

	return nextNum, nil
}

func (r *JWTokensRepo) GetJWTokenSecret(
	ctx context.Context,
	transaction Transaction,
	id int64,
	role domain.Role,
	number int,
	purpose jwtoken.Purpose,
) (string, error) {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return "", errors.New("GetJWTokenSecret: error: type assertion failed on interface Transaction")
	}
	row := tx.QueryRow(ctx, `SELECT secret FROM jwtokens_chain WHERE id=$1 AND role=$2 AND number=$3 AND purpose=$4`,
		id, role, number, purpose)

	var secret string
	if err := row.Scan(&secret); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNoRows
		}
		return "", err
	}

	return secret, nil

}

func (r *JWTokensRepo) DropJWTokens(
	ctx context.Context,
	transaction Transaction,
	id int64,
	role domain.Role,
	number int,
) error {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return errors.New("DropJWTokens: error: type assertion failed on interface Transaction")
	}
	if _, err := tx.Exec(ctx, `DELETE FROM jwtokens_chain WHERE id=$1 AND role=$2 AND number=$3`,
		id, role, number); err != nil {
		return err
	}

	return nil
}

func (r *JWTokensRepo) DropAllJWTokens(
	ctx context.Context,
	transaction Transaction,
	id int64,
	role domain.Role,
) error {
	tx, ok := transaction.(pgx.Tx)
	if !ok {
		return errors.New("DropAllJWTokens: error: type assertion failed on interface Transaction")
	}
	if _, err := tx.Exec(ctx, `DELETE FROM jwtokens_chain WHERE id=$1 and role=$2`,
		id, role); err != nil {
		return err
	}

	return nil
}

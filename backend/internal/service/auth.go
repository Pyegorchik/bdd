package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/Pyegorchik/bdd/backend/internal/config"
	"github.com/Pyegorchik/bdd/backend/internal/domain"
	"github.com/Pyegorchik/bdd/backend/internal/repository"
	"github.com/Pyegorchik/bdd/backend/models"
	"github.com/Pyegorchik/bdd/backend/pkg/hash"
	"github.com/Pyegorchik/bdd/backend/pkg/jwtoken"
	"github.com/Pyegorchik/bdd/backend/pkg/logger"
	"github.com/Pyegorchik/bdd/backend/pkg/now"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type AuthService struct {
	cfg              *config.ServiceConfig
	repoUsers        repository.Users
	repoJWTokens     repository.JWTokens
	repoTransactions repository.Transactions
	jwtManager       jwtoken.JWTokenManager
	hashManager      hash.HashManager
	logging          logger.Logger
}

func NewAuthService(
	cfg *config.ServiceConfig,
	repoUsers repository.Users,
	repoJWTokens repository.JWTokens,
	repoTransactions repository.Transactions,
	jwtManager jwtoken.JWTokenManager,
	hashManager hash.HashManager,
	logging logger.Logger) Auth {

	return &AuthService{
		cfg:              cfg,
		repoUsers:        repoUsers,
		repoJWTokens:     repoJWTokens,
		repoTransactions: repoTransactions,
		jwtManager:       jwtManager,
		hashManager:      hashManager,
		logging:          logging,
	}
}

const (
	alphabet    = "abcdefghijklmnopqrstuvwxyz1234567890"
	authMessage = "Hello, %s! Please, sign this message with random param %s to sign in!"
)

func (s *AuthService) GetUserById(
	ctx context.Context,
	id int64,
) (*domain.UserChain, error) {
	tx, err := s.repoTransactions.BeginTransaction(ctx)
	if err != nil {
		return nil, newServiceError(code500,
			fmt.Errorf("GetUserById/BeginTransaction: %w", err), InternalError, "")
	}
	defer tx.Rollback(context.Background())

	user, err := s.repoUsers.GetUserById(ctx, tx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, newServiceError(code400,
				fmt.Errorf("GetUserByToken/GetUserById: %w", err), UserNotExist, "")
		}
		return nil, newServiceError(code500,
			fmt.Errorf("GetUserByToken/GetUserById: %w", err), InternalError, "")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, newServiceError(code500,
			fmt.Errorf("GetCampaign/Commit: %w", err), InternalError, "")
	}
	return user, nil
}

func (s *AuthService) GetUserByJWToken(
	ctx context.Context,
	purpose jwtoken.Purpose,
	token string,
) (*domain.UserWithTokenNumber, error) {
	tx, err := s.repoTransactions.BeginTransaction(ctx)
	if err != nil {
		return nil, newServiceError(code500,
			fmt.Errorf("GetUserByJWToken/BeginTransaction: %w", err), InternalError, "")
	}
	defer tx.Rollback(context.Background())

	tokenData, err := s.jwtManager.ParseJWToken(token)
	if err != nil {
		return nil, newServiceError(code400,
			fmt.Errorf("GetUserByJWToken/ParseJWToken: %w", err), ParseTokenFailed, "")
	}

	secret, err := s.repoJWTokens.GetJWTokenSecret(ctx, tx, tokenData.ID, domain.Role(tokenData.Role), tokenData.Number, purpose)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, newServiceError(code401,
				fmt.Errorf("GetUserByJWToken/GetJWTokenSecret: %w", err), ParseTokenFailed, "")
		}
		return nil, newServiceError(code500,
			fmt.Errorf("GetUGetUserByJWTokenserByToken/GetJWTokenSecret: %w", err), InternalError, "")
	}

	if tokenData.Secret != secret {
		return nil, newServiceError(code401,
			fmt.Errorf("GetUserByJWToken: %s", TokenWrongSecret), TokenWrongSecret, "")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, newServiceError(code500,
			fmt.Errorf("GetCampaign/Commit: %w", err), InternalError, "")
	}

	return &domain.UserWithTokenNumber{
		ID:     tokenData.ID,
		Role:   domain.Role(tokenData.Role),
		Number: tokenData.Number,
	}, nil
}

func (s *AuthService) RefreshJWTokens(
	ctx context.Context,
	id, number int64,
	role domain.Role,
) (*models.AuthResponse, *jwtoken.JWTokenData, *jwtoken.JWTokenData, error) {
	tx, err := s.repoTransactions.BeginTransaction(ctx)
	if err != nil {
		return nil, nil, nil, newServiceError(code500,
			fmt.Errorf("RefreshJWTokens/BeginTransaction: %w", err), InternalError, "")
	}
	defer tx.Rollback(context.Background())

	if err := s.repoJWTokens.DropJWTokens(ctx, tx, id, role, int(number)); err != nil {
		return nil, nil, nil, newServiceError(code500,
			fmt.Errorf("RefreshJWTokens/DropJWTokens: %w", err), InternalError, "")
	}

	resp, err := s.getAuthRespWithUserById(ctx, tx, role, id)
	if err != nil {
		return nil, nil, nil, newServiceError(code500, fmt.Errorf("GetUserAndRefreshTokens/getAuthRespWithUserById: %w", err), InternalError, "")
	}

	accessToken, refreshToken, err := s.generateJWTokensWithNumber(ctx, tx, id, role, int(number))
	if err != nil {
		return nil, nil, nil, newServiceError(code500,
			fmt.Errorf("RefreshJWTokens/generateTokensWithNumber: %w", err), InternalError, "")
	}

	resp.ServerTime = now.Now().UnixMilli()

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, nil, newServiceError(code500,
			fmt.Errorf("RefreshJWTokens/Commit: %w", err), InternalError, "")
	}
	return resp, accessToken, refreshToken, nil
}

func (s *AuthService) Logout(ctx context.Context, id, number int64, role domain.Role) error {
	tx, err := s.repoTransactions.BeginTransaction(ctx)
	if err != nil {
		return newServiceError(code500,
			fmt.Errorf("Logout/BeginTransaction: %w", err), InternalError, "")
	}
	defer tx.Rollback(context.Background())

	if err := s.repoJWTokens.DropJWTokens(ctx, tx, id, role, int(number)); err != nil {
		return newServiceError(code500,
			fmt.Errorf("Logout/DropJWTokens: %w", err), InternalError, "")
	}

	if err := tx.Commit(ctx); err != nil {
		return newServiceError(code500,
			fmt.Errorf("Logout/Commit: %w", err), InternalError, "")
	}
	return nil
}

func (s *AuthService) FullLogout(ctx context.Context, id int64, role domain.Role) error {
	tx, err := s.repoTransactions.BeginTransaction(ctx)
	if err != nil {
		return newServiceError(code500,
			fmt.Errorf("FullLogout/BeginTransaction: %w", err), InternalError, "")
	}
	defer tx.Rollback(context.Background())

	if err := s.repoJWTokens.DropAllJWTokens(ctx, tx, id, role); err != nil {
		return newServiceError(code500,
			fmt.Errorf("FullLogout/DropAllJWTokens: %w", err), InternalError, "")
	}

	if err := tx.Commit(ctx); err != nil {
		return newServiceError(code500,
			fmt.Errorf("FullLogout/Commit: %w", err), InternalError, "")
	}
	return nil
}

func (s *AuthService) GetAuthMessage(
	ctx context.Context,
	req *models.AuthMessageRequest,
) (*models.AuthMessageResponse, error) {
	tx, err := s.repoTransactions.BeginTransaction(ctx)
	if err != nil {
		return nil, newServiceError(code500,
			fmt.Errorf("GetAuthMessage/BeginTransaction: %w", err), InternalError, "")
	}
	defer tx.Rollback(context.Background())

	msg, err := s.repoUsers.GetAuthMessageByAddress(ctx, tx, *req.Address)
	if err != nil && !errors.Is(err, repository.ErrNoRows) {
		return nil, newServiceError(code500,
			fmt.Errorf("GetAuthMessage/GetAuthMessageByAddress: %w", err), InternalError, "")
	}
	if msg != nil {
		if now.Now().Sub(time.UnixMilli(msg.CreatedAt)) < 5*time.Minute {
			return &models.AuthMessageResponse{
				Message: &msg.Message,
			}, nil
		}
	}
	if err := s.repoUsers.DeleteAuthMessage(ctx, tx, *req.Address); err != nil {
		return nil, newServiceError(code500,
			fmt.Errorf("GetAuthMessage/DeleteAuthMessage: %w", err), InternalError, "")
	}
	message := fmt.Sprintf(authMessage, strings.ToLower(*req.Address), randomString(64))
	if err := s.repoUsers.InsertAuthMessage(ctx, tx, &domain.AuthMessage{
		Address:   strings.ToLower(*req.Address),
		Message:   message,
		CreatedAt: now.Now().UnixMilli(),
	}); err != nil {
		return nil, newServiceError(code500,
			fmt.Errorf("GetAuthMessage/InsertAuthMessage: %w", err), InternalError, "")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, newServiceError(code500,
			fmt.Errorf("GetAuthMessage/Commit: %w", err), InternalError, "")
	}
	return &models.AuthMessageResponse{
		Message: &message,
	}, nil
}

func (s *AuthService) AuthByMessage(
	ctx context.Context,
	req *models.AuthBySignatureRequest,
) (*models.AuthResponse, *jwtoken.JWTokenData, *jwtoken.JWTokenData, error) {
	tx, err := s.repoTransactions.BeginTransaction(ctx)
	if err != nil {
		return nil, nil, nil, newServiceError(code500,
			fmt.Errorf("AuthByMessage/BeginTransaction: %w", err), InternalError, "")
	}
	defer tx.Rollback(context.Background())

	msg, err := s.repoUsers.GetAuthMessageByAddress(ctx, tx, *req.Address)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, nil, nil, newServiceError(code400,
				fmt.Errorf("AuthByMessage/GetAuthMessageByAddress: %w", err), AuthMessageNotExist, "")
		}
		return nil, nil, nil, newServiceError(code500,
			fmt.Errorf("AuthByMessage/GetAuthMessageByAddress: %w", err), InternalError, "")
	}
	if now.Now().Sub(time.UnixMilli(msg.CreatedAt)) >= 5*time.Minute {
		return nil, nil, nil, newServiceError(code400,
			fmt.Errorf("AuthByMessage: %s", AuthMessageExpired), AuthMessageExpired, "")
	}

	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg.Message), msg.Message)))
	sig := common.FromHex(*req.Signature)
	if len(sig) > 0 && sig[len(sig)-1] > 4 {
		sig[len(sig)-1] -= 27
	}
	pubKey, err := crypto.Ecrecover(hash, sig)
	if err != nil {
		return nil, nil, nil, newServiceError(code401,
			fmt.Errorf("AuthByMessage: %s", EcrecoverFailed), EcrecoverFailed, "")
	}
	pkey, err := crypto.UnmarshalPubkey(pubKey)
	if err != nil {
		return nil, nil, nil, newServiceError(code401,
			fmt.Errorf("AuthByMessage/UnmarshalPubkey: %w", err), EcrecoverFailed, "")
	}
	signedAddress := crypto.PubkeyToAddress(*pkey)
	if !strings.EqualFold(signedAddress.Hex(), *req.Address) {
		return nil, nil, nil, newServiceError(code401,
			fmt.Errorf("AuthByMessage: %s", WrongSignature), WrongSignature, "")
	}

	user, err := s.repoUsers.GetUserByAddress(ctx, tx, strings.ToLower(*req.Address))
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			user, err = createUser(ctx, tx, s.repoUsers, strings.ToLower(*req.Address), 0)
			if err != nil {
				return nil, nil, nil, newServiceError(code400,
					fmt.Errorf("AuthByMessage/createUser: %w", err), InternalError, "")
			}
		} else {
			return nil, nil, nil, newServiceError(code400,
				fmt.Errorf("AuthByMessage/GetUserByAddress: %w", err), InternalError, "")
		}

	}

	resp, err := s.getAuthRespWithUserById(ctx, tx, user.Role, user.ID)
	if err != nil {
		return nil, nil, nil, newServiceError(code500, fmt.Errorf("AuthByMessage/getAuthRespWithUserById: %w", err), InternalError, "")
	}

	accessToken, refreshToken, err := s.generateJWTokens(ctx, tx, user.ID, user.Role)
	if err != nil {
		return nil, nil, nil, newServiceError(code500,
			fmt.Errorf("AuthByMessage/generateJWTokens: %w", err), InternalError, "")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, nil, newServiceError(code500,
			fmt.Errorf("AuthByMessage/Commit: %w", err), InternalError, "")
	}
	return resp, accessToken, refreshToken, nil
}

func createUser(
	ctx context.Context,
	tx repository.Transaction,
	repoUsers repository.Users,
	address string,
	d_id int64,
) (*domain.UserChain, error) {
	u := &domain.UserChain{
		Address: common.HexToAddress(address),
	}
	var err error
	u.ID, err = repoUsers.InsertUser(ctx, tx, u)
	if err != nil {
		return nil, fmt.Errorf("createUser/InsertUser: %w", err)
	}

	return u, nil
}

func (s *AuthService) getAuthRespWithUserById(ctx context.Context, tx repository.Transaction, role domain.Role, id int64) (*models.AuthResponse, error) {
	var resp models.AuthResponse

	user, err := s.repoUsers.GetUserById(ctx, tx, id)
	if err != nil {

	}
	resp.User = domain.UserToModel(user)

	return &resp, nil
}

func (s *AuthService) generateJWTokens(
	ctx context.Context,
	tx repository.Transaction,
	id int64,
	role domain.Role,
) (*jwtoken.JWTokenData, *jwtoken.JWTokenData, error) {
	number, err := s.repoJWTokens.GetJWTokenNumber(ctx, tx, id, role, jwtoken.PurposeAccess)
	if err != nil {
		return nil, nil, fmt.Errorf("generateJWTokens/GetJWTokenNumber: %w", err)
	}

	return s.generateJWTokensWithNumber(ctx, tx, id, role, number)
}

func (s *AuthService) generateJWTokensWithNumber(
	ctx context.Context,
	tx repository.Transaction,
	id int64,
	role domain.Role,
	number int,
) (*jwtoken.JWTokenData, *jwtoken.JWTokenData, error) {
	accessExpiresAt := now.Now().Add(time.Duration(s.cfg.AccessTokenTTL) * time.Millisecond)
	accessTokenData := &jwtoken.JWTokenData{
		Purpose:   jwtoken.PurposeAccess,
		ID:        id,
		Role:      int(role),
		Number:    number,
		ExpiresAt: accessExpiresAt,
		Secret:    generateSecret(0, id, number, jwtoken.PurposeAccess),
	}
	accessToken, err := s.jwtManager.GenerateJWToken(accessTokenData)
	if err != nil {
		return nil, nil, fmt.Errorf("generateJWTokensWithNumber/GenerateJWTokenAccess: %w", err)
	}

	refreshExpiresAt := now.Now().Add(time.Duration(s.cfg.RefreshTokenTTL) * time.Millisecond)
	refreshTokenData := &jwtoken.JWTokenData{
		Purpose:   jwtoken.PurposeRefresh,
		ID:        id,
		Role:      int(role),
		Number:    number,
		ExpiresAt: refreshExpiresAt,
		Secret:    generateSecret(int(0), id, number, jwtoken.PurposeRefresh),
	}
	refreshToken, err := s.jwtManager.GenerateJWToken(refreshTokenData)
	if err != nil {
		return nil, nil, fmt.Errorf("generateJWTokensWithNumber/GenerateJWTokenRefresh: %w", err)
	}

	if err = s.repoJWTokens.InsertJWToken(ctx, tx, *accessTokenData); err != nil {
		return nil, nil, fmt.Errorf("generateJWTokensWithNumber/InsertJWTokenAccess: %w", err)
	}
	if err = s.repoJWTokens.InsertJWToken(ctx, tx, *refreshTokenData); err != nil {
		return nil, nil, fmt.Errorf("generateJWTokensWithNumber/InsertJWTokenRefresh: %w", err)
	}
	return accessToken, refreshToken, nil
}

func generateSecret(role int, id int64, number int, purpose jwtoken.Purpose) string {
	toHashElems := []string{
		strconv.Itoa(role),
		strconv.Itoa(int(id)),
		strconv.Itoa(number),
		strconv.Itoa(int(purpose)),
		randomString(20),
	}

	toHash := strings.Join(toHashElems, "_")
	hash := sha256.Sum256([]byte(toHash))

	return hex.EncodeToString(hash[:])
}

func randomString(l int) string {
	res := make([]byte, l)
	for i := 0; i < l; i++ {
		res[i] = alphabet[rand.Intn(len(alphabet))]
	}

	return string(res)
}

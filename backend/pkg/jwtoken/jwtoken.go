package jwtoken

import "time"

type Purpose int

const (
	PurposeAccess = Purpose(iota)
	PurposeRefresh
)

type JWTokenData struct {
	Token     string
	Purpose   Purpose
	Role      int
	ID        int64
	Number    int
	ExpiresAt time.Time
	Secret    string
}

type jwtokenManager struct {
	signingKey string
}

type JWTokenManager interface {
	GenerateJWToken(data *JWTokenData) (*JWTokenData, error)
	ParseJWToken(jwtoken string) (*JWTokenData, error)
}

func NewTokenManager(signingKey string) JWTokenManager {
	return &jwtokenManager{
		signingKey: signingKey,
	}
}
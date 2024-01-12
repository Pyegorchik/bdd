package jwtoken

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

func (t *jwtokenManager) GenerateJWToken(data *JWTokenData) (*JWTokenData, error) {
	claims := jwt.MapClaims{
		"id":      data.ID,
		"role":    data.Role,
		"purpose": int64(data.Purpose),
		"secret":  data.Secret,
		"exp":     data.ExpiresAt.Unix(),
		"number":  data.Number,
	}
	jwtoken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := jwtoken.SignedString([]byte(t.signingKey))
	if err != nil {
		return nil, fmt.Errorf("GenerateJWToken/SignedString: sign jwtoken failed: %w", err)
	}
	data.Token = token
	return data, nil
}
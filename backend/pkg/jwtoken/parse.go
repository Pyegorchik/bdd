package jwtoken

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func parseJWTokenIntClaim(claims jwt.MapClaims, key string) (int64, error) {
	if parsedValue, ok := claims[key].(float64); !ok {
		return 0, fmt.Errorf("parseJWTokenIntClaim: error: invalid JWToken claim: %s", key)
	} else {
		return int64(parsedValue), nil
	}
}

func parseJWTokenStringClaim(claims jwt.MapClaims, key string) (string, error) {
	if stringValue, ok := claims[key].(string); !ok {
		return "", fmt.Errorf("parseJWTokenStringClaim: error: invalid JWToken claim: %s", key)
	} else {
		return stringValue, nil
	}
}

func (t *jwtokenManager) ParseJWToken(JWToken string) (*JWTokenData, error) {
	parsedJWToken, err := jwt.Parse(JWToken, func(JWToken *jwt.Token) (i interface{}, e error) {
		if _, ok := JWToken.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("ParseJWToken/Parse: error: unexpected signing method: %v", JWToken.Header["alg"])
		}
		return []byte(t.signingKey), nil
	})
	if err != nil && parsedJWToken == nil {
		return nil, fmt.Errorf("ParseJWToken/Parse: parse JWToken failed: %w", err)
	}
	claims, ok := parsedJWToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("ParseJWToken/parsedJWToken.Claims: error: JWToken wrong claims")
	}
	purpose, err := parseJWTokenIntClaim(claims, "purpose")
	if err != nil {
		return nil, fmt.Errorf("ParseJWToken/parseJWTokenIntClaim/purpose: %w", err)
	}
	if purpose != int64(PurposeAccess) && purpose != int64(PurposeRefresh) {
		return nil, fmt.Errorf("ParseJWToken: error: invalid purpose: %d", purpose)
	}
	id, err := parseJWTokenIntClaim(claims, "id")
	if err != nil {
		return nil, fmt.Errorf("ParseJWToken/parseJWTokenIntClaim/id: %w", err)
	}
	number, err := parseJWTokenIntClaim(claims, "number")
	if err != nil {
		return nil, fmt.Errorf("ParseJWToken/parseJWTokenIntClaim/number: %w", err)
	}
	role, err := parseJWTokenIntClaim(claims, "role")
	if err != nil {
		return nil, fmt.Errorf("ParseJWToken/parseJWTokenIntClaim/role: %w", err)
	}
	secret, err := parseJWTokenStringClaim(claims, "secret")
	if err != nil {
		return nil, fmt.Errorf("ParseJWToken/parseJWTokenIntClaim/secret: %w", err)
	}
	expiresAt, err := parseJWTokenIntClaim(claims, "exp")
	if err != nil {
		return nil, fmt.Errorf("ParseJWToken/parseJWTokenIntClaim/exp: %w", err)
	}
	return &JWTokenData{
		Purpose:   Purpose(purpose),
		Role:      int(role),
		ID:        id,
		Number:    int(number),
		ExpiresAt: time.Unix(expiresAt, 0),
		Secret:    secret,
	}, nil
}
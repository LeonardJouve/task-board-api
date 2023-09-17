package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenClaims = jwt.RegisteredClaims

func CreateToken(name string, userId uint, lifetime int64) (string, *TokenClaims, error) {
	privatePEM, _, err := getRSAKeys(name)
	if err != nil {
		return "", nil, err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		return "", nil, err
	}

	claims := &TokenClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(lifetime) * time.Minute)),
		ID:        uuid.NewString(),
		Subject:   fmt.Sprint(userId),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
	if err != nil {
		return "", nil, err
	}

	return token, claims, nil
}

func ValidateToken(name string, token string) (*TokenClaims, error) {
	_, publicPEM, err := getRSAKeys(name)

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicPEM)
	if err != nil {
		return nil, err
	}

	var claims = &TokenClaims{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	return claims, nil
}

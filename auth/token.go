package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenClaims = jwt.RegisteredClaims

func CreateToken(c *fiber.Ctx, name string, userId uint, lifetime int) (string, bool) {
	privatePEM, _, err := getRSAKeys(name)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "invalid token encryption",
		})
		return "", false
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "invalid token encryption",
		})
		return "", false
	}

	claims := &TokenClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(lifetime) * time.Minute)),
		ID:        uuid.NewString(),
		Subject:   fmt.Sprint(userId),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "invalid token encryption",
		})
		return "", false
	}

	ctx := context.TODO()
	if err := store.Redis.Set(ctx, claims.ID, userId, time.Until(claims.ExpiresAt.Time)).Err(); err != nil {
		c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": err.Error(),
		})
		return "", false
	}

	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		MaxAge:   lifetime * 60,
		Secure:   false,
		HTTPOnly: true,
		Domain:   os.Getenv("HOST"),
	})

	return token, true
}

// TODO handle expiration
func ValidateToken(c *fiber.Ctx, name string, token string) (*TokenClaims, bool) {
	_, publicPEM, err := getRSAKeys(name)
	if err != nil {
		c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
		return nil, false
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicPEM)
	if err != nil {
		c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
		return nil, false
	}

	var claims = &TokenClaims{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
		return nil, false
	}

	return claims, true
}

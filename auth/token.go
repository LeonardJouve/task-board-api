package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/LeonardJouve/task-board-api/dotenv"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims = jwt.RegisteredClaims

func createToken(c *fiber.Ctx, name string, userId uint, lifetime int) (*TokenClaims, string, bool) {
	privateKey, ok := getPrivateKey(c, name)
	if !ok {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return nil, "", false
	}

	claims := &TokenClaims{
		ID:        utils.UUIDv4(),
		Subject:   fmt.Sprint(userId),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(lifetime) * time.Minute)),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return nil, "", false
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

	return claims, token, true
}

func isExpired(c *fiber.Ctx, claims TokenClaims) (bool, bool) {
	tokenAvailableSince, ok := getTokenAvailableSince(c, claims.Subject)
	if !ok {
		return false, false
	}

	if claims.ExpiresAt.Before(time.Now().UTC()) || claims.IssuedAt.Before(tokenAvailableSince) {
		c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
		return true, true
	}

	return false, true
}

func setTokenAvailableSince(c *fiber.Ctx, userId string) bool {
	ctx := context.TODO()
	if _, err := store.Redis.Set(ctx, getTokenAvailableSinceKey(userId), time.Now().UTC().Unix(), 0).Result(); err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return false
	}

	return true
}

func getTokenAvailableSince(c *fiber.Ctx, userId string) (time.Time, bool) {
	ctx := context.TODO()
	tokenAvailableSince, err := store.Redis.Get(ctx, getTokenAvailableSinceKey(userId)).Int64()
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return time.Time{}, false
	}

	return time.Unix(tokenAvailableSince, 0).UTC(), true
}

func getTokenAvailableSinceKey(userId string) string {
	return fmt.Sprintf("token_available_since_%s", userId)
}

func CreateTokens(c *fiber.Ctx, userId uint) (string, string, bool) {
	ctx := context.TODO()

	accessClaims, accessToken, ok := createToken(c, ACCESS_TOKEN, userId, dotenv.GetInt("ACCESS_TOKEN_LIFETIME_IN_MINUTE"))
	if !ok {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return "", "", false
	}
	if err := store.Redis.Set(ctx, accessClaims.ID, userId, time.Until(accessClaims.ExpiresAt.Time)).Err(); err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return "", "", false
	}

	refreshClaims, refreshToken, ok := createToken(c, REFRESH_TOKEN, userId, dotenv.GetInt("REFRESH_TOKEN_LIFETIME_IN_MINUTE"))
	if !ok {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return "", "", false
	}
	if err := store.Redis.Set(ctx, refreshClaims.ID, accessClaims.ID, time.Until(refreshClaims.ExpiresAt.Time)).Err(); err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return "", "", false
	}

	return accessToken, refreshToken, true
}

func ValidateToken(c *fiber.Ctx, name string, token string) (TokenClaims, bool) {
	publicKey, ok := getPublicKey(c, name)
	if !ok {
		c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
		return TokenClaims{}, false
	}

	var claims = TokenClaims{}
	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
		return TokenClaims{}, false
	}

	return claims, true
}

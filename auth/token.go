package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/LeonardJouve/task-board-api/dotenv"
	"github.com/LeonardJouve/task-board-api/models"
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

	jwt.TimePrecision = time.Microsecond
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
		Secure:   true,
		HTTPOnly: true,
		Domain:   os.Getenv("HOST"),
	})

	return claims, token, true
}

func isExpired(c *fiber.Ctx, claims TokenClaims) (bool, bool) {
	var user models.User
	if err := store.Database.Model(&models.User{}).Where("id = ?", claims.Subject).First(&user).Error; err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return false, false
	}

	if claims.ExpiresAt.Before(time.Now().UTC()) || claims.IssuedAt.Before(user.TokenAvailableSince) {
		c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
		return true, true
	}

	return false, true
}

func CreateTokens(c *fiber.Ctx, userId uint) (string, string, bool) {
	ctx := context.TODO()

	accessClaims, accessToken, ok := createToken(c, ACCESS_TOKEN, userId, dotenv.GetInt("ACCESS_TOKEN_LIFETIME_IN_MINUTE"))
	if !ok {
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

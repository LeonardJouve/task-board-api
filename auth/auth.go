package auth

import (
	"context"
	"strings"
	"time"

	"github.com/LeonardJouve/task-board-api/dotenv"
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/schema"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

const (
	ACCESS_TOKEN  = "access_token"
	REFRESH_TOKEN = "refresh_token"
	CSRF_TOKEN    = "csrf_token"
)

func Protect(c *fiber.Ctx) error {
	var accessToken string
	authorization := c.Get("authorization")
	if strings.HasPrefix(authorization, "Bearer ") {
		accessToken = strings.TrimPrefix(authorization, "Bearer ")
	} else if accessTokenCookie := c.Cookies(ACCESS_TOKEN); len(accessTokenCookie) != 0 {
		accessToken = accessTokenCookie
	}

	accessTokenClaims, ok := ValidateToken(c, ACCESS_TOKEN, accessToken)
	if !ok {
		return nil
	}

	ctx := context.TODO()
	userId, err := store.Redis.Get(ctx, accessTokenClaims.ID).Result()
	if err == redis.Nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	var user models.User
	store.Database.First(&user, userId)
	if user.ID == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	c.Locals("user", user)

	return c.Next()
}

func GetCSRF(c *fiber.Ctx) error {
	csrfToken, ok := c.Locals(CSRF_TOKEN).(string)
	if !ok || len(csrfToken) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func Register(c *fiber.Ctx) error {
	user, ok := schema.GetRegisterUserInput(c)
	if !ok {
		return nil
	}

	if err := store.Database.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(schema.SanitizeUser(&user))
}

func Login(c *fiber.Ctx) error {
	user, ok := schema.GetLoginUserInput(c)
	if !ok {
		return nil
	}

	accessToken, ok := CreateToken(c, ACCESS_TOKEN, user.ID, dotenv.GetInt("ACCESS_TOKEN_LIFETIME_IN_MINUTE"))
	if !ok {
		return nil
	}
	refreshToken, ok := CreateToken(c, REFRESH_TOKEN, user.ID, dotenv.GetInt("REFRESH_TOKEN_LIFETIME_IN_MINUTE"))
	if !ok {
		return nil
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		ACCESS_TOKEN:  accessToken,
		REFRESH_TOKEN: refreshToken,
	})
}

func Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies(REFRESH_TOKEN)
	if len(refreshToken) == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	ctx := context.TODO()

	refreshTokenClaims, ok := ValidateToken(c, REFRESH_TOKEN, refreshToken)
	if !ok {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	userId, err := store.Redis.Get(ctx, refreshTokenClaims.ID).Result()
	if err == redis.Nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	var user models.User
	store.Database.First(&user, userId)

	accessToken, ok := CreateToken(c, ACCESS_TOKEN, user.ID, dotenv.GetInt("ACCESS_TOKEN_LIFETIME_IN_MINUTE"))
	if !ok {
		return nil
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		ACCESS_TOKEN: accessToken,
	})
}

func Logout(c *fiber.Ctx) error {
	accessToken := c.Cookies(ACCESS_TOKEN)
	accessTokenClaims, ok := ValidateToken(c, ACCESS_TOKEN, accessToken)
	if !ok {
		return nil
	}

	refreshToken := c.Cookies(REFRESH_TOKEN)
	refreshTokenClaims, ok := ValidateToken(c, refreshToken, refreshToken)
	if !ok {
		return nil
	}

	ctx := context.TODO()
	if err := store.Redis.Del(ctx, accessTokenClaims.ID, refreshTokenClaims.ID).Err(); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	expired := time.Now().Add(-time.Hour * 24)
	c.Cookie(&fiber.Cookie{
		Name:    ACCESS_TOKEN,
		Value:   "",
		Expires: expired,
	})
	c.Cookie(&fiber.Cookie{
		Name:    REFRESH_TOKEN,
		Value:   "",
		Expires: expired,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

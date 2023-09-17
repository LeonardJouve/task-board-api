package auth

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/LeonardJouve/task-board-api/schema"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/LeonardJouve/task-board-api/store/models"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var validate = schema.Init()

func Router(c *fiber.Ctx) error {
	switch strings.Split(c.Path(), "/")[2] {
	case "register":
		if c.Method() != "POST" {
			break
		}
		return register(c)
	case "login":
		if c.Method() != "POST" {
			break
		}
		return login(c)
	case "refresh":
		if c.Method() != "GET" {
			break
		}
		return refresh(c)
	case "logout":
		if c.Method() != "GET" {
			break
		}
		return logout(c)
	}
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"message": "not found",
	})
}

func Protect(c *fiber.Ctx) error {
	var accessToken string
	authorization := c.Get("authorization")
	if strings.HasPrefix(authorization, "Bearer ") {
		accessToken = strings.TrimPrefix(authorization, "Bearer ")
	} else if c.Cookies("access_token") != "" {
		accessToken = c.Cookies("access_token")
	}

	accessTokenClaims, err := ValidateToken("access_token", accessToken)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
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
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "unprocessable entity",
		})
	}

	c.Locals("user", user)

	return c.Next()
}

func register(c *fiber.Ctx) error {
	var input schema.RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	if err := validate.Struct(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if input.Password != input.PasswordConfirm {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid password confirmation",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "unable to hash password",
		})
	}

	user := &models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
	}
	result := store.Database.Create(user)
	if result.Error != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(user.Sanitize())
}

func login(c *fiber.Ctx) error {
	var input schema.LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	if err := validate.Struct(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	var user models.User
	store.Database.Where(&models.User{Email: input.Email}).First(&user)
	if user.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}

	accessTokenLifetimeInMinuteString := os.Getenv("ACCESS_TOKEN_LIFETIME_IN_MINUTE")
	accessTokenLifetimeInMinute, err := strconv.ParseInt(accessTokenLifetimeInMinuteString, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "invalid env",
		})
	}

	accessToken, accessTokenClaims, err := CreateToken("access_token", user.ID, accessTokenLifetimeInMinute)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "invalid token encryption",
		})
	}
	refreshToken, refreshTokenClaims, err := CreateToken("refresh_token", user.ID, accessTokenLifetimeInMinute)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "invalid token encryption",
		})
	}

	ctx := context.TODO()
	if err := store.Redis.Set(ctx, accessTokenClaims.ID, user.ID, time.Until(accessTokenClaims.ExpiresAt.Time)).Err(); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}
	if err := store.Redis.Set(ctx, refreshTokenClaims.ID, user.ID, time.Until(refreshTokenClaims.ExpiresAt.Time)).Err(); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	accessTokenMaxAgeInMinuteString := os.Getenv("ACCESS_TOKEN_MAX_AGE_IN_MINUTE")
	accessTokenMaxAgeInMinute, err := strconv.ParseInt(accessTokenMaxAgeInMinuteString, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "invalid env",
		})
	}

	refreshTokenMaxAgeInMinuteString := os.Getenv("REFRESH_TOKEN_MAX_AGE_IN_MINUTE")
	refreshTokenMaxAgeInMinute, err := strconv.ParseInt(refreshTokenMaxAgeInMinuteString, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "invalid env",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   int(accessTokenMaxAgeInMinute) * 60,
		Secure:   false,
		HTTPOnly: true,
		Domain:   os.Getenv("HOST"),
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   int(refreshTokenMaxAgeInMinute) * 60,
		Secure:   false,
		HTTPOnly: true,
		Domain:   os.Getenv("HOST"),
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if len(refreshToken) == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	ctx := context.TODO()

	refreshTokenClaims, err := ValidateToken("refresh_token", refreshToken)
	if err != nil {
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

	accessTokenLifetimeInMinuteString := os.Getenv("ACCESS_TOKEN_LIFETIME_IN_MINUTE")
	accessTokenLifetimeInMinute, err := strconv.ParseInt(accessTokenLifetimeInMinuteString, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "invalid env",
		})
	}

	accessToken, accessTokenClaims, err := CreateToken("access_token", user.ID, accessTokenLifetimeInMinute)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if err := store.Redis.Set(ctx, accessTokenClaims.ID, user.ID, time.Until(accessTokenClaims.ExpiresAt.Time)).Err(); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	accessTokenMaxAgeInMinuteString := os.Getenv("ACCESS_TOKEN_MAX_AGE_IN_MINUTE")
	accessTokenMaxAgeInMinute, err := strconv.ParseInt(accessTokenMaxAgeInMinuteString, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "invalid env",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   int(accessTokenMaxAgeInMinute) * 60,
		Secure:   false,
		HTTPOnly: true,
		Domain:   os.Getenv("HOST"),
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func logout(c *fiber.Ctx) error {
	accessToken := c.Cookies("access_token")
	accessTokenClaims, err := ValidateToken("access_token", accessToken)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	refreshToken := c.Cookies("refresh_token")
	refreshTokenClaims, err := ValidateToken("refresh_token", refreshToken)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	ctx := context.TODO()
	if err = store.Redis.Del(ctx, accessTokenClaims.ID, refreshTokenClaims.ID).Err(); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	expired := time.Now().Add(-time.Hour * 24)
	c.Cookie(&fiber.Cookie{
		Name:    "access_token",
		Value:   "",
		Expires: expired,
	})
	c.Cookie(&fiber.Cookie{
		Name:    "refresh_token",
		Value:   "",
		Expires: expired,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

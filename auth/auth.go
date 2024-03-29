package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/schema"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"gorm.io/gorm"
)

const (
	ACCESS_TOKEN  = "access_token"
	REFRESH_TOKEN = "refresh_token"
	CSRF_TOKEN    = "csrf_token"
	TOKEN_USED    = "token_used"
)

func CsrfTokenExtractor(c *fiber.Ctx) (string, error) {
	csrfToken, err := c.GetReqHeaders()[CSRF_TOKEN]
	if err || len(csrfToken) == 0 {
		csrfToken = c.Cookies(CSRF_TOKEN)

		if len(csrfToken) == 0 {
			return "", errors.New("invalid csrf token")
		}
	}

	return strings.Clone(csrfToken), nil
}

func Protect(c *fiber.Ctx) error {
	var accessToken string
	authorization := c.Get("Authorization")
	if strings.HasPrefix(authorization, "Bearer ") {
		accessToken = strings.TrimPrefix(authorization, "Bearer ")
	} else if accessTokenCookie := c.Cookies(ACCESS_TOKEN); len(accessTokenCookie) != 0 {
		accessToken = accessTokenCookie
	}

	accessTokenClaims, ok := ValidateToken(c, ACCESS_TOKEN, accessToken)
	if !ok {
		return nil
	}

	expired, ok := isExpired(c, accessTokenClaims)
	if !ok || expired {
		return nil
	}

	ctx := context.TODO()
	userId, err := store.Redis.Get(ctx, accessTokenClaims.ID).Result()
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	var user models.User
	if err := store.Database.First(&user, userId).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
	}
	if user.ID == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	c.Locals("user", user)
	c.Locals("sessionId", utils.UUIDv4())

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
		"csrfToken": csrfToken,
	})
}

func Register(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer store.RollbackTransactionIfNeeded(c, tx)

	user, ok := schema.GetRegisterUserInput(c)
	if !ok {
		return nil
	}

	if ok := store.Execute(c, tx.Create(&user).Error); !ok {
		return nil
	}

	tx.Commit()

	return c.Status(fiber.StatusCreated).JSON(models.SanitizeUser(&user))
}

func Login(c *fiber.Ctx) error {
	user, ok := schema.GetLoginUserInput(c)
	if !ok {
		return nil
	}

	accessToken, refreshToken, ok := CreateTokens(c, user.ID)
	if !ok {
		return nil
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func Refresh(c *fiber.Ctx) error {
	accessToken := c.Cookies(ACCESS_TOKEN)
	refreshToken := c.Cookies(REFRESH_TOKEN)
	if len(accessToken) == 0 || len(refreshToken) == 0 {
		var ok bool
		accessToken, refreshToken, ok = schema.GetRefreshLogoutInput(c)
		if !ok {
			return nil
		}
	}

	accessTokenClaims, ok := ValidateToken(c, ACCESS_TOKEN, accessToken)
	if !ok {
		return nil
	}

	refreshTokenClaims, ok := ValidateToken(c, REFRESH_TOKEN, refreshToken)
	if !ok {
		return nil
	}

	expired, ok := isExpired(c, refreshTokenClaims)
	if !ok || expired {
		return nil
	}

	ctx := context.TODO()
	accessTokenId, err := store.Redis.Get(ctx, refreshTokenClaims.ID).Result()
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	if accessTokenId == TOKEN_USED {
		var user models.User
		if err := store.Database.Model(&models.User{}).Where("id = ?", refreshTokenClaims.Subject).First(&user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "server error",
			})
		}

		if err := store.Database.Model(&user).Update("token_available_since", time.Now().Unix()).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "server error",
			})

		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	if accessTokenId != accessTokenClaims.ID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	var user models.User
	if err := store.Database.First(&user, accessTokenClaims.Subject).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
	}
	if user.ID == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	if _, err := store.Redis.Del(ctx, accessTokenClaims.ID).Result(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
	}
	if _, err := store.Redis.Set(ctx, refreshTokenClaims.ID, TOKEN_USED, time.Until(refreshTokenClaims.ExpiresAt.Time)).Result(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
	}

	accessToken, refreshToken, ok = CreateTokens(c, user.ID)
	if !ok {
		return nil
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func Logout(c *fiber.Ctx) error {
	accessToken := c.Cookies(ACCESS_TOKEN)
	refreshToken := c.Cookies(REFRESH_TOKEN)
	if len(accessToken) == 0 || len(refreshToken) == 0 {
		var ok bool
		accessToken, refreshToken, ok = schema.GetRefreshLogoutInput(c)
		if !ok {
			return nil
		}
	}

	accessTokenClaims, ok := ValidateToken(c, ACCESS_TOKEN, accessToken)
	if !ok {
		return nil
	}

	refreshTokenClaims, ok := ValidateToken(c, REFRESH_TOKEN, refreshToken)
	if !ok {
		return nil
	}

	expired, ok := isExpired(c, refreshTokenClaims)
	if !ok || expired {
		return nil
	}

	ctx := context.TODO()
	if err := store.Redis.Del(ctx, accessTokenClaims.ID, refreshTokenClaims.ID).Err(); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	expires := time.Now().UTC().Add(-24 * time.Hour)
	c.Cookie(&fiber.Cookie{
		Name:    ACCESS_TOKEN,
		Value:   "",
		Expires: expires,
	})
	c.Cookie(&fiber.Cookie{
		Name:    REFRESH_TOKEN,
		Value:   "",
		Expires: expires,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "ok",
	})
}

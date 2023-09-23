package schema

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type SanitizedUser struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func SanitizeUser(user *models.User) *SanitizedUser {
	return &SanitizedUser{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}

type RegisterInput struct {
	Name            string `json:"name" validate:"required"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	PasswordConfirm string `json:"passwordConfirm" validate:"required,min=8"`
}

func GetRegisterUserInput(c *fiber.Ctx) (models.User, bool) {
	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.User{}, false
	}
	if err := validate.Struct(input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.User{}, false
	}

	if input.Password != input.PasswordConfirm {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid password confirmation",
		})
		return models.User{}, false
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return models.User{}, false
	}

	return models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
	}, true
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func GetLoginUserInput(c *fiber.Ctx) (models.User, bool) {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.User{}, false
	}
	if err := validate.Struct(input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.User{}, false
	}

	var user models.User
	store.Database.Where(&models.User{Email: input.Email}).First(&user)
	if user.ID == 0 {
		c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "invalid credentials",
		})
		return models.User{}, false
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "invalid credentials",
		})
		return models.User{}, false
	}

	return user, true
}

type RefreshInput struct {
	AccessToken  string `json:"accessToken" validate:"required"`
	RefreshToken string `json:"refreshToken" validate:"required"`
}

func GetRefreshInput(c *fiber.Ctx) (string, string, bool) {
	var input RefreshInput
	if err := c.BodyParser(&input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return "", "", false
	}
	if err := validate.Struct(&input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return "", "", false
	}

	return input.AccessToken, input.RefreshToken, true
}

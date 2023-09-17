package schema

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

type RegisterInput struct {
	Name            string `json:"name" validate:"required"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	PasswordConfirm string `json:"passwordConfirm" validate:"required,min=8"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

const EMAIL_REGEX = `^.+@.+\..+$`

func Init() *validator.Validate {
	var validate = validator.New()
	validate.RegisterValidation("email", validateEmail)

	return validate
}

func validateEmail(field validator.FieldLevel) bool {
	return regexp.MustCompile(EMAIL_REGEX).MatchString(field.Field().String())
}

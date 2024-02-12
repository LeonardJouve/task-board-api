package schema

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

const EMAIL_REGEX = `^.+@.+\..+$`
const COLOR_REGEX = `^#([A-Fa-f0-9]{6})$`

var validate *validator.Validate

func Init() {
	validate = validator.New()
	validate.RegisterValidation("email", validateEmail)
	validate.RegisterValidation("color", validateColor)
}

func validateEmail(field validator.FieldLevel) bool {
	return regexp.MustCompile(EMAIL_REGEX).MatchString(field.Field().String())
}

func validateColor(field validator.FieldLevel) bool {
	return regexp.MustCompile(COLOR_REGEX).MatchString(field.Field().String())
}

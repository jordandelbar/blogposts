package validation

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

func RegisterCustomValidators(validate *validator.Validate) *validator.Validate {
	validate.RegisterValidation("no_html", validateNoHTML)
	validate.RegisterValidation("no_script_tags", validateNoScriptTags)
	validate.RegisterValidation("alphanum_hyphen", validateAlphanumHyphen)
	validate.RegisterValidation("strong_password", validateStrongPassword)

	return validate
}

func validateNoHTML(fl validator.FieldLevel) bool {
	text := fl.Field().String()
	return !strings.ContainsAny(text, "<>")
}

func validateNoScriptTags(fl validator.FieldLevel) bool {
	text := strings.ToLower(fl.Field().String())
	return !strings.Contains(text, "<script") && !strings.Contains(text, "javascript:")
}

func validateAlphanumHyphen(fl validator.FieldLevel) bool {
	slug := fl.Field().String()
	matched, _ := regexp.MatchString("^[a-zA-Z0-9-]+$", slug)
	return matched && !strings.HasPrefix(slug, "-") && !strings.HasSuffix(slug, "-")
}

func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	
	if len(password) < 8 {
		return false
	}
	
	hasUpper, _ := regexp.MatchString(`[A-Z]`, password)
	hasLower, _ := regexp.MatchString(`[a-z]`, password)
	hasDigit, _ := regexp.MatchString(`\d`, password)
	hasSpecial, _ := regexp.MatchString(`[!@#$%^&*(),.?":{}|<>]`, password)
	
	return hasUpper && hasLower && hasDigit && hasSpecial
}

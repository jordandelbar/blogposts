package dto_validation

import (
	"personal_website/pkg/validation"
	"strings"

	"github.com/go-playground/validator/v10"
)

type DtoValidator struct {
	validate *validator.Validate
	Errors   map[string][]string
}

func NewDtoValidator() *DtoValidator {
	validate := validation.RegisterCustomValidators(validator.New())

	return &DtoValidator{
		validate: validate,
		Errors:   make(map[string][]string),
	}
}

func (v *DtoValidator) ValidateStruct(s interface{}) *DtoValidator {
	err := v.validate.Struct(s)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range ve {
				message := v.getErrorMessage(fieldErr)
				fieldName := strings.ToLower(fieldErr.Field())
				v.Errors[fieldName] = append(v.Errors[fieldName], message)
			}
		}
	}
	return v
}

func (v *DtoValidator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *DtoValidator) getErrorMessage(fieldErr validator.FieldError) string {
	field := strings.ToLower(fieldErr.Field())
	tag := fieldErr.Tag()
	
	// Generic messages based on validation tag
	switch tag {
	case "required":
		return field + " is required"
	case "email":
		return "invalid email format"
	case "min":
		return field + " must be at least " + fieldErr.Param() + " characters"
	case "max":
		return field + " must not exceed " + fieldErr.Param() + " characters"
	case "strong_password":
		return "password must be at least 8 characters long and include uppercase, lowercase, digit and special character"
	case "no_html":
		return field + " cannot contain HTML tags"
	case "no_script_tags":
		return field + " cannot contain script tags"
	case "alphanum_hyphen":
		return field + " must contain only letters, numbers and hyphens"
	default:
		return "validation failed for " + field + " (" + tag + ")"
	}
}

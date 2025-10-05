package mappers

import (
	"personal_website/internal/app/core/domain"
	"personal_website/internal/infrastructure/http/dto"
)

func ContactFormToDomain(form dto.ContactForm) domain.ContactMessage {
	return domain.ContactMessage{
		Name:    form.Name,
		Email:   form.Email,
		Message: form.Message,
	}
}

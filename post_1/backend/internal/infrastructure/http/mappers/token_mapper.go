package mappers

import (
	"personal_website/internal/app/core/domain"
	"personal_website/internal/infrastructure/http/dto"
)

func ActivationTokenToDomain(activationToken dto.ActivationToken) (domain.Token, error) {
	return domain.Token{
		Plaintext: activationToken.TokenPlaintext,
	}, nil
}

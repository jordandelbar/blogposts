package mappers

import (
	"personal_website/internal/app/core/domain"
	"personal_website/internal/infrastructure/http/dto"
)

func UserRequestToDomain(req dto.UserRequest) (domain.User, error) {
	user := domain.User{
		Name:  req.Name,
		Email: req.Email,
	}

	err := user.Password.Set(req.Password)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func UserToResponse(user *domain.User) dto.UserResponse {
	response := dto.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	if !user.CreatedAt.IsZero() {
		response.CreatedAt = user.CreatedAt.Format("2006-01-02T15:04:05Z")
	}

	return response
}

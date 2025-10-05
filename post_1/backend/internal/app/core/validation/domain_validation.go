package domain_validation

import (
	"context"
	"fmt"
	"personal_website/internal/app/core/domain"
	"personal_website/internal/app/core/ports"
)

type DomainValidator struct {
	userRepo ports.UserRepository
	Errors   map[string][]string
}

func NewTokenValidator() *DomainValidator {
	return &DomainValidator{
		userRepo: nil,
		Errors:   make(map[string][]string),
	}
}

func NewUserValidator(userRepo ports.UserRepository) *DomainValidator {
	return &DomainValidator{
		userRepo: userRepo,
		Errors:   make(map[string][]string),
	}
}

func (dv *DomainValidator) Check(ok bool, key, message string) *DomainValidator {
	if !ok {
		dv.AddError(key, message)
	}

	return dv
}

func (dv *DomainValidator) AddError(key, message string) {
	dv.Errors[key] = append(dv.Errors[key], message)
}

func (dv *DomainValidator) Valid() bool {
	return len(dv.Errors) == 0
}

func (dv *DomainValidator) Error() error {
	if len(dv.Errors) > 0 {
		return fmt.Errorf("validation failed: %v", dv.Errors)
	} else {
		return nil
	}
}

func (dv *DomainValidator) ValidateUser(user *domain.User) *DomainValidator {
	if dv.userRepo == nil {
		panic("ValidateUser called on token validator - use NewUserValidator() instead")
	}

	exists, err := dv.userRepo.CheckUserExistsByEmail(context.Background(), user.Email)
	if err != nil {
		dv.AddError("database", "unable to verify email uniqueness")
		return dv
	}

	if exists {
		dv.AddError("user", "user already exists")
	}

	return dv
}

func (dv *DomainValidator) ValidateTokenPlaintext(tokenPlaintext string) *DomainValidator {
	dv.Check(tokenPlaintext != "", "token", "must be provided")
	dv.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")

	return dv
}

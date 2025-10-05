package domain_validation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"personal_website/internal/app/core/domain"
	"strings"
	"testing"
	"time"
)

type mockUserRepository struct {
	users           map[string]domain.User
	shouldReturnErr bool
	errToReturn     error
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]domain.User),
	}
}

func (m *mockUserRepository) CreateUser(ctx context.Context, user domain.User) (int, error) {
	if m.shouldReturnErr {
		return 0, m.errToReturn
	}
	m.users[user.Email] = user
	return int(user.ID), nil
}

func (m *mockUserRepository) ActivateUser(ctx context.Context, user *domain.User) error {
	if m.shouldReturnErr {
		return m.errToReturn
	}

	u, ok := m.users[user.Email]
	if !ok {
		return fmt.Errorf("user not found")
	}

	u.Activated = true
	m.users[user.Email] = u
	return nil
}

func (m *mockUserRepository) CheckUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.shouldReturnErr {
		return false, m.errToReturn
	}
	_, exists := m.users[email]
	return exists, nil
}

func (m *mockUserRepository) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	if m.shouldReturnErr {
		return domain.User{}, m.errToReturn
	}

	user, exists := m.users[email]
	if !exists {
		return domain.User{}, sql.ErrNoRows
	}
	return user, nil
}

func (m *mockUserRepository) DeactivateUser(ctx context.Context, id int) error {
	if m.shouldReturnErr {
		return m.errToReturn
	}
	for email, user := range m.users {
		if user.ID == id {
			user.Activated = false
			m.users[email] = user
			return nil
		}
	}
	return sql.ErrNoRows
}

func (m *mockUserRepository) DeleteUser(ctx context.Context, id int) error {
	if m.shouldReturnErr {
		return m.errToReturn
	}
	for email, user := range m.users {
		if user.ID == id {
			delete(m.users, email)
			return nil
		}
	}
	return sql.ErrNoRows
}

func (m *mockUserRepository) GetForToken(ctx context.Context, scope domain.TokenScope, tokenPlaintext string) (*domain.User, error) {
	if m.shouldReturnErr {
		return &domain.User{}, m.errToReturn
	}
	return &domain.User{}, nil
}

func (m *mockUserRepository) addUser(email string) {
	m.users[email] = domain.User{
		ID:        1,
		CreatedAt: time.Now(),
		Name:      "Test User",
		Email:     email,
	}
}

func (m *mockUserRepository) setError(err error) {
	m.shouldReturnErr = true
	m.errToReturn = err
}

func TestNewUserValidator(t *testing.T) {
	mockRepo := newMockUserRepository()

	validator := NewUserValidator(mockRepo)

	if validator.userRepo != mockRepo {
		t.Error("NewUserValidator() did not set userRepo correctly")
	}

	if validator.Errors == nil {
		t.Error("NewUserValidator() did not initialize Errors map")
	}

	if len(validator.Errors) != 0 {
		t.Errorf("NewUserValidator() Errors map should be empty, got %d items", len(validator.Errors))
	}
}

func TestNewTokenValidator(t *testing.T) {
	validator := NewTokenValidator()

	if validator.userRepo != nil {
		t.Error("NewTokenValidator() should set userRepo to nil")
	}

	if validator.Errors == nil {
		t.Error("NewTokenValidator() did not initialize Errors map")
	}

	if len(validator.Errors) != 0 {
		t.Errorf("NewTokenValidator() Errors map should be empty, got %d items", len(validator.Errors))
	}
}

func TestDomainValidator_ValidateUser_HappyPath(t *testing.T) {
	tests := []struct {
		name string
		user domain.User
	}{
		{
			name: "new unique user",
			user: domain.User{
				Name:  "John Doe",
				Email: "john@example.com",
			},
		},
		{
			name: "another unique user",
			user: domain.User{
				Name:  "Jane Smith",
				Email: "jane@test.org",
			},
		},
		{
			name: "valid email format",
			user: domain.User{
				Name:  "User Name",
				Email: "user123@domain.co.uk",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockUserRepository()
			validator := NewUserValidator(mockRepo)

			result := validator.ValidateUser(&tt.user)

			if result != validator {
				t.Error("ValidateUser() should return itself for method chaining")
			}

			if !validator.Valid() {
				t.Errorf("ValidateUser() with unique email should pass validation, got errors: %v", validator.Error())
			}

			if len(validator.Errors) != 0 {
				t.Errorf("ValidateUser() with unique email should have no errors, got %d", len(validator.Errors))
			}
		})
	}
}

func TestDomainValidator_ValidateUser_EmailAlreadyExists(t *testing.T) {
	tests := []struct {
		name          string
		user          domain.User
		existingEmail string
	}{
		{
			name: "exact email match",
			user: domain.User{
				Name:  "John Doe",
				Email: "test@example.com",
			},
			existingEmail: "test@example.com",
		},
		{
			name: "common email",
			user: domain.User{
				Name:  "Admin User",
				Email: "admin@site.com",
			},
			existingEmail: "admin@site.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockUserRepository()
			mockRepo.addUser(tt.existingEmail)
			validator := NewUserValidator(mockRepo)

			result := validator.ValidateUser(&tt.user)

			if result != validator {
				t.Error("ValidateUser() should return itself for method chaining")
			}
			if validator.Valid() {
				t.Errorf("ValidateUser() with existing email should fail validation")
			}

			if len(validator.Errors) != 1 {
				t.Errorf("ValidateUser() with existing email should have 1 error, got %d", len(validator.Errors))
			}

			userErrors, exists := validator.Errors["user"]
			if !exists {
				t.Error("Expected error for 'user' field, but none found")
			}

			if len(userErrors) != 1 {
				t.Errorf("Expected 1 user error, got %d", len(userErrors))
			}

			if len(userErrors) > 0 && userErrors[0] != "user already exists" {
				t.Errorf("Expected error message 'user already exists', got '%s'", userErrors[0])
			}
		})
	}
}

func TestDomainValidator_ValidateUser_DatabaseError(t *testing.T) {
	tests := []struct {
		name string
		user domain.User
		err  error
	}{
		{
			name: "database connection error",
			user: domain.User{
				Name:  "Test User",
				Email: "test@example.com",
			},
			err: errors.New("database connection failed"),
		},
		{
			name: "query timeout error",
			user: domain.User{
				Name:  "Timeout User",
				Email: "timeout@example.com",
			},
			err: errors.New("query timeout"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockUserRepository()
			mockRepo.setError(tt.err)
			validator := NewUserValidator(mockRepo)

			result := validator.ValidateUser(&tt.user)

			if result != validator {
				t.Error("ValidateUser() should return itself for method chaining")
			}

			if validator.Valid() {
				t.Error("ValidateUser() with database error should fail validation")
			}

			if len(validator.Errors) != 1 {
				t.Errorf("ValidateUser() with database error should have 1 error, got %d", len(validator.Errors))
			}

			databaseErrors, exists := validator.Errors["database"]
			if !exists {
				t.Error("Expected error for 'database' field, but none found")
			}

			if len(databaseErrors) != 1 {
				t.Errorf("Expected 1 database error, got %d", len(databaseErrors))
			}

			if len(databaseErrors) > 0 && databaseErrors[0] != "unable to verify email uniqueness" {
				t.Errorf("Expected error message 'unable to verify email uniqueness', got '%s'", databaseErrors[0])
			}
		})
	}
}

func TestDomainValidator_ValidateTokenPlaintext_ValidTokens(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "valid 26 character token",
			token: "abcdefghijklmnopqrstuvwxyz",
		},
		{
			name:  "valid 26 character numeric token",
			token: "12345678901234567890123456",
		},
		{
			name:  "valid mixed token",
			token: "a1b2c3d4e5f6g7h8i9j0k1l2m3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewTokenValidator()

			result := validator.ValidateTokenPlaintext(tt.token)

			if result != validator {
				t.Error("ValidateTokenPlaintext() should return itself for method chaining")
			}

			if !validator.Valid() {
				t.Errorf("ValidateTokenPlaintext() with valid token should pass validation, got errors: %v", validator.Error())
			}

			if len(validator.Errors) != 0 {
				t.Errorf("ValidateTokenPlaintext() with valid token should have no errors, got %d", len(validator.Errors))
			}
		})
	}
}

func TestDomainValidator_ValidateTokenPlaintext_InvalidTokens(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		expectedErrors []string
	}{
		{
			name:           "empty token",
			token:          "",
			expectedErrors: []string{"must be provided", "must be 26 bytes long"},
		},
		{
			name:           "short token",
			token:          "short",
			expectedErrors: []string{"must be 26 bytes long"},
		},
		{
			name:           "long token",
			token:          "this_token_is_way_too_long_to_be_valid",
			expectedErrors: []string{"must be 26 bytes long"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewTokenValidator()

			result := validator.ValidateTokenPlaintext(tt.token)

			if result != validator {
				t.Error("ValidateTokenPlaintext() should return itself for method chaining")
			}

			if validator.Valid() {
				t.Error("ValidateTokenPlaintext() with invalid token should fail validation")
			}

			tokenErrors, exists := validator.Errors["token"]
			if !exists {
				t.Error("Expected error for 'token' field, but none found")
			}

			if len(tokenErrors) != len(tt.expectedErrors) {
				t.Errorf("Expected %d token errors, got %d", len(tt.expectedErrors), len(tokenErrors))
			}

			for _, expectedError := range tt.expectedErrors {
				found := false
				for _, actualError := range tokenErrors {
					if actualError == expectedError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error message '%s' not found in %v", expectedError, tokenErrors)
				}
			}
		})
	}
}

func TestDomainValidator_Check(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		key       string
		message   string
		shouldErr bool
	}{
		{
			name:      "passing condition should not add error",
			condition: true,
			key:       "test",
			message:   "should not appear",
			shouldErr: false,
		},
		{
			name:      "failing condition should add error",
			condition: false,
			key:       "test",
			message:   "test error message",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewTokenValidator()

			result := validator.Check(tt.condition, tt.key, tt.message)

			if result != validator {
				t.Error("Check() should return itself for method chaining")
			}

			if tt.shouldErr {
				if validator.Valid() {
					t.Error("Check() with failing condition should make validator invalid")
				}

				errors, exists := validator.Errors[tt.key]
				if !exists {
					t.Errorf("Expected error for key '%s', but none found", tt.key)
				}

				if len(errors) != 1 {
					t.Errorf("Expected 1 error for key '%s', got %d", tt.key, len(errors))
				}

				if len(errors) > 0 && errors[0] != tt.message {
					t.Errorf("Expected error message '%s', got '%s'", tt.message, errors[0])
				}
			} else {
				if !validator.Valid() {
					t.Error("Check() with passing condition should keep validator valid")
				}

				if len(validator.Errors) != 0 {
					t.Errorf("Check() with passing condition should have no errors, got %d", len(validator.Errors))
				}
			}
		})
	}
}

func TestDomainValidator_Valid(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*DomainValidator)
		expected bool
	}{
		{
			name:     "empty validator should be valid",
			setup:    func(v *DomainValidator) {},
			expected: true,
		},
		{
			name: "validator with errors should be invalid",
			setup: func(v *DomainValidator) {
				v.Errors["test"] = []string{"error message"}
			},
			expected: false,
		},
		{
			name: "validator with multiple errors should be invalid",
			setup: func(v *DomainValidator) {
				v.Errors["field1"] = []string{"error 1"}
				v.Errors["field2"] = []string{"error 2", "error 3"}
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewTokenValidator()
			tt.setup(validator)

			result := validator.Valid()

			if result != tt.expected {
				t.Errorf("Valid() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestDomainValidator_Error(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*DomainValidator)
		hasError bool
	}{
		{
			name:     "empty validator should return nil error",
			setup:    func(v *DomainValidator) {},
			hasError: false,
		},
		{
			name: "validator with errors should return error",
			setup: func(v *DomainValidator) {
				v.Errors["test"] = []string{"error message"}
			},
			hasError: true,
		},
		{
			name: "validator with multiple errors should return error",
			setup: func(v *DomainValidator) {
				v.Errors["field1"] = []string{"error 1"}
				v.Errors["field2"] = []string{"error 2"}
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewTokenValidator()
			tt.setup(validator)

			result := validator.Error()

			if tt.hasError {
				if result == nil {
					t.Error("Error() should return error when validator has errors")
				}
				if !strings.Contains(result.Error(), "validation failed:") {
					t.Errorf("Error() should contain 'validation failed:', got %s", result.Error())
				}
			} else {
				if result != nil {
					t.Errorf("Error() should return nil when validator has no errors, got %s", result.Error())
				}
			}
		})
	}
}

func TestDomainValidator_ValidateUser_PanicOnTokenValidator(t *testing.T) {
	validator := NewTokenValidator()

	user := &domain.User{
		Name:  "Test User",
		Email: "test@example.com",
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("ValidateUser() on token validator should panic")
		} else {
			expectedMsg := "ValidateUser called on token validator - use NewUserValidator() instead"
			if r != expectedMsg {
				t.Errorf("Expected panic message '%s', got '%v'", expectedMsg, r)
			}
		}
	}()

	validator.ValidateUser(user)
}

func TestDomainValidator_MethodChaining(t *testing.T) {
	mockRepo := newMockUserRepository()
	validator := NewUserValidator(mockRepo)

	user := &domain.User{
		Name:  "Test User",
		Email: "test@example.com",
	}

	// Test chaining ValidateUser with ValidateTokenPlaintext
	result := validator.ValidateUser(user).ValidateTokenPlaintext("abcdefghijklmnopqrstuvwxyz")

	if result != validator {
		t.Error("Method chaining should return the same validator instance")
	}

	if !validator.Valid() {
		t.Errorf("Chained validation with valid data should pass, got errors: %v", validator.Error())
	}

	// Test chaining with Check method
	validator2 := NewTokenValidator()
	result2 := validator2.Check(true, "test", "message").ValidateTokenPlaintext("12345678901234567890123456")

	if result2 != validator2 {
		t.Error("Method chaining with Check should return the same validator instance")
	}

	if !validator2.Valid() {
		t.Errorf("Chained validation with valid conditions should pass, got errors: %v", validator2.Error())
	}
}

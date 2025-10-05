package dto_validation

import (
	"personal_website/internal/infrastructure/http/dto"
	"strings"
	"testing"
)

func TestNewDtoValidator(t *testing.T) {
	validator := NewDtoValidator()

	if validator == nil {
		t.Error("NewDtoValidator() returned nil")
	}

	if validator.validate == nil {
		t.Error("NewDtoValidator() did not initialize validator")
	}

	if validator.Errors == nil {
		t.Error("NewDtoValidator() did not initialize Errors map")
	}

	if len(validator.Errors) != 0 {
		t.Errorf("NewDtoValidator() Errors map should be empty, got %d items", len(validator.Errors))
	}
}

func TestDtoValidator_ValidateStruct_ValidData(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
	}{
		{
			name: "valid user request",
			data: dto.UserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "Password123!",
			},
		},
		{
			name: "valid user with minimum values",
			data: dto.UserRequest{
				Name:     "Jo",
				Email:    "j@e.co",
				Password: "Password123!",
			},
		},
		{
			name: "valid user with maximum name length",
			data: dto.UserRequest{
				Name:     strings.Repeat("a", 100),
				Email:    "test@example.com",
				Password: "Password123!",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewDtoValidator()

			result := validator.ValidateStruct(tt.data)

			if result != validator {
				t.Error("ValidateStruct() should return itself for method chaining")
			}

			if len(validator.Errors) != 0 {
				t.Errorf("ValidateStruct() with valid data should have no errors, got %d", len(validator.Errors))
			}
		})
	}
}

func TestDtoValidator_ValidateStruct_InvalidData(t *testing.T) {
	tests := []struct {
		name           string
		data           interface{}
		expectedFields []string
	}{
		{
			name:           "empty user request",
			data:           dto.UserRequest{},
			expectedFields: []string{"name", "email", "password"},
		},
		{
			name: "invalid email format",
			data: dto.UserRequest{
				Name:     "John Doe",
				Email:    "invalid-email",
				Password: "Password123!",
			},
			expectedFields: []string{"email"},
		},
		{
			name: "short name",
			data: dto.UserRequest{
				Name:     "J",
				Email:    "john@example.com",
				Password: "Password123!",
			},
			expectedFields: []string{"name"},
		},
		{
			name: "short password",
			data: dto.UserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "1234567",
			},
			expectedFields: []string{"password"},
		},
		{
			name: "multiple validation failures",
			data: dto.UserRequest{
				Name:     "",
				Email:    "invalid",
				Password: "123",
			},
			expectedFields: []string{"name", "email", "password"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewDtoValidator()

			result := validator.ValidateStruct(tt.data)

			if result != validator {
				t.Error("ValidateStruct() should return itself for method chaining")
			}

			if validator.Valid() {
				t.Error("ValidateStruct() with invalid data should fail validation")
			}

			if len(validator.Errors) == 0 {
				t.Error("ValidateStruct() with invalid data should have errors")
			}

			// Check that expected fields have errors
			for _, expectedField := range tt.expectedFields {
				if _, hasError := validator.Errors[expectedField]; !hasError {
					t.Errorf("Expected validation error for field %s, but none found", expectedField)
				}
			}
		})
	}
}

func TestDtoValidator_Valid(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*DtoValidator)
		expected bool
	}{
		{
			name:     "empty validator should be valid",
			setup:    func(v *DtoValidator) {},
			expected: true,
		},
		{
			name: "validator with errors should be invalid",
			setup: func(v *DtoValidator) {
				v.Errors["test"] = []string{"error message"}
			},
			expected: false,
		},
		{
			name: "validator with multiple errors should be invalid",
			setup: func(v *DtoValidator) {
				v.Errors["field1"] = []string{"error 1"}
				v.Errors["field2"] = []string{"error 2", "error 3"}
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewDtoValidator()
			tt.setup(validator)

			result := validator.Valid()

			if result != tt.expected {
				t.Errorf("Valid() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestDtoValidator_Errors(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*DtoValidator)
		expectedCount int
	}{
		{
			name:          "empty validator should have no errors",
			setup:         func(v *DtoValidator) {},
			expectedCount: 0,
		},
		{
			name: "validator with errors should have errors",
			setup: func(v *DtoValidator) {
				v.Errors["test"] = []string{"error message"}
			},
			expectedCount: 1,
		},
		{
			name: "validator with multiple errors should have multiple errors",
			setup: func(v *DtoValidator) {
				v.Errors["field1"] = []string{"error 1"}
				v.Errors["field2"] = []string{"error 2"}
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewDtoValidator()
			tt.setup(validator)

			if len(validator.Errors) != tt.expectedCount {
				t.Errorf("Expected %d errors, got %d", tt.expectedCount, len(validator.Errors))
			}
		})
	}
}

func TestDtoValidator_ErrorMessages(t *testing.T) {
	tests := []struct {
		name            string
		data            dto.UserRequest
		expectedField   string
		expectedMessage string
	}{
		{
			name: "invalid email should have correct message",
			data: dto.UserRequest{
				Name:     "John Doe",
				Email:    "invalid-email",
				Password: "ValidPass123!",
			},
			expectedField:   "email",
			expectedMessage: "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewDtoValidator()
			validator.ValidateStruct(tt.data)

			if validator.Valid() {
				t.Error("Expected validation to fail")
				return
			}

			errors, exists := validator.Errors[tt.expectedField]
			if !exists {
				t.Errorf("Expected error for field %s, but none found", tt.expectedField)
				return
			}

			if len(errors) == 0 {
				t.Errorf("Expected error messages for field %s, but none found", tt.expectedField)
				return
			}

			found := false
			for _, msg := range errors {
				if msg == tt.expectedMessage {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected error message %q for field %s, but got %v", tt.expectedMessage, tt.expectedField, errors)
			}
		})
	}
}

func TestDtoValidator_MethodChaining(t *testing.T) {
	validator := NewDtoValidator()

	// Test that we can chain method calls
	result := validator.ValidateStruct(dto.UserRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	})

	if result != validator {
		t.Error("ValidateStruct() should return the same validator instance for method chaining")
	}

	// Test chaining with validation check
	validator2 := NewDtoValidator()
	isValid := validator2.ValidateStruct(dto.UserRequest{}).Valid()

	if isValid {
		t.Error("Expected chained validation to fail for empty struct")
	}
}

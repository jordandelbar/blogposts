package domain

import (
	"testing"
	"time"
)

func TestUser_IsAnonymous(t *testing.T) {
	tests := []struct {
		name   string
		userID int
		want   bool
	}{
		{
			name:   "anonymous user with ID 0",
			userID: 0,
			want:   true,
		},
		{
			name:   "authenticated user with positive ID",
			userID: 1,
			want:   false,
		},
		{
			name:   "authenticated user with larger ID",
			userID: 12345,
			want:   false,
		},
		{
			name:   "user with negative ID (edge case)",
			userID: -1,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				ID: tt.userID,
			}

			result := u.IsAnonymous()

			if result != tt.want {
				t.Errorf("User.IsAnonymous() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestUser_Fields(t *testing.T) {
	// Test that User struct fields can be properly set and retrieved
	now := time.Now()
	testPassword := password{}
	err := testPassword.Set("testpassword123")
	if err != nil {
		t.Fatalf("Failed to set test password: %v", err)
	}

	user := User{
		ID:        42,
		CreatedAt: now,
		Name:      "John Doe",
		Email:     "john.doe@example.com",
		Password:  testPassword,
		Activated: true,
	}

	// Verify all fields are set correctly
	if user.ID != 42 {
		t.Errorf("User.ID = %v, want %v", user.ID, 42)
	}

	if !user.CreatedAt.Equal(now) {
		t.Errorf("User.CreatedAt = %v, want %v", user.CreatedAt, now)
	}

	if user.Name != "John Doe" {
		t.Errorf("User.Name = %v, want %v", user.Name, "John Doe")
	}

	if user.Email != "john.doe@example.com" {
		t.Errorf("User.Email = %v, want %v", user.Email, "john.doe@example.com")
	}

	if !user.Activated {
		t.Errorf("User.Activated = %v, want %v", user.Activated, true)
	}

	// Test password functionality
	match, err := user.Password.Matches("testpassword123")
	if err != nil {
		t.Errorf("User.Password.Matches() error = %v", err)
	}

	if !match {
		t.Error("User.Password should match the set password")
	}
}

func TestUser_ZeroValues(t *testing.T) {
	// Test zero values behavior
	var user User

	if !user.IsAnonymous() {
		t.Error("Zero-value User should be anonymous")
	}

	if user.ID != 0 {
		t.Errorf("Zero-value User.ID = %v, want 0", user.ID)
	}

	if user.Name != "" {
		t.Errorf("Zero-value User.Name = %v, want empty string", user.Name)
	}

	if user.Email != "" {
		t.Errorf("Zero-value User.Email = %v, want empty string", user.Email)
	}

	if user.Activated {
		t.Error("Zero-value User.Activated should be false")
	}

	// Test zero-value time
	zeroTime := time.Time{}
	if !user.CreatedAt.Equal(zeroTime) {
		t.Errorf("Zero-value User.CreatedAt should be zero time")
	}
}

func TestUser_PasswordIntegration(t *testing.T) {
	// Test password integration with User struct
	user := User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
	}

	// Set password
	err := user.Password.Set("mypassword123")
	if err != nil {
		t.Fatalf("Failed to set user password: %v", err)
	}

	// Verify password works
	match, err := user.Password.Matches("mypassword123")
	if err != nil {
		t.Errorf("User password matching failed: %v", err)
	}

	if !match {
		t.Error("User password should match")
	}

	// Verify wrong password fails
	match, err = user.Password.Matches("wrongpassword")
	if err != nil {
		t.Errorf("User password matching failed: %v", err)
	}

	if match {
		t.Error("User password should not match wrong password")
	}

	// Test user is not anonymous after setting ID
	if user.IsAnonymous() {
		t.Error("User with ID should not be anonymous")
	}
}
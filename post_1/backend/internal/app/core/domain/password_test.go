package domain

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestPassword_Set(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		wantErr   bool
	}{
		{
			name:      "valid password",
			plaintext: "validpassword123",
			wantErr:   false,
		},
		{
			name:      "empty password",
			plaintext: "",
			wantErr:   false, // bcrypt allows empty passwords
		},
		{
			name:      "long password",
			plaintext: "this_is_a_very_long_password_that_should_still_work_fine_123456789",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &password{}
			err := p.Set(tt.plaintext)

			if (err != nil) != tt.wantErr {
				t.Errorf("password.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify plaintext is stored
				if p.plaintext == nil || *p.plaintext != tt.plaintext {
					t.Errorf("password.Set() plaintext not stored correctly")
				}

				// Verify hash is generated
				if p.hash == nil || len(p.hash) == 0 {
					t.Errorf("password.Set() hash not generated")
				}

				// Verify hash can be used with bcrypt
				err = bcrypt.CompareHashAndPassword(p.hash, []byte(tt.plaintext))
				if err != nil {
					t.Errorf("password.Set() generated invalid hash: %v", err)
				}
			}
		})
	}
}

func TestPassword_Matches(t *testing.T) {
	tests := []struct {
		name           string
		setupPassword  string
		testPassword   string
		wantMatch      bool
		wantErr        bool
	}{
		{
			name:          "matching passwords",
			setupPassword: "mypassword123",
			testPassword:  "mypassword123",
			wantMatch:     true,
			wantErr:       false,
		},
		{
			name:          "non-matching passwords",
			setupPassword: "mypassword123",
			testPassword:  "wrongpassword",
			wantMatch:     false,
			wantErr:       false,
		},
		{
			name:          "empty test password",
			setupPassword: "mypassword123",
			testPassword:  "",
			wantMatch:     false,
			wantErr:       false,
		},
		{
			name:          "both empty passwords",
			setupPassword: "",
			testPassword:  "",
			wantMatch:     true,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &password{}
			err := p.Set(tt.setupPassword)
			if err != nil {
				t.Fatalf("Failed to set up password: %v", err)
			}

			match, err := p.Matches(tt.testPassword)

			if (err != nil) != tt.wantErr {
				t.Errorf("password.Matches() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if match != tt.wantMatch {
				t.Errorf("password.Matches() = %v, want %v", match, tt.wantMatch)
			}
		})
	}
}

func TestPassword_Matches_WithInvalidHash(t *testing.T) {
	p := &password{}
	p.hash = []byte("invalid_hash")

	match, err := p.Matches("anypassword")

	if err == nil {
		t.Error("password.Matches() expected error with invalid hash, got nil")
	}

	if match {
		t.Error("password.Matches() expected false match with invalid hash, got true")
	}
}

func TestPassword_Hash(t *testing.T) {
	p := &password{}
	testPassword := "testpassword123"

	err := p.Set(testPassword)
	if err != nil {
		t.Fatalf("Failed to set password: %v", err)
	}

	hash := p.Hash()

	if hash == nil || len(hash) == 0 {
		t.Error("password.Hash() returned empty hash")
	}

	// Verify returned hash matches internal hash
	if string(hash) != string(p.hash) {
		t.Error("password.Hash() returned different hash than stored internally")
	}
}

func TestPassword_SetHash(t *testing.T) {
	p := &password{}
	testHash := []byte("test_hash_bytes")

	p.SetHash(testHash)

	if string(p.hash) != string(testHash) {
		t.Error("password.SetHash() did not store hash correctly")
	}

	returnedHash := p.Hash()
	if string(returnedHash) != string(testHash) {
		t.Error("password.SetHash() stored hash cannot be retrieved correctly")
	}
}

func TestPassword_SetHash_Integration(t *testing.T) {
	// Test setting a real bcrypt hash
	plaintext := "realpassword123"

	// Generate a real bcrypt hash
	realHash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		t.Fatalf("Failed to generate bcrypt hash: %v", err)
	}

	p := &password{}
	p.SetHash(realHash)

	// Verify the hash works for matching
	match, err := p.Matches(plaintext)
	if err != nil {
		t.Errorf("password.Matches() with SetHash failed: %v", err)
	}

	if !match {
		t.Error("password.Matches() with SetHash should match original plaintext")
	}

	// Verify it rejects wrong password
	match, err = p.Matches("wrongpassword")
	if err != nil {
		t.Errorf("password.Matches() with SetHash failed: %v", err)
	}

	if match {
		t.Error("password.Matches() with SetHash should not match wrong plaintext")
	}
}
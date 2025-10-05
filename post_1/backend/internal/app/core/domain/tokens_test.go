package domain

import (
	"crypto/sha256"
	"testing"
	"time"
)

func TestTokenScope_String(t *testing.T) {
	tests := []struct {
		name    string
		scope   TokenScope
		want    string
		wantErr bool
	}{
		{
			name:    "activation scope",
			scope:   ScopeActivation,
			want:    "activation",
			wantErr: false,
		},
		{
			name:    "authentication scope",
			scope:   ScopeAuthentication,
			want:    "authentication",
			wantErr: false,
		},
		{
			name:    "invalid scope",
			scope:   TokenScope(999),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.scope.String()

			if (err != nil) != tt.wantErr {
				t.Errorf("TokenScope.String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != tt.want {
				t.Errorf("TokenScope.String() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestTokenScope_Constants(t *testing.T) {
	// Test that constants have expected values
	if ScopeActivation != 0 {
		t.Errorf("ScopeActivation = %v, want 0", ScopeActivation)
	}

	if ScopeAuthentication != 1 {
		t.Errorf("ScopeAuthentication = %v, want 1", ScopeAuthentication)
	}
}

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name   string
		userID int
		scope  TokenScope
	}{
		{
			name:   "activation token",
			userID: 123,
			scope:  ScopeActivation,
		},
		{
			name:   "authentication token",
			userID: 456,
			scope:  ScopeAuthentication,
		},
		{
			name:   "zero user ID",
			userID: 0,
			scope:  ScopeActivation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeGeneration := time.Now()

			token := GenerateToken(tt.userID, tt.scope)

			afterGeneration := time.Now()

			// Verify token is not nil
			if token == nil {
				t.Fatal("GenerateToken() returned nil")
			}

			// Verify user ID is set correctly
			if token.UserID != tt.userID {
				t.Errorf("Token.UserID = %v, want %v", token.UserID, tt.userID)
			}

			// Verify scope is set correctly
			if token.Scope != tt.scope {
				t.Errorf("Token.Scope = %v, want %v", token.Scope, tt.scope)
			}

			// Verify plaintext is generated and not empty
			if token.Plaintext == "" {
				t.Error("Token.Plaintext should not be empty")
			}

			// Verify hash is generated
			if token.Hash == nil || len(token.Hash) == 0 {
				t.Error("Token.Hash should not be empty")
			}

			// Verify hash matches plaintext
			expectedHash := sha256.Sum256([]byte(token.Plaintext))
			if string(token.Hash) != string(expectedHash[:]) {
				t.Error("Token.Hash does not match expected SHA256 of plaintext")
			}

			// Verify expiry is set to ~24 hours from now
			if token.Expiry.Before(beforeGeneration) {
				t.Error("Token.Expiry should be after generation time")
			}

			if token.Expiry.After(afterGeneration.Add(24*time.Hour + time.Minute)) {
				t.Error("Token.Expiry should be approximately 24 hours from generation")
			}

			// Should be close to 24 hours (within 1 minute tolerance for test execution time)
			timeDiff := token.Expiry.Sub(beforeGeneration)
			expectedDuration := 24 * time.Hour

			if timeDiff < expectedDuration-time.Minute || timeDiff > expectedDuration+time.Minute {
				t.Errorf("Token expiry duration = %v, want ~%v", timeDiff, expectedDuration)
			}
		})
	}
}

func TestGenerateToken_UniquePlaintexts(t *testing.T) {
	// Generate multiple tokens and verify they have unique plaintexts
	const numTokens = 100
	plaintexts := make(map[string]bool)

	for i := 0; i < numTokens; i++ {
		token := GenerateToken(1, ScopeActivation)

		if plaintexts[token.Plaintext] {
			t.Errorf("Generated duplicate plaintext: %s", token.Plaintext)
		}

		plaintexts[token.Plaintext] = true
	}

	if len(plaintexts) != numTokens {
		t.Errorf("Expected %d unique plaintexts, got %d", numTokens, len(plaintexts))
	}
}

func TestGenerateToken_UniqueHashes(t *testing.T) {
	// Generate multiple tokens and verify they have unique hashes
	const numTokens = 100
	hashes := make(map[string]bool)

	for i := 0; i < numTokens; i++ {
		token := GenerateToken(1, ScopeActivation)
		hashString := string(token.Hash)

		if hashes[hashString] {
			t.Errorf("Generated duplicate hash")
		}

		hashes[hashString] = true
	}

	if len(hashes) != numTokens {
		t.Errorf("Expected %d unique hashes, got %d", numTokens, len(hashes))
	}
}

func TestToken_Fields(t *testing.T) {
	// Test Token struct field assignment and retrieval
	now := time.Now()
	plaintext := "test_plaintext_123"
	hash := sha256.Sum256([]byte(plaintext))

	token := Token{
		Plaintext: plaintext,
		Hash:      hash[:],
		UserID:    42,
		Expiry:    now.Add(time.Hour),
		Scope:     ScopeAuthentication,
	}

	// Verify all fields
	if token.Plaintext != plaintext {
		t.Errorf("Token.Plaintext = %v, want %v", token.Plaintext, plaintext)
	}

	if string(token.Hash) != string(hash[:]) {
		t.Error("Token.Hash not set correctly")
	}

	if token.UserID != 42 {
		t.Errorf("Token.UserID = %v, want %v", token.UserID, 42)
	}

	if token.Scope != ScopeAuthentication {
		t.Errorf("Token.Scope = %v, want %v", token.Scope, ScopeAuthentication)
	}

	expectedExpiry := now.Add(time.Hour)
	if !token.Expiry.Equal(expectedExpiry) {
		t.Errorf("Token.Expiry = %v, want %v", token.Expiry, expectedExpiry)
	}
}

func TestToken_ZeroValues(t *testing.T) {
	// Test zero values behavior
	var token Token

	if token.Plaintext != "" {
		t.Errorf("Zero-value Token.Plaintext = %v, want empty string", token.Plaintext)
	}

	if token.Hash != nil {
		t.Errorf("Zero-value Token.Hash = %v, want nil", token.Hash)
	}

	if token.UserID != 0 {
		t.Errorf("Zero-value Token.UserID = %v, want 0", token.UserID)
	}

	if token.Scope != ScopeActivation {
		t.Errorf("Zero-value Token.Scope = %v, want %v (ScopeActivation)", token.Scope, ScopeActivation)
	}

	zeroTime := time.Time{}
	if !token.Expiry.Equal(zeroTime) {
		t.Error("Zero-value Token.Expiry should be zero time")
	}
}
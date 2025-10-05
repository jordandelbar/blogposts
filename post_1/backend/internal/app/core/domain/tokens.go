package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"time"
)

type TokenScope int

const (
	ScopeActivation TokenScope = iota
	ScopeAuthentication
	ScopeRefresh
)

func (t TokenScope) String() (string, error) {
	switch t {
	case ScopeActivation:
		return "activation", nil
	case ScopeAuthentication:
		return "authentication", nil
	case ScopeRefresh:
		return "refresh", nil
	default:
		return "", errors.New("Incorrect token scope")
	}
}

type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int
	Expiry    time.Time
	Scope     TokenScope
}

func GenerateToken(userID int, scope TokenScope) *Token {
	ttl := 24 * time.Hour
	return GenerateTokenWithTTL(userID, scope, ttl)
}

func GenerateTokenWithTTL(userID int, scope TokenScope, ttl time.Duration) *Token {
	plaintext := rand.Text()
	if plaintext == "" {
		panic("failed to generate random token text")
	}

	token := &Token{
		Plaintext: plaintext,
		UserID:    userID,
		Expiry:    time.Now().Add(ttl),
		Scope:     scope,
	}

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token
}

func GenerateAccessToken(userID int) *Token {
	return GenerateTokenWithTTL(userID, ScopeAuthentication, 10*time.Minute)
}

func GenerateRefreshToken(userID int) *Token {
	return GenerateTokenWithTTL(userID, ScopeRefresh, 4*24*time.Hour)
}

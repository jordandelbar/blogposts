package valkey_adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"personal_website/internal/app/core/domain"

	valkey "github.com/valkey-io/valkey-go"
)

type sessionAdapter struct {
	client valkey.Client
}

func NewSessionAdapter(client valkey.Client) *sessionAdapter {
	return &sessionAdapter{
		client: client,
	}
}

type sessionData struct {
	UserID      int                `json:"user_id"`
	Email       string             `json:"email"`
	Permissions domain.Permissions `json:"permissions"`
	Activated   bool               `json:"activated"`
	ExpiresAt   time.Time          `json:"expires_at"`
}

func (s *sessionAdapter) buildKey(token string, scope domain.TokenScope) (string, error) {
	scopeStr, err := scope.String()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:token:%s", scopeStr, token), nil
}

func (s *sessionAdapter) buildUserIndexKey(userID int, scope domain.TokenScope) string {
	scopeStr, _ := scope.String()
	return fmt.Sprintf("user:%d:%s:sessions", userID, scopeStr)
}

func (s *sessionAdapter) StoreSession(ctx context.Context, token string, scope domain.TokenScope, session *domain.Session) error {
	key, err := s.buildKey(token, scope)
	if err != nil {
		return domain.NewInternalError(err)
	}

	// Set TTL based on token scope
	var ttl time.Duration
	switch scope {
	case domain.ScopeAuthentication:
		ttl = 10 * time.Minute // Access tokens
	case domain.ScopeRefresh:
		ttl = 4 * 24 * time.Hour // Refresh tokens (4 days)
	default:
		ttl = 24 * time.Hour // Default fallback
	}

	expiresAt := time.Now().Add(ttl)
	data := sessionData{
		UserID:      session.UserID,
		Email:       session.Email,
		Permissions: session.Permissions,
		Activated:   session.Activated,
		ExpiresAt:   expiresAt,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return domain.NewInternalError(err)
	}

	setCmd := s.client.B().Set().Key(key).Value(string(jsonData)).Ex(ttl).Build()
	if err := s.client.Do(ctx, setCmd).Error(); err != nil {
		return domain.NewInternalError(err)
	}

	// Add to user index (no TTL - let individual sessions handle expiry)
	userIndexKey := s.buildUserIndexKey(session.UserID, scope)
	saddCmd := s.client.B().Sadd().Key(userIndexKey).Member(token).Build()
	if err := s.client.Do(ctx, saddCmd).Error(); err != nil {
		return domain.NewInternalError(err)
	}

	return nil
}

func (s *sessionAdapter) GetSession(ctx context.Context, token string, scope domain.TokenScope) (*domain.Session, error) {
	key, err := s.buildKey(token, scope)
	if err != nil {
		return nil, domain.NewInternalError(err)
	}

	cmd := s.client.B().Get().Key(key).Build()
	result := s.client.Do(ctx, cmd)

	if result.Error() != nil {
		if valkey.IsValkeyNil(result.Error()) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, domain.NewInternalError(result.Error())
	}

	jsonStr, err := result.ToString()
	if err != nil {
		return nil, domain.NewInternalError(err)
	}

	var data sessionData
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, domain.NewInternalError(err)
	}

	return &domain.Session{
		UserID:      data.UserID,
		Email:       data.Email,
		Permissions: data.Permissions,
		Activated:   data.Activated,
	}, nil
}

func (s *sessionAdapter) DeleteAllSessionsForUser(ctx context.Context, userID int, scope domain.TokenScope) error {
	userIndexKey := s.buildUserIndexKey(userID, scope)

	// Get all tokens for this user
	cmd := s.client.B().Smembers().Key(userIndexKey).Build()
	result := s.client.Do(ctx, cmd)

	if result.Error() != nil {
		if valkey.IsValkeyNil(result.Error()) {
			return nil // No sessions to delete
		}
		return domain.NewInternalError(result.Error())
	}

	tokens, err := result.AsStrSlice()
	if err != nil {
		return domain.NewInternalError(err)
	}

	if len(tokens) == 0 {
		return nil
	}

	// Delete all session keys
	for _, token := range tokens {
		sessionKey, err := s.buildKey(token, scope)
		if err != nil {
			continue
		}
		delCmd := s.client.B().Del().Key(sessionKey).Build()
		s.client.Do(ctx, delCmd)
	}

	// Delete the user index
	delIndexCmd := s.client.B().Del().Key(userIndexKey).Build()
	return s.client.Do(ctx, delIndexCmd).Error()
}

func (s *sessionAdapter) DeleteSession(ctx context.Context, token string) error {
	// Try to delete with authentication scope (most common)
	key, err := s.buildKey(token, domain.ScopeAuthentication)
	if err != nil {
		return domain.NewInternalError(err)
	}

	// Get session first to find user ID for index cleanup
	session, err := s.GetSession(ctx, token, domain.ScopeAuthentication)
	if err != nil && err != domain.ErrSessionNotFound {
		return err
	}

	// Delete the session key
	delCmd := s.client.B().Del().Key(key).Build()
	if err := s.client.Do(ctx, delCmd).Error(); err != nil {
		return domain.NewInternalError(err)
	}

	// Remove from user index if session was found
	if session != nil {
		userIndexKey := s.buildUserIndexKey(session.UserID, domain.ScopeAuthentication)
		sremCmd := s.client.B().Srem().Key(userIndexKey).Member(token).Build()
		s.client.Do(ctx, sremCmd) // Ignore error for index cleanup
	}

	return nil
}

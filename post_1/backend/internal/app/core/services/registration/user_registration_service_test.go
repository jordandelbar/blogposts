package registration

import (
	"context"
	"database/sql"
	"errors"
	"personal_website/internal/app/core/domain"
	"personal_website/internal/app/core/ports"
	"strings"
	"testing"
)

type mockEmailService struct {
	shouldFailActivation   bool
	shouldFailNotification bool
	activationError        error
	notificationError      error
	sentEmails             []sentEmail
}

type sentEmail struct {
	emailType string
	token     string
	email     string
	baseURL   string
	user      *domain.User
}

func (m *mockEmailService) SendActivationEmail(ctx context.Context, activationToken, recipientEmail, baseURL string) error {
	if m.shouldFailActivation {
		return m.activationError
	}
	m.sentEmails = append(m.sentEmails, sentEmail{
		emailType: "activation",
		token:     activationToken,
		email:     recipientEmail,
		baseURL:   baseURL,
	})
	return nil
}

func (m *mockEmailService) SendNewUserNotification(ctx context.Context, user *domain.User) error {
	if m.shouldFailNotification {
		return m.notificationError
	}
	m.sentEmails = append(m.sentEmails, sentEmail{
		emailType: "notification",
		user:      user,
	})
	return nil
}

func (m *mockEmailService) SendContactEmail(ctx context.Context, form domain.ContactMessage) error {
	return nil
}

type mockUserRepo struct {
	users              map[string]domain.User
	shouldFailCreate   bool
	shouldFailActivate bool
	createError        error
	activateError      error
	nextUserID         int
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user domain.User) (int, error) {
	if m.shouldFailCreate {
		return 0, m.createError
	}
	m.nextUserID++
	user.ID = m.nextUserID
	m.users[user.Email] = user
	return m.nextUserID, nil
}

func (m *mockUserRepo) ActivateUser(ctx context.Context, user *domain.User) error {
	if m.shouldFailActivate {
		return m.activateError
	}
	if existingUser, ok := m.users[user.Email]; ok {
		existingUser.Activated = true
		m.users[user.Email] = existingUser
		user.Activated = true
	}
	return nil
}

func (m *mockUserRepo) CheckUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, exists := m.users[email]
	return exists, nil
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	user, exists := m.users[email]
	if !exists {
		return domain.User{}, sql.ErrNoRows
	}
	return user, nil
}

func (m *mockUserRepo) DeactivateUser(ctx context.Context, id int) error {
	if m.shouldFailCreate {
		return m.createError
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

func (m *mockUserRepo) DeleteUser(ctx context.Context, id int) error {
	if m.shouldFailCreate {
		return m.createError
	}
	for email, user := range m.users {
		if user.ID == id {
			delete(m.users, email)
			return nil
		}
	}
	return sql.ErrNoRows
}

type mockSessionRepo struct {
	sessions         map[string]*domain.Session
	shouldFailStore  bool
	shouldFailGet    bool
	shouldFailDelete bool
	storeError       error
	getError         error
	deleteError      error
}

func (m *mockSessionRepo) StoreSession(ctx context.Context, token string, scope domain.TokenScope, session *domain.Session) error {
	if m.shouldFailStore {
		return m.storeError
	}
	scopeStr, _ := scope.String()
	key := scopeStr + ":" + token
	m.sessions[key] = session
	return nil
}

func (m *mockSessionRepo) GetSession(ctx context.Context, token string, scope domain.TokenScope) (*domain.Session, error) {
	if m.shouldFailGet {
		return nil, m.getError
	}
	scopeStr, _ := scope.String()
	key := scopeStr + ":" + token
	session, exists := m.sessions[key]
	if !exists {
		return nil, domain.ErrSessionNotFound
	}
	return session, nil
}

func (m *mockSessionRepo) DeleteSession(ctx context.Context, token string) error {
	if m.shouldFailDelete {
		return m.deleteError
	}
	// Try to find and delete the session with any scope
	for key := range m.sessions {
		if strings.HasSuffix(key, ":"+token) {
			delete(m.sessions, key)
			return nil
		}
	}
	return domain.ErrSessionNotFound
}

func (m *mockSessionRepo) DeleteAllSessionsForUser(ctx context.Context, userID int, scope domain.TokenScope) error {
	if m.shouldFailDelete {
		return m.deleteError
	}
	scopeStr, _ := scope.String()
	// Remove all sessions for this user and scope
	for key, session := range m.sessions {
		if session.UserID == userID && strings.HasPrefix(key, scopeStr+":") {
			delete(m.sessions, key)
		}
	}
	return nil
}

type mockTransaction struct {
	userRepo         *mockUserRepo
	shouldFailCommit bool
	commitError      error
	committed        bool
	rolledBack       bool
}

func (m *mockTransaction) UserRepo() ports.UserRepository {
	return m.userRepo
}

func (m *mockTransaction) Commit() error {
	if m.shouldFailCommit {
		return m.commitError
	}
	m.committed = true
	return nil
}

func (m *mockTransaction) Rollback() error {
	m.rolledBack = true
	return nil
}

type mockDatabase struct {
	transaction     *mockTransaction
	shouldFailBegin bool
	beginError      error
}

func (m *mockDatabase) UserRepo() ports.UserRepository {
	return nil
}

func (m *mockDatabase) PermissionRepo() ports.PermissionRepository {
	return nil
}

func (m *mockDatabase) ArticleRepo() ports.ArticleRepository {
	return nil
}

func (m *mockDatabase) Begin(ctx context.Context) (ports.Transaction, error) {
	if m.shouldFailBegin {
		return nil, m.beginError
	}
	return m.transaction, nil
}

type mockDatastore struct {
	database    *mockDatabase
	sessionRepo *mockSessionRepo
}

func (m *mockDatastore) UserRepo() ports.UserRepository {
	return m.database.UserRepo()
}

func (m *mockDatastore) PermissionRepo() ports.PermissionRepository {
	return m.database.PermissionRepo()
}

func (m *mockDatastore) ArticleRepo() ports.ArticleRepository {
	return m.database.ArticleRepo()
}

func (m *mockDatastore) SessionRepo() ports.SessionRepository {
	return m.sessionRepo
}

func (m *mockDatastore) Begin(ctx context.Context) (ports.Transaction, error) {
	return m.database.Begin(ctx)
}

func TestNewUserService(t *testing.T) {
	emailService := &mockEmailService{}
	sessionRepo := &mockSessionRepo{sessions: make(map[string]*domain.Session)}
	database := &mockDatabase{}
	datastore := &mockDatastore{
		database:    database,
		sessionRepo: sessionRepo,
	}

	service := NewUserService(emailService, datastore)

	if service.emailService != emailService {
		t.Error("NewUserService() did not set emailService correctly")
	}

	if service.datastore != datastore {
		t.Error("NewUserService() did not set datastore correctly")
	}
}

func TestUserService_RegisterUser_Success(t *testing.T) {
	emailService := &mockEmailService{}
	userRepo := &mockUserRepo{
		users:      make(map[string]domain.User),
		nextUserID: 0,
	}
	sessionRepo := &mockSessionRepo{sessions: make(map[string]*domain.Session)}
	transaction := &mockTransaction{
		userRepo: userRepo,
	}
	database := &mockDatabase{
		transaction: transaction,
	}
	datastore := &mockDatastore{
		database:    database,
		sessionRepo: sessionRepo,
	}

	service := NewUserService(emailService, datastore)

	user := domain.User{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	activationURL := "http://example.com/activate"

	err := service.RegisterUser(context.Background(), user, activationURL)

	if err != nil {
		t.Errorf("RegisterUser() should succeed, got error: %v", err)
	}

	if !transaction.committed {
		t.Error("RegisterUser() should commit transaction")
	}

	if len(userRepo.users) != 1 {
		t.Errorf("RegisterUser() should create 1 user, got %d", len(userRepo.users))
	}

	if len(sessionRepo.sessions) != 1 {
		t.Errorf("RegisterUser() should create 1 session, got %d", len(sessionRepo.sessions))
	}

	if len(emailService.sentEmails) != 1 {
		t.Errorf("RegisterUser() should send 1 email, got %d", len(emailService.sentEmails))
	}

	sentEmail := emailService.sentEmails[0]
	if sentEmail.emailType != "activation" {
		t.Errorf("RegisterUser() should send activation email, got %s", sentEmail.emailType)
	}
	if sentEmail.email != user.Email {
		t.Errorf("RegisterUser() should send email to %s, got %s", user.Email, sentEmail.email)
	}
	if sentEmail.baseURL != activationURL {
		t.Errorf("RegisterUser() should use baseURL %s, got %s", activationURL, sentEmail.baseURL)
	}

	// Verify session contains correct data
	var storedSession *domain.Session
	for _, session := range sessionRepo.sessions {
		storedSession = session
		break
	}
	if storedSession == nil {
		t.Error("RegisterUser() should store a session")
	} else {
		if storedSession.UserID != 1 {
			t.Errorf("RegisterUser() should store session with UserID 1, got %d", storedSession.UserID)
		}
		if storedSession.Email != user.Email {
			t.Errorf("RegisterUser() should store session with email %s, got %s", user.Email, storedSession.Email)
		}
		if storedSession.Activated {
			t.Error("RegisterUser() should store session with Activated=false")
		}
	}
}

func TestUserService_RegisterUser_CreateSessionError(t *testing.T) {
	emailService := &mockEmailService{}
	userRepo := &mockUserRepo{
		users:      make(map[string]domain.User),
		nextUserID: 0,
	}
	sessionRepo := &mockSessionRepo{
		sessions:        make(map[string]*domain.Session),
		shouldFailStore: true,
		storeError:      errors.New("store session failed"),
	}
	transaction := &mockTransaction{
		userRepo: userRepo,
	}
	database := &mockDatabase{
		transaction: transaction,
	}
	datastore := &mockDatastore{
		database:    database,
		sessionRepo: sessionRepo,
	}

	service := NewUserService(emailService, datastore)

	user := domain.User{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	err := service.RegisterUser(context.Background(), user, "http://example.com")

	if err == nil {
		t.Error("RegisterUser() should fail when StoreSession() fails")
	}

	// Should be wrapped in domain.DomainError with internal type
	var domainErr domain.DomainError
	if !errors.As(err, &domainErr) {
		t.Error("RegisterUser() should return domain.DomainError when StoreSession() fails")
	} else if domainErr.Type != domain.ErrorTypeInternal {
		t.Error("RegisterUser() should return internal error type when StoreSession() fails")
	}

	if transaction.committed {
		t.Error("RegisterUser() should not commit when StoreSession() fails")
	}

	if len(emailService.sentEmails) != 0 {
		t.Error("RegisterUser() should not send email when StoreSession() fails")
	}
}

func TestUserService_ActivateUser_Success(t *testing.T) {
	emailService := &mockEmailService{}
	userRepo := &mockUserRepo{
		users: map[string]domain.User{
			"test@example.com": {
				ID:        1,
				Email:     "test@example.com",
				Name:      "Test User",
				Activated: false,
			},
		},
	}
	sessionRepo := &mockSessionRepo{
		sessions: map[string]*domain.Session{
			"activation:test-activation-token": {
				UserID:      1,
				Email:       "test@example.com",
				Permissions: domain.Permissions{},
				Activated:   false,
			},
		},
	}
	transaction := &mockTransaction{
		userRepo: userRepo,
	}
	database := &mockDatabase{
		transaction: transaction,
	}
	datastore := &mockDatastore{
		database:    database,
		sessionRepo: sessionRepo,
	}

	service := NewUserService(emailService, datastore)

	tokenPlaintext := "test-activation-token"

	user, err := service.ActivateUser(context.Background(), tokenPlaintext)

	if err != nil {
		t.Errorf("ActivateUser() should succeed, got error: %v", err)
	}

	if user == nil {
		t.Fatal("ActivateUser() should return user")
	}

	if user.ID != 1 {
		t.Errorf("ActivateUser() should return user with ID 1, got %d", user.ID)
	}

	if !transaction.committed {
		t.Error("ActivateUser() should commit transaction")
	}

	if len(emailService.sentEmails) != 1 {
		t.Errorf("ActivateUser() should send notification email, got %d emails", len(emailService.sentEmails))
	}

	sentEmail := emailService.sentEmails[0]
	if sentEmail.emailType != "notification" {
		t.Errorf("ActivateUser() should send notification email, got %s", sentEmail.emailType)
	}

	// Verify session was deleted
	if len(sessionRepo.sessions) != 0 {
		t.Error("ActivateUser() should delete activation session")
	}
}

func TestUserService_ActivateUser_GetSessionError(t *testing.T) {
	emailService := &mockEmailService{}
	userRepo := &mockUserRepo{
		users: make(map[string]domain.User),
	}
	sessionRepo := &mockSessionRepo{
		sessions:      make(map[string]*domain.Session),
		shouldFailGet: true,
		getError:      domain.ErrSessionNotFound,
	}
	transaction := &mockTransaction{
		userRepo: userRepo,
	}
	database := &mockDatabase{
		transaction: transaction,
	}
	datastore := &mockDatastore{
		database:    database,
		sessionRepo: sessionRepo,
	}

	service := NewUserService(emailService, datastore)

	user, err := service.ActivateUser(context.Background(), "invalid-token")

	if err == nil {
		t.Error("ActivateUser() should fail when GetSession() fails")
	}

	if user != nil {
		t.Error("ActivateUser() should not return user when GetSession() fails")
	}

	if transaction.committed {
		t.Error("ActivateUser() should not commit when GetSession() fails")
	}
}

package tests

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"personal_website/cmd/app"
	"personal_website/config"
	"personal_website/internal/app/core/ports"
	datastore_adapter "personal_website/internal/infrastructure/adapters/repository/datastore"
	postgres_adapter "personal_website/internal/infrastructure/adapters/repository/postgres"
	"personal_website/internal/infrastructure/adapters/repository/postgres/sqlc"
	valkey_adapter "personal_website/internal/infrastructure/adapters/repository/valkey"
	"personal_website/pkg/telemetry"
	"testing"
	"time"

	"github.com/awnumar/memguard"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	valkeycontainer "github.com/testcontainers/testcontainers-go/modules/valkey"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	pgContainer     *pgcontainer.PostgresContainer
	valkeyContainer *valkeycontainer.ValkeyContainer
	db              *sql.DB
	pgDatabase      ports.PostgresDatabase
	vkDatabase      ports.ValkeyDatabase
	datastore       ports.Datastore
	queries         *sqlc.Queries
	testCfg         *config.Config
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start PostgreSQL container
	pgc, err := pgcontainer.Run(
		ctx,
		"postgres:17-alpine",
		pgcontainer.WithDatabase("testdb"),
		pgcontainer.WithUsername("testuser"),
		pgcontainer.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		panic(err)
	}
	pgContainer = pgc

	// Start Valkey container without auth for tests
	vc, err := valkeycontainer.Run(
		ctx,
		"valkey/valkey:8.1.3-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections"),
		),
	)
	if err != nil {
		panic(err)
	}
	valkeyContainer = vc

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	pgHost, _ := pgContainer.Host(ctx)
	pgPort, _ := pgContainer.MappedPort(ctx, "5432")

	valkeyHost, _ := valkeyContainer.Host(ctx)
	valkeyPort, _ := valkeyContainer.MappedPort(ctx, "6379")

	testCfg = NewTestConfig(map[string]string{
		"pg_user":         "testuser",
		"pg_password":     "testpass",
		"pg_dbname":       "testdb",
		"pg_host":         pgHost,
		"pg_port":         pgPort.Port(),
		"valkey_host":     valkeyHost,
		"valkey_port":     valkeyPort.Port(),
		"valkey_password": "",
	}, randomPort())

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}

	pgDatabase, err = postgres_adapter.NewDatabase(&testCfg.Postgres)
	if err != nil {
		panic(err)
	}

	vkDatabase, err = valkey_adapter.NewDatabase(&testCfg.Valkey)
	if err != nil {
		panic(err)
	}

	datastore = datastore_adapter.NewDatastore(pgDatabase, vkDatabase)

	queries = sqlc.New(db)

	code := m.Run()

	// Teardown
	db.Close()
	pgDatabase.Close()
	vkDatabase.Close()
	_ = pgContainer.Terminate(ctx)
	_ = valkeyContainer.Terminate(ctx)
	os.Exit(code)
}

func setupTestDB(t *testing.T) {
	err := runMigrationsUp(db)
	require.NoError(t, err)
}

func cleanupDB(t *testing.T) {
	err := runMigrationsDown(db)
	require.NoError(t, err)
}

func startTestServer(t *testing.T) string {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	testMockEmailSender = &MockEmailSender{}
	testMockResumeService = NewMockResumeService()

	// Create a mock telemetry for tests
	testTelemetry, err := telemetry.NewTelemetry(logger)
	if err != nil {
		// If telemetry setup fails, create a nil telemetry for tests
		testTelemetry = nil
	}

	srv, err := app.NewServer(app.ServerDeps{
		Logger:        logger,
		Config:        testCfg,
		Datastore:     datastore,
		EmailSender:   testMockEmailSender,
		ResumeService: testMockResumeService,
		Telemetry:     testTelemetry,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		_ = srv.Serve(ctx)
	}()

	t.Cleanup(func() {
		cancel()
		_ = srv.Shutdown(context.Background())
	})

	baseURL := fmt.Sprintf("http://localhost:%d", testCfg.App.Port)

	// Wait for server to be ready
	waitForServerReady(t, baseURL)

	return baseURL
}

func waitForServerReady(t *testing.T, baseURL string) {
	client := &http.Client{Timeout: 100 * time.Millisecond}

	for i := 0; i < 50; i++ { // Try for up to 5 seconds
		resp, err := client.Get(baseURL + "/health")
		if err == nil {
			resp.Body.Close()
			return // Server is ready
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("Server failed to start within timeout period")
}

func runMigrationsUp(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../sql/schemas",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func runMigrationsDown(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../sql/schemas",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	err = m.Down()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run down migrations: %w", err)
	}

	return nil
}

func randomPort() int {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func newLocked(s string) *memguard.LockedBuffer {
	return memguard.NewBufferFromBytes([]byte(s))
}

// Helper functions to create TestSuites with proper dependencies
func NewTestSuite(t *testing.T) *TestSuite {
	setupTestDB(t)
	t.Cleanup(func() { cleanupDB(t) })

	serverAddr := startTestServer(t)
	authToken := createTestUser(t, queries, datastore)

	return &TestSuite{
		ServerAddr: serverAddr,
		AuthToken:  authToken,
	}
}

func NewUnauthenticatedTestSuite(t *testing.T) *TestSuite {
	setupTestDB(t)
	t.Cleanup(func() { cleanupDB(t) })

	serverAddr := startTestServer(t)

	return &TestSuite{
		ServerAddr: serverAddr,
	}
}

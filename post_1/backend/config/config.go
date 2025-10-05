package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/awnumar/memguard"
)

type PostgresConfig struct {
	User            *memguard.LockedBuffer
	Password        *memguard.LockedBuffer
	Database        *memguard.LockedBuffer
	Host            *memguard.LockedBuffer
	Port            *memguard.LockedBuffer
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	// Retry configuration
	MaxRetries     int
	RetryDelay     time.Duration
	MaxRetryDelay  time.Duration
	ConnectTimeout time.Duration
}

func (p *PostgresConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", p.User.String(), p.Password.String(), p.Host.String(), p.Port.String(), p.Database.String())
}

type ValkeyConfig struct {
	Host     *memguard.LockedBuffer
	Port     *memguard.LockedBuffer
	Password *memguard.LockedBuffer
}

func (v *ValkeyConfig) DSN() string {
	if v.Password != nil && v.Password.String() != "" {
		return fmt.Sprintf(":%s@%s:%s", v.Password.String(), v.Host.String(), v.Port.String())
	}
	return fmt.Sprintf("%s:%s", v.Host.String(), v.Port.String())
}

type LimiterConfig struct {
	Rps     int
	Burst   int
	Enabled bool
}

type SMTPConfig struct {
	Username  *memguard.LockedBuffer
	Password  *memguard.LockedBuffer
	Host      *memguard.LockedBuffer
	Port      *memguard.LockedBuffer
	Recipient *memguard.LockedBuffer
}

type CORSConfig struct {
	TrustedOrigins []string
}

type MinioConfig struct {
	Endpoint  *memguard.LockedBuffer
	AccessKey *memguard.LockedBuffer
	SecretKey *memguard.LockedBuffer
	Bucket    *memguard.LockedBuffer
	UseSSL    bool
}

type AppConfig struct {
	Environment     string
	Version         string
	Port            int
	MetricsPort     int
	Limiter         LimiterConfig
	Cors            CORSConfig
	ShutdownTimeout time.Duration
	ActivationUrl   string
}

type Config struct {
	Postgres PostgresConfig
	Valkey   ValkeyConfig
	SMTP     SMTPConfig
	Minio    MinioConfig
	App      AppConfig
}

func readSecret(filename string) *memguard.LockedBuffer {
	filepath := "/run/secrets/" + filename
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// Try case insensitive environment variable lookup
		envVar := getEnvCaseInsensitive(filename)
		if envVar != "" {
			return memguard.NewBufferFromBytes([]byte(envVar))
		}
		log.Fatalf("Secret %s not found and no %s environment variable set", filepath, filename)
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf("Failed to read secret %s: %v", filepath, err)
	}
	trimmed := strings.TrimSpace(string(data))

	return memguard.NewBufferFromBytes([]byte(trimmed))
}

func getEnvCaseInsensitive(key string) string {
	// Try original key first
	if val := os.Getenv(key); val != "" {
		return val
	}

	// Try uppercase
	if val := os.Getenv(strings.ToUpper(key)); val != "" {
		return val
	}

	// Try lowercase
	if val := os.Getenv(strings.ToLower(key)); val != "" {
		return val
	}

	return ""
}

func InitConfig() Config {
	var config Config

	memguard.CatchInterrupt()

	flag.StringVar(&config.App.Environment, "env", os.Getenv("ENVIRONMENT"), "Environment (development|staging|production)")
	flag.IntVar(&config.App.Port, "port", 5000, "API server port")
	flag.IntVar(&config.App.MetricsPort, "metrics-port", 2112, "Metrics server port")
	flag.IntVar(&config.App.Limiter.Rps, "rate-limiter", 500, "Rate limiter")
	flag.IntVar(&config.App.Limiter.Burst, "rate-limiter-burst", 20, "Rate limiter burst")
	flag.BoolVar(&config.App.Limiter.Enabled, "rate-limiter-enabled", true, "Enable rate limiter")
	flag.StringVar(&config.App.ActivationUrl, "activation-url", "", "User activation base url")
	flag.Parse()

	if corsOrigins := getEnvCaseInsensitive("CORS_TRUSTED_ORIGINS"); corsOrigins != "" {
		config.App.Cors.TrustedOrigins = strings.Fields(corsOrigins)
	}

	config.Postgres.User = readSecret("db_user")
	config.Postgres.Password = readSecret("db_password")
	config.Postgres.Database = readSecret("db_database_name")
	config.Postgres.Host = readSecret("db_host")
	config.Postgres.Port = readSecret("db_port")
	config.Postgres.MaxOpenConns = 25
	config.Postgres.MaxIdleConns = 25
	config.Postgres.ConnMaxLifetime = 5 * time.Minute
	config.Postgres.ConnMaxIdleTime = 5 * time.Minute
	config.Postgres.MaxRetries = 5
	config.Postgres.RetryDelay = 1 * time.Second
	config.Postgres.MaxRetryDelay = 30 * time.Second
	config.Postgres.ConnectTimeout = 10 * time.Second

	config.Valkey.Host = readSecret("valkey_host")
	config.Valkey.Port = readSecret("valkey_port")
	config.Valkey.Password = readSecret("valkey_password")

	config.SMTP.Username = readSecret("smtp_username")
	config.SMTP.Password = readSecret("smtp_password")
	config.SMTP.Host = readSecret("smtp_host")
	config.SMTP.Port = readSecret("smtp_port")
	config.SMTP.Recipient = readSecret("smtp_recipient")

	config.Minio.Endpoint = readSecret("minio_endpoint")
	config.Minio.AccessKey = readSecret("minio_access_key")
	config.Minio.SecretKey = readSecret("minio_secret_key")
	config.Minio.Bucket = readSecret("minio_bucket")
	config.Minio.UseSSL = getEnvCaseInsensitive("MINIO_USE_SSL") == "true"

	return config
}

# Personal Website Backend

Go backend for personal website with articles, user management, and file uploads.

## Requirements

- Go 1.24+
- Docker & Docker Compose
- golang-migrate CLI

## Quick Start

```bash
# Start services (from project root)
make up

# Setup database (from backend/)
cd backend
make migrate

# Run backend locally
make run
```

## Development Setup

### Option 1: Full Docker Stack
```bash
# From project root
make up        # Start all services
make build     # Rebuild images
make down      # Stop all services
```

### Option 2: Backend Only (Local Development)
```bash
# Start dependencies only
docker compose up postgres valkey minio -d

# From backend/
make migrate   # Run database migrations
make run       # Start backend locally
```

## Configuration

Backend uses `.env` file in `backend/` directory:

```bash
# Database
db_user=myuser
db_password=mypassword
db_database_name=postgres
db_host=localhost
db_port=5432

# Cache
valkey_host=localhost
valkey_port=6379
valkey_password=valkeypassword123

# Server
PORT=5000
ENVIRONMENT=development
CORS_TRUSTED_ORIGINS="http://localhost:3000 http://localhost:3001"

# Email
smtp_username=your_email
smtp_password=your_password
smtp_host=smtp.gmail.com
smtp_port=587
smtp_recipient=your_recipient

# S3/MinIO
minio_endpoint=localhost:9000
minio_access_key=testuser
minio_secret_key=testpassword123
minio_bucket=documents
MINIO_USE_SSL=false
```

## Available Commands

```bash
# Development
make run              # Start backend locally
make migrate          # Run database migrations
make test             # Run all tests
make up               # Start docker services
make down             # Stop docker services

# Code generation
make swagger-generate # Generate API docs
make sqlc-generate    # Generate database code

# Testing
make unit-test        # Unit tests only
make integration-test # Integration tests only
make coverage         # Test coverage report
```

## API Documentation

Swagger UI available at `/swagger/index.html` when running.

## Project Structure

```
cmd/                 # Application entry points
config/              # Configuration management
internal/
├── app/
│   ├── core/        # Domain logic and entities
│   ├── ports/       # Interface definitions
│   └── services/    # Business logic
└── infrastructure/  # External concerns (HTTP, DB, etc.)
pkg/                 # Reusable utilities
sql/                 # Database schemas and migrations
tests/               # Integration tests
```

## Key Endpoints

```
POST   /v1/users                    # Register user
POST   /v1/auth/authenticate        # Login
GET    /v1/articles                 # List published articles
POST   /v1/articles                 # Create article (auth required)
GET    /health                      # Health check
GET    /metrics                     # Prometheus metrics
```

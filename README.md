# Portfolios

A portfolio management backend application built with Go and PostgreSQL.

## Tech Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: PostgreSQL 15+
- **ORM**: GORM
- **Authentication**: JWT (golang-jwt/jwt)
- **Migrations**: golang-migrate

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 15 or higher
- Docker and Docker Compose (optional)

## Getting Started

### Local Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/alenon/portfolios.git
   cd portfolios
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Install dependencies**
   ```bash
   make install
   # Or manually:
   # go mod download
   ```

4. **Set up PostgreSQL database**
   ```bash
   createdb portfolios
   ```

5. **Run database migrations**
   ```bash
   make migrate-up
   # Or manually:
   # migrate -path migrations -database "postgresql://username:password@localhost:5432/portfolios?sslmode=disable" up
   ```

6. **Run the backend server**
   ```bash
   make run
   # Or manually:
   # go run cmd/api/main.go
   ```

7. **Access the API**
   - Backend API: http://localhost:8080

### Docker Development Setup

1. **Start all services**
   ```bash
   make docker-dev
   # Or manually:
   # docker-compose -f docker-compose.dev.yml up
   ```

2. **Access the application**
   - Backend API: http://localhost:8080
   - PostgreSQL: localhost:5432

3. **Stop services**
   ```bash
   docker-compose -f docker-compose.dev.yml down
   ```

## Available Commands

Run `make help` to see all available commands:

```bash
make help              # Show available commands
make install           # Install all dependencies
make migrate-up        # Run database migrations
make migrate-down      # Rollback database migrations
make migrate-create    # Create a new migration
make build             # Build backend
make run               # Run backend server
make test              # Run all tests
make test-coverage     # Run tests with coverage report
make docker-up         # Start production Docker containers
make docker-down       # Stop Docker containers
make docker-dev        # Start development containers
make docker-logs       # View container logs
make clean             # Clean build artifacts
make lint              # Run linters
make format            # Format code
```

## Project Structure

```
portfolios/
├── cmd/
│   └── api/              # Application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── database/         # Database connection
│   ├── dto/              # Data Transfer Objects
│   ├── handlers/         # HTTP handlers
│   ├── middleware/       # HTTP middleware
│   ├── models/           # Database models
│   ├── repository/       # Data access layer
│   ├── services/         # Business logic
│   └── utils/            # Utility functions
├── pkg/                  # Public packages
├── migrations/           # Database migrations
├── configs/              # Configuration files
├── .env.example          # Environment variables template
├── docker-compose.yml    # Production Docker setup
├── docker-compose.dev.yml # Development Docker setup
├── Dockerfile            # Backend Docker image
├── Makefile              # Build automation
└── README.md
```

## Environment Variables

See `.env.example` for all available configuration options. Key variables:

- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret key for JWT token signing (must be at least 32 characters)
- `SMTP_*`: Email service configuration for password reset
- `CORS_ALLOWED_ORIGINS`: Allowed origins for CORS
- `SERVER_PORT`: Backend server port (default: 8080)

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run backend tests only
go test -v ./...
```

## Database Migrations

```bash
# Create a new migration
make migrate-create NAME=create_users_table

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down
```

## API Documentation

Once the backend is running, API documentation is available at:
- Swagger UI: http://localhost:8080/api/docs (coming soon)

## Security

- All passwords are hashed using bcrypt
- JWT tokens for authentication and authorization
- CORS configured for allowed origins
- Rate limiting on authentication endpoints
- Input validation on all endpoints
- HTTPS enforced in production

## License

MIT

## Contributing

Contributions are welcome! Please read the contributing guidelines before submitting PRs.

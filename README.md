# Portfolios

A comprehensive portfolio management application built with Go, PostgreSQL, React, and TypeScript.

## Tech Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: PostgreSQL 15+
- **ORM**: GORM
- **Authentication**: JWT (golang-jwt/jwt)
- **Migrations**: golang-migrate

### Frontend
- **Framework**: React 18+
- **Language**: TypeScript
- **Build Tool**: Vite
- **UI Library**: Material-UI (MUI)
- **State Management**: React Query + Zustand
- **Form Handling**: React Hook Form
- **HTTP Client**: Axios
- **Routing**: React Router v6

## Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- PostgreSQL 15 or higher
- Docker and Docker Compose (optional)

## Getting Started

### Local Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/lenon/portfolios.git
   cd portfolios
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Install dependencies**
   ```bash
   # Backend dependencies
   go mod download

   # Frontend dependencies
   cd frontend
   npm install
   cd ..
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

7. **Run the frontend dev server** (in a new terminal)
   ```bash
   make run-frontend
   # Or manually:
   # cd frontend && npm run dev
   ```

8. **Access the application**
   - Frontend: http://localhost:5173
   - Backend API: http://localhost:8080

### Docker Development Setup

1. **Start all services with hot-reload**
   ```bash
   make docker-dev
   # Or manually:
   # docker-compose -f docker-compose.dev.yml up
   ```

2. **Access the application**
   - Frontend: http://localhost:5173
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
make build             # Build backend and frontend
make run               # Run backend server
make run-frontend      # Run frontend dev server
make test              # Run all tests
make test-coverage     # Run tests with coverage report
make docker-up         # Start production Docker containers
make docker-down       # Stop Docker containers
make docker-dev        # Start development containers with hot-reload
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
├── frontend/
│   └── src/
│       ├── components/   # React components
│       ├── pages/        # Page components
│       ├── services/     # API services
│       ├── contexts/     # React contexts
│       ├── utils/        # Utility functions
│       └── types/        # TypeScript types
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
- `CORS_ALLOWED_ORIGINS`: Allowed frontend origins
- `SERVER_PORT`: Backend server port (default: 8080)

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run backend tests only
go test -v ./...

# Run frontend tests only
cd frontend && npm test
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

# Technical Stack

## Backend

### Language & Runtime
- **Go 1.21+**: Primary backend language for performance, concurrency support, and strong typing

### Web Framework
- **Gin**: Lightweight, fast HTTP web framework for building REST APIs
  - Alternative: Echo or Chi if you prefer different routing patterns
  - Justification: Gin provides excellent performance, middleware support, and is widely adopted in the Go community

### Database
- **PostgreSQL 15+**: Primary relational database for transactional data
  - Robust support for complex queries needed for portfolio calculations
  - JSONB support for flexible corporate action metadata
  - Excellent data integrity and ACID compliance
  - Strong Go driver ecosystem

### Database Tooling
- **golang-migrate/migrate**: Database migration management
  - Version-controlled schema changes
  - CLI and library support for migrations
- **sqlc**: Generate type-safe Go code from SQL queries
  - Compile-time query validation
  - Eliminates runtime query errors
  - Alternative: Use GORM ORM if you prefer traditional ORM patterns

### API Design
- **REST API**: Primary API architecture
  - Well-understood patterns for CRUD operations
  - Easy to document with OpenAPI/Swagger
  - Straightforward to consume from frontend
- **JSON**: Data serialization format
- **go-playground/validator**: Request validation
  - Declarative validation rules
  - Comprehensive error messages

### Authentication & Authorization
- **JWT (JSON Web Tokens)**: Stateless authentication
  - golang-jwt/jwt library for token generation and validation
- **bcrypt**: Password hashing (golang.org/x/crypto/bcrypt)
- **Middleware-based authorization**: Role-based access control
  - Resource-level authorization (users can only access their own portfolios)

### Real-Time Data
- **WebSockets**: Live portfolio value updates during market hours
  - gorilla/websocket for WebSocket connections
  - Efficient push of price updates to connected clients
- **Market Data API Integration**: Third-party financial data provider
  - **Alpha Vantage**: Free tier available, good for MVP
  - **Twelve Data**: Better rate limits and coverage for production
  - **Polygon.io**: Professional option with real-time data
  - Strategy: Abstract market data behind interface for easy provider switching

### Background Jobs
- **Go routines + time.Ticker**: Scheduled tasks for price updates and corporate action detection
  - Alternative: Use asynq (Redis-backed job queue) if you need persistent job queues
- **Context-based cancellation**: Graceful shutdown of background workers

### CSV Processing
- **encoding/csv**: Standard library CSV parsing
- **Custom mapper**: Column header mapping logic for different broker formats
- **Validation pipeline**: Multi-stage validation before transaction import

### External APIs & HTTP Clients
- **net/http**: Standard library HTTP client for market data API calls
- **golang.org/x/time/rate**: Rate limiting for external API calls
- **Context with timeout**: Timeout handling for external requests

### Testing
- **testing**: Standard library testing framework
- **testify**: Assertion library for more readable tests
  - testify/assert for assertions
  - testify/mock for mocking
- **httptest**: HTTP handler testing
- **dockertest**: Integration testing with real PostgreSQL in Docker

### Logging & Monitoring
- **zap** or **zerolog**: Structured logging
  - High performance
  - JSON output for log aggregation
- **Prometheus metrics**: Application metrics (request counts, latencies, error rates)
  - Use promhttp for HTTP handler instrumentation
- **OpenTelemetry**: Distributed tracing (optional, for production observability)

### Configuration
- **viper**: Configuration management
  - Support for environment variables, config files, and defaults
  - 12-factor app compliance
- **Environment variables**: Production secrets and configuration

## Frontend

### Recommendation: React with TypeScript
- **Justification**:
  - Largest ecosystem and community support
  - Excellent TypeScript integration
  - Rich library ecosystem for charts and data visualization
  - Strong real-time data handling with hooks
  - Material-UI or Ant Design provide comprehensive component libraries

### Alternative Options:
- **Svelte/SvelteKit**: If you prefer simpler, more performant option with less boilerplate
- **Vue 3 + TypeScript**: Middle ground between React and Svelte

### Frontend Framework & Build Tools
- **React 18+**: UI library
- **TypeScript**: Type safety for large codebase
- **Vite**: Fast build tool and dev server
  - Faster than Create React App
  - Great HMR (Hot Module Replacement)

### UI Component Library
- **Material-UI (MUI)**: Comprehensive React component library
  - Alternative: Ant Design, Chakra UI, or shadcn/ui
  - Provides tables, forms, modals, date pickers out of the box

### State Management
- **React Query (TanStack Query)**: Server state management
  - Perfect for API data fetching, caching, and synchronization
  - Real-time refetching and optimistic updates
- **Zustand** or **Context API**: Client-only state management
  - Lightweight compared to Redux
  - Sufficient for UI state, theme, user preferences

### Data Visualization
- **Recharts** or **Victory**: React chart libraries
  - Declarative API matching React patterns
  - Good for line charts (portfolio performance over time)
- **Alternative: D3.js**: If you need highly customized visualizations

### Form Handling
- **React Hook Form**: Performant form library
  - Minimal re-renders
  - Built-in validation
  - Works well with TypeScript

### HTTP Client
- **Axios**: HTTP client for API calls
  - Interceptors for JWT token handling
  - Better error handling than fetch
  - Alternative: Use native fetch with React Query

### WebSocket Client
- **Native WebSocket API**: For real-time price updates
  - Wrapped in React hook for easy component integration
  - Automatic reconnection logic

### Routing
- **React Router v6**: Client-side routing
  - Declarative routing
  - Nested routes for complex layouts

### CSV Handling
- **PapaParse**: CSV parsing in browser
  - Parse uploaded files before sending to backend
  - Client-side validation and preview

### Date Handling
- **date-fns**: Modern date manipulation library
  - Smaller bundle size than Moment.js
  - Immutable and functional

### Testing
- **Vitest**: Unit and integration testing
  - Fast, Vite-native test runner
- **React Testing Library**: Component testing
  - Encourages testing user behavior, not implementation
- **Playwright** or **Cypress**: End-to-end testing

## DevOps & Infrastructure

### Containerization
- **Docker**: Container runtime
  - Multi-stage builds for optimized images
  - Separate images for backend and frontend
- **Docker Compose**: Local development environment
  - PostgreSQL, backend, frontend services
  - Seed data for development

### Orchestration
- **Kubernetes**: Container orchestration
  - Deployments for backend and frontend
  - StatefulSet for PostgreSQL (or use managed database)
  - Services for internal communication
  - Ingress for external traffic routing

### Package Management
- **Helm**: Kubernetes package manager
  - Helm charts for application deployment
  - Values files for environment-specific configuration (dev, staging, prod)
  - Chart versioning aligned with application versions

### CI/CD
- **GitHub Actions**: CI/CD pipelines
  - Automated testing on pull requests
  - Docker image building and pushing to registry
  - Helm chart deployment to Kubernetes
- **Docker Registry**: Container image storage
  - Docker Hub, GitHub Container Registry, or private registry

### Secrets Management
- **Kubernetes Secrets**: Runtime secrets
  - Database credentials, JWT signing keys, API keys
- **Sealed Secrets** or **External Secrets Operator**: GitOps-friendly secret management
  - Encrypt secrets in Git
  - Sync from external secret stores (AWS Secrets Manager, HashiCorp Vault)

### Database Management
- **Kubernetes CronJob**: Automated database backups
- **PostgreSQL Operator** (optional): Advanced database management in Kubernetes
- **Alternative: Managed PostgreSQL**: AWS RDS, Google Cloud SQL, or Azure Database for PostgreSQL for production

### Monitoring & Observability
- **Prometheus**: Metrics collection
  - Scrape application metrics
  - Alert on error rates, latency, resource usage
- **Grafana**: Metrics visualization
  - Dashboards for application and infrastructure metrics
- **Loki**: Log aggregation (optional)
  - Centralized logging from all pods
- **Kubernetes-native observability**: kubectl logs, describe, top commands

### Ingress & Load Balancing
- **NGINX Ingress Controller**: HTTP(S) routing to services
  - TLS termination
  - Rate limiting
  - WebSocket support for real-time features
- **cert-manager**: Automatic TLS certificate management
  - Let's Encrypt integration

## Development Tools & Practices

### Version Control
- **Git**: Source control
- **GitHub** or **GitLab**: Code hosting and collaboration
  - Pull request workflow
  - Code review process

### Code Quality
- **golangci-lint**: Go linting
  - Multiple linters in one tool
  - CI integration to fail builds on lint errors
- **ESLint + Prettier**: Frontend linting and formatting
  - TypeScript-aware linting rules
- **Pre-commit hooks**: Automated formatting and linting before commits
  - husky for Git hooks

### API Documentation
- **Swagger/OpenAPI**: REST API documentation
  - swaggo/swag for generating Swagger docs from Go comments
  - Swagger UI for interactive API exploration

### Database Tools
- **pgAdmin** or **DBeaver**: GUI database management
- **psql**: Command-line PostgreSQL client

### Local Development
- **Air**: Live reload for Go applications during development
  - Watches files and rebuilds on changes
- **Hot Module Replacement (Vite)**: Live frontend updates without full reload

### CLI Development
- **Cobra**: Go CLI framework for admin interface
  - Subcommand structure (e.g., `portfolios user create`, `portfolios backup`)
  - Flag parsing and help generation
  - Easy to test and extend

## Data & Integration

### Market Data Providers
- **MVP (Free Tier)**: Alpha Vantage
  - 5 requests per minute, 500 per day
  - Good for initial development
- **Production**: Twelve Data or Polygon.io
  - Higher rate limits
  - More reliable data
  - Real-time updates

### Corporate Actions Data
- **Market Data API**: Many providers include corporate action data
  - Stock splits, dividends in API responses
- **Alternative: Web Scraping**: Nasdaq or Yahoo Finance (use cautiously, respect ToS)
- **Manual Entry Fallback**: UI for users to manually record corporate actions if API doesn't catch them

## Deployment Model

### Web Application
- **SPA (Single Page Application)**: React frontend served statically
  - NGINX container for frontend static files
  - API calls to backend service

### CLI Application
- **Go Binary**: Compiled CLI tool
  - Distributed via releases (GitHub Releases)
  - Connects to same PostgreSQL database
  - Requires database credentials (secure configuration)

### Architecture Pattern
- **Microservices-lite**: Single backend service initially
  - Monolithic API with domain-separated packages
  - Can split into microservices later if needed (e.g., separate market data service)

## Summary

This stack provides:
- **Performance**: Go backend with PostgreSQL for fast data processing
- **Type Safety**: TypeScript frontend and Go backend minimize runtime errors
- **Scalability**: Kubernetes orchestration allows horizontal scaling
- **Developer Experience**: Modern tooling with hot reload, comprehensive testing
- **Production Ready**: Structured logging, metrics, CI/CD, and secret management
- **Flexibility**: Abstract market data provider for easy switching
- **User Control**: CSV import and manual entry (no broker API dependencies)

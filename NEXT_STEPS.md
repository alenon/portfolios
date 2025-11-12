# Next Steps - Portfolios Backend

This document tracks the implementation roadmap for the portfolios backend application.

## ✅ Completed

- [x] Authentication & Authorization (JWT, refresh tokens, password reset)
- [x] Portfolio Management (CRUD operations)
- [x] Transaction Management (buy/sell tracking)
- [x] Holdings Management (current positions)
- [x] Tax Lot Tracking (FIFO/LIFO, tax-loss harvesting)
- [x] Corporate Actions (splits, dividends, mergers, spinoffs)
- [x] Corporate Action Detection (background job)
- [x] Performance Analytics Service (TWR, MWR, annualized returns, benchmark comparison)
- [x] Market Data Service (Alpha Vantage integration with caching)
- [x] Performance Snapshot Service (daily performance tracking)
- [x] Runtime Home Directory (configuration and logging infrastructure)
- [x] Comprehensive Test Coverage
  - DTOs: 97.3% coverage (30+ tests for all conversion functions)
  - Middleware: 61.2% coverage (17 tests for auth, CORS, rate limiting)
  - Handlers: 68.4% coverage (23 tests for new handlers)
  - Services: 78.7% coverage (up from 57.2%)
    - Performance Analytics: ~90% coverage (new comprehensive tests)
    - Corporate Actions: CRUD methods at 100%, complex operations at 80%+
  - Database: 40.5% coverage (new basic tests)
  - Logger: 88.1% coverage (new comprehensive tests)
  - Overall project coverage: 55.3% (up from 49.5%)
  - All tests passing with proper mocking and edge case handling

## 🔥 High Priority

### 1. CSV Import Functionality
**Status:** ✅ COMPLETED
**Priority:** Critical - Core user-facing feature

Manual transaction import is a key differentiator per the product spec.

**Tasks:**
- [x] Create standard CSV format parser (generic imports)
- [x] Implement broker-specific parsers:
  - [x] Fidelity
  - [x] Schwab
  - [x] TD Ameritrade
  - [x] E*TRADE
  - [x] Interactive Brokers
  - [x] Robinhood
- [x] Add import validation and error handling
- [x] Implement bulk import endpoint: `POST /api/v1/portfolios/:id/transactions/bulk`
- [x] Add CSV import endpoint: `POST /api/v1/portfolios/:id/transactions/import/csv`
- [x] Add import batch tracking (uses existing `import_batch_id` field)
- [x] Add batch management endpoints (list, delete)
- [x] Create import service and handler
- [x] Wire up routes in main application
- [x] Fix import cycles (moved Quote and PerformanceMetrics types to dto package)
- [x] All tests passing

**Files created:**
- `internal/dto/import.go` - Import request/response DTOs
- `internal/dto/market_data_types.go` - Quote and HistoricalPrice types
- `internal/dto/performance_analytics_types.go` - Performance metric types
- `internal/services/csv_import_service.go` - Import service implementation
- `internal/services/csv_parsers/parser.go` - Base parser and utilities
- `internal/services/csv_parsers/generic_parser.go` - Standard CSV format
- `internal/services/csv_parsers/fidelity_parser.go` - Fidelity format
- `internal/services/csv_parsers/schwab_parser.go` - Schwab format
- `internal/services/csv_parsers/td_ameritrade_parser.go` - TD Ameritrade format
- `internal/services/csv_parsers/etrade_parser.go` - E*TRADE format
- `internal/services/csv_parsers/interactive_brokers_parser.go` - Interactive Brokers format
- `internal/services/csv_parsers/robinhood_parser.go` - Robinhood format
- `internal/handlers/import_handler.go` - Import HTTP handlers

**API Endpoints:**
- `POST /api/v1/portfolios/:id/transactions/import/csv` - Import from CSV file
- `POST /api/v1/portfolios/:id/transactions/import/bulk` - Bulk import transactions
- `GET /api/v1/portfolios/:id/imports/batches` - List import batches
- `DELETE /api/v1/portfolios/:id/imports/batches/:batch_id` - Delete import batch

### 2. Background Jobs for Market Data
**Status:** ✅ COMPLETED (Initial Implementation)
**Priority:** High - Enables automated updates

**Tasks:**
- [x] Create end-of-day price update job (cache refresh)
- [x] Create performance snapshot generation job (placeholder)
- [x] Create stale data cleanup job (cache cleanup)
- [x] Wire up jobs in scheduler
- [x] Add job monitoring/logging (built into scheduler)
- [x] Jobs run on @daily schedule

**Files created:**
- `internal/jobs/price_update_job.go` - Market data cache refresh
- `internal/jobs/snapshot_generation_job.go` - Snapshot generation (placeholder)
- `internal/jobs/cleanup_job.go` - Cache and data cleanup

**Notes:**
- Jobs are initialized automatically when market data service is available
- Current implementation focuses on cache management
- Full implementation requires repository enhancements:
  - Add `FindAll()` method to PortfolioRepository for batch operations
  - Add `FindExpired()` methods to token repositories for cleanup
  - Consider simplifying CreateSnapshot API or adding system user context
  - Optionally add current price field to Holding model for price updates

### 3. Export Functionality
**Status:** Not Started
**Priority:** High - User data portability

**Tasks:**
- [ ] Implement CSV export for portfolios
- [ ] Implement CSV export for transactions
- [ ] Implement CSV export for holdings
- [ ] Implement PDF report generation (performance reports)
- [ ] Create export service
- [ ] Add export API endpoints:
  - `GET /api/v1/portfolios/:id/export/csv`
  - `GET /api/v1/portfolios/:id/export/transactions`
  - `GET /api/v1/portfolios/:id/export/holdings`
  - `POST /api/v1/portfolios/:id/export/report`
- [ ] Add comprehensive tests

**Files to create:**
- `internal/services/export_service.go`
- `internal/handlers/export_handler.go`
- `internal/dto/export.go`

## 📊 Medium Priority

### 4. CLI Tool
**Status:** ✅ COMPLETED
**Priority:** Medium - Extensive CLI described in product spec

A comprehensive CLI has been implemented with modern TUI features using Cobra, Bubble Tea, and Lipgloss.

**Tasks:**
- [x] Set up Cobra CLI framework
- [x] Implement authentication commands:
  - `portfolios auth login`
  - `portfolios auth logout`
  - `portfolios auth register`
  - `portfolios auth whoami`
- [x] Implement portfolio commands:
  - `portfolios portfolio list`
  - `portfolios portfolio create`
  - `portfolios portfolio show <id>`
  - `portfolios portfolio delete <id>`
  - `portfolios portfolio holdings <id>`
  - `portfolios portfolio select` (interactive)
- [x] Implement transaction commands:
  - `portfolios transaction add <portfolio-id>`
  - `portfolios transaction import <portfolio-id> <file>` (supports 7 broker formats)
  - `portfolios transaction list <portfolio-id>`
  - `portfolios transaction delete <portfolio-id> <tx-id>`
  - `portfolios transaction batches <portfolio-id>`
  - `portfolios transaction delete-batch <portfolio-id> <batch-id>`
- [x] Implement performance commands:
  - `portfolios performance show <portfolio-id>`
  - `portfolios performance compare <id1> <id2>`
  - `portfolios performance benchmark <portfolio-id> <symbol>`
  - `portfolios performance snapshots <portfolio-id>`
- [x] Add configuration file support (`~/.portfolios/config.yaml`)
- [x] Add output formatting (table, JSON, CSV)
- [x] Add interactive TUI components (Bubble Tea portfolio selector)
- [x] Add comprehensive help text and documentation
- [x] Add version command with build info
- [x] Add config management commands
- [x] Add Makefile targets for building and installing

**Files created:**
- `cmd/portfolios/main.go` - CLI entry point
- `cmd/portfolios/cmd/root.go` - Root command and banner
- `cmd/portfolios/cmd/auth.go` - Authentication commands
- `cmd/portfolios/cmd/portfolio.go` - Portfolio management
- `cmd/portfolios/cmd/transaction.go` - Transaction management
- `cmd/portfolios/cmd/performance.go` - Performance analytics
- `cmd/portfolios/cmd/config.go` - Configuration management
- `cmd/portfolios/cmd/version.go` - Version command
- `cmd/portfolios/cmd/interactive.go` - Interactive selector
- `cmd/portfolios/README.md` - Comprehensive CLI documentation
- `internal/cli/config.go` - Configuration management
- `internal/cli/client.go` - API client
- `internal/cli/output.go` - Output formatting (table, JSON, CSV)
- `internal/cli/selector.go` - Bubble Tea interactive selector

**Technologies Used:**
- **Cobra v1.10.1** - CLI framework
- **Bubble Tea v1.3.10** - Terminal UI framework
- **Lipgloss v1.1.0** - Styling and table rendering
- **Bubbles v0.21.0** - TUI components
- **Viper v1.21.0** - Configuration management

**Features:**
- Beautiful styled output with colors and tables
- Interactive portfolio selector with keyboard navigation
- Support for 7 broker CSV formats (Fidelity, Schwab, TD Ameritrade, E*TRADE, Interactive Brokers, Robinhood, Generic)
- Multiple output formats (table, JSON, CSV)
- Persistent authentication with token storage
- Comprehensive error handling and validation
- Command aliases for faster workflows
- Date range filtering for performance analytics
- Dry-run mode for CSV imports
- Import batch management

### 5. Portfolio Comparison
**Status:** Not Started
**Priority:** Medium

**Tasks:**
- [ ] Implement portfolio comparison service
- [ ] Add side-by-side performance comparison
- [ ] Compare allocations, returns, risk metrics
- [ ] Add API endpoint: `GET /api/v1/portfolios/compare`
- [ ] Add tests

### 6. API Documentation
**Status:** Not Started
**Priority:** Medium

**Tasks:**
- [ ] Add Swagger/OpenAPI annotations to handlers
- [ ] Generate OpenAPI spec
- [ ] Add Swagger UI endpoint: `GET /api/docs`
- [ ] Document all request/response schemas
- [ ] Add example requests/responses
- [ ] Add authentication documentation

## 🚀 Lower Priority / Future Enhancements

### 7. Real-time Features
**Status:** Not Started
**Priority:** Low - Advanced feature

**Tasks:**
- [ ] Implement WebSocket support
- [ ] Add real-time portfolio value updates
- [ ] Add real-time price feeds
- [ ] Add push notifications for significant events

### 8. Advanced Analytics
**Status:** Not Started
**Priority:** Low - Advanced feature

**Tasks:**
- [ ] Implement risk analysis (Sharpe ratio, beta, volatility)
- [ ] Add portfolio optimization suggestions
- [ ] Implement drawdown analysis
- [ ] Add correlation analysis between holdings

### 9. Extended Asset Support
**Status:** Not Started
**Priority:** Low - Expansion feature

**Tasks:**
- [ ] Add options support
- [ ] Add futures support
- [ ] Add cryptocurrency tracking
- [ ] Add bond tracking
- [ ] Add mutual fund support

## 📝 Notes

### Current State
- **Commit:** c0ef85e - Configured deployment to trigger only on release tags
- **Branch:** `claude/deploy-script-release-only-011CV3tnMm3G3Z2oTmDNd9MW`
- **All tests passing:** ✅
- **Test Coverage:** ~56% overall

### Recent Changes (2025-11-12)
- **Deployment Configuration:** Modified GitHub Actions workflow to deploy only on release tags
  - Staging: Automatically deploys on pre-release tags (e.g., v1.0.0-rc.1, v1.0.0-beta.1)
  - Production: Automatically deploys on stable release tags (e.g., v1.0.0, v1.2.3)
  - Build job still runs on main/develop branches for CI validation
  - No automatic deployment on regular branch pushes
  - Updated documentation with clear deployment strategy and examples
- **Runtime Home Directory:** ✅ COMPLETED
  - Created runtime home directory structure (`~/.portfolios`)
  - YAML configuration file support (`~/.portfolios/config.yaml`)
  - Separate log files for server and request logs
  - Enhanced logger with multi-output support
  - Environment variable overrides for all configuration
  - Comprehensive documentation in `docs/RUNTIME_HOME_DIRECTORY.md`
  - Example configuration file (`config.example.yaml`)
  - Full test coverage for runtime package
  - Updated main.go with proper initialization and logging

**Previous Changes (2025-11-12):
- **CLI Tool:** Completed full implementation
  - Comprehensive CLI using Cobra v1.10.1 framework
  - Beautiful TUI with Bubble Tea v1.3.10 and Lipgloss v1.1.0
  - Interactive portfolio selector with keyboard navigation
  - Authentication commands (login, logout, register, whoami)
  - Portfolio management commands (list, create, show, delete, holdings, select)
  - Transaction commands (add, list, import, delete, batches)
  - Performance analytics commands (show, compare, benchmark, snapshots)
  - Configuration management with Viper v1.21.0
  - Multiple output formats (table, JSON, CSV)
  - Persistent token storage in ~/.portfolios/config.yaml
  - Makefile targets for building and installing
  - Comprehensive documentation in cmd/portfolios/README.md
  - 14 new files created for CLI implementation
- **Previous Changes (2025-11-11):**
  - CSV Import Feature with 7 broker-specific parsers
  - Background Jobs for market data updates and cleanup
  - Architecture improvements for import cycle resolution

### Environment Variables Needed
```bash
# Required
DATABASE_URL=postgresql://user:password@localhost:5432/portfolios
JWT_SECRET=your-secret-key

# Optional (for market data features)
MARKET_DATA_API_KEY=your-alpha-vantage-key
MARKET_DATA_PROVIDER=alphavantage

# Optional (for email features)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your-username
SMTP_PASSWORD=your-password
SMTP_FROM=noreply@example.com
```

### Architecture Patterns
- **Service Layer:** Business logic in `internal/services/`
- **Handler Layer:** HTTP handlers in `internal/handlers/`
- **Repository Layer:** Data access in `internal/repository/`
- **DTO Layer:** Request/response objects in `internal/dto/`
- **Models:** Database models in `internal/models/`

### Testing Guidelines
- Unit tests for services with mocks
- Integration tests in `tests/integration/`
- Security tests in `tests/security/`
- Run all tests: `go test ./...`
- Run with coverage: `go test -cover ./...`

## 🎯 Immediate Recommendation

**Start with #3 (Export Functionality)** because:
1. Complements the recently completed CSV import feature
2. Critical for user data portability and compliance
3. Relatively self-contained implementation
4. Natural follow-up to import functionality
5. Users can now import data and should be able to export it

**Alternative: Portfolio Comparison (#5)** if you prefer working on analytics features:
1. API endpoint already partially implemented in the CLI
2. Would complete the performance analytics suite
3. Useful for users managing multiple portfolios

---

*Last Updated: 2025-11-12*
*Last Commit: feat: configure deployment to trigger only on release tags*

### CLI Implementation Highlights
- **14 new files** created for full-featured CLI
- **Latest library versions**: Cobra v1.10.1, Bubble Tea v1.3.10, Lipgloss v1.1.0, Viper v1.21.0
- **30+ commands** across auth, portfolio, transaction, and performance domains
- **Interactive TUI** with keyboard navigation and beautiful styled output
- **Multi-format output** supporting table, JSON, and CSV formats
- **Comprehensive docs** with usage examples and troubleshooting guide
- **Build automation** via Makefile with version injection

### Previous Test Coverage Improvements
- Added comprehensive tests for `performance_analytics_service.go` (0% → ~90%)
- Added all CRUD method tests for `corporate_action_service.go` (0% → 100%)
- Added complex corporate action tests (ApplySpinoff, ApplyTickerChange)
- Added database package tests (0% → 40.5%)
- Added logger package tests (0% → 88.1%)
- Services package improved from 57.2% to 78.7% coverage

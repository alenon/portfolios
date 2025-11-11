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
**Status:** Not Started
**Priority:** High - Enables automated updates

**Tasks:**
- [ ] Create end-of-day price update job
  - Fetch closing prices for all held symbols
  - Update holdings with current market values
  - Create/update performance snapshots
- [ ] Create performance snapshot generation job
  - Run nightly for all active portfolios
  - Calculate daily performance metrics
- [ ] Create stale data cleanup job
  - Remove old cache entries
  - Clean up expired tokens
  - Archive old snapshots
- [ ] Wire up jobs in scheduler
- [ ] Add job monitoring/logging
- [ ] Add configuration for job schedules

**Files to create:**
- `internal/jobs/price_update_job.go`
- `internal/jobs/snapshot_generation_job.go`
- `internal/jobs/cleanup_job.go`

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
**Status:** Not Started
**Priority:** Medium - Extensive CLI described in product spec

The product spec describes a comprehensive CLI but none exists yet.

**Tasks:**
- [ ] Set up Cobra CLI framework
- [ ] Implement authentication commands:
  - `portfolios auth login`
  - `portfolios auth logout`
  - `portfolios auth register`
- [ ] Implement portfolio commands:
  - `portfolios portfolio list`
  - `portfolios portfolio create`
  - `portfolios portfolio show <id>`
  - `portfolios portfolio delete <id>`
- [ ] Implement transaction commands:
  - `portfolios transaction add`
  - `portfolios transaction import <file>`
  - `portfolios transaction list`
- [ ] Implement performance commands:
  - `portfolios performance show <portfolio-id>`
  - `portfolios performance compare <id1> <id2>`
  - `portfolios performance benchmark <portfolio-id> <symbol>`
- [ ] Implement tax lot commands
- [ ] Implement market data query commands
- [ ] Add configuration file support (`~/.portfolios/config.yaml`)
- [ ] Add output formatting (table, JSON, CSV)
- [ ] Add comprehensive help text

**Files to create:**
- `cmd/portfolios/main.go`
- `cmd/portfolios/cmd/*.go` (command files)
- `internal/cli/*.go` (CLI utilities)

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
- **Commit:** [Latest] - Implemented CSV import functionality with broker-specific parsers
- **Branch:** `claude/implement-next-steps-011CV2mfhyT66RhFApDVJ9nZ`
- **All tests passing:** ✅
- **Test Coverage:** 55.3% overall (services: 78.7%, dto: 97.3%, logger: 88.1%, models: 97.0%, utils: 96.9%)

### Recent Changes (2025-11-11)
- **CSV Import Feature:** Completed full implementation
  - 7 broker-specific CSV parsers (Fidelity, Schwab, TD Ameritrade, E*TRADE, Interactive Brokers, Robinhood, Generic)
  - Import validation and error handling with detailed error reporting
  - Batch tracking and management capabilities
  - Dry-run mode for validation before import
  - 4 new API endpoints for import operations
- **Architecture:** Fixed import cycles by moving Quote and PerformanceMetrics types to dto package
- **All tests:** Passing (13 test suites, 0 failures)

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

**Start with #1 (CSV Import Functionality)** because:
1. It's a core user-facing feature per the product spec
2. It's relatively self-contained
3. Enables users to populate portfolios with real data
4. Required before the CLI tool would be truly useful
5. Unblocks user testing and feedback

---

*Last Updated: 2025-11-11*
*Last Commit: feat: implement CSV import functionality with broker-specific parsers*

### Latest Test Coverage Improvements
- Added comprehensive tests for `performance_analytics_service.go` (0% → ~90%)
- Added all CRUD method tests for `corporate_action_service.go` (0% → 100%)
- Added complex corporate action tests (ApplySpinoff, ApplyTickerChange)
- Added database package tests (0% → 40.5%)
- Added logger package tests (0% → 88.1%)
- Services package improved from 57.2% to 78.7% coverage

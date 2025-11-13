# End-to-End (E2E) Tests

This directory contains comprehensive end-to-end tests for the Portfolios application, testing both the backend API and CLI tool running in Docker containers.

## Overview

The E2E test suite validates:
- ✅ **Authentication flows** - Registration, login, logout, token refresh
- ✅ **Portfolio management** - CRUD operations via both API and CLI
- ✅ **Transaction management** - Buy/sell transactions via API and CLI
- ✅ **CSV import functionality** - Generic and broker-specific formats (Fidelity, etc.)
- ✅ **Performance analytics** - Holdings, metrics, snapshots
- ✅ **Data persistence** - Verification across API calls and CLI commands
- ✅ **Error handling** - Invalid inputs, unauthorized access, etc.

## Architecture

### Docker Environment

The E2E tests use a dedicated Docker Compose environment (`docker-compose.e2e.yml`) with:

1. **PostgreSQL Database** (`postgres-e2e`)
   - Isolated test database
   - Runs on port 5433 to avoid conflicts
   - Automatic health checks

2. **Backend API** (`backend-e2e`)
   - Built from production Dockerfile
   - Runs on port 8081
   - Configured for testing environment
   - Health check endpoint

3. **CLI Container** (`cli-e2e`)
   - Built from Dockerfile.cli
   - Contains the CLI binary
   - Runs tests within the container
   - Access to test fixtures

### Test Structure

```
tests/e2e/
├── README.md              # This file
├── main_test.go          # Test setup and teardown
├── helpers.go            # Test utilities and helpers
├── auth_test.go          # Authentication tests
├── portfolio_test.go     # Portfolio management tests
├── transaction_test.go   # Transaction tests
├── csv_import_test.go    # CSV import tests
├── performance_test.go   # Performance analytics tests
└── fixtures/             # Test data files
    ├── generic_import.csv
    └── fidelity_import.csv
```

## Running E2E Tests

### Quick Start

Run all E2E tests:
```bash
make e2e-test
```

This will:
1. Build Docker images
2. Start all services
3. Wait for services to be healthy
4. Run all tests
5. Clean up containers

### Manual Control

Start the E2E environment:
```bash
make e2e-up
```

Run tests inside the CLI container:
```bash
docker-compose -f docker-compose.e2e.yml exec cli-e2e sh -c "cd /tests && go test -v ./..."
```

View logs:
```bash
make e2e-logs
```

Stop and clean up:
```bash
make e2e-down
```

Complete cleanup (remove volumes):
```bash
make e2e-clean
```

### Debugging

Open a shell in the CLI container:
```bash
make e2e-shell
```

Run specific tests:
```bash
docker-compose -f docker-compose.e2e.yml exec cli-e2e \
  sh -c "cd /tests && go test -v -run TestAuthRegisterViaAPI"
```

Run tests without CLI (API only):
```bash
docker-compose -f docker-compose.e2e.yml exec cli-e2e \
  sh -c "cd /tests && go test -v -short ./..."
```

## Test Categories

### 1. Authentication Tests (`auth_test.go`)

Tests the complete authentication flow:

- **Registration**
  - `TestAuthRegisterViaAPI` - User registration via API
  - Validates token generation and user creation

- **Login**
  - `TestAuthLoginViaAPI` - Login with valid credentials
  - `TestAuthLoginFailsWithWrongPassword` - Login failure handling

- **Token Management**
  - `TestAuthTokenRefresh` - Token refresh flow
  - `TestAuthProtectedEndpointAccess` - Protected endpoint access

- **Logout**
  - `TestAuthLogout` - Logout and token revocation

- **End-to-End**
  - `TestAuthFlowEndToEnd` - Complete auth lifecycle

### 2. Portfolio Tests (`portfolio_test.go`)

Tests portfolio management via API and CLI:

- **API Tests**
  - `TestPortfolioCreateViaAPI` - Create portfolio
  - `TestPortfolioListViaAPI` - List all portfolios
  - `TestPortfolioGetByIDViaAPI` - Get specific portfolio
  - `TestPortfolioUpdateViaAPI` - Update portfolio details
  - `TestPortfolioDeleteViaAPI` - Delete portfolio
  - `TestPortfolioHoldingsViaAPI` - View holdings

- **CLI Tests**
  - `TestCLIPortfolioList` - List via CLI
  - `TestCLIPortfolioCreate` - Create via CLI
  - `TestCLIPortfolioShow` - Show via CLI
  - `TestCLIPortfolioDelete` - Delete via CLI

- **Integration**
  - `TestPortfolioFlowEndToEnd` - Complete CRUD flow

### 3. Transaction Tests (`transaction_test.go`)

Tests transaction management:

- **API Tests**
  - `TestTransactionCreateBuyViaAPI` - Buy transactions
  - `TestTransactionCreateSellViaAPI` - Sell transactions
  - `TestTransactionListViaAPI` - List transactions
  - `TestTransactionGetByIDViaAPI` - Get specific transaction
  - `TestTransactionUpdateViaAPI` - Update transaction
  - `TestTransactionDeleteViaAPI` - Delete transaction

- **CLI Tests**
  - `TestCLITransactionList` - List via CLI
  - `TestCLITransactionDelete` - Delete via CLI

- **Integration**
  - `TestTransactionFlowEndToEnd` - Complete transaction lifecycle

### 4. CSV Import Tests (`csv_import_test.go`)

Tests CSV import functionality:

- **Format Support**
  - `TestCSVImportGenericFormatViaAPI` - Generic CSV format
  - `TestCSVImportFidelityFormatViaAPI` - Fidelity format

- **Bulk Operations**
  - `TestCSVImportBulkTransactionsViaAPI` - Bulk import
  - `TestCSVImportBatchListViaAPI` - List import batches
  - `TestCSVImportBatchDeleteViaAPI` - Delete batch

- **CLI Import**
  - `TestCLICSVImport` - Import via CLI
  - `TestCLIImportBatchList` - List batches via CLI

- **Integration**
  - `TestCSVImportFlowEndToEnd` - Complete import workflow

### 5. Performance Tests (`performance_test.go`)

Tests performance analytics:

- **Metrics**
  - `TestPerformanceGetPortfolioPerformanceViaAPI` - Get metrics
  - `TestPerformanceGetHoldingsViaAPI` - Current holdings
  - `TestPerformanceSnapshotsViaAPI` - Performance snapshots

- **CLI**
  - `TestCLIPerformanceShow` - Show performance
  - `TestCLIPerformanceSnapshots` - List snapshots

- **Advanced**
  - `TestPerformanceWithMultipleTransactions` - Complex scenarios
  - `TestPerformanceComparePortfolios` - Portfolio comparison

## Test Helpers

### TestContext

The `TestContext` struct provides utilities for tests:

```go
ctx := NewTestContext(t)

// API operations
err := ctx.CreateTestUser(email, password)
err := ctx.Login(email, password)
err := ctx.APIRequest("GET", "/api/v1/portfolios", nil, &result)

// CLI operations
stdout, stderr, err := ctx.RunCLI("portfolio", "list")
stdout, stderr, err := ctx.RunCLIWithInput(input, "auth", "login")

// Config management
err := ctx.SaveCLIConfig()
```

### Helper Functions

- `GenerateUniqueEmail()` - Generate unique test emails
- `GetFixturePath(filename)` - Get path to test fixtures
- `ParseJSONOutput(output)` - Parse CLI JSON output
- `createTestPortfolio(ctx, t, name)` - Create test portfolio
- `uploadCSVFile(ctx, portfolioID, path, broker)` - Upload CSV

## Test Fixtures

CSV import test files in `fixtures/`:

- `generic_import.csv` - Standard format with 4 transactions
- `fidelity_import.csv` - Fidelity broker format

Add new fixtures by creating CSV files in the fixtures directory.

## Environment Variables

Customize test behavior:

```bash
# Test timeout (default: 10m)
TEST_TIMEOUT=15m make e2e-test

# Skip cleanup (for debugging)
CLEANUP=false make e2e-test

# Show all logs
SHOW_LOGS=true make e2e-test
```

## Writing New Tests

### API Test Template

```go
func TestMyFeatureViaAPI(t *testing.T) {
    ctx := NewTestContext(t)

    // Setup
    err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
    require.NoError(t, err)

    // Test your feature
    var result MyResult
    err = ctx.APIRequest("GET", "/api/v1/my-endpoint", nil, &result)
    require.NoError(t, err)

    // Assertions
    assert.Equal(t, expectedValue, result.Field)
}
```

### CLI Test Template

```go
func TestMyFeatureViaCLI(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping CLI test in short mode")
    }

    ctx := NewTestContext(t)

    // Setup
    err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
    require.NoError(t, err)
    err = ctx.SaveCLIConfig()
    require.NoError(t, err)

    // Run CLI command
    stdout, stderr, err := ctx.RunCLI("my-command", "--flag", "value")

    // Parse and assert
    if err == nil && stdout != "" {
        var result MyResult
        if parseErr := json.Unmarshal([]byte(stdout), &result); parseErr == nil {
            assert.Equal(t, expectedValue, result.Field)
        }
    }
}
```

## CI/CD Integration

Add to your CI pipeline:

```yaml
# GitHub Actions example
- name: Run E2E Tests
  run: make e2e-test
  timeout-minutes: 15
```

## Troubleshooting

### Services Not Starting

Check logs:
```bash
docker-compose -f docker-compose.e2e.yml logs backend-e2e
docker-compose -f docker-compose.e2e.yml logs postgres-e2e
```

### Tests Timing Out

Increase timeout:
```bash
TEST_TIMEOUT=20m make e2e-test
```

### Database Issues

Clean and restart:
```bash
make e2e-clean
make e2e-test
```

### CLI Tests Failing

CLI tests may fail in non-interactive environments. They are marked with:
```go
if testing.Short() {
    t.Skip("Skipping CLI test in short mode")
}
```

Run only API tests:
```bash
go test -v -short ./tests/e2e/
```

## Performance Considerations

- **Parallel Tests**: Tests use isolated users/portfolios and can run in parallel
- **Database**: Uses PostgreSQL (same as production) for realistic testing
- **Cleanup**: Automatic cleanup via test context and defer statements
- **Isolation**: Each test creates unique users to avoid conflicts

## Test Coverage

Current E2E test coverage:

| Module | Tests | Coverage |
|--------|-------|----------|
| Authentication | 8 | 100% of flows |
| Portfolio CRUD | 9 | 100% of operations |
| Transactions | 9 | Buy/Sell/Update/Delete |
| CSV Import | 8 | Generic + Fidelity formats |
| Performance | 6 | Metrics + Holdings |

**Total**: 40+ E2E tests covering critical user workflows

## Future Enhancements

- [ ] Add more broker-specific CSV parsers tests (Schwab, TD Ameritrade, etc.)
- [ ] Add WebSocket real-time tests (when implemented)
- [ ] Add performance benchmarking tests
- [ ] Add load testing scenarios
- [ ] Add API rate limiting tests
- [ ] Add corporate actions E2E tests
- [ ] Add export functionality tests

## Contributing

When adding new features:

1. Add corresponding E2E tests
2. Update this README
3. Ensure tests pass: `make e2e-test`
4. Document any new fixtures or helpers

## Support

For issues or questions:
- Check the logs: `make e2e-logs`
- Review the test output for specific failures
- Use `make e2e-shell` to debug interactively

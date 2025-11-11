# Quality Assurance Rules

## Test Coverage Requirements

Before completing any feature implementation or code changes, you MUST verify:

### 1. Test Coverage Verification

Run the following command and verify coverage meets requirements:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -1
```

**Requirements:**
- **Business Logic Coverage**: Minimum 80% for `internal/services`, `internal/repository`, and `internal/models` packages
- **Overall Project Coverage**: Target 80%, but minimum 70% is acceptable if infrastructure code is excluded

To check business logic coverage specifically:

```bash
go test -v ./internal/services/... ./internal/repository/... ./internal/models/... -coverprofile=coverage_business.out
go tool cover -func=coverage_business.out | tail -1
```

**What counts toward coverage:**
- ✅ Services (business logic)
- ✅ Repositories (data access)
- ✅ Models (domain logic, validation, calculations)
- ✅ Handlers (HTTP request handling)
- ✅ Middleware (where testable)

**What typically has lower coverage (acceptable):**
- ⚠️ `cmd/api/main.go` - Application entry point
- ⚠️ `internal/database` - Database connection utilities
- ⚠️ `internal/dto` - Data transfer objects (simple structs)
- ⚠️ `internal/logger` - Logging utilities
- ⚠️ `internal/utils` - Simple utility functions
- ⚠️ `scripts` - Seed/migration scripts

### 2. Code Formatting Verification

Before committing any code changes, ensure all Go files are properly formatted:

```bash
go fmt ./...
```

**Expected output:**
- No output means all files are already formatted correctly
- If files are reformatted, the command will list them

**Alternative**: Use `gofmt` to check without modifying files:
```bash
# Check if any files need formatting (should return nothing)
gofmt -l .
```

**Format specific files:**
```bash
# Format a specific file or directory
go fmt ./internal/services/...
```

**Requirements:**
- ✅ All Go files must be formatted using `go fmt`
- ✅ No unformatted files in the codebase
- ✅ Formatting must be applied BEFORE running the linter (linter may catch formatting issues)

### 3. Linter Verification

Run the linter and ensure it passes with zero issues:

```bash
golangci-lint run --timeout=5m
```

**Expected output:**
```
0 issues.
```

If golangci-lint is not installed, check using:
```bash
command -v golangci-lint >/dev/null 2>&1 && echo "installed" || echo "not installed"
```

### 4. Test Execution

Ensure all tests pass:

```bash
go test ./...
```

All tests must pass with no failures or panics.

## When to Run These Checks

You MUST run these verification steps:

1. **Before committing code** - Always verify tests pass and linter is clean
2. **After adding new features** - Ensure new code has adequate test coverage
3. **After refactoring** - Verify coverage hasn't decreased
4. **Before creating pull requests** - Final verification that everything passes
5. **When user requests verification** - When explicitly asked to verify quality

## Reporting Results

When reporting completion of a feature or task, include:

1. **Code Formatting**:
   - ✅ All files formatted
   - ❌ X files need formatting

2. **Linter Status**:
   - ✅ Passing (0 issues)
   - ❌ Failing (X issues)

3. **Test Coverage**:
   - Business Logic: X%
   - Overall Project: X%

4. **Test Results**:
   - All tests passing: ✅ / ❌
   - Number of test files added/modified

## Handling Coverage Below 80%

If coverage is below 80% for business logic:

1. **Identify gaps**: Use `go tool cover -html=coverage.out` to see uncovered lines
2. **Write tests**: Add tests for critical paths and business logic
3. **Focus on**:
   - Error handling paths
   - Edge cases
   - Validation logic
   - Business calculations
   - Authorization checks

## Example Workflow

```bash
# 1. Write code and tests

# 2. Format all code
go fmt ./...

# 3. Run tests
go test ./...

# 4. Check coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -1

# 5. Check business logic coverage
go test -coverprofile=coverage_business.out ./internal/services/... ./internal/repository/... ./internal/models/...
go tool cover -func=coverage_business.out | tail -1

# 6. Run linter
golangci-lint run --timeout=5m

# 7. If all pass, commit and push
git add .
git commit -m "your commit message"
git push
```

## CI/CD Integration

These same checks run in CI/CD:
- GitHub Actions workflow runs `go fmt`, `golangci-lint`, and `go test`
- Pull requests must pass all checks
- Coverage reports are generated automatically
- Formatting violations will cause CI to fail

## Non-Negotiable

⚠️ **CRITICAL**: Never skip formatting, linter, or test verification before committing code. These checks:
- Catch bugs before they reach production
- Maintain consistent code quality and style
- Ensure the codebase remains readable and maintainable
- Prevent merge conflicts from formatting inconsistencies

**The order matters:**
1. Format first (`go fmt`) - ensures consistent style
2. Test second (`go test`) - verifies functionality
3. Lint last (`golangci-lint`) - catches code quality issues

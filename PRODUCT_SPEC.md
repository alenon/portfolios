# Product Specification: Portfolios
## CLI and Server-Side Implementation Focus

---

## Product Mission

Portfolios is a portfolio management platform that helps retail investors track and analyze their stock market investments by providing a centralized, real-time view of all their portfolios with automatic corporate action handling and comprehensive performance analytics.

Unlike competitors like Sharesight that require broker API integrations, Portfolios uses manual transaction import (CSV) and entry, giving users complete control over their data without requiring broker credentials.

---

## Target Users

### Primary User Personas

**Sarah - The Organized Investor** (35-45)
- Manages 3-4 portfolios across different brokers
- Needs consolidated view of all portfolio performance
- Struggles with manual spreadsheet tracking
- Misses corporate actions affecting cost basis

**Michael - The Active Trader** (28-40)
- Trades weekly with regular buy/sell transactions
- Needs detailed performance analytics
- Wants to track performance of different trading strategies
- Requires quick transaction entry

**Linda - The Long-Term Planner** (50-65)
- Building wealth for retirement with dividend-paying stocks
- Multiple accounts (401k rollovers, IRAs, taxable accounts)
- Loses track of dividend reinvestments and stock splits
- Needs accurate dividend tracking and automatic corporate action handling

---

## Core Problems & Solutions

### 1. Fragmented Portfolio Visibility
**Problem:** Investors with multiple brokerage accounts lack unified view of total investment performance.

**Solution:** Manual transaction import (CSV) and CLI-based entry provides comprehensive portfolio analytics without broker API access.

### 2. Inaccurate Performance Tracking
**Problem:** Corporate actions (splits, mergers, dividend reinvestments) complicate tracking and cause cost basis errors.

**Solution:** Automatic corporate action detection and application ensures accurate portfolio data.

### 3. No Real-Time Monitoring
**Problem:** Checking multiple broker websites is tedious and doesn't provide consolidated views.

**Solution:** Live portfolio monitoring with real-time market data shows current portfolio values and daily changes.

---

## Server-Side Architecture

### Technology Stack
- **Language:** Go 1.21+
- **Framework:** Gin
- **Database:** PostgreSQL 15+
- **ORM:** GORM
- **Authentication:** JWT
- **Migrations:** golang-migrate

### Core System Components

#### 1. Transaction Management Service
**Responsibilities:**
- Process buy/sell transactions
- Validate transaction data
- Calculate cost basis using appropriate methods (FIFO, LIFO, specific lot)
- Maintain transaction history audit trail
- Support bulk import from CSV

**Key Functions:**
```go
// Core transaction operations
CreateTransaction(userID, portfolioID, transaction) error
UpdateTransaction(transactionID, updates) error
DeleteTransaction(transactionID) error
GetTransactionHistory(portfolioID, filters) ([]Transaction, error)
BulkImportTransactions(portfolioID, csvData) (ImportResult, error)
```

#### 2. Portfolio Management Service
**Responsibilities:**
- Create and manage multiple portfolios per user
- Track current holdings and positions
- Calculate portfolio-level metrics
- Support portfolio comparison
- Handle portfolio-level settings (cost basis method, reporting currency)

**Key Functions:**
```go
CreatePortfolio(userID, portfolio) error
GetPortfolio(portfolioID) (Portfolio, error)
ListPortfolios(userID) ([]Portfolio, error)
GetCurrentHoldings(portfolioID) ([]Holding, error)
ComparePortfolios(portfolioIDs) (Comparison, error)
```

#### 3. Corporate Actions Service
**Responsibilities:**
- Detect and process stock splits
- Handle dividend payments and reinvestments
- Process mergers, acquisitions, and ticker changes
- Support spinoffs
- Automatically adjust cost basis and share quantities

**Key Functions:**
```go
DetectCorporateActions(symbol, dateRange) ([]CorporateAction, error)
ApplyCorporateAction(portfolioID, action) error
ProcessStockSplit(portfolioID, symbol, ratio, date) error
ProcessDividend(portfolioID, symbol, amount, type, date) error
ProcessMerger(portfolioID, oldSymbol, newSymbol, ratio, date) error
ProcessSpinoff(portfolioID, parentSymbol, spinoffSymbol, ratio, date) error
```

#### 4. Market Data Service
**Responsibilities:**
- Fetch real-time stock quotes
- Retrieve historical price data
- Cache market data efficiently
- Support multiple exchanges
- Handle currency conversion for international stocks

**Key Functions:**
```go
GetQuote(symbol) (Quote, error)
GetQuotes(symbols) (map[string]Quote, error)
GetHistoricalPrices(symbol, dateRange) ([]Price, error)
GetExchangeRate(fromCurrency, toCurrency) (float64, error)
```

#### 5. Performance Analytics Service
**Responsibilities:**
- Calculate total return (absolute and percentage)
- Calculate time-weighted return (TWR)
- Calculate money-weighted return (MWR/IRR)
- Calculate annualized returns
- Generate performance reports over custom date ranges
- Support benchmark comparison

**Key Functions:**
```go
CalculateTotalReturn(portfolioID, dateRange) (Return, error)
CalculateTimeWeightedReturn(portfolioID, dateRange) (float64, error)
CalculateMoneyWeightedReturn(portfolioID, dateRange) (float64, error)
CalculateAnnualizedReturn(portfolioID, dateRange) (float64, error)
CompareToBenchmark(portfolioID, benchmark, dateRange) (Comparison, error)
GetPerformanceMetrics(portfolioID, dateRange) (Metrics, error)
```

#### 6. Tax Lot Tracking Service
**Responsibilities:**
- Track individual purchase lots
- Calculate lot-specific cost basis
- Support multiple cost basis methods (FIFO, LIFO, specific lot)
- Identify tax-loss harvesting opportunities
- Generate tax reports

**Key Functions:**
```go
GetTaxLots(portfolioID, symbol) ([]TaxLot, error)
AllocateSale(portfolioID, symbol, quantity, method) ([]LotAllocation, error)
IdentifyTaxLossOpportunities(portfolioID, threshold) ([]Opportunity, error)
GenerateTaxReport(portfolioID, taxYear) (TaxReport, error)
```

---

## Data Models

### Core Entities

#### User
```go
type User struct {
    ID           uint      `gorm:"primaryKey"`
    Email        string    `gorm:"unique;not null"`
    PasswordHash string    `gorm:"not null"`
    FirstName    string
    LastName     string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

#### Portfolio
```go
type Portfolio struct {
    ID               uint      `gorm:"primaryKey"`
    UserID           uint      `gorm:"not null;index"`
    Name             string    `gorm:"not null"`
    Description      string
    BaseCurrency     string    `gorm:"default:'USD'"`
    CostBasisMethod  string    `gorm:"default:'FIFO'"` // FIFO, LIFO, SPECIFIC_LOT
    CreatedAt        time.Time
    UpdatedAt        time.Time
    User             User      `gorm:"foreignKey:UserID"`
}
```

#### Transaction
```go
type Transaction struct {
    ID            uint      `gorm:"primaryKey"`
    PortfolioID   uint      `gorm:"not null;index"`
    Type          string    `gorm:"not null"` // BUY, SELL, DIVIDEND, SPLIT, MERGER, SPINOFF
    Symbol        string    `gorm:"not null;index"`
    Date          time.Time `gorm:"not null;index"`
    Quantity      float64   `gorm:"not null"`
    Price         float64
    Commission    float64   `gorm:"default:0"`
    Currency      string    `gorm:"default:'USD'"`
    Notes         string
    ImportBatchID *uint     // Track which import batch this came from
    CreatedAt     time.Time
    UpdatedAt     time.Time
    Portfolio     Portfolio `gorm:"foreignKey:PortfolioID"`
}
```

#### Holding
```go
type Holding struct {
    ID           uint      `gorm:"primaryKey"`
    PortfolioID  uint      `gorm:"not null;index"`
    Symbol       string    `gorm:"not null;index"`
    Quantity     float64   `gorm:"not null"`
    CostBasis    float64   `gorm:"not null"` // Total cost basis
    AvgCostPrice float64   `gorm:"not null"` // Average cost per share
    UpdatedAt    time.Time
    Portfolio    Portfolio `gorm:"foreignKey:PortfolioID"`
}
```

#### TaxLot
```go
type TaxLot struct {
    ID           uint      `gorm:"primaryKey"`
    PortfolioID  uint      `gorm:"not null;index"`
    Symbol       string    `gorm:"not null;index"`
    PurchaseDate time.Time `gorm:"not null"`
    Quantity     float64   `gorm:"not null"`
    CostBasis    float64   `gorm:"not null"`
    TransactionID uint     `gorm:"not null"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

#### CorporateAction
```go
type CorporateAction struct {
    ID          uint      `gorm:"primaryKey"`
    Symbol      string    `gorm:"not null;index"`
    Type        string    `gorm:"not null"` // SPLIT, DIVIDEND, MERGER, SPINOFF
    Date        time.Time `gorm:"not null;index"`
    Ratio       float64   // For splits and mergers
    Amount      float64   // For dividends
    NewSymbol   string    // For mergers and ticker changes
    Currency    string
    Description string
    Applied     bool      `gorm:"default:false"` // Has this been applied to portfolios?
    CreatedAt   time.Time
}
```

#### PerformanceSnapshot
```go
type PerformanceSnapshot struct {
    ID               uint      `gorm:"primaryKey"`
    PortfolioID      uint      `gorm:"not null;index"`
    Date             time.Time `gorm:"not null;index"`
    TotalValue       float64   `gorm:"not null"`
    TotalCostBasis   float64   `gorm:"not null"`
    TotalReturn      float64   `gorm:"not null"`
    TotalReturnPct   float64   `gorm:"not null"`
    DayChange        float64
    DayChangePct     float64
    CreatedAt        time.Time
    Portfolio        Portfolio `gorm:"foreignKey:PortfolioID"`
}
```

---

## CLI Implementation

### Command Structure

The CLI should follow a hierarchical command structure for intuitive usage:

```bash
portfolios [global-flags] <command> [command-flags] [arguments]
```

### Global Flags
```
--config, -c     Configuration file path (default: ~/.portfolios/config.yaml)
--profile, -p    Profile to use (default: default)
--output, -o     Output format: json, table, csv (default: table)
--verbose, -v    Verbose output
--quiet, -q      Suppress non-essential output
```

### Core CLI Commands

#### 1. Authentication Commands
```bash
# Login and get JWT token
portfolios auth login --email user@example.com --password <password>

# Logout (clear stored token)
portfolios auth logout

# Register new user
portfolios auth register --email user@example.com --password <password> --name "John Doe"

# Refresh token
portfolios auth refresh
```

#### 2. Portfolio Commands
```bash
# List all portfolios
portfolios portfolio list

# Create new portfolio
portfolios portfolio create --name "Retirement IRA" --currency USD --cost-basis-method FIFO

# Get portfolio details
portfolios portfolio get <portfolio-id>

# Update portfolio
portfolios portfolio update <portfolio-id> --name "New Name" --description "Updated description"

# Delete portfolio
portfolios portfolio delete <portfolio-id>

# Show current holdings
portfolios portfolio holdings <portfolio-id>

# Show portfolio performance
portfolios portfolio performance <portfolio-id> --from 2024-01-01 --to 2024-12-31
```

#### 3. Transaction Commands
```bash
# Add buy transaction
portfolios transaction buy <portfolio-id> --symbol AAPL --date 2024-01-15 --quantity 10 --price 180.50 --commission 1.00

# Add sell transaction
portfolios transaction sell <portfolio-id> --symbol AAPL --date 2024-06-15 --quantity 5 --price 195.25 --commission 1.00

# Add dividend
portfolios transaction dividend <portfolio-id> --symbol AAPL --date 2024-03-15 --amount 25.50 --type CASH

# Import from CSV
portfolios transaction import <portfolio-id> --file transactions.csv --format <broker-format>

# List transactions
portfolios transaction list <portfolio-id> --symbol AAPL --from 2024-01-01 --to 2024-12-31

# Delete transaction
portfolios transaction delete <transaction-id>

# Bulk import
portfolios transaction import <portfolio-id> --file /path/to/transactions.csv --broker fidelity
```

#### 4. Corporate Actions Commands
```bash
# List corporate actions for a symbol
portfolios corpaction list --symbol AAPL --from 2024-01-01

# Apply stock split manually
portfolios corpaction split <portfolio-id> --symbol AAPL --ratio 4:1 --date 2024-08-28

# Apply dividend
portfolios corpaction dividend <portfolio-id> --symbol AAPL --amount 0.25 --type REINVEST --date 2024-05-15

# Apply merger
portfolios corpaction merger <portfolio-id> --old-symbol XYZ --new-symbol ABC --ratio 1.5 --date 2024-07-01

# Auto-detect and apply corporate actions
portfolios corpaction auto-apply <portfolio-id> --from 2024-01-01
```

#### 5. Performance & Analytics Commands
```bash
# Show performance metrics
portfolios performance metrics <portfolio-id> --from 2024-01-01 --to 2024-12-31

# Calculate time-weighted return
portfolios performance twr <portfolio-id> --from 2024-01-01 --to 2024-12-31

# Calculate money-weighted return
portfolios performance mwr <portfolio-id> --from 2024-01-01 --to 2024-12-31

# Compare portfolios
portfolios performance compare <portfolio-id-1> <portfolio-id-2> --from 2024-01-01

# Compare to benchmark
portfolios performance benchmark <portfolio-id> --index SPY --from 2024-01-01

# Generate performance report
portfolios performance report <portfolio-id> --from 2024-01-01 --to 2024-12-31 --output report.pdf
```

#### 6. Tax Lot Commands
```bash
# List tax lots for a symbol
portfolios taxlot list <portfolio-id> --symbol AAPL

# Show tax-loss harvesting opportunities
portfolios taxlot harvest <portfolio-id> --threshold -3.0

# Generate tax report
portfolios taxlot report <portfolio-id> --year 2024 --output tax-report.pdf
```

#### 7. Market Data Commands
```bash
# Get current quote
portfolios market quote AAPL

# Get multiple quotes
portfolios market quote AAPL MSFT GOOGL

# Get historical prices
portfolios market history AAPL --from 2024-01-01 --to 2024-12-31

# Get exchange rate
portfolios market exchange --from USD --to EUR
```

#### 8. Export Commands
```bash
# Export portfolio to CSV
portfolios export csv <portfolio-id> --output portfolio.csv

# Export transactions
portfolios export transactions <portfolio-id> --output transactions.csv --from 2024-01-01

# Export holdings
portfolios export holdings <portfolio-id> --output holdings.csv

# Export performance report
portfolios export performance <portfolio-id> --from 2024-01-01 --to 2024-12-31 --format pdf --output report.pdf
```

### CLI Configuration

Configuration file location: `~/.portfolios/config.yaml`

```yaml
default_profile: default

profiles:
  default:
    api_url: http://localhost:8080
    auth_token: <jwt-token>
    default_currency: USD
    default_cost_basis_method: FIFO
    output_format: table

  production:
    api_url: https://api.portfolios.example.com
    auth_token: <jwt-token>
```

### CLI Output Formats

#### Table Format (Default)
```
PORTFOLIO: Retirement IRA (ID: 123)
Currency: USD | Cost Basis Method: FIFO

HOLDINGS:
+--------+----------+------------+-----------+----------+----------+
| Symbol | Quantity | Avg Cost   | Market    | Total    | Gain/Loss|
|        |          | per Share  | Price     | Value    | (%)      |
+--------+----------+------------+-----------+----------+----------+
| AAPL   | 100.00   | $150.25    | $180.50   | $18,050  | +20.15%  |
| MSFT   | 50.00    | $280.00    | $310.25   | $15,512  | +10.80%  |
+--------+----------+------------+-----------+----------+----------+
TOTAL                                         | $33,562  | +15.85%  |
+--------+----------+------------+-----------+----------+----------+
```

#### JSON Format
```json
{
  "portfolio": {
    "id": 123,
    "name": "Retirement IRA",
    "currency": "USD",
    "cost_basis_method": "FIFO"
  },
  "holdings": [
    {
      "symbol": "AAPL",
      "quantity": 100.0,
      "avg_cost_per_share": 150.25,
      "market_price": 180.50,
      "total_value": 18050.00,
      "gain_loss_pct": 20.15
    }
  ]
}
```

#### CSV Format
```csv
Symbol,Quantity,AvgCostPerShare,MarketPrice,TotalValue,GainLossPct
AAPL,100.00,150.25,180.50,18050.00,20.15
MSFT,50.00,280.00,310.25,15512.50,10.80
```

---

## API Endpoints

### Authentication
```
POST   /api/v1/auth/register          Register new user
POST   /api/v1/auth/login             Login and get JWT token
POST   /api/v1/auth/refresh           Refresh JWT token
POST   /api/v1/auth/logout            Logout (invalidate token)
POST   /api/v1/auth/password/reset    Request password reset
POST   /api/v1/auth/password/confirm  Confirm password reset
```

### Portfolios
```
GET    /api/v1/portfolios             List user's portfolios
POST   /api/v1/portfolios             Create new portfolio
GET    /api/v1/portfolios/:id         Get portfolio details
PUT    /api/v1/portfolios/:id         Update portfolio
DELETE /api/v1/portfolios/:id         Delete portfolio
GET    /api/v1/portfolios/:id/holdings           Get current holdings
GET    /api/v1/portfolios/:id/performance        Get performance metrics
GET    /api/v1/portfolios/compare                Compare multiple portfolios
```

### Transactions
```
GET    /api/v1/portfolios/:id/transactions       List transactions
POST   /api/v1/portfolios/:id/transactions       Create transaction
POST   /api/v1/portfolios/:id/transactions/bulk  Bulk import transactions
GET    /api/v1/transactions/:id                  Get transaction details
PUT    /api/v1/transactions/:id                  Update transaction
DELETE /api/v1/transactions/:id                  Delete transaction
```

### Corporate Actions
```
GET    /api/v1/corporate-actions                 List corporate actions
POST   /api/v1/corporate-actions                 Create corporate action
GET    /api/v1/corporate-actions/:symbol         Get actions for symbol
POST   /api/v1/portfolios/:id/corporate-actions/apply   Apply corporate action
POST   /api/v1/portfolios/:id/corporate-actions/auto    Auto-detect and apply
```

### Market Data
```
GET    /api/v1/market/quote/:symbol              Get current quote
POST   /api/v1/market/quotes                     Get multiple quotes
GET    /api/v1/market/history/:symbol            Get historical prices
GET    /api/v1/market/exchange                   Get exchange rate
```

### Performance Analytics
```
GET    /api/v1/portfolios/:id/performance/metrics     Get performance metrics
GET    /api/v1/portfolios/:id/performance/twr         Calculate TWR
GET    /api/v1/portfolios/:id/performance/mwr         Calculate MWR
GET    /api/v1/portfolios/:id/performance/benchmark   Compare to benchmark
POST   /api/v1/portfolios/:id/performance/report      Generate report
```

### Tax Lots
```
GET    /api/v1/portfolios/:id/tax-lots            List all tax lots
GET    /api/v1/portfolios/:id/tax-lots/:symbol    List tax lots for symbol
GET    /api/v1/portfolios/:id/tax-lots/harvest    Get tax-loss harvest opportunities
POST   /api/v1/portfolios/:id/tax-lots/report     Generate tax report
```

### Export
```
GET    /api/v1/portfolios/:id/export/csv          Export portfolio to CSV
GET    /api/v1/portfolios/:id/export/transactions Export transactions to CSV
GET    /api/v1/portfolios/:id/export/holdings     Export holdings to CSV
POST   /api/v1/portfolios/:id/export/report       Generate and export report
```

---

## CSV Import Specification

### Standard CSV Format

The system should support a standard CSV format with the following columns:

```csv
Date,Type,Symbol,Quantity,Price,Commission,Currency,Notes
2024-01-15,BUY,AAPL,10,180.50,1.00,USD,Initial purchase
2024-02-20,BUY,MSFT,5,310.25,1.00,USD,
2024-03-15,DIVIDEND,AAPL,,0.25,,USD,Quarterly dividend
2024-06-15,SELL,AAPL,5,195.50,1.00,USD,Partial sale
2024-08-28,SPLIT,AAPL,4:1,,,USD,Stock split
```

### Broker-Specific Parsers

The system should include parsers for common broker export formats:

- **Fidelity**
- **Charles Schwab**
- **TD Ameritrade**
- **E*TRADE**
- **Interactive Brokers**
- **Robinhood**

Each parser should:
1. Map broker columns to standard format
2. Handle broker-specific transaction types
3. Validate and normalize data
4. Generate import report with errors/warnings

### Import Validation

The import process should validate:
- Date format and range
- Symbol format (ticker symbols)
- Quantity > 0 for buy/sell transactions
- Price > 0 for buy/sell transactions
- Valid transaction type
- Valid currency code
- No duplicate transactions

### Import Response

```json
{
  "success": true,
  "total_rows": 100,
  "imported": 98,
  "errors": 2,
  "warnings": 5,
  "batch_id": "abc123",
  "details": {
    "errors": [
      {
        "row": 15,
        "message": "Invalid symbol format: AAPL.XYZ"
      }
    ],
    "warnings": [
      {
        "row": 42,
        "message": "Missing commission, defaulting to 0.00"
      }
    ]
  }
}
```

---

## Corporate Actions Processing

### Stock Split Processing

When a stock split occurs:

1. **Adjust Share Quantity:** Multiply all holdings by split ratio
2. **Adjust Cost Basis:** Divide cost basis per share by split ratio
3. **Update Tax Lots:** Adjust each tax lot's quantity and cost per share
4. **Create Split Transaction:** Record the split in transaction history
5. **Maintain Total Cost Basis:** Ensure total cost basis remains unchanged

Example: 4-for-1 split of AAPL
- Before: 100 shares @ $180/share = $18,000 total cost
- After: 400 shares @ $45/share = $18,000 total cost

### Dividend Processing

#### Cash Dividend
1. Record dividend payment
2. Update cash balance (if tracked)
3. Do not affect share count or cost basis

#### Dividend Reinvestment (DRIP)
1. Record dividend payment
2. Calculate shares purchased (dividend amount / share price)
3. Create buy transaction for reinvested shares
4. Update holdings and cost basis
5. Create new tax lot

### Merger/Acquisition Processing

When Company A is acquired by Company B:

1. **Close Position in A:** Mark all A shares as sold/converted
2. **Open Position in B:** Create new position based on merger ratio
3. **Transfer Cost Basis:** Cost basis from A transfers to B
4. **Handle Cash Portion:** If merger includes cash, record appropriately
5. **Update Tax Lots:** Convert A tax lots to B lots with preserved acquisition dates

### Spinoff Processing

When Company A spins off division as Company B:

1. **Maintain A Position:** Original shares remain
2. **Create B Position:** New shares based on spinoff ratio
3. **Allocate Cost Basis:** Split original cost basis between A and B based on fair market value
4. **Create Tax Lots:** New tax lots for B with original acquisition date from A
5. **Record Transaction:** Create spinoff transaction

---

## Performance Calculations

### Total Return

**Absolute Return:**
```
Total Return = Current Value - Total Cost Basis
```

**Percentage Return:**
```
Total Return % = (Current Value - Total Cost Basis) / Total Cost Basis × 100
```

### Time-Weighted Return (TWR)

TWR removes the effect of cash flows, showing the true investment performance:

```
TWR = [(1 + R1) × (1 + R2) × ... × (1 + Rn)] - 1

where Ri = (Ending Value - Beginning Value - Cash Flow) / (Beginning Value + Cash Flow)
```

Algorithm:
1. Divide time period into sub-periods at each cash flow
2. Calculate return for each sub-period
3. Compound the sub-period returns
4. Annualize if period > 1 year

### Money-Weighted Return (MWR/IRR)

MWR accounts for the timing and size of cash flows:

```
0 = NPV = Σ(CFt / (1 + IRR)^t) + (Ending Value / (1 + IRR)^n)

where:
- CFt = cash flow at time t (negative for purchases, positive for sales)
- n = total number of periods
- IRR = internal rate of return (solve iteratively)
```

Use Newton-Raphson method to solve for IRR.

### Annualized Return

```
Annualized Return = (1 + Total Return)^(365/days) - 1

where days = number of days in measurement period
```

### Benchmark Comparison

```
Alpha = Portfolio Return - Benchmark Return

Relative Return = (Portfolio Return / Benchmark Return - 1) × 100
```

---

## Technical Implementation Requirements

### 1. Database Requirements

**Indexes:**
- User email (unique)
- Portfolio user_id
- Transaction portfolio_id, symbol, date
- Holding portfolio_id, symbol
- TaxLot portfolio_id, symbol
- CorporateAction symbol, date

**Constraints:**
- Foreign key relationships with CASCADE delete for portfolios
- CHECK constraints for positive quantities and prices
- Unique constraints on portfolio name per user

### 2. API Design Principles

- RESTful endpoints with proper HTTP verbs
- JWT authentication required for all endpoints except auth
- Request validation with meaningful error messages
- Pagination for list endpoints (default 50, max 500)
- Rate limiting (100 requests per minute per user)
- API versioning (/api/v1/)

### 3. Error Handling

Standard error response format:
```json
{
  "error": {
    "code": "INVALID_TRANSACTION",
    "message": "Transaction date cannot be in the future",
    "details": {
      "field": "date",
      "value": "2025-12-31"
    }
  }
}
```

Error codes:
- `AUTHENTICATION_REQUIRED` (401)
- `INSUFFICIENT_PERMISSIONS` (403)
- `RESOURCE_NOT_FOUND` (404)
- `VALIDATION_ERROR` (400)
- `DUPLICATE_RESOURCE` (409)
- `INTERNAL_SERVER_ERROR` (500)

### 4. Performance Considerations

**Caching Strategy:**
- Cache market quotes (5-minute TTL during market hours, 1-hour after close)
- Cache performance calculations (15-minute TTL)
- Cache user portfolios list (5-minute TTL)
- Invalidate cache on write operations

**Optimization:**
- Use database indexes on frequently queried columns
- Batch market data API calls
- Precompute daily performance snapshots
- Use read replicas for analytics queries
- Implement connection pooling

### 5. Market Data Integration

**Data Sources:**
- Primary: Alpha Vantage, Yahoo Finance, or IEX Cloud
- Backup: Secondary provider for failover
- Historical data: CSV import for backtesting

**Update Frequency:**
- Real-time quotes: 5-minute delayed (free tier) or real-time (paid)
- Daily updates: End-of-day prices, corporate actions
- Corporate actions: Daily scan for all held symbols

### 6. Background Jobs

**Scheduled Tasks:**
- Daily EOD price updates (after market close)
- Corporate action detection (daily)
- Performance snapshot generation (daily)
- Stale data cleanup (weekly)
- Email notifications (as needed)

**Queue System:**
- Use job queue for CSV imports
- Background processing for large reports
- Async corporate action processing

---

## Security Requirements

### Authentication
- JWT tokens with 24-hour expiration
- Refresh tokens with 30-day expiration
- Secure password hashing (bcrypt, cost factor 12)
- Password requirements: min 12 chars, mixed case, numbers, special chars

### Authorization
- Users can only access their own portfolios
- Admin role for system management
- API key support for CLI tool

### Data Protection
- HTTPS only in production
- Database encryption at rest
- Sensitive data (tokens) not logged
- SQL injection prevention (parameterized queries)
- XSS prevention (input sanitization)

### Rate Limiting
- 100 requests/minute per user (general)
- 10 requests/minute for auth endpoints
- 1000 requests/hour for market data

---

## Testing Requirements

### Unit Tests
- Test all service layer functions
- Test calculation accuracy (performance, cost basis)
- Test corporate action processing
- Test CSV parsing and validation
- Target: >80% code coverage

### Integration Tests
- Test API endpoints end-to-end
- Test database transactions
- Test authentication flows
- Test CSV import pipeline

### Performance Tests
- Load test API endpoints (1000 concurrent users)
- Test large portfolio performance (10,000 transactions)
- Test bulk import (10,000 row CSV)

---

## Deployment Requirements

### Environment Variables
```bash
# Database
DATABASE_URL=postgresql://user:pass@host:5432/portfolios

# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
GIN_MODE=release

# JWT
JWT_SECRET=<32+ character secret>
JWT_EXPIRATION=24h
REFRESH_TOKEN_EXPIRATION=720h

# Market Data
MARKET_DATA_API_KEY=<api-key>
MARKET_DATA_PROVIDER=alphavantage

# CORS
CORS_ALLOWED_ORIGINS=https://app.portfolios.com

# Email (for notifications)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=<email>
SMTP_PASSWORD=<password>

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# Background Jobs
ENABLE_BACKGROUND_JOBS=true
EOD_UPDATE_TIME=18:00
```

### Docker Deployment
- Multi-stage Docker build for small image size
- Docker Compose for development
- Production deployment to cloud (AWS, GCP, Azure)
- Database migrations on container start
- Health check endpoint: /health

### Monitoring
- Structured JSON logging
- Health check endpoint
- Metrics export (Prometheus format)
- Error tracking (Sentry integration)

---

## Future Enhancements (Post-MVP)

### Phase 2 Features
- Real-time WebSocket updates for portfolio values
- Mobile app (React Native)
- Options and futures support
- Cryptocurrency portfolio tracking
- Automated broker integration (Plaid)

### Phase 3 Features
- Social features (share portfolios, leaderboards)
- Portfolio recommendations using ML
- Risk analysis and optimization
- Robo-advisor features
- Tax optimization suggestions

---

## Success Metrics

### Technical Metrics
- API response time < 200ms (p95)
- System uptime > 99.9%
- Database query time < 50ms (p95)
- CLI command execution < 500ms

### Business Metrics
- User registration rate
- Active users (DAU/MAU)
- Average portfolios per user
- Average transactions per portfolio
- CSV import success rate
- User retention (30-day, 90-day)

---

## Glossary

**Cost Basis:** The original purchase price of a security, including commissions, used to calculate capital gains/losses.

**FIFO (First In, First Out):** Cost basis method where the first shares purchased are the first sold.

**LIFO (Last In, First Out):** Cost basis method where the most recently purchased shares are sold first.

**TWR (Time-Weighted Return):** Return calculation that removes the effect of cash flows to measure pure investment performance.

**MWR (Money-Weighted Return):** Return calculation that accounts for timing and size of cash flows, similar to IRR.

**Tax Lot:** A record of a specific purchase of shares, including date, quantity, and cost basis.

**Corporate Action:** An event initiated by a company that affects its shareholders (split, dividend, merger, spinoff).

**Stock Split:** Corporate action that increases share count and proportionally decreases share price.

**Dividend:** Payment made by corporation to shareholders, either in cash or additional shares.

**Spinoff:** When a company creates a new independent company by distributing shares to existing shareholders.

**IRR (Internal Rate of Return):** The discount rate that makes NPV of all cash flows equal to zero.

---

## Appendix: Example CLI Workflows

### Workflow 1: New User Setup
```bash
# Register
portfolios auth register --email john@example.com --password SecurePass123!

# Login
portfolios auth login --email john@example.com --password SecurePass123!

# Create first portfolio
portfolios portfolio create --name "Retirement 401k" --currency USD

# Import transactions
portfolios transaction import 1 --file fidelity-export.csv --broker fidelity

# Check holdings
portfolios portfolio holdings 1

# View performance
portfolios performance metrics 1
```

### Workflow 2: Daily Monitoring
```bash
# Check all portfolios
portfolios portfolio list

# View holdings for main portfolio
portfolios portfolio holdings 1

# Check today's performance
portfolios performance metrics 1 --from today

# Get latest quotes
portfolios market quote AAPL MSFT GOOGL
```

### Workflow 3: Tax Preparation
```bash
# Generate tax report for 2024
portfolios taxlot report 1 --year 2024 --output tax-2024.pdf

# Export transactions for CPA
portfolios export transactions 1 --from 2024-01-01 --to 2024-12-31 --output transactions-2024.csv

# Check tax-loss harvest opportunities
portfolios taxlot harvest 1 --threshold -5.0
```

### Workflow 4: Corporate Action Handling
```bash
# Auto-detect and apply corporate actions
portfolios corpaction auto-apply 1 --from 2024-01-01

# Manually apply a stock split
portfolios corpaction split 1 --symbol NVDA --ratio 10:1 --date 2024-06-10

# List recent corporate actions
portfolios corpaction list --symbol AAPL --from 2024-01-01
```

---

*This specification focuses on CLI and server-side implementation. Web UI specifications are maintained separately.*

**Version:** 1.0
**Last Updated:** 2025-11-11
**Status:** Initial Draft for Implementation

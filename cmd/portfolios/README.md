# Portfolios CLI

A powerful command-line interface for managing your investment portfolios.

## Features

- 🔐 **Authentication** - Secure login with JWT tokens
- 📊 **Portfolio Management** - Create, view, and manage multiple portfolios
- 💰 **Transaction Tracking** - Record buys, sells, and dividends
- 📈 **Performance Analytics** - TWR, MWR, annualized returns, and benchmarking
- 📥 **CSV Import** - Import transactions from major brokers (Fidelity, Schwab, TD Ameritrade, E*TRADE, Interactive Brokers, Robinhood)
- 🎨 **Beautiful UI** - Interactive TUI with styled tables and colorful output
- 📤 **Multiple Output Formats** - Table, JSON, or CSV output

## Installation

### Build from source

```bash
go build -o portfolios ./cmd/portfolios
sudo mv portfolios /usr/local/bin/
```

### Using Make (if available)

```bash
make build
make install
```

## Quick Start

### 1. Register a new account

```bash
portfolios auth register
```

### 2. Login

```bash
portfolios auth login
```

### 3. Create your first portfolio

```bash
portfolios portfolio create
```

### 4. Add a transaction

```bash
portfolios transaction add <portfolio-id>
```

### 5. View portfolio performance

```bash
portfolios performance show <portfolio-id>
```

## Command Reference

### Authentication

```bash
# Register a new account
portfolios auth register

# Login
portfolios auth login

# Logout
portfolios auth logout

# Show current user
portfolios auth whoami
```

### Portfolio Management

```bash
# List all portfolios
portfolios portfolio list
portfolios p ls                    # Short alias

# Create a new portfolio
portfolios portfolio create

# Show portfolio details
portfolios portfolio show <id>

# Show portfolio holdings
portfolios portfolio holdings <id>
portfolios p hold <id>             # Short alias

# Delete a portfolio
portfolios portfolio delete <id>
portfolios p rm <id>               # Short alias

# Interactive portfolio selector
portfolios portfolio select
portfolios interactive             # Alias
```

### Transaction Management

```bash
# List transactions
portfolios transaction list <portfolio-id>
portfolios tx ls <portfolio-id>    # Short alias

# Add a transaction manually
portfolios transaction add <portfolio-id> --type buy

# Import from CSV
portfolios transaction import <portfolio-id> transactions.csv --broker fidelity

# Import with dry-run (validation only)
portfolios transaction import <portfolio-id> transactions.csv --broker schwab --dry-run

# Supported brokers:
# - generic (standard CSV format)
# - fidelity
# - schwab
# - tdameritrade
# - etrade
# - interactivebrokers
# - robinhood

# List import batches
portfolios transaction batches <portfolio-id>

# Delete a transaction
portfolios transaction delete <portfolio-id> <transaction-id>

# Delete an import batch
portfolios transaction delete-batch <portfolio-id> <batch-id>
```

### Performance Analytics

```bash
# Show performance metrics
portfolios performance show <portfolio-id>
portfolios perf show <portfolio-id>   # Short alias

# Show performance for date range
portfolios performance show <portfolio-id> --start 2024-01-01 --end 2024-12-31

# Compare two portfolios
portfolios performance compare <id1> <id2>

# Compare against benchmark (e.g., S&P 500)
portfolios performance benchmark <portfolio-id> SPY

# View performance snapshots
portfolios performance snapshots <portfolio-id>
```

### Configuration

```bash
# Show current configuration
portfolios config show

# Set API base URL
portfolios config set api_base_url http://localhost:8080

# Set output format
portfolios config set output_format json    # table, json, or csv

# Show config file path
portfolios config path
```

### Other Commands

```bash
# Show version information
portfolios version

# Show help
portfolios --help
portfolios <command> --help
```

## Output Formats

The CLI supports three output formats:

### Table (default)

Beautiful formatted tables with colors:

```bash
portfolios portfolio list
# or
portfolios portfolio list --output table
```

### JSON

Machine-readable JSON output:

```bash
portfolios portfolio list --output json
portfolios portfolio list -o json
```

### CSV

CSV format for spreadsheet import:

```bash
portfolios portfolio list --output csv
```

## Global Flags

These flags work with all commands:

- `--api-url <url>` - Override the API base URL
- `--output <format>` or `-o <format>` - Output format (table, json, csv)
- `--config <path>` - Custom config file path

## Configuration

The CLI stores configuration in `~/.portfolios/config.yaml`:

```yaml
api_base_url: http://localhost:8080
output_format: table
access_token: <your-token>
refresh_token: <your-refresh-token>
```

## Interactive Mode

The CLI includes an interactive portfolio selector:

```bash
portfolios interactive
# or
portfolios portfolio select
```

Use arrow keys (↑/↓) or j/k to navigate, Enter to select, and q to quit.

## CSV Import Format

### Generic CSV Format

```csv
date,type,symbol,quantity,price,commission,notes
2024-01-15,buy,AAPL,100,150.50,0,Initial purchase
2024-02-20,sell,AAPL,50,155.75,0,Partial sale
2024-03-10,dividend,AAPL,0,0.25,0,Quarterly dividend
```

### Broker-Specific Formats

The CLI automatically detects and parses formats from:
- **Fidelity** - Trade confirmation exports
- **Schwab** - Transaction history CSV
- **TD Ameritrade** - Account history
- **E*TRADE** - Transaction downloads
- **Interactive Brokers** - Activity statement CSV
- **Robinhood** - Transaction export

## Examples

### Complete Workflow Example

```bash
# 1. Register and login
portfolios auth register
portfolios auth login

# 2. Create a portfolio
portfolios portfolio create
# Enter name: "Tech Stocks"
# Enter description: "Technology sector investments"

# 3. Import transactions from broker
portfolios transaction import 1 fidelity-export.csv --broker fidelity

# 4. View holdings
portfolios portfolio holdings 1

# 5. Check performance
portfolios performance show 1

# 6. Compare to S&P 500
portfolios performance benchmark 1 SPY

# 7. Export data
portfolios transaction list 1 --output csv > transactions.csv
```

### Batch Operations

```bash
# Import multiple CSV files
portfolios tx import 1 jan-2024.csv --broker schwab
portfolios tx import 1 feb-2024.csv --broker schwab
portfolios tx import 1 mar-2024.csv --broker schwab

# List all import batches
portfolios tx batches 1

# Remove an incorrect import batch
portfolios tx delete-batch 1 <batch-id>
```

## Tips and Tricks

### 1. Use Aliases

The CLI supports short aliases for common commands:

- `portfolios p` = `portfolios portfolio`
- `portfolios tx` = `portfolios transaction`
- `portfolios perf` = `portfolios performance`

### 2. JSON Output for Scripting

Use JSON output for automation:

```bash
# Get portfolio ID programmatically
ID=$(portfolios portfolio list -o json | jq '.[0].id')
portfolios performance show $ID
```

### 3. Configure Default Output Format

Set your preferred output format permanently:

```bash
portfolios config set output_format json
```

### 4. Use Dry-Run for Validation

Always validate CSV imports first:

```bash
portfolios tx import 1 data.csv --broker fidelity --dry-run
```

## Troubleshooting

### Authentication Issues

If you get authentication errors:

```bash
# Check current user
portfolios auth whoami

# Re-login
portfolios auth logout
portfolios auth login
```

### API Connection Issues

If the CLI can't connect to the API:

```bash
# Check current API URL
portfolios config show

# Update API URL
portfolios config set api_base_url http://your-api-server:8080
```

### Config File Issues

If config is corrupted:

```bash
# Show config path
portfolios config path

# Delete and recreate
rm ~/.portfolios/config.yaml
portfolios version  # Will recreate config
```

## Development

### Build

```bash
go build -o portfolios ./cmd/portfolios
```

### Build with Version Info

```bash
VERSION=1.0.0
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD)

go build -ldflags "\
  -X 'github.com/lenon/portfolios/cmd/portfolios/cmd.Version=${VERSION}' \
  -X 'github.com/lenon/portfolios/cmd/portfolios/cmd.BuildDate=${BUILD_DATE}' \
  -X 'github.com/lenon/portfolios/cmd/portfolios/cmd.GitCommit=${GIT_COMMIT}'" \
  -o portfolios ./cmd/portfolios
```

### Run Tests

```bash
go test ./cmd/portfolios/...
go test ./internal/cli/...
```

## License

See the main project LICENSE file.

## Contributing

Contributions are welcome! Please see the main project CONTRIBUTING.md.

## Support

For issues, questions, or feature requests, please open an issue on GitHub.

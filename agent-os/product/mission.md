# Product Mission

## Pitch
Portfolios is a web-based portfolio management platform that helps retail investors track and analyze their stock market investments by providing a centralized, real-time view of all their portfolios with automatic corporate action handling and comprehensive performance analytics.

## Users

### Primary Customers
- **Individual Retail Investors**: People managing their own stock portfolios who need better visibility into performance across multiple accounts
- **Active Traders**: Users who frequently buy and sell securities and need accurate, up-to-date performance tracking
- **Multi-Portfolio Managers**: Investors who manage separate portfolios for different goals (retirement, growth, income) and want unified monitoring

### User Personas

**Sarah - The Organized Investor** (35-45)
- **Role:** Professional with multiple investment accounts
- **Context:** Manages 3-4 portfolios across different brokers (retirement, taxable, experimental trading account)
- **Pain Points:** Struggles to see total portfolio performance across all accounts; manually tracks transactions in spreadsheets; misses corporate actions like stock splits affecting cost basis
- **Goals:** Consolidate all portfolio data in one place; see real-time total portfolio value; track historical performance accurately

**Michael - The Active Trader** (28-40)
- **Role:** Tech-savvy investor who trades weekly
- **Context:** Actively manages portfolio with regular buy/sell transactions; follows market closely
- **Pain Points:** Current broker platforms don't provide adequate performance analytics; needs to manually calculate returns; wants to track performance of different trading strategies
- **Goals:** Quick transaction entry; real-time portfolio monitoring; detailed performance metrics to evaluate trading decisions

**Linda - The Long-Term Planner** (50-65)
- **Role:** Retirement-focused investor
- **Context:** Building wealth for retirement with dividend-paying stocks
- **Context:** Multiple accounts (401k rollovers, IRAs, taxable accounts)
- **Pain Points:** Loses track of dividend reinvestments; doesn't understand true portfolio cost basis after stock splits; can't easily see total retirement savings across accounts
- **Goals:** Accurate dividend tracking; automatic corporate action handling; clear visualization of retirement portfolio growth over time

## The Problem

### Fragmented Portfolio Visibility
Retail investors with multiple brokerage accounts lack a unified view of their total investment performance. Each broker provides their own reporting, but investors can't see the big picture. Third-party solutions like Sharesight exist but require broker integrations that many users find invasive or don't support their brokers.

**Our Solution:** Manual transaction import (CSV) and entry gives users complete control over their data without requiring broker API access, while still providing comprehensive portfolio analytics.

### Inaccurate Performance Tracking
Corporate actions like stock splits, mergers, and dividend reinvestments complicate portfolio tracking. Most investors manually track these in spreadsheets, leading to errors in cost basis calculations and performance metrics.

**Our Solution:** Automatic corporate action detection and application ensures portfolio data remains accurate without manual intervention.

### No Real-Time Monitoring
Investors want to see how their portfolios perform during market hours, but checking multiple broker websites is tedious and doesn't provide consolidated views.

**Our Solution:** Live portfolio monitoring aggregates real-time market data to show current portfolio values and daily changes across all holdings.

## Differentiators

### Privacy-First Data Control
Unlike Sharesight and other competitors that require broker API integrations, Portfolios uses manual transaction import. This gives users complete control over their data without sharing broker credentials or API access. Users choose what data to import and when.

This results in increased user trust and works with any broker, including international brokers not supported by API-based solutions.

### Automatic Corporate Action Handling
Unlike spreadsheet tracking or basic portfolio apps, Portfolios automatically detects and applies corporate actions (stock splits, mergers, spinoffs, dividend reinvestments). This eliminates manual cost basis adjustments and ensures accurate performance calculations.

This results in significant time savings and improved accuracy for long-term investors who experience multiple corporate actions over time.

### Multi-Portfolio Architecture
Unlike single-portfolio tools, Portfolios is designed from the ground up for users managing multiple portfolios. Track retirement accounts separately from taxable accounts, or separate conservative holdings from speculative trades.

This results in better organization and more meaningful performance analytics tailored to each portfolio's investment strategy.

## Key Features

### Core Features
- **CSV Transaction Import:** Bulk import buy/sell transactions from broker statements with intelligent parsing and validation
- **Manual Transaction Entry:** Quick UI-based transaction entry for users who prefer direct input or have only a few transactions
- **Real-Time Portfolio Valuation:** Live market data integration to show current portfolio values and intraday changes
- **Performance Analytics:** Comprehensive metrics including total return, time-weighted return, money-weighted return, and annualized returns
- **Holdings Dashboard:** Clear visualization of current holdings with cost basis, current value, gain/loss, and allocation percentages

### Portfolio Management Features
- **Multiple Portfolio Support:** Create and manage unlimited portfolios with separate performance tracking for each
- **Portfolio Comparison:** Side-by-side comparison of portfolio performance to evaluate different investment strategies
- **Historical Performance Charts:** Visual representation of portfolio growth over time with configurable date ranges
- **Transaction History:** Complete audit trail of all buy/sell transactions with filtering and search capabilities

### Corporate Action Features
- **Automatic Stock Split Detection:** System detects stock splits and automatically adjusts share quantities and cost basis
- **Dividend Tracking:** Record cash dividends and dividend reinvestments with proper cost basis adjustments
- **Merger and Acquisition Handling:** Support for stock mergers, acquisitions, and ticker symbol changes
- **Spinoff Support:** Handle corporate spinoffs where shareholders receive new company shares

### Advanced Features
- **Tax Lot Tracking:** Track individual purchase lots for accurate cost basis reporting and tax-loss harvesting opportunities
- **Currency Support:** Handle multi-currency portfolios for international investments
- **Benchmark Comparison:** Compare portfolio performance against market indices (S&P 500, NASDAQ, etc.)
- **Export and Reporting:** Generate PDF reports and CSV exports for tax preparation or external analysis

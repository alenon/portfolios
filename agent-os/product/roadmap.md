# Product Roadmap

1. [ ] User Authentication & Authorization — Implement secure user registration, login, JWT-based authentication, and user session management with password reset functionality. Each user should have isolated access to only their portfolios. `M`

2. [ ] Portfolio CRUD Operations — Build complete portfolio management with create, read, update, delete operations. Users can create multiple portfolios, name them, and view a list of all their portfolios with basic metadata. `S`

3. [ ] Manual Transaction Entry — Develop UI and backend API for manually entering buy/sell transactions with fields for ticker symbol, date, quantity, price, and transaction type. Include form validation and immediate reflection in portfolio holdings. `M`

4. [ ] Holdings Calculation Engine — Implement core engine that calculates current holdings from transaction history. Calculate total shares owned per security, average cost basis, and support multiple purchases of the same security. Must handle sells that reduce position sizes. `L`

5. [ ] Real-Time Price Integration — Integrate with a market data API (Alpha Vantage, Twelve Data, or similar) to fetch current stock prices. Display real-time portfolio valuation and daily gain/loss for each holding and total portfolio. `M`

6. [ ] Holdings Dashboard — Create comprehensive portfolio view showing all current holdings with columns for ticker, shares, cost basis, current value, gain/loss (dollar and percent), and allocation percentage. Include portfolio total value and total gain/loss summary. `M`

7. [ ] Transaction History View — Build UI to display complete transaction history with filtering by portfolio, date range, ticker, and transaction type. Include search functionality and pagination for large transaction lists. `S`

8. [ ] CSV Transaction Import — Develop CSV parser and import workflow that allows users to upload transaction files. Include column mapping UI, data validation, error reporting, and preview before final import. Support common broker CSV formats. `L`

9. [ ] Stock Split Detection & Application — Build automated system to detect historical stock splits using market data API. Automatically adjust historical transaction quantities and prices when splits are detected. Display split notifications to users. `L`

10. [ ] Dividend Tracking — Add dividend recording functionality supporting both cash dividends and dividend reinvestments (DRIP). Include dividend history view and total dividend income reporting per portfolio and per security. `M`

11. [ ] Performance Metrics Calculation — Implement comprehensive return calculations including total return, time-weighted return (TWR), money-weighted return (MWR/IRR), and annualized returns. Display metrics on portfolio dashboard with explanatory tooltips. `L`

12. [ ] Historical Performance Charts — Create interactive charts showing portfolio value over time. Support multiple time ranges (1M, 3M, 6M, 1Y, ALL). Include option to overlay multiple portfolios for comparison. `M`

13. [ ] Tax Lot Tracking — Implement individual tax lot tracking for each purchase. Support FIFO, LIFO, and specific lot identification methods for sell transactions. Display unrealized gains per lot for tax-loss harvesting opportunities. `L`

14. [ ] Merger & Acquisition Handling — Add support for corporate mergers and acquisitions. Allow users to record when one security is exchanged for another, handling both stock-for-stock and cash-plus-stock deals. Maintain accurate cost basis through corporate actions. `M`

15. [ ] Portfolio Comparison View — Build side-by-side portfolio comparison showing relative performance metrics. Include visual comparisons of returns, allocation differences, and growth trajectories. Useful for evaluating different investment strategies. `M`

16. [ ] Benchmark Comparison — Integrate market index data (S&P 500, NASDAQ, Russell 2000) and add benchmark comparison to portfolio performance. Show portfolio performance relative to selected benchmark over various time periods. `S`

17. [ ] Multi-Currency Support — Extend system to support international stocks with foreign currencies. Implement currency conversion, display holdings in user's home currency, and track currency gain/loss separately from investment gain/loss. `L`

18. [ ] Export & Reporting — Build PDF and CSV export functionality for portfolios, transactions, and performance reports. Include customizable report templates for tax preparation and external analysis. `M`

19. [ ] Spinoff Support — Add functionality to handle corporate spinoffs where shareholders receive shares in a newly created company. Allocate original cost basis between parent and spinoff shares according to IRS guidelines or user-specified allocation. `M`

20. [ ] CLI Administrative Interface — Develop command-line interface for administrative tasks such as user management, database migrations, bulk data operations, and system health monitoring. Include data backup/restore commands. `S`

> Notes
> - Items ordered by technical dependencies and architectural considerations
> - Core foundation (auth, portfolios, transactions) must be built before advanced features
> - Holdings calculation engine is central to all subsequent features
> - Real-time pricing enables many user-facing features but can be implemented after basic holdings work
> - Corporate actions (splits, dividends, mergers, spinoffs) build on the holdings engine
> - Each item represents an end-to-end (frontend + backend) functional and testable feature
> - Performance metrics require complete transaction history and holdings data
> - Advanced features (tax lots, multi-currency, benchmarks) can be added after core MVP is stable

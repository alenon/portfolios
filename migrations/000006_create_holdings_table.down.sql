-- Drop holdings table and related objects
DROP TRIGGER IF EXISTS update_holdings_updated_at ON holdings;
DROP INDEX IF EXISTS idx_holdings_updated_at;
DROP INDEX IF EXISTS idx_holdings_symbol;
DROP INDEX IF EXISTS idx_holdings_portfolio_id;
DROP INDEX IF EXISTS idx_holdings_portfolio_symbol;
DROP TABLE IF EXISTS holdings;

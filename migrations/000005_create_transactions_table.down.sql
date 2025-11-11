-- Drop transactions table and related objects
DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;
DROP INDEX IF EXISTS idx_transactions_import_batch;
DROP INDEX IF EXISTS idx_transactions_portfolio_date;
DROP INDEX IF EXISTS idx_transactions_portfolio_symbol;
DROP INDEX IF EXISTS idx_transactions_date;
DROP INDEX IF EXISTS idx_transactions_symbol;
DROP INDEX IF EXISTS idx_transactions_portfolio_id;
DROP TABLE IF EXISTS transactions;

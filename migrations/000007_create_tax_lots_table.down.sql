-- Drop tax_lots table and related objects
DROP TRIGGER IF EXISTS update_tax_lots_updated_at ON tax_lots;
DROP INDEX IF EXISTS idx_tax_lots_transaction_id;
DROP INDEX IF EXISTS idx_tax_lots_purchase_date;
DROP INDEX IF EXISTS idx_tax_lots_portfolio_symbol;
DROP INDEX IF EXISTS idx_tax_lots_symbol;
DROP INDEX IF EXISTS idx_tax_lots_portfolio_id;
DROP TABLE IF EXISTS tax_lots;

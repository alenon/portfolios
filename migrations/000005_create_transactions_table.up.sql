-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    date TIMESTAMP NOT NULL,
    quantity NUMERIC(20, 8) NOT NULL,
    price NUMERIC(20, 8),
    commission NUMERIC(20, 8) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    notes TEXT,
    import_batch_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_transaction_type CHECK (type IN ('BUY', 'SELL', 'DIVIDEND', 'SPLIT', 'MERGER', 'SPINOFF', 'DIVIDEND_REINVEST')),
    CONSTRAINT chk_quantity_positive CHECK (quantity > 0),
    CONSTRAINT chk_price_non_negative CHECK (price IS NULL OR price >= 0),
    CONSTRAINT chk_commission_non_negative CHECK (commission >= 0)
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_transactions_portfolio_id ON transactions(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_transactions_symbol ON transactions(symbol);
CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_portfolio_symbol ON transactions(portfolio_id, symbol);
CREATE INDEX IF NOT EXISTS idx_transactions_portfolio_date ON transactions(portfolio_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_import_batch ON transactions(import_batch_id);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_transactions_updated_at
    BEFORE UPDATE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

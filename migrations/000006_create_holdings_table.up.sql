-- Create holdings table
CREATE TABLE IF NOT EXISTS holdings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    quantity NUMERIC(20, 8) NOT NULL,
    cost_basis NUMERIC(20, 8) NOT NULL,
    avg_cost_price NUMERIC(20, 8) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_holding_quantity CHECK (quantity >= 0),
    CONSTRAINT chk_holding_cost_basis CHECK (cost_basis >= 0),
    CONSTRAINT chk_holding_avg_cost CHECK (avg_cost_price >= 0)
);

-- Create unique index to prevent duplicate holdings
CREATE UNIQUE INDEX IF NOT EXISTS idx_holdings_portfolio_symbol ON holdings(portfolio_id, symbol);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_holdings_portfolio_id ON holdings(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_holdings_symbol ON holdings(symbol);
CREATE INDEX IF NOT EXISTS idx_holdings_updated_at ON holdings(updated_at DESC);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_holdings_updated_at
    BEFORE UPDATE ON holdings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

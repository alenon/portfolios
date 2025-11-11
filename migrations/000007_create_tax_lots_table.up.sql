-- Create tax_lots table
CREATE TABLE IF NOT EXISTS tax_lots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    purchase_date TIMESTAMP NOT NULL,
    quantity NUMERIC(20, 8) NOT NULL,
    cost_basis NUMERIC(20, 8) NOT NULL,
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_tax_lot_quantity CHECK (quantity >= 0),
    CONSTRAINT chk_tax_lot_cost_basis CHECK (cost_basis >= 0)
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_tax_lots_portfolio_id ON tax_lots(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_tax_lots_symbol ON tax_lots(symbol);
CREATE INDEX IF NOT EXISTS idx_tax_lots_portfolio_symbol ON tax_lots(portfolio_id, symbol);
CREATE INDEX IF NOT EXISTS idx_tax_lots_purchase_date ON tax_lots(purchase_date);
CREATE INDEX IF NOT EXISTS idx_tax_lots_transaction_id ON tax_lots(transaction_id);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_tax_lots_updated_at
    BEFORE UPDATE ON tax_lots
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

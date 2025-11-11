-- Create portfolios table
CREATE TABLE IF NOT EXISTS portfolios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    base_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    cost_basis_method VARCHAR(20) NOT NULL DEFAULT 'FIFO',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_cost_basis_method CHECK (cost_basis_method IN ('FIFO', 'LIFO', 'SPECIFIC_LOT'))
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_portfolios_user_id ON portfolios(user_id);
CREATE INDEX IF NOT EXISTS idx_portfolios_created_at ON portfolios(created_at DESC);

-- Add unique constraint on portfolio name per user
CREATE UNIQUE INDEX IF NOT EXISTS idx_portfolios_user_name ON portfolios(user_id, name);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_portfolios_updated_at
    BEFORE UPDATE ON portfolios
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

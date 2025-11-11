-- Create corporate_actions table
CREATE TABLE IF NOT EXISTS corporate_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(20) NOT NULL,
    type VARCHAR(20) NOT NULL,
    date TIMESTAMP NOT NULL,
    ratio NUMERIC(20, 8),
    amount NUMERIC(20, 8),
    new_symbol VARCHAR(20),
    currency VARCHAR(3),
    description TEXT,
    applied BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_corporate_action_type CHECK (type IN ('SPLIT', 'DIVIDEND', 'MERGER', 'SPINOFF', 'TICKER_CHANGE'))
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_corporate_actions_symbol ON corporate_actions(symbol);
CREATE INDEX IF NOT EXISTS idx_corporate_actions_date ON corporate_actions(date DESC);
CREATE INDEX IF NOT EXISTS idx_corporate_actions_type ON corporate_actions(type);
CREATE INDEX IF NOT EXISTS idx_corporate_actions_applied ON corporate_actions(applied);
CREATE INDEX IF NOT EXISTS idx_corporate_actions_symbol_date ON corporate_actions(symbol, date DESC);

-- Create performance_snapshots table
CREATE TABLE IF NOT EXISTS performance_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    date TIMESTAMP NOT NULL,
    total_value NUMERIC(20, 8) NOT NULL,
    total_cost_basis NUMERIC(20, 8) NOT NULL,
    total_return NUMERIC(20, 8) NOT NULL,
    total_return_pct NUMERIC(10, 4) NOT NULL,
    day_change NUMERIC(20, 8),
    day_change_pct NUMERIC(10, 4),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create unique index to prevent duplicate snapshots for same portfolio and date
CREATE UNIQUE INDEX IF NOT EXISTS idx_performance_snapshots_portfolio_date ON performance_snapshots(portfolio_id, date);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_performance_snapshots_portfolio_id ON performance_snapshots(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_performance_snapshots_date ON performance_snapshots(date DESC);
CREATE INDEX IF NOT EXISTS idx_performance_snapshots_created_at ON performance_snapshots(created_at DESC);

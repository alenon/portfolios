-- Drop performance_snapshots table and related objects
DROP INDEX IF EXISTS idx_performance_snapshots_created_at;
DROP INDEX IF EXISTS idx_performance_snapshots_date;
DROP INDEX IF EXISTS idx_performance_snapshots_portfolio_id;
DROP INDEX IF EXISTS idx_performance_snapshots_portfolio_date;
DROP TABLE IF EXISTS performance_snapshots;

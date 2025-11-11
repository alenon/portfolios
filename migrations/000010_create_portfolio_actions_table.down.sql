-- Drop portfolio_actions table
DROP INDEX IF EXISTS idx_portfolio_actions_unique_pending;
DROP INDEX IF EXISTS idx_portfolio_actions_detected_at;
DROP INDEX IF EXISTS idx_portfolio_actions_portfolio_status;
DROP INDEX IF EXISTS idx_portfolio_actions_status;
DROP INDEX IF EXISTS idx_portfolio_actions_corporate_action_id;
DROP INDEX IF EXISTS idx_portfolio_actions_portfolio_id;
DROP TABLE IF EXISTS portfolio_actions;

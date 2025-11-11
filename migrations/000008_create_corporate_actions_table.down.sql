-- Drop corporate_actions table and related objects
DROP INDEX IF EXISTS idx_corporate_actions_symbol_date;
DROP INDEX IF EXISTS idx_corporate_actions_applied;
DROP INDEX IF EXISTS idx_corporate_actions_type;
DROP INDEX IF EXISTS idx_corporate_actions_date;
DROP INDEX IF EXISTS idx_corporate_actions_symbol;
DROP TABLE IF EXISTS corporate_actions;

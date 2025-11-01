-- Drop indexes
DROP INDEX IF EXISTS idx_refresh_tokens_token_hash;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;

-- Drop refresh_tokens table
DROP TABLE IF EXISTS refresh_tokens;

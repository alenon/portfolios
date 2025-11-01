-- Drop indexes
DROP INDEX IF EXISTS idx_password_reset_tokens_user_id;
DROP INDEX IF EXISTS idx_password_reset_tokens_token_hash;

-- Drop password_reset_tokens table
DROP TABLE IF EXISTS password_reset_tokens;

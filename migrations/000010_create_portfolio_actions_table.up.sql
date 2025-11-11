-- Create portfolio_actions table
CREATE TABLE IF NOT EXISTS portfolio_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    corporate_action_id UUID NOT NULL REFERENCES corporate_actions(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    affected_symbol VARCHAR(20) NOT NULL,
    shares_affected BIGINT NOT NULL,
    detected_at TIMESTAMP NOT NULL,
    reviewed_at TIMESTAMP,
    applied_at TIMESTAMP,
    reviewed_by_user_id UUID REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_portfolio_action_status CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED', 'APPLIED'))
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_portfolio_actions_portfolio_id ON portfolio_actions(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_portfolio_actions_corporate_action_id ON portfolio_actions(corporate_action_id);
CREATE INDEX IF NOT EXISTS idx_portfolio_actions_status ON portfolio_actions(status);
CREATE INDEX IF NOT EXISTS idx_portfolio_actions_portfolio_status ON portfolio_actions(portfolio_id, status);
CREATE INDEX IF NOT EXISTS idx_portfolio_actions_detected_at ON portfolio_actions(detected_at DESC);

-- Create unique constraint to prevent duplicate pending actions for the same portfolio and corporate action
CREATE UNIQUE INDEX IF NOT EXISTS idx_portfolio_actions_unique_pending
    ON portfolio_actions(portfolio_id, corporate_action_id)
    WHERE status = 'PENDING';

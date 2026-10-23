-- Migration: 0010_investment_valuation_history.up.sql
-- Description: Create investment_valuation_history table for tracking valuations over time

CREATE TABLE IF NOT EXISTS investment_valuation_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investment_id UUID NOT NULL REFERENCES investments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    valuation BIGINT NOT NULL CHECK (valuation >= 0),
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_val_history_inv_created
    ON investment_valuation_history(investment_id, created_at ASC);

CREATE INDEX IF NOT EXISTS idx_val_history_user_created
    ON investment_valuation_history(user_id, created_at ASC);

-- Backfill initial valuation for existing investments
INSERT INTO investment_valuation_history (investment_id, user_id, valuation, notes, created_at)
SELECT id, user_id, current_value, 'Initial valuation baseline', created_at
FROM investments
WHERE deleted_at IS NULL;

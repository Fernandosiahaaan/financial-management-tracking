-- Migration: 0005_cycle.up.sql
-- Description: Ensure cycle_start_day constraints and index on user_settings

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_settings' AND column_name = 'cycle_start_day'
    ) THEN
        ALTER TABLE user_settings ADD COLUMN cycle_start_day SMALLINT NOT NULL DEFAULT 1 CHECK (cycle_start_day >= 1 AND cycle_start_day <= 31);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_user_settings_cycle_start ON user_settings(user_id, cycle_start_day);

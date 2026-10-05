-- Migration: 0005_cycle.down.sql
-- Description: Drop index for cycle_start_day

DROP INDEX IF EXISTS idx_user_settings_cycle_start;

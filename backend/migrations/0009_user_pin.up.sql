-- 0009_user_pin.up.sql
-- Add pin_hash to users table for database-backed PIN authentication

ALTER TABLE users ADD COLUMN IF NOT EXISTS pin_hash TEXT;

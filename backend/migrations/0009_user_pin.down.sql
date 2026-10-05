-- 0009_user_pin.down.sql
-- Remove pin_hash from users table

ALTER TABLE users DROP COLUMN IF EXISTS pin_hash;

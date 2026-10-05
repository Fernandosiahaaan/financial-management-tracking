-- Migration: 0007_receivables_investments.down.sql
-- Description: Drop tables created in Phase 08

DROP TABLE IF EXISTS receivable_payments CASCADE;
DROP TABLE IF EXISTS receivables CASCADE;
DROP TABLE IF EXISTS investments CASCADE;

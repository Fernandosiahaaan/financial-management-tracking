-- Migration: 0007_receivables_investments.up.sql
-- Description: Create receivables, receivable_payments, and investments tables for Phase 08

-- Receivables table: tracking money lent to counterparties
CREATE TABLE IF NOT EXISTS receivables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    source_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    counterparty VARCHAR(255) NOT NULL,
    principal BIGINT NOT NULL CHECK (principal > 0),
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'PARTIALLY_PAID', 'PAID', 'OVERDUE', 'WRITTEN_OFF')),
    due_date DATE,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_receivables_user_status
    ON receivables(user_id, status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_receivables_user
    ON receivables(user_id)
    WHERE deleted_at IS NULL;

-- Receivable payments: tracking partial and full repayments
CREATE TABLE IF NOT EXISTS receivable_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    receivable_id UUID NOT NULL REFERENCES receivables(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    target_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    amount BIGINT NOT NULL CHECK (amount > 0),
    payment_date DATE NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_receivable_payments_receivable
    ON receivable_payments(receivable_id);

CREATE INDEX IF NOT EXISTS idx_receivable_payments_user_date
    ON receivable_payments(user_id, payment_date DESC);

-- Investments table: tracking investment capital and manual valuations
CREATE TABLE IF NOT EXISTS investments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    source_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    type VARCHAR(50) NOT NULL CHECK (type IN ('STOCK', 'MUTUAL_FUND', 'GOLD', 'CRYPTO', 'OTHER')),
    name VARCHAR(255) NOT NULL,
    capital BIGINT NOT NULL CHECK (capital >= 0),
    current_value BIGINT NOT NULL CHECK (current_value >= 0),
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_investments_user_type
    ON investments(user_id, type)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_investments_user
    ON investments(user_id)
    WHERE deleted_at IS NULL;

-- Migration: 0006_budgets_allocations.up.sql
-- Description: Create budgets and allocations tables for Phase 06

-- Budgets: planned spending limits per category per cycle
CREATE TABLE IF NOT EXISTS budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    cycle_start DATE NOT NULL,
    cycle_end DATE NOT NULL,
    planned_amount BIGINT NOT NULL CHECK (planned_amount >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_budget_cycle_dates CHECK (cycle_start < cycle_end)
);

-- One active budget per user/category/cycle
CREATE UNIQUE INDEX IF NOT EXISTS idx_budgets_user_category_cycle
    ON budgets(user_id, category_id, cycle_start, cycle_end)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_budgets_user_cycle
    ON budgets(user_id, cycle_start, cycle_end)
    WHERE deleted_at IS NULL;

-- Allocations: income distribution to purpose categories per cycle
CREATE TABLE IF NOT EXISTS allocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    cycle_start DATE NOT NULL,
    cycle_end DATE NOT NULL,
    allocated_amount BIGINT NOT NULL CHECK (allocated_amount >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_allocation_cycle_dates CHECK (cycle_start < cycle_end)
);

-- One active allocation per user/category/cycle
CREATE UNIQUE INDEX IF NOT EXISTS idx_allocations_user_category_cycle
    ON allocations(user_id, category_id, cycle_start, cycle_end)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_allocations_user_cycle
    ON allocations(user_id, cycle_start, cycle_end)
    WHERE deleted_at IS NULL;

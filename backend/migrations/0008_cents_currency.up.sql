-- Convert monetary columns from whole Rupiah to Cents / Sen (x100)
-- 1 IDR = 100 Sen (Cents)

UPDATE accounts SET 
    opening_balance = opening_balance * 100,
    current_balance = current_balance * 100;

UPDATE transactions SET 
    amount = amount * 100;

UPDATE budgets SET 
    planned_amount = planned_amount * 100;

UPDATE allocations SET 
    allocated_amount = allocated_amount * 100;

UPDATE receivables SET 
    principal = principal * 100;

UPDATE receivable_payments SET 
    amount = amount * 100;

UPDATE investments SET 
    capital = capital * 100,
    current_value = current_value * 100;

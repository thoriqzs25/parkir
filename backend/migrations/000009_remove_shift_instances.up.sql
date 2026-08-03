-- Remove shift instances, keep only shift configs (templates)
-- Shifts are now computed from check_in_at + location_shift_configs

-- Drop foreign key constraints first
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS sessions_shift_id_fkey;
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_shift_id_fkey;

-- Drop indexes on shift_id
DROP INDEX IF EXISTS idx_sessions_shift;
DROP INDEX IF EXISTS idx_transactions_shift;

-- Change sessions.shift_id (UUID) to shift_number (INTEGER)
ALTER TABLE sessions DROP COLUMN shift_id;
ALTER TABLE sessions ADD COLUMN shift_number INTEGER;

CREATE INDEX idx_sessions_shift_number ON sessions (location_id, shift_number);

-- Change transactions.shift_id (UUID) to shift_number (INTEGER)
ALTER TABLE transactions DROP COLUMN shift_id;
ALTER TABLE transactions ADD COLUMN shift_number INTEGER;

CREATE INDEX idx_transactions_shift_number ON transactions (location_id, shift_number);

-- Drop the shifts table entirely
DROP TABLE IF EXISTS shifts;

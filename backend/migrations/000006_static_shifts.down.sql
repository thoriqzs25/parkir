-- Revert static shift management changes

-- Remove new columns from shifts
ALTER TABLE shifts DROP COLUMN IF EXISTS shift_number;
ALTER TABLE shifts DROP COLUMN IF EXISTS shift_date;
ALTER TABLE shifts DROP COLUMN IF EXISTS void_count;
ALTER TABLE shifts DROP COLUMN IF EXISTS incident_count;
ALTER TABLE shifts DROP CONSTRAINT IF EXISTS unique_location_shift_date;

-- Restore old columns (nullable for compatibility, data is gone anyway)
ALTER TABLE shifts ADD COLUMN operator_id UUID REFERENCES users(id);
ALTER TABLE shifts ADD COLUMN status VARCHAR(20) DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'CLOSED', 'FLAGGED', 'FORCE_CLOSED'));
ALTER TABLE shifts ADD COLUMN started_at TIMESTAMPTZ;
ALTER TABLE shifts ADD COLUMN ended_at TIMESTAMPTZ;
ALTER TABLE shifts ADD COLUMN expected_cash NUMERIC(12,2);
ALTER TABLE shifts ADD COLUMN cash_handover_amount NUMERIC(12,2);
ALTER TABLE shifts ADD COLUMN discrepancy NUMERIC(12,2);
ALTER TABLE shifts ADD COLUMN discrepancy_notes TEXT;
ALTER TABLE shifts ADD COLUMN force_closed_by UUID REFERENCES users(id);
ALTER TABLE shifts ADD COLUMN force_closed_reason TEXT;

-- Recreate old indexes
CREATE INDEX idx_shifts_operator_status ON shifts (operator_id, status);
CREATE INDEX idx_shifts_started_at ON shifts (started_at DESC);
DROP INDEX IF EXISTS idx_shifts_date;

-- Drop shift configs table
DROP TRIGGER IF EXISTS trg_check_shift_config_overlap ON location_shift_configs;
DROP FUNCTION IF EXISTS check_shift_config_overlap();
DROP TABLE IF EXISTS location_shift_configs;

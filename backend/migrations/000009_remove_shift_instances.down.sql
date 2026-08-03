-- Restore shift instances table

-- Recreate shifts table
CREATE TABLE shifts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id     UUID NOT NULL REFERENCES locations(id),
    shift_number    INTEGER NOT NULL,
    shift_date      DATE NOT NULL,
    void_count      INTEGER DEFAULT 0,
    incident_count  INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE(location_id, shift_number, shift_date)
);

CREATE INDEX idx_shifts_date ON shifts(shift_date DESC);
CREATE INDEX idx_shifts_location ON shifts(location_id);

-- Drop new columns
DROP INDEX IF EXISTS idx_sessions_shift_number;
DROP INDEX IF EXISTS idx_transactions_shift_number;

ALTER TABLE sessions DROP COLUMN shift_id;
ALTER TABLE sessions ADD COLUMN shift_id UUID REFERENCES shifts(id);
ALTER TABLE transactions DROP COLUMN shift_id;
ALTER TABLE transactions ADD COLUMN shift_id UUID NOT NULL REFERENCES shifts(id);

CREATE INDEX idx_sessions_shift ON sessions (shift_id);
CREATE INDEX idx_transactions_shift ON transactions (shift_id);

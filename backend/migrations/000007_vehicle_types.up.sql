-- Customizable vehicle types
-- Replaces hardcoded CHECK constraints with FK references to vehicle_types table

CREATE TABLE vehicle_types (
    name         VARCHAR(20) PRIMARY KEY,
    display_name VARCHAR(100) NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Seed default vehicle types (match existing CHECK constraint values)
INSERT INTO vehicle_types (name, display_name, description) VALUES
    ('CAR', 'Car', 'Four-wheeled motor vehicle'),
    ('MOTO', 'Motorcycle', 'Two-wheeled motor vehicle'),
    ('TRUCK', 'Truck', 'Large cargo vehicle');

-- Update trigger function to update updated_at
CREATE OR REPLACE FUNCTION update_vehicle_types_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_vehicle_types_updated_at
    BEFORE UPDATE ON vehicle_types
    FOR EACH ROW
    EXECUTE FUNCTION update_vehicle_types_updated_at();

-- location_rates: drop CHECK, add FK
ALTER TABLE location_rates DROP CONSTRAINT IF EXISTS location_rates_vehicle_type_check;
ALTER TABLE location_rates ADD CONSTRAINT fk_location_rates_vehicle_type
    FOREIGN KEY (vehicle_type) REFERENCES vehicle_types(name) ON DELETE RESTRICT;

-- sessions: drop CHECK, add FK
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS sessions_vehicle_type_check;
ALTER TABLE sessions ADD CONSTRAINT fk_sessions_vehicle_type
    FOREIGN KEY (vehicle_type) REFERENCES vehicle_types(name) ON DELETE RESTRICT;

-- transactions: drop CHECK, add FK
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_vehicle_type_check;
ALTER TABLE transactions ADD CONSTRAINT fk_transactions_vehicle_type
    FOREIGN KEY (vehicle_type) REFERENCES vehicle_types(name) ON DELETE RESTRICT;

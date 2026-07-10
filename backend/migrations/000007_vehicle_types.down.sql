-- Revert vehicle types customization
-- Must clean up any non-standard vehicle types before restoring CHECK constraints

-- Remove rows that reference custom vehicle types (not CAR, MOTO, TRUCK)
DELETE FROM transactions WHERE vehicle_type NOT IN ('CAR', 'MOTO', 'TRUCK');
DELETE FROM sessions WHERE vehicle_type NOT IN ('CAR', 'MOTO', 'TRUCK');
DELETE FROM location_rates WHERE vehicle_type NOT IN ('CAR', 'MOTO', 'TRUCK');
DELETE FROM vehicle_types WHERE name NOT IN ('CAR', 'MOTO', 'TRUCK');

-- Drop FK constraints
ALTER TABLE location_rates DROP CONSTRAINT IF EXISTS fk_location_rates_vehicle_type;
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS fk_sessions_vehicle_type;
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS fk_transactions_vehicle_type;

-- Restore CHECK constraints
ALTER TABLE location_rates ADD CONSTRAINT location_rates_vehicle_type_check
    CHECK (vehicle_type IN ('CAR', 'MOTO', 'TRUCK'));
ALTER TABLE sessions ADD CONSTRAINT sessions_vehicle_type_check
    CHECK (vehicle_type IN ('CAR', 'MOTO', 'TRUCK'));
ALTER TABLE transactions ADD CONSTRAINT transactions_vehicle_type_check
    CHECK (vehicle_type IN ('CAR', 'MOTO', 'TRUCK'));

-- Drop vehicle_types table
DROP TRIGGER IF EXISTS trg_vehicle_types_updated_at ON vehicle_types;
DROP FUNCTION IF EXISTS update_vehicle_types_updated_at();
DROP TABLE IF EXISTS vehicle_types;

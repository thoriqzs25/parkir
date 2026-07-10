CREATE TABLE gates (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id     VARCHAR(64) UNIQUE NOT NULL,
    name          VARCHAR(100) NOT NULL DEFAULT '',
    location_id   UUID REFERENCES locations(id) ON DELETE SET NULL,
    ip_address    VARCHAR(45) NOT NULL DEFAULT '',
    last_seen_at  TIMESTAMPTZ,
    registered_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_gates_device_id ON gates (device_id);
CREATE INDEX idx_gates_location ON gates (location_id);

CREATE OR REPLACE FUNCTION update_gates_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_gates_updated_at
    BEFORE UPDATE ON gates
    FOR EACH ROW
    EXECUTE FUNCTION update_gates_updated_at();

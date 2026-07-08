-- Add soft delete to roles
ALTER TABLE roles ADD COLUMN deleted_at TIMESTAMPTZ;

-- Physical parking locations
CREATE TABLE locations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(150) NOT NULL,
    code            VARCHAR(20) UNIQUE NOT NULL,
    address         TEXT,
    city            VARCHAR(100),
    status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                        CHECK (status IN ('ACTIVE', 'INACTIVE')),
    capacity        JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Rate configuration per location and vehicle type
CREATE TABLE location_rates (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id             UUID NOT NULL REFERENCES locations(id),
    vehicle_type            VARCHAR(10) NOT NULL
                                CHECK (vehicle_type IN ('CAR', 'MOTO', 'TRUCK')),
    first_hour_rate         NUMERIC(12,2) NOT NULL,
    subsequent_hourly_rate  NUMERIC(12,2) NOT NULL,
    daily_flat_rate         NUMERIC(12,2) NOT NULL,
    effective_from          DATE NOT NULL,
    effective_until         DATE,
    created_by              UUID REFERENCES users(id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT no_overlap UNIQUE (location_id, vehicle_type, effective_from)
);

CREATE INDEX idx_rates_location_type ON location_rates (location_id, vehicle_type);

-- Function to prevent overlapping effective date ranges for rates
CREATE OR REPLACE FUNCTION check_rate_overlap()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM location_rates
        WHERE location_id = NEW.location_id
          AND vehicle_type = NEW.vehicle_type
          AND id != NEW.id
          AND (
              NEW.effective_from <= COALESCE(effective_until, '9999-12-31'::date)
              AND COALESCE(NEW.effective_until, '9999-12-31'::date) >= effective_from
          )
    ) THEN
        RAISE EXCEPTION 'Overlapping rate effective dates for location_id=%, vehicle_type=%', NEW.location_id, NEW.vehicle_type;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_check_rate_overlap
BEFORE INSERT OR UPDATE ON location_rates
FOR EACH ROW
EXECUTE FUNCTION check_rate_overlap();

-- Role applies at these locations
CREATE TABLE user_role_locations (
    user_id         UUID NOT NULL REFERENCES users(id),
    location_id     UUID NOT NULL REFERENCES locations(id),
    PRIMARY KEY (user_id, location_id)
);

-- Independent permission grants (additive)
CREATE TABLE user_permission_grants (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    location_id     UUID REFERENCES locations(id),
    permission      VARCHAR(100) NOT NULL,
    granted_by      UUID NOT NULL REFERENCES users(id),
    granted_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ,
    revoked_at      TIMESTAMPTZ,
    revoked_by      UUID REFERENCES users(id),
    UNIQUE (user_id, location_id, permission)
);

CREATE INDEX idx_grants_user ON user_permission_grants (user_id);
CREATE INDEX idx_grants_location ON user_permission_grants (location_id);
CREATE INDEX idx_grants_permission ON user_permission_grants (permission);
CREATE INDEX idx_grants_active ON user_permission_grants (user_id, location_id)
    WHERE revoked_at IS NULL;

-- Immutable audit trail
CREATE TABLE audit_logs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action        VARCHAR(100) NOT NULL,
    actor_id      UUID REFERENCES users(id),
    actor_role    VARCHAR(50),
    entity_type   VARCHAR(50) NOT NULL,
    entity_id     UUID NOT NULL,
    location_id   UUID REFERENCES locations(id),
    ip_address    INET,
    metadata      JSONB,
    timestamp     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_actor ON audit_logs (actor_id);
CREATE INDEX idx_audit_logs_entity ON audit_logs (entity_type, entity_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs (timestamp);
CREATE INDEX idx_audit_logs_action ON audit_logs (action);

-- Revoke UPDATE/DELETE on audit_logs from application user (append-only enforcement)
-- NOTE: Run this as a superuser or in a separate privileged migration.
-- DO $$
-- BEGIN
--     EXECUTE 'REVOKE UPDATE, DELETE ON audit_logs FROM app_user';
-- EXCEPTION WHEN undefined_object THEN
--     NULL;
-- END $$;

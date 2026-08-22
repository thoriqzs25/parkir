-- PARKIR v2 — Flattened Initial Schema
-- Based on PRD v2 Automated Gate System

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- ROLES & USERS
-- ============================================

CREATE TABLE roles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) UNIQUE NOT NULL,
    permissions     JSONB NOT NULL DEFAULT '[]',
    created_tz      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_tz      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id         UUID NOT NULL REFERENCES roles(id),
    name            VARCHAR(100) NOT NULL,
    username        VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                        CHECK (status IN ('ACTIVE', 'DEACTIVATED')),
    created_tz      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_tz      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_role ON users (role_id);
CREATE INDEX idx_users_status ON users (status);

-- ============================================
-- LOCATIONS
-- ============================================

CREATE TABLE locations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(150) NOT NULL,
    code            VARCHAR(20) UNIQUE NOT NULL,
    address         TEXT,
    city            VARCHAR(100),
    status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                        CHECK (status IN ('ACTIVE', 'DRAFT', 'INACTIVE')),
    created_tz      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_tz      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_locations_city ON locations (city);

-- ============================================
-- GATES
-- ============================================

CREATE TABLE gates (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id         UUID NOT NULL REFERENCES locations(id),
    gate_id             VARCHAR(100) UNIQUE NOT NULL,
    type                VARCHAR(10) NOT NULL CHECK (type IN ('ENTRY', 'EXIT')),
    vehicle_type        VARCHAR(10) NOT NULL CHECK (vehicle_type IN ('CAR', 'MOTO', 'TRUCK')),
    status              VARCHAR(20) NOT NULL DEFAULT 'UNREGISTERED'
                            CHECK (status IN ('UNREGISTERED', 'REGISTERED', 'OPERATIONAL')),
    config              JSONB,
    branch_service_url  VARCHAR(255),
    presence_tz         TIMESTAMPTZ,
    created_tz          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_tz          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_gates_location ON gates (location_id);
CREATE INDEX idx_gates_status ON gates (status);

-- ============================================
-- RATE CONFIGS (versioned, integer)
-- ============================================

CREATE TABLE rate_configs (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id             UUID NOT NULL REFERENCES locations(id),
    vehicle_type            VARCHAR(10) NOT NULL CHECK (vehicle_type IN ('CAR', 'MOTO', 'TRUCK')),
    version                 VARCHAR(50) NOT NULL,
    first_hour_rate         INTEGER NOT NULL,
    subsequent_hourly_rate  INTEGER NOT NULL,
    daily_flat_rate         INTEGER NOT NULL,
    effective_from          DATE NOT NULL,
    effective_until         DATE,
    created_tz              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_rate_configs_location_type ON rate_configs (location_id, vehicle_type);
CREATE INDEX idx_rate_configs_version ON rate_configs (version);

CREATE OR REPLACE FUNCTION check_rate_overlap()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM rate_configs
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
BEFORE INSERT OR UPDATE ON rate_configs
FOR EACH ROW
EXECUTE FUNCTION check_rate_overlap();

-- ============================================
-- SHIFT CONFIGS (versioned shift definitions)
-- ============================================

CREATE TABLE shift_configs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id     UUID NOT NULL REFERENCES locations(id),
    version         VARCHAR(50) NOT NULL,
    shift_code      VARCHAR(20) NOT NULL,
    shift_number    INTEGER NOT NULL,
    start_time      TIME NOT NULL,
    end_time        TIME NOT NULL,
    is_overnight    BOOLEAN NOT NULL DEFAULT false,
    created_tz      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_shift_configs_location ON shift_configs (location_id);
CREATE INDEX idx_shift_configs_version ON shift_configs (version);

CREATE OR REPLACE FUNCTION check_shift_config_overlap()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM shift_configs
        WHERE location_id = NEW.location_id
          AND id != NEW.id
          AND (
              (NEW.start_time < end_time AND NEW.end_time > start_time)
              OR (NEW.is_overnight OR is_overnight)
          )
    ) THEN
        RAISE EXCEPTION 'Overlapping shift configuration for location_id=%', NEW.location_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_check_shift_config_overlap
BEFORE INSERT OR UPDATE ON shift_configs
FOR EACH ROW
EXECUTE FUNCTION check_shift_config_overlap();

-- ============================================
-- SESSIONS
-- ============================================

CREATE TABLE sessions (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id             UUID NOT NULL REFERENCES locations(id),
    gate_id                 VARCHAR(100) NOT NULL,
    vehicle_type            VARCHAR(10) NOT NULL CHECK (vehicle_type IN ('CAR', 'MOTO', 'TRUCK')),
    check_in_tz             TIMESTAMPTZ NOT NULL DEFAULT now(),
    check_out_tz            TIMESTAMPTZ,
    fee_amount              INTEGER,
    rate_config_version     VARCHAR(50) NOT NULL,
    shift_config_version    VARCHAR(50) NOT NULL,
    shift_number            INTEGER NOT NULL,
    state                   VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                                CHECK (state IN ('ACTIVE', 'PENDING_PAYMENT', 'CLOSED', 'VOIDED')),
    qr_data                 TEXT NOT NULL,
    created_tz              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_tz              TIMESTAMPTZ NOT NULL DEFAULT now(),
    synced_tz               TIMESTAMPTZ
);

CREATE INDEX idx_sessions_location ON sessions (location_id);
CREATE INDEX idx_sessions_state ON sessions (state);
CREATE INDEX idx_sessions_check_in ON sessions (check_in_tz DESC);
CREATE INDEX idx_sessions_synced ON sessions (synced_tz);

-- ============================================
-- TRANSACTIONS (1:1 with sessions)
-- ============================================

CREATE TABLE transactions (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id              UUID UNIQUE NOT NULL REFERENCES sessions(id),
    location_id             UUID NOT NULL REFERENCES locations(id),
    gate_id                 VARCHAR(100) NOT NULL,
    amount                  INTEGER NOT NULL,
    payment_method          VARCHAR(20) NOT NULL
                                CHECK (payment_method IN ('EMONEY', 'FLAZZ', 'CASH', 'OFFLINE_SOP')),
    payment_reference       VARCHAR(100),
    rate_config_version     VARCHAR(50) NOT NULL,
    shift_config_version    VARCHAR(50) NOT NULL,
    shift_number            INTEGER NOT NULL,
    created_tz              TIMESTAMPTZ NOT NULL DEFAULT now(),
    synced_tz               TIMESTAMPTZ,
    voided                  BOOLEAN NOT NULL DEFAULT false,
    voided_tz               TIMESTAMPTZ,
    voided_by               UUID REFERENCES users(id),
    void_reason             TEXT
);

CREATE INDEX idx_transactions_session ON transactions (session_id);
CREATE INDEX idx_transactions_location ON transactions (location_id);
CREATE INDEX idx_transactions_created ON transactions (created_tz DESC);
CREATE INDEX idx_transactions_synced ON transactions (synced_tz);
CREATE INDEX idx_transactions_voided ON transactions (voided);

-- ============================================
-- INCIDENTS
-- ============================================

CREATE TABLE incidents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id     UUID NOT NULL REFERENCES locations(id),
    gate_id         VARCHAR(100),
    type            VARCHAR(50) NOT NULL CHECK (type IN (
                        'HARDWARE_FAILURE', 'PAYMENT_FAILURE', 'QR_UNREADABLE',
                        'PRINTER_JAM', 'GATE_MOTOR_FAILURE', 'OTHER'
                    )),
    state           VARCHAR(20) NOT NULL DEFAULT 'OPEN'
                        CHECK (state IN ('OPEN', 'IN_PROGRESS', 'RESOLVED')),
    description     TEXT NOT NULL,
    resolved_tz     TIMESTAMPTZ,
    resolved_by     UUID REFERENCES users(id),
    created_tz      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_tz      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_incidents_location ON incidents (location_id);
CREATE INDEX idx_incidents_state ON incidents (state);

-- ============================================
-- AUDIT LOGS (immutable trail)
-- ============================================

CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID REFERENCES users(id),
    location_id     UUID REFERENCES locations(id),
    action          VARCHAR(100) NOT NULL,
    entity_type     VARCHAR(50) NOT NULL,
    entity_id       UUID,
    metadata        JSONB,
    ip_address      VARCHAR(45),
    timestamp_tz    TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_tz      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_user ON audit_logs (user_id);
CREATE INDEX idx_audit_logs_entity ON audit_logs (entity_type, entity_id);
CREATE INDEX idx_audit_logs_location ON audit_logs (location_id);
CREATE INDEX idx_audit_logs_created ON audit_logs (created_tz DESC);

-- ============================================
-- UPDATE TRIGGERS (auto-update updated_tz)
-- ============================================

CREATE OR REPLACE FUNCTION update_updated_tz()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_tz = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_roles_updated_tz
    BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_tz();

CREATE TRIGGER trg_users_updated_tz
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_tz();

CREATE TRIGGER trg_locations_updated_tz
    BEFORE UPDATE ON locations
    FOR EACH ROW EXECUTE FUNCTION update_updated_tz();

CREATE TRIGGER trg_gates_updated_tz
    BEFORE UPDATE ON gates
    FOR EACH ROW EXECUTE FUNCTION update_updated_tz();

CREATE TRIGGER trg_sessions_updated_tz
    BEFORE UPDATE ON sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_tz();

CREATE TRIGGER trg_incidents_updated_tz
    BEFORE UPDATE ON incidents
    FOR EACH ROW EXECUTE FUNCTION update_updated_tz();

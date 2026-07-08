-- Operator work periods
CREATE TABLE shifts (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id             UUID NOT NULL REFERENCES users(id),
    location_id             UUID NOT NULL REFERENCES locations(id),
    status                  VARCHAR(20) NOT NULL DEFAULT 'OPEN'
                                CHECK (status IN ('OPEN', 'CLOSED', 'FLAGGED', 'FORCE_CLOSED')),
    started_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at                TIMESTAMPTZ,
    expected_cash           NUMERIC(12,2),
    cash_handover_amount    NUMERIC(12,2),
    discrepancy             NUMERIC(12,2),
    discrepancy_notes       TEXT,
    force_closed_by         UUID REFERENCES users(id),
    force_closed_reason     TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_shifts_operator_status ON shifts (operator_id, status);
CREATE INDEX idx_shifts_location ON shifts (location_id);
CREATE INDEX idx_shifts_started_at ON shifts (started_at DESC);

-- Parking sessions
CREATE TABLE sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id     UUID NOT NULL REFERENCES locations(id),
    operator_id     UUID NOT NULL REFERENCES users(id),
    shift_id        UUID REFERENCES shifts(id),
    plate           VARCHAR(20) NOT NULL,
    city_code       VARCHAR(10) NOT NULL DEFAULT 'UNKNOWN',
    vehicle_type    VARCHAR(10) NOT NULL
                        CHECK (vehicle_type IN ('CAR', 'MOTO', 'TRUCK')),
    state           VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                        CHECK (state IN ('ACTIVE', 'PENDING_PAYMENT', 'CLOSED', 'VOIDED')),
    check_in_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    check_out_at    TIMESTAMPTZ,
    fee_amount      NUMERIC(12,2),
    rate_snapshot   JSONB,
    offline_sync    BOOLEAN NOT NULL DEFAULT false,
    sync_conflict   BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_location_state ON sessions (location_id, state);
CREATE INDEX idx_sessions_plate_state ON sessions (plate, state);
CREATE INDEX idx_sessions_operator ON sessions (operator_id);
CREATE INDEX idx_sessions_shift ON sessions (shift_id);
CREATE INDEX idx_sessions_check_in_at ON sessions (check_in_at DESC);

-- Payment records
CREATE TABLE transactions (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id              UUID UNIQUE NOT NULL REFERENCES sessions(id),
    location_id             UUID NOT NULL REFERENCES locations(id),
    shift_id                UUID NOT NULL REFERENCES shifts(id),
    operator_id             UUID NOT NULL REFERENCES users(id),
    vehicle_type            VARCHAR(10) NOT NULL
                                CHECK (vehicle_type IN ('CAR', 'MOTO', 'TRUCK')),
    plate                   VARCHAR(20) NOT NULL,
    check_in_at             TIMESTAMPTZ NOT NULL,
    check_out_at            TIMESTAMPTZ NOT NULL,
    duration_hours          INTEGER NOT NULL,
    rate_first_hour         NUMERIC(12,2) NOT NULL,
    rate_subsequent_hourly  NUMERIC(12,2) NOT NULL,
    rate_daily              NUMERIC(12,2) NOT NULL,
    fee_amount              NUMERIC(12,2) NOT NULL,
    payment_method          VARCHAR(10) NOT NULL
                                CHECK (payment_method IN ('CASH', 'DIGITAL')),
    amount_tendered         NUMERIC(12,2),
    change_amount           NUMERIC(12,2),
    payment_reference       VARCHAR(100),
    receipt_number          VARCHAR(50) UNIQUE NOT NULL,
    voided                  BOOLEAN NOT NULL DEFAULT false,
    voided_at               TIMESTAMPTZ,
    voided_by               UUID REFERENCES users(id),
    void_reason             TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_transactions_session ON transactions (session_id);
CREATE INDEX idx_transactions_receipt ON transactions (receipt_number);
CREATE INDEX idx_transactions_shift ON transactions (shift_id);
CREATE INDEX idx_transactions_location_created ON transactions (location_id, created_at DESC);

-- Per-location daily receipt sequence (race-safe with row-level locking)
CREATE TABLE receipt_sequences (
    location_id     UUID NOT NULL REFERENCES locations(id),
    sequence_date   DATE NOT NULL,
    last_number     INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (location_id, sequence_date)
);

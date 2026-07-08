-- Operational incidents
CREATE TABLE incidents (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id       UUID NOT NULL REFERENCES locations(id),
    type              VARCHAR(30) NOT NULL
                        CHECK (type IN ('STUCK_AT_GATE', 'PAYMENT_DISPUTE', 'OPERATOR_ERROR', 'SYSTEM_DOWNTIME')),
    state             VARCHAR(20) NOT NULL DEFAULT 'OPEN'
                        CHECK (state IN ('OPEN', 'IN_PROGRESS', 'RESOLVED')),
    session_id        UUID REFERENCES sessions(id),
    reported_by       UUID NOT NULL REFERENCES users(id),
    reported_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    description       TEXT NOT NULL,
    resolved_by       UUID REFERENCES users(id),
    resolved_at       TIMESTAMPTZ,
    resolution_notes  TEXT,
    adjustment_action VARCHAR(30)
                        CHECK (adjustment_action IN ('VOID_TRANSACTION', 'REASSIGN_SESSION')),
    adjustment_entity_id UUID,
    offline_sync      BOOLEAN NOT NULL DEFAULT false,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_incidents_location ON incidents (location_id);
CREATE INDEX idx_incidents_state ON incidents (state);
CREATE INDEX idx_incidents_type ON incidents (type);
CREATE INDEX idx_incidents_reported_at ON incidents (reported_at DESC);

CREATE TABLE incident_notes (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id   UUID NOT NULL REFERENCES incidents(id),
    author_id     UUID NOT NULL REFERENCES users(id),
    note          TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_incident_notes_incident ON incident_notes (incident_id);

-- Anomaly alerts
CREATE TABLE alerts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code              VARCHAR(50) NOT NULL,
    location_id       UUID REFERENCES locations(id),
    state             VARCHAR(20) NOT NULL DEFAULT 'TRIGGERED'
                        CHECK (state IN ('TRIGGERED', 'ACKNOWLEDGED', 'RESOLVED')),
    entity_type       VARCHAR(50),
    entity_id         UUID,
    triggered_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    acknowledged_by   UUID REFERENCES users(id),
    acknowledged_at   TIMESTAMPTZ,
    resolved_by       UUID REFERENCES users(id),
    resolved_at       TIMESTAMPTZ,
    resolution_notes  TEXT,
    metadata          JSONB,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_alerts_location ON alerts (location_id);
CREATE INDEX idx_alerts_state ON alerts (state);
CREATE INDEX idx_alerts_code ON alerts (code);
CREATE INDEX idx_alerts_triggered_at ON alerts (triggered_at DESC);

CREATE TABLE alert_configs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id   UUID REFERENCES locations(id),
    code          VARCHAR(50) NOT NULL,
    enabled       BOOLEAN NOT NULL DEFAULT true,
    threshold     JSONB,
    updated_by    UUID REFERENCES users(id),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (location_id, code)
);

CREATE INDEX idx_alert_configs_location ON alert_configs (location_id);

-- Composite index for audit_logs query performance
CREATE INDEX IF NOT EXISTS idx_audit_logs_location_timestamp ON audit_logs (location_id, timestamp DESC);
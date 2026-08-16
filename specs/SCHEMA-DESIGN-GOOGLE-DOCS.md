# PARKIR v2 — Schema Design for Google Docs

## Tables and Relationships

### LOCATIONS
| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | |
| name | VARCHAR(150) | NOT NULL | |
| code | VARCHAR(20) | UNIQUE, NOT NULL | |
| address | TEXT | | |
| city | VARCHAR(100) | | |
| status | VARCHAR(20) | DEFAULT 'ACTIVE', CHECK (ACTIVE, INACTIVE) | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

---

### GATES
| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| location_id | UUID | FK → locations(id) | |
| gate_id | VARCHAR(100) | UNIQUE, NOT NULL | Hardware serial/MAC |
| gate_type | VARCHAR(10) | CHECK (ENTRY, EXIT) | |
| vehicle_type | VARCHAR(10) | CHECK (CAR, MOTO, TRUCK, ALL) | |
| status | VARCHAR(20) | DEFAULT 'UNREGISTERED', CHECK (UNREGISTERED, REGISTERED, OPERATIONAL) | |
| config | JSONB | | Gate-specific settings |
| server_room_url | VARCHAR(255) | | Learned during bootstrap |
| last_seen_at | TIMESTAMPTZ | | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

---

### RATE_CONFIGS
| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| location_id | UUID | FK → locations(id) | |
| vehicle_type | VARCHAR(10) | CHECK (CAR, MOTO, TRUCK) | |
| version | VARCHAR(50) | NOT NULL | e.g., "rate_v42" |
| first_hour_rate | INTEGER | NOT NULL | In cents |
| subsequent_hourly_rate | INTEGER | NOT NULL | In cents |
| daily_flat_rate | INTEGER | NOT NULL | In cents |
| effective_from | DATE | NOT NULL | |
| effective_until | DATE | | Nullable |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

---

### SHIFT_CONFIGS
| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| location_id | UUID | FK → locations(id) | |
| version | VARCHAR(50) | NOT NULL | e.g., "shift_v15" |
| shift_code | VARCHAR(20) | NOT NULL | e.g., "06-14" |
| shift_number | INTEGER | NOT NULL | Display number |
| start_time | TIME | NOT NULL | |
| end_time | TIME | NOT NULL | |
| is_overnight | BOOLEAN | DEFAULT false | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

---

### SESSIONS
| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| location_id | UUID | FK → locations(id) | |
| gate_id | VARCHAR(100) | NOT NULL | Entry gate hardware ID (STRING, not FK) |
| vehicle_type | VARCHAR(10) | CHECK (CAR, MOTO, TRUCK) | |
| check_in_at | TIMESTAMPTZ | NOT NULL | |
| check_out_at | TIMESTAMPTZ | | Nullable |
| fee_amount | INTEGER | | In cents, nullable until checkout |
| rate_config_version | VARCHAR(50) | | Which rate version used |
| shift_config_version | VARCHAR(50) | | Which shift config used |
| shift_number | INTEGER | NOT NULL | Assigned at check-in |
| state | VARCHAR(20) | DEFAULT 'ACTIVE', CHECK (ACTIVE, PENDING_PAYMENT, CLOSED, VOIDED) | |
| qr_data | TEXT | NOT NULL | QR code content |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |
| synced_at | TIMESTAMPTZ | | When synced to cloud |

---

### TRANSACTIONS
| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| session_id | UUID | FK → sessions(id), UNIQUE | 1:1 with sessions |
| location_id | UUID | FK → locations(id) | |
| gate_id | VARCHAR(100) | NOT NULL | Exit gate hardware ID (STRING, not FK) |
| amount | INTEGER | NOT NULL | In cents |
| payment_method | VARCHAR(20) | CHECK (EMONEY, FLAZZ, CASH, OFFLINE_SOP) | |
| payment_reference | VARCHAR(100) | | Vendor reference |
| config_version | VARCHAR(50) | NOT NULL | Rate version used |
| shift_number | INTEGER | NOT NULL | Assigned at checkout |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| synced_at | TIMESTAMPTZ | | When synced to cloud |
| voided | BOOLEAN | DEFAULT false | |
| voided_at | TIMESTAMPTZ | | |
| voided_by | UUID | FK → users(id) | Manager who voided |
| void_reason | TEXT | | |

---

### ROLES
| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| name | VARCHAR(100) | UNIQUE, NOT NULL | |
| permissions | JSONB | DEFAULT '[]' | e.g., ["sessions:view", "gates:view"] |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

---

### USERS
| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| role_id | UUID | FK → roles(id), NOT NULL | |
| name | VARCHAR(100) | NOT NULL | |
| email | VARCHAR(255) | UNIQUE, NOT NULL | |
| password_hash | VARCHAR(255) | NOT NULL | |
| status | VARCHAR(20) | DEFAULT 'ACTIVE', CHECK (ACTIVE, DEACTIVATED) | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

---

### INCIDENTS
| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| location_id | UUID | FK → locations(id) | |
| gate_id | VARCHAR(100) | | Hardware ID if applicable |
| type | VARCHAR(50) | CHECK (HARDWARE_FAILURE, PAYMENT_FAILURE, QR_UNREADABLE, PRINTER_JAM, GATE_MOTOR_FAILURE, OTHER) | |
| state | VARCHAR(20) | DEFAULT 'OPEN', CHECK (OPEN, IN_PROGRESS, RESOLVED) | |
| description | TEXT | NOT NULL | |
| resolved_at | TIMESTAMPTZ | | |
| resolved_by | UUID | FK → users(id) | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

---

### AUDIT_LOGS
| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| user_id | UUID | FK → users(id) | Null for system actions |
| location_id | UUID | FK → locations(id) | |
| action | VARCHAR(100) | NOT NULL | e.g., "session.create" |
| entity_type | VARCHAR(50) | NOT NULL | "session", "transaction", "gate" |
| entity_id | UUID | | |
| metadata | JSONB | | Additional context |
| ip_address | VARCHAR(45) | | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

---

## Relationships

| From Table | To Table | Relationship | Foreign Key | Cardinality |
|------------|----------|--------------|-------------|-------------|
| GATES | LOCATIONS | belongs to | location_id | Many-to-One |
| RATE_CONFIGS | LOCATIONS | belongs to | location_id | Many-to-One |
| SHIFT_CONFIGS | LOCATIONS | belongs to | location_id | Many-to-One |
| SESSIONS | LOCATIONS | belongs to | location_id | Many-to-One |
| TRANSACTIONS | LOCATIONS | belongs to | location_id | Many-to-One |
| INCIDENTS | LOCATIONS | belongs to | location_id | Many-to-One |
| AUDIT_LOGS | LOCATIONS | belongs to | location_id | Many-to-One |
| USERS | ROLES | belongs to | role_id | Many-to-One |
| TRANSACTIONS | SESSIONS | belongs to | session_id | One-to-One |
| INCIDENTS | USERS | resolved by | resolved_by | Many-to-One |
| AUDIT_LOGS | USERS | performed by | user_id | Many-to-One |
| TRANSACTIONS | USERS | voided by | voided_by | Many-to-One |

---

## Key Design Decisions

### 1. gate_id is a STRING, not a UUID FK
- `gate_id` in SESSIONS and TRANSACTIONS is a VARCHAR(100) storing the hardware serial/MAC
- This is NOT a foreign key to the GATES table
- Reason: Hardware identifier assigned by device, not database
- Allows tracking sessions even if gate is not yet registered in GATES table

### 2. No operator_id in SESSIONS
- PRD v2 is fully automated (no operators at gates)
- All actions tracked via `gate_id` (hardware) and `audit_logs` (user actions)

### 3. Currency uses INTEGER (cents)
- No floating point precision issues
- Example: Rp 5,000 = 500000 (integer)
- Standard practice for financial systems

### 4. Configs are versioned
- RATE_CONFIGS has `version` field (e.g., "rate_v42")
- SHIFT_CONFIGS has `version` field (e.g., "shift_v15")
- SESSIONS and TRANSACTIONS record which config version was used
- Provides audit trail and supports offline operation

### 5. Sync tracking
- SESSIONS and TRANSACTIONS have `synced_at` field
- NULL = not yet synced to cloud
- Server Room App syncs to Cloud Backend every 1 minute

### 6. 1:1 relationship: SESSIONS ↔ TRANSACTIONS
- One transaction per closed session
- Enforced by UNIQUE constraint on `session_id` in TRANSACTIONS

---

## Data Type Reference

| Type | PostgreSQL | Usage |
|------|-----------|-------|
| UUID | UUID | Primary keys, foreign keys |
| TIMESTAMPTZ | TIMESTAMPTZ | All timestamps (stored in UTC) |
| INTEGER | INTEGER | Currency in cents |
| JSONB | JSONB | Flexible JSON data (permissions, configs) |
| VARCHAR(N) | VARCHAR(N) | Fixed max length strings |
| TEXT | TEXT | Unlimited length strings |
| BOOLEAN | BOOLEAN | True/false flags |
| DATE | DATE | Calendar dates |
| TIME | TIME | Time of day |

---

## Indexes (Recommended)

```sql
-- LOCATIONS
CREATE INDEX idx_locations_city ON locations (city);

-- GATES
CREATE INDEX idx_gates_location ON gates (location_id);
CREATE INDEX idx_gates_status ON gates (status);

-- RATE_CONFIGS
CREATE INDEX idx_rate_configs_location_type ON rate_configs (location_id, vehicle_type);
CREATE INDEX idx_rate_configs_version ON rate_configs (version);

-- SHIFT_CONFIGS
CREATE INDEX idx_shift_configs_location ON shift_configs (location_id);
CREATE INDEX idx_shift_configs_version ON shift_configs (version);

-- SESSIONS
CREATE INDEX idx_sessions_location ON sessions (location_id);
CREATE INDEX idx_sessions_state ON sessions (state);
CREATE INDEX idx_sessions_check_in ON sessions (check_in_at);
CREATE INDEX idx_sessions_synced ON sessions (synced_at);

-- TRANSACTIONS
CREATE INDEX idx_transactions_session ON transactions (session_id);
CREATE INDEX idx_transactions_location ON transactions (location_id);
CREATE INDEX idx_transactions_created ON transactions (created_at);
CREATE INDEX idx_transactions_synced ON transactions (synced_at);
CREATE INDEX idx_transactions_voided ON transactions (voided);

-- USERS
CREATE INDEX idx_users_role ON users (role_id);
CREATE INDEX idx_users_status ON users (status);

-- INCIDENTS
CREATE INDEX idx_incidents_location ON incidents (location_id);
CREATE INDEX idx_incidents_state ON incidents (state);

-- AUDIT_LOGS
CREATE INDEX idx_audit_logs_user ON audit_logs (user_id);
CREATE INDEX idx_audit_logs_entity ON audit_logs (entity_type, entity_id);
CREATE INDEX idx_audit_logs_location ON audit_logs (location_id);
CREATE INDEX idx_audit_logs_created ON audit_logs (created_at);
```

# PARKIR v2 — PostgreSQL Schema Design Diagram

## Entity Relationship Diagram (Target State per PRD v2)

```mermaid
erDiagram
    %% ==========================================
    %% CORE ENTITIES
    %% ==========================================
    
    locations {
        uuid id PK
        varchar(150) name
        varchar(20) code UK
        text address
        varchar(100) city
        varchar(20) status "ACTIVE|INACTIVE"
        timestamptz created_at
        timestamptz updated_at
    }
    
    gates {
        uuid id PK
        uuid location_id FK
        varchar(100) gate_id UK "hardware serial/MAC"
        varchar(10) gate_type "ENTRY|EXIT"
        varchar(10) vehicle_type "CAR|MOTO|TRUCK|ALL"
        varchar(20) status "UNREGISTERED|REGISTERED|OPERATIONAL"
        jsonb config
        varchar(255) server_room_url
        timestamptz last_seen_at
        timestamptz created_at
        timestamptz updated_at
    }
    
    rate_configs {
        uuid id PK
        uuid location_id FK
        varchar(10) vehicle_type "CAR|MOTO|TRUCK"
        varchar(50) version "e.g., rate_v42"
        integer first_hour_rate "in cents"
        integer subsequent_hourly_rate "in cents"
        integer daily_flat_rate "in cents"
        date effective_from
        date effective_until
        timestamptz created_at
    }
    
    shift_configs {
        uuid id PK
        uuid location_id FK
        varchar(50) version "e.g., shift_v15"
        varchar(20) shift_code "e.g., 06-14"
        integer shift_number
        time start_time
        time end_time
        boolean is_overnight
        timestamptz created_at
    }
    
    %% ==========================================
    %% USER MANAGEMENT
    %% ==========================================
    
    roles {
        uuid id PK
        varchar(100) name UK
        jsonb permissions "e.g., [sessions:view, gates:view]"
        timestamptz created_at
        timestamptz updated_at
    }
    
    users {
        uuid id PK
        uuid role_id FK
        varchar(100) name
        varchar(255) email UK
        varchar(255) password_hash
        varchar(20) status "ACTIVE|DEACTIVATED"
        timestamptz created_at
        timestamptz updated_at
    }
    
    %% ==========================================
    %% PARKING OPERATIONS
    %% ==========================================
    
    sessions {
        uuid id PK
        uuid location_id FK
        varchar(100) gate_id "entry gate hardware ID"
        varchar(10) vehicle_type "CAR|MOTO|TRUCK"
        timestamptz check_in_at
        timestamptz check_out_at
        integer fee_amount "in cents, nullable"
        varchar(50) rate_config_version
        varchar(50) shift_config_version
        integer shift_number
        varchar(20) state "ACTIVE|PENDING_PAYMENT|CLOSED|VOIDED"
        text qr_data
        timestamptz created_at
        timestamptz updated_at
        timestamptz synced_at
    }
    
    transactions {
        uuid id PK
        uuid session_id FK UK
        uuid location_id FK
        varchar(100) gate_id "exit gate hardware ID"
        integer amount "in cents"
        varchar(20) payment_method "EMONEY|FLAZZ|CASH|OFFLINE_SOP"
        varchar(100) payment_reference
        varchar(50) config_version
        integer shift_number
        timestamptz created_at
        timestamptz synced_at
        boolean voided
        timestamptz voided_at
        uuid voided_by FK "manager who voided"
        text void_reason
    }
    
    %% ==========================================
    %% INCIDENTS & AUDIT
    %% ==========================================
    
    incidents {
        uuid id PK
        uuid location_id FK
        varchar(100) gate_id "hardware ID if applicable"
        varchar(50) type "HARDWARE_FAILURE|PAYMENT_FAILURE|QR_UNREADABLE|PRINTER_JAM|GATE_MOTOR_FAILURE|OTHER"
        varchar(20) state "OPEN|IN_PROGRESS|RESOLVED"
        text description
        timestamptz resolved_at
        uuid resolved_by FK
        timestamptz created_at
        timestamptz updated_at
    }
    
    audit_logs {
        uuid id PK
        uuid user_id FK "null for system actions"
        uuid location_id FK
        varchar(100) action "e.g., session.create"
        varchar(50) entity_type "session|transaction|gate"
        uuid entity_id
        jsonb metadata
        varchar(45) ip_address
        timestamptz created_at
    }
    
    %% ==========================================
    %% RELATIONSHIPS
    %% ==========================================
    
    locations ||--o{ gates : "has"
    locations ||--o{ rate_configs : "configures"
    locations ||--o{ shift_configs : "configures"
    locations ||--o{ sessions : "hosts"
    locations ||--o{ transactions : "processes"
    locations ||--o{ incidents : "experiences"
    locations ||--o{ audit_logs : "tracked in"
    
    roles ||--o{ users : "assigned to"
    
    sessions ||--o| transactions : "results in"
    sessions }o--|| users : "voided_by (manager)"
    
    incidents }o--|| users : "resolved_by"
    audit_logs }o--|| users : "performed_by"
    
    %% ==========================================
    %% NOTES
    %% ==========================================
    
    %% gate_id in sessions/transactions is a STRING (hardware ID), not a FK
    %% This is intentional: hardware identifier from device, not database UUID
    %% sessions have NO operator_id (v2 is fully automated)
    %% Currency values use INTEGER (cents) for precision
    %% synced_at tracks when data synced to cloud (null = not synced)
    %% Configs are versioned for audit trail
```

## Clear Schema Diagram with Relationships

```
╔══════════════════════════════════════════════════════════════════════════════╗
║                              LOCATIONS                                       ║
╠══════════════════════════════════════════════════════════════════════════════╣
║  id (UUID PK)  │  name  │  code (UNIQUE)  │  address  │  city  │  status   ║
║  created_at    │  updated_at                                                   ║
╚══════════════════════════════════════════════════════════════════════════════╝
        │
        │ 1:N
        ├──────────────────────────────────────────────────────────┐
        │              │              │              │              │
        ▼              ▼              ▼              ▼              ▼
   ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐
   │ GATES   │   │  RATE   │   │  SHIFT  │   │SESSIONS │   │   TX    │
   │         │   │CONFIGS  │   │CONFIGS  │   │         │   │         │
   ├─────────┤   ├─────────┤   ├─────────┤   ├─────────┤   ├─────────┤
   │id (UUID)│   │id (UUID)│   │id (UUID)│   │id (UUID)│   │id (UUID)│
   │location │◄──│location │◄──│location │◄──│location │◄──│session  │
   │  _id    │   │  _id    │   │  _id    │   │  _id    │   │  _id    │
   │gate_id  │   │vehicle  │   │version  │   │gate_id  │   │location │
   │gate_type│   │  _type  │   │shift_   │   │vehicle  │   │  _id    │
   │vehicle  │   │version  │   │ code    │   │  _type  │   │gate_id  │
   │  _type  │   │first_hr │   │shift_   │   │check_in │   │amount   │
   │status   │   │subseq_hr│   │ number  │   │check_out│   │payment  │
   │config   │   │daily    │   │start    │   │fee_amt  │   │method   │
   │server_  │   │eff_from │   │end_time │   │rate_ver │   │config   │
   │  url    │   │eff_until│   │overnight│   │shift_ver│   │version  │
   │last_seen│   │created  │   │created  │   │shift_num│   │shift_num│
   │created  │   └─────────┘   └─────────┘   │state    │   │created  │
   │updated  │                                │qr_data  │   │synced   │
   └─────────┘                                │created  │   │voided   │
                                              │updated  │   │voided_at│
                                              │synced   │   │voided_by│◄──┐
                                              └────┬────┘   └─────────┘   │
                                                   │                       │
                                                   │ 1:1                   │
                                                   └───────────────────────┘


╔══════════════════════════════════════════════════════════════════════════════╗
║                                ROLES                                         ║
╠══════════════════════════════════════════════════════════════════════════════╣
║  id (UUID PK)  │  name (UNIQUE)  │  permissions (JSONB)                     ║
║  created_at    │  updated_at                                                 ║
╚══════════════════════════════════════════════════════════════════════════════╝
        │
        │ 1:N
        ▼
╔══════════════════════════════════════════════════════════════════════════════╗
║                                USERS                                         ║
╠══════════════════════════════════════════════════════════════════════════════╣
║  id (UUID PK)  │  role_id (FK→roles)  │  name  │  email (UNIQUE)            ║
║  password_hash │  status              │  created_at  │  updated_at          ║
╚══════════════════════════════════════════════════════════════════════════════╝
        │
        │ referenced by
        ├─────────────────────────────────┐
        │                                 │
        ▼                                 ▼
   ┌──────────┐                    ┌──────────┐
   │INCIDENTS │                    │AUDIT_LOGS│
   ├──────────┤                    ├──────────┤
   │id (UUID) │                    │id (UUID) │
   │location  │                    │user_id   │◄─── (who)
   │  _id     │                    │location  │
   │gate_id   │                    │  _id     │
   │type      │                    │action    │
   │state     │                    │entity_   │
   │desc      │                    │  type    │
   │resolved  │                    │entity_id │
   │  _at     │                    │metadata  │
   │resolved  │◄───────────────────│ip_address│
   │  _by     │   (resolved_by)    │created   │
   │created   │                    └──────────┘
   │updated   │
   └──────────┘


┌──────────────────────────────────────────────────────────────────────────────┐
│                         RELATIONSHIP SUMMARY                                 │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  LOCATIONS (1) ──────► (N) GATES              [location_id FK]              │
│  LOCATIONS (1) ──────► (N) RATE_CONFIGS       [location_id FK]              │
│  LOCATIONS (1) ──────► (N) SHIFT_CONFIGS      [location_id FK]              │
│  LOCATIONS (1) ──────► (N) SESSIONS           [location_id FK]              │
│  LOCATIONS (1) ──────► (N) TRANSACTIONS       [location_id FK]              │
│  LOCATIONS (1) ──────► (N) INCIDENTS          [location_id FK]              │
│  LOCATIONS (1) ──────► (N) AUDIT_LOGS         [location_id FK]              │
│                                                                              │
│  ROLES (1) ──────────► (N) USERS              [role_id FK]                  │
│                                                                              │
│  SESSIONS (1) ─────────► (1) TRANSACTIONS     [session_id FK, 1:1]          │
│                                                                              │
│  USERS (1) ────────────► (N) INCIDENTS        [resolved_by FK]              │
│  USERS (1) ────────────► (N) AUDIT_LOGS       [user_id FK]                  │
│  USERS (1) ────────────► (N) TRANSACTIONS     [voided_by FK]                │
│                                                                              │
│  IMPORTANT: gate_id in SESSIONS/TRANSACTIONS is a STRING (hardware ID)      │
│             NOT a UUID foreign key to GATES table                           │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

### ✅ Already Implemented (matches spec)

| Entity | Status | Notes |
|--------|--------|-------|
| `roles` | ✅ Match | Permissions as JSONB |
| `users` | ⚠️ Partial | Has `pin_hash` (not in spec), missing some indexes |
| `locations` | ⚠️ Partial | Has `capacity` JSONB (not in spec) |
| `audit_logs` | ⚠️ Partial | Field names differ slightly |

### ❌ Not Implemented / Needs Migration

| Entity | Status | What's Missing |
|--------|--------|----------------|
| `gates` | ❌ Incomplete | Missing: `gate_type`, `vehicle_type`, `status`, `server_room_url`, `config` JSONB. Current has: `device_id`, `name`, `ip_address` (not in spec) |
| `rate_configs` | ❌ Different | Current: `location_rates` (DECIMAL rates, no `version`). Spec: `rate_configs` (INTEGER cents, versioned) |
| `shift_configs` | ❌ Missing | Current has `shifts` (tracks actual shift instances). Spec needs `shift_configs` (versioned shift definitions) |
| `sessions` | ❌ Different | Current: has `operator_id`, `plate`, `city_code`, `rate_snapshot`. Spec: no operator, uses `gate_id` string, `rate_config_version`, `shift_config_version`, `qr_data`, `synced_at` |
| `transactions` | ❌ Different | Current: complex with `receipt_number`, many rate fields. Spec: simpler, uses `gate_id` string, `config_version`, `shift_number`, `synced_at`, different payment methods |
| `incidents` | ❌ Missing | Current: `incidents` + `incident_notes` (different structure). Spec: single `incidents` table with different fields |
| `sync_queue` | ❌ Missing | Server Room App only (not in cloud backend) |

### 🗑️ In Current Migrations but NOT in Spec

| Entity | Reason to Remove |
|--------|------------------|
| `shifts` | v2 has no operator shifts. Replaced by `shift_configs` (definitions) |
| `location_rates` | Replaced by `rate_configs` (versioned, integer cents) |
| `receipt_sequences` | Not in spec (receipts handled differently in v2) |
| `user_role_locations` | Not in spec (v2 uses simpler role model) |
| `user_permission_grants` | Not in spec (v2 uses role-level permissions only) |
| `sessions.operator_id` | v2 is fully automated, no operators |
| `sessions.plate`, `city_code` | Not in spec (gates are vehicle-type-specific, no manual plate entry) |
| `transactions.receipt_number` | Not in spec (receipts printed by gate app, not tracked in DB) |

## Key Differences: v1 → v2

| Aspect | v1 (Current) | v2 (PRD Spec) |
|--------|--------------|---------------|
| **Operators** | Human operators at gates | Fully automated, no operators |
| **Sessions** | Track `operator_id`, `plate`, manual entry | Track `gate_id` (hardware), `qr_data`, automated |
| **Rates** | DECIMAL (e.g., 5000.00) | INTEGER cents (e.g., 500000) |
| **Rate Configs** | Single table, no versioning | Versioned (`rate_v1`, `rate_v2`, ...) |
| **Shifts** | Track actual shift instances | Versioned shift definitions (`shift_configs`) |
| **Gate ID** | UUID FK to gates table | String (hardware serial/MAC) |
| **Payment Methods** | CASH, DIGITAL | EMONEY, FLAZZ, CASH, OFFLINE_SOP |
| **Sync** | Not tracked | `synced_at` timestamps for cloud sync |
| **Currency** | NUMERIC(12,2) | INTEGER (cents) |

## Migration Path

To align with PRD v2, you need:

1. **Rename `location_rates` → `rate_configs`**
   - Change DECIMAL to INTEGER (multiply by 100)
   - Add `version` column
   - Update trigger logic

2. **Create `shift_configs` table** (versioned shift definitions)
   - Keep `shifts` table temporarily if needed for v1 compatibility
   - Or drop if fully migrating to v2

3. **Update `gates` table**
   - Add: `gate_type`, `vehicle_type`, `status`, `server_room_url`, `config`
   - Rename: `device_id` → `gate_id`
   - Remove: `name`, `ip_address` (not in spec)

4. **Update `sessions` table**
   - Remove: `operator_id`, `plate`, `city_code`, `rate_snapshot`, `offline_sync`, `sync_conflict`
   - Add: `gate_id` (string), `rate_config_version`, `shift_config_version`, `shift_number`, `qr_data`, `synced_at`
   - Change: `fee_amount` from NUMERIC to INTEGER

5. **Update `transactions` table**
   - Remove: `receipt_number`, `rate_first_hour`, `rate_subsequent_hourly`, `rate_daily`, `duration_hours`, `check_in_at`, `check_out_at`, `amount_tendered`, `change_amount`
   - Add: `gate_id` (string), `amount` (INTEGER), `config_version`, `shift_number`, `synced_at`
   - Change: `payment_method` values (EMONEY, FLAZZ instead of DIGITAL)

6. **Update `users` table**
   - Remove: `pin_hash` (no operators in v2)

7. **Create `sync_queue` table** (Server Room App only, not in cloud backend)

8. **Create new `incidents` table** (replace current structure)

## Recommendation

**Option A: Gradual Migration**
- Keep v1 tables for backward compatibility
- Add v2 tables alongside (e.g., `rate_configs_v2`, `sessions_v2`)
- Migrate data in batches
- Switch applications to v2 tables
- Drop v1 tables

**Option B: Clean Break**
- Create new migration that drops v1-specific tables/columns
- Add v2 tables
- Migrate essential data (locations, users, roles)
- Update all application code to use v2 schema

**Option C: Dual Schema**
- Maintain both schemas temporarily
- v1 for current operations
- v2 for new automated system
- Sync between them during transition

Given that PRD v2 is a **complete redesign** (automated vs manual), **Option B (Clean Break)** is recommended if you're fully committing to v2. If you need to maintain v1 operations during transition, use **Option A (Gradual Migration)**.

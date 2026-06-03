# Chapter 3 — Users & Roles

## 3.1 User Types

The system has four distinct user types, each with a different scope of access and a different primary interface.

| User Type | Primary Interface | Scope |
|-----------|-----------------|-------|
| Parking Operator / Attendant | Desktop App | Day-to-day operations at a location |
| Facility Manager | Web Dashboard | Oversight of one or more locations |
| System Administrator | Web Dashboard | System configuration and technical management |
| Owner | Web Dashboard | Full business and financial oversight |

### Parking Operator / Attendant
- Works at a parking booth or gate.
- Responsible for: vehicle check-in, check-out, payment collection, receipt printing, and incident reporting.
- Access is limited to the location(s) they are assigned to.
- Cannot access reports, billing configuration, or user management unless explicitly granted.

### Facility Manager
- Oversees the operation of one or more locations.
- Responsible for: reviewing reports, managing operators, resolving incidents, and performing manual adjustments.
- Cannot modify system-wide configuration (rates across all locations, role definitions) unless explicitly granted.

### System Administrator
- Has global access to all locations for technical configuration.
- Responsible for: creating and managing roles, managing users, configuring locations, and monitoring system health.
- **Does not have access to financial data** (revenue reports, transaction details, payment summaries) unless explicitly granted.
- There should always be at least one active admin account.

### Owner
- Has full access to all locations, all configuration, and all financial data.
- Responsible for: business oversight, financial reporting, revenue analysis, and strategic decisions.
- Can view and export all financial reports across all locations.
- Can grant or revoke finance permissions for other roles.
- There should always be at least one active owner account.

---

## 3.2 Role-Based Access Control (RBAC)

### Overview
The system uses **role-based access control with fully custom permission sets**. Roles are not predefined templates — they are created and configured by system administrators, then assigned to users.

Key properties:
- A role is a named set of permissions.
- A user is assigned exactly one role.
- Permissions can optionally be scoped to specific locations (a manager may have `view_revenue` for Location A but not Location B).
- Roles can be created, edited, and deactivated by owners and admins.
- **Admins cannot grant `finance:*` permissions** — only owners can assign finance permissions to roles.
- Changes to a role take effect on the user's next action (or session refresh).

### 3.2.1 Access Control Architecture Options

Two architecture options are available for access control:

#### Option A: Pure Role-Based (Simple)

Permissions are tied entirely to the role. User inherits all permissions from their assigned role at their assigned locations.

```
User: Budi
  └── Role: manager
        └── Permissions: [reports:*, incidents:resolve, adjustments:*]
        └── Locations: [Location A, Location C]
```

**Pros:** Simple to understand and manage.
**Cons:** Cannot grant partial access (e.g., finance-only at Location B without full manager access).

#### Option B: Hybrid — Role-Location + Independent Grants (Recommended)

Two separate concepts:
1. **Role-Location Assignment** — Which locations the user works at with their role's permissions
2. **Permission Grants** — Additional permissions at ANY location (independent of role)

```
User: Budi
  └── Role: manager
  └── Role applies at: [Location A, Location C]
  └── Independent Grants:
        └── Location B: [finance:view_transactions]  ← ONLY this permission
```

**Effective permissions:**

| Location | Source | Permissions |
|----------|--------|-------------|
| Location A | Role (manager) | sessions:*, reports:*, incidents:*, adjustments:* |
| Location B | Grant only | finance:view_transactions |
| Location C | Role (manager) | sessions:*, reports:*, incidents:*, adjustments:* |

**Permission Resolution Logic:**
```
getPermissions(user, location):
    permissions = []

    # 1. If user's role applies at this location, add role permissions
    if location in user.role_locations:
        permissions += user.role.permissions

    # 2. Add any independent grants for this location
    permissions += user.grants.where(location_id = location)

    # 3. Add any global grants (location_id = null)
    permissions += user.grants.where(location_id = null)

    return unique(permissions)
```

**Example Scenarios:**

| User | Role | Role Locations | Independent Grants | Result |
|------|------|----------------|-------------------|--------|
| Budi | manager | A, C | Location B: finance:view_transactions | Manager at A & C, finance viewer only at B |
| Siti | operator | A | Location B: reports:view_revenue | Operator at A, reports viewer only at B |
| Andi | owner | — | Global: all permissions | Full access everywhere |

**Pros:**
- Role handles 90% of cases (standard permissions)
- Grants handle exceptions (specific access without full role)
- Supports temporary access (expires_at field)
- Easy to audit: "Show all users with finance:* grants"

**Cons:**
- Slightly more complex permission resolution
- Need to track both role permissions and user grants

> **Decision:** TBD — choose based on operational complexity needs.

### Permission Dimensions

Permissions are organized by module. Each permission is a string in the format `module:action`.

#### Sessions Module
| Permission | Description |
|-----------|-------------|
| `sessions:view` | View session list and details |
| `sessions:create` | Perform vehicle check-in |
| `sessions:close` | Perform vehicle check-out and close session |
| `sessions:void` | Void / cancel a session (requires adjustment permission) |

#### Payments Module
| Permission | Description |
|-----------|-------------|
| `payments:view` | View transaction records |
| `payments:collect_cash` | Record cash payment |
| `payments:collect_digital` | Record digital payment |
| `payments:void` | Void a transaction |

#### Manual Adjustments Module
| Permission | Description |
|-----------|-------------|
| `adjustments:void_transaction` | Authorize transaction void |
| `adjustments:reassign_session` | Reassign session to another operator |

#### Incidents Module
| Permission | Description |
|-----------|-------------|
| `incidents:view` | View incident list |
| `incidents:create` | File a new incident |
| `incidents:resolve` | Close an incident with resolution notes |

#### Reports Module
| Permission | Description |
|-----------|-------------|
| `reports:view_revenue` | Access revenue reports |
| `reports:view_occupancy` | Access occupancy reports |
| `reports:view_operators` | Access operator activity reports |

#### Finance Module
| Permission | Description |
|-----------|-------------|
| `finance:view_transactions` | View detailed transaction records with amounts |
| `finance:view_revenue_summary` | View aggregated revenue summaries across locations |
| `finance:view_cash_handover` | View shift cash handover and discrepancy reports |
| `finance:export` | Export financial data (CSV, PDF) |
| `finance:view_all_locations` | View financial data for all locations (overrides location scoping) |

#### User Management Module
| Permission | Description |
|-----------|-------------|
| `users:view` | View user list |
| `users:create` | Create new users |
| `users:edit` | Edit user details and role assignment |
| `users:deactivate` | Deactivate a user account |

#### Locations Module
| Permission | Description |
|-----------|-------------|
| `locations:view` | View location details and assigned operators |
| `locations:create` | Create a new location |
| `locations:edit` | Edit location name, address, code, capacity |
| `locations:deactivate` | Activate or deactivate a location |
| `locations:assign_operators` | Assign or remove operators from a location |

#### Rates Module
| Permission | Description |
|-----------|-------------|
| `rates:view` | View rate configurations |
| `rates:create` | Create new rate configurations |
| `rates:edit` | Edit existing rate configurations |

#### Observability Module
| Permission | Description |
|-----------|-------------|
| `observability:view_health` | View system health dashboard |
| `observability:view_audit` | View audit logs |
| `observability:view_alerts` | View anomaly alerts |
| `observability:manage_alerts` | Configure alert thresholds |

---

## 3.3 Default Role Suggestions

These are recommended starting roles. Owners can customize or replace them.

| Role Name | Typical Permissions |
|-----------|-------------------|
| `operator` | sessions:view, sessions:create, sessions:close, payments:collect_cash, payments:collect_digital, incidents:create, shifts:start, shifts:end |
| `manager` | reports:*, incidents:*, adjustments:*, users:view, locations:view, rates:view, observability:view_health, observability:view_alerts, shifts:view, shifts:force_close, shifts:resolve_discrepancy |
| `admin` | All permissions **except** finance:* — includes: users:*, locations:*, rates:*, observability:*, roles management |
| `owner` | All permissions across all modules including finance:* — full business and financial oversight |

### Manager Role Clarification

Managers are **supervisors**, not administrators. They can:
- View location details and rates (read-only)
- Authorize void transactions (sign-off with PIN)
- Resolve incidents (including manual gate authorization)
- View reports for assigned locations
- Manage shifts and discrepancies

Managers **cannot**:
- Create, edit, or deactivate locations
- Create or modify rate configurations
- Create or edit users
- Access finance data (unless explicitly granted)

### Permission Hierarchy

```
owner
  └── All permissions (including finance:*)
        │
admin   │
  └── All permissions EXCEPT finance:*
  └── locations:*, rates:*, users:*, observability:*
        │
manager │
  └── Reports, incidents, adjustments, shifts (location-scoped)
  └── locations:view, rates:view (read-only)
        │
operator
  └── Sessions, payments, incidents:create, shifts (location-scoped)
```

---

## 3.4 Location Scoping

- A user's permissions can be scoped to one or more specific locations.
- A user with `reports:view_revenue` scoped to Location A cannot view revenue for Location B.
- **Owners** have implicit access to all locations and all financial data.
- **System admins** have implicit access to all locations for technical configuration, but not financial data.
- When a user is assigned to a location, this is stored as a `user_location` association with the user's role context.

```
User
 └── Role (set of permissions)
      └── Location scope [Location A, Location B, ...]

Exception: Owner and Admin roles bypass location scoping for their respective domains.
  - Owner: all locations + all finance
  - Admin: all locations + all config (no finance)
```

---

## 3.5 User Account Lifecycle

| State | Description |
|-------|-------------|
| `ACTIVE` | User can log in and use the system |
| `DEACTIVATED` | User cannot log in; historical records are preserved |

- Accounts are never hard-deleted to preserve audit trail integrity.
- Deactivating a user does not void any of their historical sessions or transactions.
- A deactivated user's name still appears in audit logs and reports.

---

## 3.6 Authentication

- Users authenticate with **email + password**.
- Session tokens are used for subsequent requests (stateless JWT recommended).
- Token expiry: configurable, default 8 hours (aligned to a typical operator shift).
- The desktop app may support PIN-based re-authentication for quick unlock after inactivity.
- Password reset is handled via admin action (v1); self-service email reset is out of scope.

---

## 3.7 User Data Model

### Option A: Pure Role-Based

```
users
  id                UUID, primary key
  name              VARCHAR(100), not null
  email             VARCHAR(255), unique, not null
  password_hash     VARCHAR, not null
  role_id           UUID, FK → roles.id
  status            ENUM('ACTIVE', 'DEACTIVATED'), default ACTIVE
  created_at        TIMESTAMP
  updated_at        TIMESTAMP

roles
  id                UUID, primary key
  name              VARCHAR(100), unique
  permissions       JSONB  -- array of permission strings
  created_at        TIMESTAMP
  updated_at        TIMESTAMP

user_locations
  user_id           UUID, FK → users.id
  location_id       UUID, FK → locations.id
  PRIMARY KEY (user_id, location_id)
```

### Option B: Hybrid — Role-Location + Independent Grants

```sql
-- Users table
users
  id                UUID PRIMARY KEY
  name              VARCHAR(100) NOT NULL
  email             VARCHAR(255) UNIQUE NOT NULL
  password_hash     VARCHAR NOT NULL
  pin_hash          VARCHAR              -- 6-digit manager PIN (hashed)
  role_id           UUID, FK → roles.id  -- user's primary role
  status            VARCHAR(20) DEFAULT 'ACTIVE'
                      CHECK (status IN ('ACTIVE', 'DEACTIVATED'))
  created_at        TIMESTAMPTZ DEFAULT now()
  updated_at        TIMESTAMPTZ DEFAULT now()

-- Roles table (permission templates)
roles
  id                UUID PRIMARY KEY
  name              VARCHAR(100) UNIQUE NOT NULL
  permissions       JSONB NOT NULL DEFAULT '[]'
  created_at        TIMESTAMPTZ DEFAULT now()
  updated_at        TIMESTAMPTZ DEFAULT now()

-- Where the user's role applies (role-based access)
user_role_locations
  user_id           UUID, FK → users.id
  location_id       UUID, FK → locations.id
  PRIMARY KEY (user_id, location_id)
  -- User gets role.permissions at this location

-- Independent permission grants (additive, any location)
user_permission_grants
  id                UUID PRIMARY KEY
  user_id           UUID NOT NULL, FK → users.id
  location_id       UUID, FK → locations.id  -- NULL = global grant
  permission        VARCHAR(100) NOT NULL
  granted_by        UUID NOT NULL, FK → users.id
  granted_at        TIMESTAMPTZ DEFAULT now()
  expires_at        TIMESTAMPTZ              -- optional: temporary access
  revoked_at        TIMESTAMPTZ              -- soft delete
  revoked_by        UUID, FK → users.id

  UNIQUE (user_id, location_id, permission)

-- Indexes
CREATE INDEX idx_grants_user ON user_permission_grants (user_id);
CREATE INDEX idx_grants_location ON user_permission_grants (location_id);
CREATE INDEX idx_grants_permission ON user_permission_grants (permission);
CREATE INDEX idx_grants_active ON user_permission_grants (user_id, location_id)
  WHERE revoked_at IS NULL AND (expires_at IS NULL OR expires_at > now());
```

**Query: Get all effective permissions for a user at a location**

```sql
-- Effective permissions = role permissions (if assigned) + grants
WITH role_perms AS (
  SELECT jsonb_array_elements_text(r.permissions) AS permission
  FROM users u
  JOIN roles r ON r.id = u.role_id
  JOIN user_role_locations url ON url.user_id = u.id
  WHERE u.id = :user_id
    AND url.location_id = :location_id
),
grant_perms AS (
  SELECT permission
  FROM user_permission_grants
  WHERE user_id = :user_id
    AND (location_id = :location_id OR location_id IS NULL)
    AND revoked_at IS NULL
    AND (expires_at IS NULL OR expires_at > now())
)
SELECT DISTINCT permission FROM (
  SELECT permission FROM role_perms
  UNION
  SELECT permission FROM grant_perms
) all_perms;
```

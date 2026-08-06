# Chapter 3 — Users & Roles

## 3.1 User Types

The system has four distinct user types. **Note: No operators exist in v2** — gates are fully automated. Staff only handle exceptions.

| User Type | Primary Interface | Scope |
|-----------|-----------------|-------|
| Server Room Staff | Server Room App (local) | Monitor gates, handle exceptions at a location |
| Facility Manager | Web Dashboard | Oversight of one or more locations |
| System Administrator | Web Dashboard | System configuration and technical management |
| Owner (Superadmin) | Web Dashboard | Full business and financial oversight |

### Server Room Staff
- Works in server room at a location (2-3 staff + 1 leader per location).
- **Does NOT process transactions** — gates are fully automated.
- Responsible for: monitoring gate status, responding to alerts, handling exceptions (hardware failures, payment issues), running offline SOP when needed.
- Uses server room app (local desktop app) to view gate status and respond to alerts.
- Cannot access reports, billing configuration, or user management.

### Facility Manager
- Oversees the operation of one or more locations.
- Responsible for: reviewing reports, monitoring gate health, managing staff, resolving incidents.
- Uses web dashboard to view live monitoring, reports, gate status.
- Cannot modify system-wide configuration (rates across all locations, role definitions) unless explicitly granted.

### System Administrator
- Has global access to all locations for technical configuration.
- Responsible for: creating and managing roles, managing users, configuring locations, monitoring system health, managing gate configurations.
- **Does not have access to financial data** (revenue reports, transaction details, payment summaries) unless explicitly granted.
- There should always be at least one active admin account.

### Owner (Superadmin)
- Has full access to all locations, all configuration, and all financial data.
- Responsible for: business oversight, financial reporting, revenue analysis, and strategic decisions.
- Can view and export all financial reports across all locations.
- Can grant or revoke finance permissions for other roles.
- Initial accounts created by developer (bootstrap). Subsequent accounts created by superadmin.
- There should always be at least one active owner account.

---

## 3.2 Role-Based Access Control (RBAC)

### Overview
The system uses **role-based access control with fully custom permission sets**. Roles are created and configured by system administrators, then assigned to users.

Key properties:
- A role is a named set of permissions.
- A user is assigned exactly one role.
- Permissions can optionally be scoped to specific locations.
- Roles can be created, edited, and deactivated by owners and admins.
- **Admins cannot grant `finance:*` permissions** — only owners can assign finance permissions to roles.

### Seeded Roles

| Role | Permissions Summary |
|------|-------------------|
| **owner** | Full access: sessions, payments, incidents, reports, finance, users, locations, rates, gates, vehicle-types, observability, shifts |
| **admin** | Same as owner **minus** `finance:*` |
| **manager** | sessions:view, reports:*, incidents:*, locations:view, gates:view, payments:view, users:view, shifts:view |
| **staff** | gates:view, incidents:view, locations:view (server room staff, minimal permissions) |

### Permission Categories

**Gates:**
- `gates:view` — View gate status and configuration
- `gates:configure` — Configure new gates (vehicle type, gate type)
- `gates:deactivate` — Deactivate gates

**Sessions:**
- `sessions:view` — View sessions (active + closed)
- `sessions:void` — Void a session

**Payments:**
- `payments:view` — View transactions
- `payments:void` — Void a transaction (requires manager PIN)

**Incidents:**
- `incidents:view` — View incidents
- `incidents:resolve` — Resolve incidents

**Reports:**
- `reports:view_revenue` — View revenue reports
- `reports:view_occupancy` — View occupancy reports
- `reports:view_operators` — View operator activity (N/A in v2, kept for compatibility)

**Finance:**
- `finance:view_transactions` — View transaction details
- `finance:view_revenue_summary` — View revenue summaries
- `finance:export` — Export financial data

**Users:**
- `users:view` — View users
- `users:create` — Create users
- `users:edit` — Edit users
- `users:deactivate` — Deactivate users

**Locations:**
- `locations:view` — View locations
- `locations:create` — Create locations
- `locations:edit` — Edit locations
- `locations:deactivate` — Deactivate locations

**Rates:**
- `rates:view` — View rate configurations
- `rates:create` — Create rate configurations
- `rates:edit` — Edit rate configurations

**Observability:**
- `observability:view_health` — View system health
- `observability:view_audit` — View audit logs
- `observability:view_alerts` — View alerts
- `observability:manage_alerts` — Manage alert configurations

**Shifts:**
- `shifts:view` — View shift configurations

---

## 3.3 Authentication

### Dashboard Authentication
- JWT RS256 tokens with 8-hour expiry
- httpOnly cookie (`access_token`)
- Refresh token flow (automatic refresh before expiry)
- Initial superadmin account created by developer (bootstrap)
- Superadmin can create new accounts and assign roles

### Gate App Authentication
- **No authentication required** — gate app is stateless, runs on trusted LAN
- Communicates with server room app via HTTP (no auth)
- Server room app communicates with cloud backend (API key or service account)

### Server Room App Authentication
- **No user-facing authentication** — runs unattended in server room
- Communicates with cloud backend (API key or service account)
- Communicates with gate apps via LAN (trusted network, no auth)

---

## 3.4 Gate App vs Server Room App Access

| Component | Authentication | Access Control |
|-----------|---------------|----------------|
| **Gate App** | None | LAN-only, trusted network |
| **Server Room App** | API key (to cloud) | LAN-only (gate apps), API key (cloud) |
| **Dashboard** | JWT (user accounts) | RBAC (roles + permissions) |
| **Cloud Backend** | API key (server room apps), JWT (dashboard) | RBAC for dashboard users |

---

## 3.5 User Account Lifecycle

1. **Bootstrap:** Developer creates initial superadmin account (one-time setup)
2. **Creation:** Superadmin creates new user accounts via dashboard
3. **Assignment:** Superadmin assigns role + locations to user
4. **Login:** User logs in with email + password → receives JWT
5. **Access:** User accesses dashboard with role-based permissions
6. **Deactivation:** Superadmin deactivates user (soft delete, preserves audit trail)

---

## 3.6 Design Decisions

**Why no operators?**
- Fully automated gates eliminate need for human operators
- Reduces operational costs for AMB
- Faster throughput (no human delay)
- Staff only handle exceptions (hardware failures, payment issues)

**Why minimal staff permissions?**
- Staff only monitor gates and handle exceptions
- Don't need access to financial data or configuration
- Can view gate status and incidents only
- Reduces security risk (limited access)

**Why API key for server room app?**
- Server room app runs unattended (no user login)
- Needs to authenticate to cloud backend for sync
- API key is simpler than JWT for service-to-service auth
- Can be rotated if compromised

---

*End of Chapter 3 — Users & Roles (v2)*

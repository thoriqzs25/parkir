# Milestone 1: Backend — Gates Table & Endpoints

## Objective

Create the database table and API endpoints for gate registration and gate info display.

## Files

| # | File | Action |
|---|------|--------|
| 1 | `backend/migrations/000008_gates.up.sql` | **New** — `CREATE TABLE gates` |
| 2 | `backend/migrations/000008_gates.down.sql` | **New** — `DROP TABLE gates` |
| 3 | `backend/internal/store/gate.go` | **New** — Store CRUD + `GetGateInfo` |
| 4 | `backend/internal/domain/gate/handler.go` | **New** — Public + admin handlers |
| 5 | `backend/internal/permissions/permissions.go` | **Edit** — Add `gates:*` permissions |
| 6 | `backend/cmd/api/main.go` | **Edit** — Register routes |

## 1.1 Migration — `000008_gates.up.sql`

```sql
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
BEGIN NEW.updated_at = now(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_gates_updated_at
    BEFORE UPDATE ON gates FOR EACH ROW
    EXECUTE FUNCTION update_gates_updated_at();
```

### 1.1a Down migration

```sql
DROP TRIGGER IF EXISTS trg_gates_updated_at ON gates;
DROP FUNCTION IF EXISTS update_gates_updated_at();
DROP TABLE IF EXISTS gates;
```

## 1.2 Store — `backend/internal/store/gate.go`

### Structs

```go
type Gate struct {
    ID           string     `json:"id"`
    DeviceID     string     `json:"device_id"`
    Name         string     `json:"name"`
    LocationID   *string    `json:"location_id,omitempty"`
    IPAddress    string     `json:"ip_address"`
    LastSeenAt   *time.Time `json:"last_seen_at,omitempty"`
    RegisteredAt time.Time  `json:"registered_at"`
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
}

type RegisterGateInput struct {
    DeviceID   string
    Name       string
    LocationID *string
    IPAddress  string
}

type UpdateGateInput struct {
    Name       *string
    LocationID *string
    IPAddress  *string
}

type GateInfo struct {
    Location struct {
        Name string `json:"name"`
        Code string `json:"code"`
    } `json:"location"`
    Rates    []RateSummary      `json:"rates"`
    Capacity map[string]int64   `json:"capacity"`
}

type RateSummary struct {
    VehicleType          string  `json:"vehicle_type"`
    FirstHourRate        float64 `json:"first_hour_rate"`
    SubsequentHourlyRate float64 `json:"subsequent_hourly_rate"`
    DailyFlatRate        float64 `json:"daily_flat_rate"`
}
```

### Methods

| Method | SQL | Returns |
|--------|-----|---------|
| `RegisterGate(ctx, RegisterGateInput)` | `INSERT INTO gates ... RETURNING *` | `*Gate`, error on unique violation |
| `GetGateByID(ctx, id)` | `SELECT * FROM gates WHERE id = $1` | `*Gate`, ErrNotFound |
| `GetGateByDeviceID(ctx, deviceID)` | `SELECT * FROM gates WHERE device_id = $1` | `*Gate`, ErrNotFound |
| `ListGates(ctx, locationID)` | `SELECT * FROM gates WHERE location_id = $1 ORDER BY name` | `[]Gate` |
| `UpdateGate(ctx, id, UpdateGateInput)` | `UPDATE gates SET ... WHERE id = $1 RETURNING *` | `*Gate`, ErrNotFound |
| `DeleteGate(ctx, id)` | `DELETE FROM gates WHERE id = $1` | error (check RowsAffected for 404) |
| `GetGateInfo(ctx, locationID)` | Multi-query: location + active rates + capacity | `*GateInfo`, ErrNotFound |

### GetGateInfo details

1. Query location by ID:
   ```sql
   SELECT name, code, capacity FROM locations WHERE id = $1
   ```
2. Query currently active rates:
   ```sql
   SELECT vehicle_type, first_hour_rate, subsequent_hourly_rate, daily_flat_rate
   FROM location_rates
   WHERE location_id = $1
     AND effective_from <= CURRENT_DATE
     AND (effective_until IS NULL OR effective_until >= CURRENT_DATE)
   ORDER BY vehicle_type ASC
   ```
3. If location not found → return `apperrors.ErrNotFound`
4. If no active rates → return empty rates array (not an error)
5. Capacity is JSONB → scan into `map[string]int64`

## 1.3 Handler — `backend/internal/domain/gate/handler.go`

### Structs

```go
type Handler struct {
    store *store.Store
}

type RegisterGateRequest struct {
    DeviceID   string  `json:"device_id" binding:"required"`
    Name       string  `json:"name"`
    LocationID *string `json:"location_id,omitempty"`
    IPAddress  string  `json:"ip_address"`
}

type UpdateGateRequest struct {
    Name       *string `json:"name,omitempty"`
    LocationID *string `json:"location_id,omitempty"`
    IPAddress  *string `json:"ip_address,omitempty"`
}
```

### Route registration

```go
func (h *Handler) RegisterPublicRoutes(r *gin.RouterGroup) {
    r.GET("/gate/:id/info", h.GetGateInfo)
}

func (h *Handler) RegisterAdminRoutes(r *gin.RouterGroup) {
    gates := r.Group("/gates")
    gates.Use(middleware.RequirePermission("gates:view"))
    {
        gates.GET("", h.ListGates)
        gates.GET("/:id", h.GetGate)
    }

    gatesCreate := r.Group("/gates")
    gatesCreate.Use(middleware.RequirePermission("gates:register"))
    {
        gatesCreate.POST("", h.RegisterGate)
    }

    gatesEdit := r.Group("/gates")
    gatesEdit.Use(middleware.RequirePermission("gates:edit"))
    {
        gatesEdit.PATCH("/:id", h.UpdateGate)
    }

    gatesDelete := r.Group("/gates")
    gatesDelete.Use(middleware.RequirePermission("gates:delete"))
    {
        gatesDelete.DELETE("/:id", h.DeleteGate)
    }
}
```

### Handler methods

**`GetGateInfo(c)`**
1. `id := c.Param("id")` — this is the location ID
2. `info, err := h.store.GetGateInfo(ctx, id)`
3. `ErrNotFound` → `response.NotFound(c, "location")`
4. Otherwise → `response.OK(c, info)`

**`RegisterGate(c)`**
1. Bind JSON → `RegisterGateRequest`
2. `gate, err := h.store.RegisterGate(ctx, ...)`
3. Unique violation (detect pgx error code 23505) → `response.Conflict(c, "GATE_EXISTS", "device_id already registered")`
4. Otherwise internal error or `response.Created(c, gate)`

**`ListGates(c)`**
1. `locationID := c.Query("location_id")` — optional filter
2. `gates, err := h.store.ListGates(ctx, locationID)`
3. `response.OK(c, gates)`

**`GetGate(c)`**
1. `id := c.Param("id")`
2. Try `GetGateByID` — if `ErrNotFound`, try `GetGateByDeviceID`
3. Both not found → 404

**`UpdateGate(c)`**
1. `id := c.Param("id")`, bind JSON
2. `gate, err := h.store.UpdateGate(ctx, id, input)`
3. `ErrNotFound` → 404

**`DeleteGate(c)`**
1. `id := c.Param("id")`
2. `err := h.store.DeleteGate(ctx, id)`
3. Check `ErrNotFound` on the error type or `RowsAffected == 0`

## 1.4 Permissions — `backend/internal/permissions/permissions.go`

Add before the `shifts:*` block:
```go
// Gates
"gates:view":     true,
"gates:register": true,
"gates:edit":     true,
"gates:delete":   true,
```

## 1.5 Routes — `cmd/api/main.go`

```go
import gatedomain "github.com/thoriqzs/PARKIR/backend/internal/domain/gate"

// After health routes:
gateHandler := gatedomain.NewHandler(store)
gateHandler.RegisterPublicRoutes(public)

// Inside the protected `api` group:
gateHandler.RegisterAdminRoutes(api)
```

## 1.6 Tests

### Unit test — `backend/internal/store/gate_test.go`

| Test | What it validates |
|------|-------------------|
| `TestStoreRegisterGate_Success` | Register a gate → returned with ID, non-empty timestamps |
| `TestStoreRegisterGate_DuplicateDeviceID` | Register twice with same device_id → error |
| `TestStoreGetGateByID_NotFound` | Random UUID → ErrNotFound |
| `TestStoreUpdateGate` | Register → update name → name changed |
| `TestStoreDeleteGate` | Register → delete → GetByID returns ErrNotFound |
| `TestStoreDeleteGate_NotFound` | Delete random ID → error (RowsAffected == 0) |

### Integration test — `backend/internal/domain/gate/integration_test.go`

| Test | What it validates |
|------|-------------------|
| `TestGetGateInfo` | Seed location + 2 active rates. Call endpoint. Verify location name + 2 rates returned. |
| `TestGetGateInfo_FutureRateExcluded` | Seed location + 1 active rate + 1 future rate. Only active rate returned. |
| `TestGetGateInfo_NoRates` | Seed location with no rates. Empty rates array. |
| `TestGetGateInfo_LocationNotFound` | Random UUID → 404. |
| `TestRegisterGate` | POST valid gate → 201 + gate with ID. |
| `TestRegisterGate_Duplicate` | Same device_id → 409. |
| `TestListGates` | Register 2 gates → list returns 2. |
| `TestGateCRUD` | Register → update name → PATCH → name changed; list → updated; delete → list empty. |

### Test helpers (same pattern as `reports/integration_test.go`)

```go
func seedLocation(ctx, t, s) string
func seedRate(ctx, t, s, locationID, vehicleType, effectiveFrom, effectiveUntil string)
```

### Run

```bash
# Unit tests (no DB needed)
cd backend && go test ./internal/store/ -run TestStoreGate -v

# Integration tests (need DB)
cd backend && PARKIR_TEST_DATABASE_URL=... go test ./internal/domain/gate/ -v
```

## 1.7 Manual verification

```bash
make migrate-up

# Test public endpoint
curl http://localhost:8080/api/v1/gate/<location-uuid>/info
# → { "location": { "name": "...", "code": "..." }, "rates": [...], "capacity": {...} }

# Test admin endpoints (with JWT cookie)
curl -X POST http://localhost:8080/api/v1/gates \
  -H "Content-Type: application/json" \
  -b "access_token=..." \
  -d '{"device_id":"gate-001","name":"Gate A","location_id":"<uuid>","ip_address":"192.168.1.100"}'
# → 201 + gate object

curl http://localhost:8080/api/v1/gates -b "access_token=..."
# → [gate objects]
```

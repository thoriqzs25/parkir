# PARKIR — System Flow

## Initial Startup (Seed)

1. **Seed** is run (`make seed` / `cmd/seed/main.go`) after migrations.
2. **4 roles** are created (upsert by name):
   - `owner` — full permissions across all domains
   - `admin` — full permissions minus finance and adjustments
   - `manager` — view + operational permissions, no payments collect
   - `operator` — minimal: sessions CRUD, payments collect, shifts start/end
3. **Default owner account** is created (upsert by email):
   - Name: `Root Owner`
   - Email: `owner@parkir.local`
   - Password: `owner123`
   - PIN: `123456`
   - Role: `owner`
4. **Default location** is created (upsert by code):
   - Name: `Main Location`
   - Code: `MAIN`
   - Address: `123 Main St`, Jakarta
5. Owner is **assigned to the Main Location** via `user_role_locations`.

## Owner Creates a Location

After login, the owner can create additional locations:

1. **POST** `/api/v1/locations`
2. **Auth:** JWT (cookie or Bearer) required.
3. **Permission:** `locations:create` (owner role has this via `locations:*`).
4. **Request body** (JSON):
   - `name` (required) — location display name
   - `code` (required) — unique short code
   - `address` (optional)
   - `city` (optional)
   - `capacity` (optional, JSONB)
5. **Validation:** handler binds and validates; missing `name`/`code` → `400 INVALID_INPUT`.
6. **Store:** raw SQL `INSERT INTO locations ... RETURNING *`.
7. **Response:** `201 Created` with the location object (id, name, code, address, city, status, capacity, created_at, updated_at).
8. **Status** defaults to `ACTIVE`.

The owner is NOT automatically assigned to the new location — they must assign users (including themselves) via `user_role_locations`.
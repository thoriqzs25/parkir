# Auth API Verification Report

## 1. Setup

- Started the full dev stack with `docker compose up -d --build`.
- Confirmed backend was healthy via `GET /health/ready`.
- Ran `make seed` to create the default owner user.

## 2. Test Flow & Results

| Step | Endpoint | Method | Status | Result |
|------|----------|--------|--------|--------|
| Health check | `/health/ready` | GET | 200 | Database connected |
| Owner login | `/api/v1/auth/login` | POST | 200 | Returned user + JWT token |
| List roles | `/api/v1/roles` | GET | 200 | Got role IDs |
| Create user (manager) | `/api/v1/users` | POST | 201 | Created `testmanager@parkir.local` |
| Create user (operator) | `/api/v1/users` | POST | 201 | Created `testoperator@parkir.local` |
| Login as new operator | `/api/v1/auth/login` | POST | 200 | Returned user + JWT token |
| Fetch current user | `/api/v1/auth/me` | GET | 200 | Returned operator profile |
| Logout | `/api/v1/auth/logout` | POST | 200 | Returned `{ "message": "logged out" }` |
| `/auth/me` after logout | `/api/v1/auth/me` | GET | 401 | Correctly rejected |
| Refresh without token | `/api/v1/auth/refresh` | POST | 401 | Correctly rejected |
| Refresh with valid token | `/api/v1/auth/refresh` | POST | 200 | Returned new token |

**Conclusion:** The auth API is now working for login, create user, fetch current user, refresh, and logout.

## 3. Bug Found & Fixed

During testing, I discovered that newly created users without a PIN could not log in. `GET /api/v1/auth/me` also failed for them.

### Root Cause

`users.pin_hash` is nullable in PostgreSQL, but the Go `User.PINHash` field was a non-nullable `string`. `GetUserByEmail` and `GetUserByID` tried to scan `NULL` into a `string`, causing an internal server error.

### Files Changed

1. `backend/internal/store/user.go`
   - Changed `PINHash` from `string` to `*string` so null values scan correctly.

2. `backend/internal/domain/transactions/handler.go`
   - Updated `validateManagerPIN` to dereference the pointer safely.

### Verification

- `go build ./...` in `backend/` passes.
- Backend compiled and restarted successfully.
- Login + `/auth/me` now work for users without a PIN.

## 4. Cleanup

- Stopped all containers with `docker compose down -v`.
- Removed temp cookie files and test script.
- No test containers or volumes left behind.

## 5. Unrelated Observations

- `/auth/login` response returns `created_at` and `updated_at` as `"0001-01-01T00:00:00Z"` because `GetUserByEmail` does not select those columns. `/auth/me` returns the real timestamps. This does not break auth but is inconsistent.
- `/auth/me` returned `"permissions": []` for the operator user, even though the operator role has permissions. You may want to check the permissions resolver if the dashboard relies on this list.

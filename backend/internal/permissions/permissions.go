package permissions

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ValidPermissions is the hardcoded allow-list of all permissions in the system.
var ValidPermissions = map[string]bool{
	// Sessions
	"sessions:view":   true,
	"sessions:create": true,
	"sessions:close":  true,
	"sessions:void":   true,

	// Payments
	"payments:view":            true,
	"payments:collect_cash":    true,
	"payments:collect_digital": true,
	"payments:void":            true,

	// Adjustments
	"adjustments:void_transaction":  true,
	"adjustments:reassign_session":  true,

	// Incidents
	"incidents:view":    true,
	"incidents:create":  true,
	"incidents:resolve": true,

	// Reports
	"reports:view_revenue":    true,
	"reports:view_occupancy":  true,
	"reports:view_operators":  true,

	// Finance
	"finance:view_transactions":    true,
	"finance:view_revenue_summary": true,
	"finance:view_cash_handover":   true,
	"finance:export":               true,
	"finance:view_all_locations":   true,

	// Users
	"users:view":        true,
	"users:create":      true,
	"users:edit":        true,
	"users:deactivate":  true,

	// Locations
	"locations:view":            true,
	"locations:create":          true,
	"locations:edit":            true,
	"locations:deactivate":      true,
	"locations:assign_operators": true,

	// Rates
	"rates:view":   true,
	"rates:create": true,
	"rates:edit":   true,

	// Observability
	"observability:view_health":  true,
	"observability:view_audit":   true,
	"observability:view_alerts":  true,
	"observability:manage_alerts": true,

	// Gates
	"gates:view":     true,
	"gates:register": true,
	"gates:edit":     true,
	"gates:delete":   true,

	// Vehicle Types
	"vehicle-types:view":   true,
	"vehicle-types:create": true,
	"vehicle-types:edit":   true,
	"vehicle-types:delete": true,

	// Shifts
	"shifts:start":               true,
	"shifts:end":                 true,
	"shifts:view":                true,
	"shifts:force_close":         true,
	"shifts:resolve_discrepancy": true,
}

// IsValid checks if a permission string is in the allow-list.
func IsValid(permission string) bool {
	return ValidPermissions[permission]
}

// ValidateList returns an error if any permission is unknown.
// Wildcard patterns like "reports:*" are accepted if the module exists.
func ValidateList(perms []string) error {
	for _, p := range perms {
		if IsValid(p) {
			continue
		}
		if strings.HasSuffix(p, ":*") {
			module := strings.TrimSuffix(p, ":*")
			moduleExists := false
			for valid := range ValidPermissions {
				if strings.HasPrefix(valid, module+":") {
					moduleExists = true
					break
				}
			}
			if !moduleExists {
				return fmt.Errorf("invalid permission module: %s", p)
			}
			continue
		}
		return fmt.Errorf("invalid permission: %s", p)
	}
	return nil
}

// Has checks if a permission is present in a list.
func Has(permissions []string, permission string) bool {
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAny checks if any of the target permissions is present.
func HasAny(permissions []string, targets ...string) bool {
	for _, t := range targets {
		if Has(permissions, t) {
			return true
		}
	}
	return false
}

// HasAll checks if all target permissions are present.
func HasAll(permissions []string, targets ...string) bool {
	for _, t := range targets {
		if !Has(permissions, t) {
			return false
		}
	}
	return true
}

// MatchWildcard checks if a permission matches a wildcard pattern like "reports:*".
func MatchWildcard(permission, pattern string) bool {
	if !strings.HasSuffix(pattern, ":*") {
		return permission == pattern
	}

	prefix := strings.TrimSuffix(pattern, ":*")
	module := strings.Split(permission, ":")[0]
	return module == prefix
}

// Expand expands wildcard permissions like "reports:*" into all matching permissions.
func Expand(permissions []string) []string {
	result := make(map[string]bool)

	for _, p := range permissions {
		if strings.HasSuffix(p, ":*") {
			prefix := strings.TrimSuffix(p, ":*")
			for valid := range ValidPermissions {
				if strings.HasPrefix(valid, prefix+":") {
					result[valid] = true
				}
			}
		} else {
			result[p] = true
		}
	}

	expanded := make([]string, 0, len(result))
	for p := range result {
		expanded = append(expanded, p)
	}
	sort.Strings(expanded)
	return expanded
}

// Resolver resolves effective permissions for a user at a location.
type Resolver struct {
	pool *pgxpool.Pool
}

func NewResolver(pool *pgxpool.Pool) *Resolver {
	return &Resolver{pool: pool}
}

// EffectivePermissions returns all permissions a user has at a specific location.
// It combines: role permissions (if role applies at location) + active grants for that location + active global grants.
// Owners bypass location scoping and get all permissions. Admins bypass location scoping and get all except finance:*.
func (r *Resolver) EffectivePermissions(ctx context.Context, userID string, locationID *string) ([]string, error) {
	set := make(map[string]bool)

	// Check for owner/admin bypass
	var roleName string
	err := r.pool.QueryRow(ctx, `
		SELECT r.name
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1 AND u.status = 'ACTIVE' AND r.deleted_at IS NULL
	`, userID).Scan(&roleName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query user role: %w", err)
	}

	if roleName == "owner" {
		for p := range ValidPermissions {
			set[p] = true
		}
		return sortedPermissions(set), nil
	}

	if roleName == "admin" {
		for p := range ValidPermissions {
			if !strings.HasPrefix(p, "finance:") {
				set[p] = true
			}
		}
		return sortedPermissions(set), nil
	}

	// 1. Role permissions
	var rolePerms []string
	if locationID != nil {
		// When a specific location is given, scope to that location.
		err = r.pool.QueryRow(ctx, `
			SELECT r.permissions
			FROM users u
			JOIN roles r ON r.id = u.role_id
			JOIN user_role_locations url ON url.user_id = u.id
			WHERE u.id = $1
			  AND r.deleted_at IS NULL
			  AND url.location_id = $2
		`, userID, locationID).Scan(&rolePerms)
	} else {
		// Without a location context, return role permissions (the user must
		// be assigned to at least one location for the role to be active).
		err = r.pool.QueryRow(ctx, `
			SELECT r.permissions
			FROM users u
			JOIN roles r ON r.id = u.role_id
			JOIN user_role_locations url ON url.user_id = u.id
			WHERE u.id = $1
			  AND r.deleted_at IS NULL
			LIMIT 1
		`, userID).Scan(&rolePerms)
	}
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("query role permissions: %w", err)
	}

	for _, p := range Expand(rolePerms) {
		set[p] = true
	}

	// 2 & 3. Active independent grants for this location and global grants
	queryGrants := `
		SELECT permission
		FROM user_permission_grants
		WHERE user_id = $1
		  AND (location_id = $2 OR location_id IS NULL)
		  AND revoked_at IS NULL
		  AND (expires_at IS NULL OR expires_at > now())
	`

	rows, err := r.pool.Query(ctx, queryGrants, userID, locationID)
	if err != nil {
		return nil, fmt.Errorf("query grants: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			return nil, fmt.Errorf("scan grant: %w", err)
		}
		for _, p := range Expand([]string{perm}) {
			set[p] = true
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("grant rows: %w", err)
	}

	return sortedPermissions(set), nil
}

func sortedPermissions(set map[string]bool) []string {
	result := make([]string, 0, len(set))
	for p := range set {
		result = append(result, p)
	}
	sort.Strings(result)
	return result
}

// IsOwner checks if a user has the owner role.
func (r *Resolver) IsOwner(ctx context.Context, userID string) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1 AND r.name = 'owner' AND r.deleted_at IS NULL
	`, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check owner: %w", err)
	}
	return count > 0, nil
}

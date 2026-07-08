package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
)

type Role struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Permissions []string   `json:"permissions"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateRoleInput struct {
	Name        string
	Permissions []string
}

type UpdateRoleInput struct {
	Name        *string
	Permissions []string
}

func (s *Store) CreateRole(ctx context.Context, input CreateRoleInput) (*Role, error) {
	var role Role
	err := s.pool.QueryRow(ctx, `
		INSERT INTO roles (name, permissions)
		VALUES ($1, $2)
		RETURNING id, name, permissions, created_at, updated_at
	`, input.Name, input.Permissions).Scan(
		&role.ID, &role.Name, &role.Permissions, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert role: %w", err)
	}

	return &role, nil
}

func (s *Store) GetRoleByID(ctx context.Context, id string) (*Role, error) {
	var role Role
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, permissions, created_at, updated_at, deleted_at
		FROM roles
		WHERE id = $1
	`, id).Scan(
		&role.ID, &role.Name, &role.Permissions, &role.CreatedAt, &role.UpdatedAt, &role.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get role by id: %w", err)
	}

	return &role, nil
}

func (s *Store) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	var role Role
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, permissions, created_at, updated_at, deleted_at
		FROM roles
		WHERE name = $1
	`, name).Scan(
		&role.ID, &role.Name, &role.Permissions, &role.CreatedAt, &role.UpdatedAt, &role.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get role by name: %w", err)
	}

	return &role, nil
}

func (s *Store) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, permissions, created_at, updated_at, deleted_at
		FROM roles
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(
			&role.ID, &role.Name, &role.Permissions, &role.CreatedAt, &role.UpdatedAt, &role.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}

func (s *Store) UpdateRole(ctx context.Context, id string, input UpdateRoleInput) (*Role, error) {
	var role Role
	err := s.pool.QueryRow(ctx, `
		UPDATE roles
		SET
			name = COALESCE($2, name),
			permissions = COALESCE($3, permissions),
			updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, name, permissions, created_at, updated_at
	`, id, input.Name, input.Permissions).Scan(
		&role.ID, &role.Name, &role.Permissions, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("update role: %w", err)
	}

	return &role, nil
}

func (s *Store) SoftDeleteRole(ctx context.Context, id string) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE roles SET deleted_at = $2, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`, id, time.Now())
	if err != nil {
		return fmt.Errorf("soft delete role: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}
	return nil
}

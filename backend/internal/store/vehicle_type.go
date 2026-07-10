package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	apperrors "github.com/thoriqzs/PARKIR/backend/internal/errors"
)

type VehicleType struct {
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateVehicleTypeInput struct {
	Name        string
	DisplayName string
	Description string
}

type UpdateVehicleTypeInput struct {
	DisplayName *string
	Description *string
}

func (s *Store) ListVehicleTypes(ctx context.Context) ([]VehicleType, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT name, display_name, description, created_at, updated_at
		FROM vehicle_types
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list vehicle types: %w", err)
	}
	defer rows.Close()

	var types []VehicleType
	for rows.Next() {
		var vt VehicleType
		if err := rows.Scan(&vt.Name, &vt.DisplayName, &vt.Description, &vt.CreatedAt, &vt.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan vehicle type: %w", err)
		}
		types = append(types, vt)
	}
	return types, rows.Err()
}

func (s *Store) GetVehicleType(ctx context.Context, name string) (*VehicleType, error) {
	var vt VehicleType
	err := s.pool.QueryRow(ctx, `
		SELECT name, display_name, description, created_at, updated_at
		FROM vehicle_types
		WHERE name = $1
	`, name).Scan(&vt.Name, &vt.DisplayName, &vt.Description, &vt.CreatedAt, &vt.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("get vehicle type: %w", err)
	}
	return &vt, nil
}

func (s *Store) CreateVehicleType(ctx context.Context, input CreateVehicleTypeInput) (*VehicleType, error) {
	input.Name = strings.ToUpper(strings.TrimSpace(input.Name))

	var vt VehicleType
	err := s.pool.QueryRow(ctx, `
		INSERT INTO vehicle_types (name, display_name, description)
		VALUES ($1, $2, $3)
		RETURNING name, display_name, description, created_at, updated_at
	`, input.Name, input.DisplayName, input.Description).Scan(
		&vt.Name, &vt.DisplayName, &vt.Description, &vt.CreatedAt, &vt.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperrors.ErrConflict
		}
		return nil, fmt.Errorf("create vehicle type: %w", err)
	}
	return &vt, nil
}

func (s *Store) UpdateVehicleType(ctx context.Context, name string, input UpdateVehicleTypeInput) (*VehicleType, error) {
	var vt VehicleType
	err := s.pool.QueryRow(ctx, `
		UPDATE vehicle_types
		SET
			display_name = COALESCE($2, display_name),
			description  = COALESCE($3, description)
		WHERE name = $1
		RETURNING name, display_name, description, created_at, updated_at
	`, name, input.DisplayName, input.Description).Scan(
		&vt.Name, &vt.DisplayName, &vt.Description, &vt.CreatedAt, &vt.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("update vehicle type: %w", err)
	}
	return &vt, nil
}

func (s *Store) DeleteVehicleType(ctx context.Context, name string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM vehicle_types WHERE name = $1`, name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return apperrors.ErrInUse
		}
		return fmt.Errorf("delete vehicle type: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (s *Store) ValidateVehicleTypeExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM vehicle_types WHERE name = $1)`, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("validate vehicle type: %w", err)
	}
	return exists, nil
}

package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
)

type Location struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Code      string                 `json:"code"`
	Address   string                 `json:"address,omitempty"`
	City      string                 `json:"city,omitempty"`
	Status    string                 `json:"status"`
	Capacity  map[string]interface{} `json:"capacity,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type CreateLocationInput struct {
	Name     string
	Code     string
	Address  string
	City     string
	Capacity map[string]interface{}
}

type UpdateLocationInput struct {
	Name     *string
	Address  *string
	City     *string
	Status   *string
	Capacity map[string]interface{}
}

func (s *Store) CreateLocation(ctx context.Context, input CreateLocationInput) (*Location, error) {
	var loc Location
	err := s.pool.QueryRow(ctx, `
		INSERT INTO locations (name, code, address, city, capacity)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, code, address, city, status, capacity, created_at, updated_at
	`, input.Name, input.Code, input.Address, input.City, input.Capacity).Scan(
		&loc.ID, &loc.Name, &loc.Code, &loc.Address, &loc.City, &loc.Status, &loc.Capacity, &loc.CreatedAt, &loc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert location: %w", err)
	}

	return &loc, nil
}

func (s *Store) GetLocationByID(ctx context.Context, id string) (*Location, error) {
	var loc Location
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, code, address, city, status, capacity, created_at, updated_at
		FROM locations
		WHERE id = $1
	`, id).Scan(
		&loc.ID, &loc.Name, &loc.Code, &loc.Address, &loc.City, &loc.Status, &loc.Capacity, &loc.CreatedAt, &loc.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get location by id: %w", err)
	}

	return &loc, nil
}

func (s *Store) ListLocations(ctx context.Context) ([]Location, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, code, address, city, status, capacity, created_at, updated_at
		FROM locations
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list locations: %w", err)
	}
	defer rows.Close()

	var locations []Location
	for rows.Next() {
		var loc Location
		if err := rows.Scan(
			&loc.ID, &loc.Name, &loc.Code, &loc.Address, &loc.City, &loc.Status, &loc.Capacity, &loc.CreatedAt, &loc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}
		locations = append(locations, loc)
	}

	return locations, rows.Err()
}

func (s *Store) UpdateLocation(ctx context.Context, id string, input UpdateLocationInput) (*Location, error) {
	var loc Location
	err := s.pool.QueryRow(ctx, `
		UPDATE locations
		SET
			name = COALESCE($2, name),
			address = COALESCE($3, address),
			city = COALESCE($4, city),
			status = COALESCE($5, status),
			capacity = COALESCE($6, capacity),
			updated_at = now()
		WHERE id = $1
		RETURNING id, name, code, address, city, status, capacity, created_at, updated_at
	`, id, input.Name, input.Address, input.City, input.Status, input.Capacity).Scan(
		&loc.ID, &loc.Name, &loc.Code, &loc.Address, &loc.City, &loc.Status, &loc.Capacity, &loc.CreatedAt, &loc.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("update location: %w", err)
	}

	return &loc, nil
}

func (s *Store) DeactivateLocation(ctx context.Context, id string) (*Location, error) {
	var loc Location
	err := s.pool.QueryRow(ctx, `
		UPDATE locations
		SET status = 'INACTIVE', updated_at = now()
		WHERE id = $1
		RETURNING id, name, code, address, city, status, capacity, created_at, updated_at
	`, id).Scan(
		&loc.ID, &loc.Name, &loc.Code, &loc.Address, &loc.City, &loc.Status, &loc.Capacity, &loc.CreatedAt, &loc.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("deactivate location: %w", err)
	}

	return &loc, nil
}

func (s *Store) AssignOperatorToLocation(ctx context.Context, locationID, userID string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO user_role_locations (user_id, location_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, locationID)
	if err != nil {
		return fmt.Errorf("assign operator: %w", err)
	}
	return nil
}

func (s *Store) RemoveOperatorFromLocation(ctx context.Context, locationID, userID string) error {
	result, err := s.pool.Exec(ctx, `
		DELETE FROM user_role_locations
		WHERE user_id = $1 AND location_id = $2
	`, userID, locationID)
	if err != nil {
		return fmt.Errorf("remove operator: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}
	return nil
}

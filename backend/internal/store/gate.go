package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	apperrors "github.com/thoriqzs/PARKIR/backend/internal/errors"
)

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
	Rates    []RateSummary   `json:"rates"`
	Capacity map[string]int64 `json:"capacity"`
}

type RateSummary struct {
	VehicleType          string  `json:"vehicle_type"`
	FirstHourRate        float64 `json:"first_hour_rate"`
	SubsequentHourlyRate float64 `json:"subsequent_hourly_rate"`
	DailyFlatRate        float64 `json:"daily_flat_rate"`
}

func (s *Store) RegisterGate(ctx context.Context, input RegisterGateInput) (*Gate, error) {
	var g Gate
	err := s.pool.QueryRow(ctx, `
		INSERT INTO gates (device_id, name, location_id, ip_address)
		VALUES ($1, $2, $3, $4)
		RETURNING id, device_id, name, location_id, ip_address, last_seen_at, registered_at, created_at, updated_at
	`, input.DeviceID, input.Name, input.LocationID, input.IPAddress).Scan(
		&g.ID, &g.DeviceID, &g.Name, &g.LocationID, &g.IPAddress,
		&g.LastSeenAt, &g.RegisteredAt, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperrors.ErrConflict
		}
		return nil, fmt.Errorf("register gate: %w", err)
	}
	return &g, nil
}

func (s *Store) GetGateByID(ctx context.Context, id string) (*Gate, error) {
	var g Gate
	err := s.pool.QueryRow(ctx, `
		SELECT id, device_id, name, location_id, ip_address, last_seen_at, registered_at, created_at, updated_at
		FROM gates WHERE id = $1
	`, id).Scan(
		&g.ID, &g.DeviceID, &g.Name, &g.LocationID, &g.IPAddress,
		&g.LastSeenAt, &g.RegisteredAt, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("get gate by id: %w", err)
	}
	return &g, nil
}

func (s *Store) GetGateByDeviceID(ctx context.Context, deviceID string) (*Gate, error) {
	var g Gate
	err := s.pool.QueryRow(ctx, `
		SELECT id, device_id, name, location_id, ip_address, last_seen_at, registered_at, created_at, updated_at
		FROM gates WHERE device_id = $1
	`, deviceID).Scan(
		&g.ID, &g.DeviceID, &g.Name, &g.LocationID, &g.IPAddress,
		&g.LastSeenAt, &g.RegisteredAt, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("get gate by device id: %w", err)
	}
	return &g, nil
}

func (s *Store) ListGates(ctx context.Context, locationID string) ([]Gate, error) {
	var rows pgx.Rows
	var err error

	if locationID != "" {
		rows, err = s.pool.Query(ctx, `
			SELECT id, device_id, name, location_id, ip_address, last_seen_at, registered_at, created_at, updated_at
			FROM gates
			WHERE location_id = $1
			ORDER BY name ASC
		`, locationID)
	} else {
		rows, err = s.pool.Query(ctx, `
			SELECT id, device_id, name, location_id, ip_address, last_seen_at, registered_at, created_at, updated_at
			FROM gates
			ORDER BY name ASC
		`)
	}
	if err != nil {
		return nil, fmt.Errorf("list gates: %w", err)
	}
	defer rows.Close()

	var gates []Gate
	for rows.Next() {
		var g Gate
		if err := rows.Scan(
			&g.ID, &g.DeviceID, &g.Name, &g.LocationID, &g.IPAddress,
			&g.LastSeenAt, &g.RegisteredAt, &g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan gate: %w", err)
		}
		gates = append(gates, g)
	}
	return gates, rows.Err()
}

func (s *Store) UpdateGate(ctx context.Context, id string, input UpdateGateInput) (*Gate, error) {
	var g Gate
	err := s.pool.QueryRow(ctx, `
		UPDATE gates
		SET
			name       = COALESCE($2, name),
			location_id = COALESCE($3, location_id),
			ip_address = COALESCE($4, ip_address)
		WHERE id = $1
		RETURNING id, device_id, name, location_id, ip_address, last_seen_at, registered_at, created_at, updated_at
	`, id, input.Name, input.LocationID, input.IPAddress).Scan(
		&g.ID, &g.DeviceID, &g.Name, &g.LocationID, &g.IPAddress,
		&g.LastSeenAt, &g.RegisteredAt, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("update gate: %w", err)
	}
	return &g, nil
}

func (s *Store) DeleteGate(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM gates WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete gate: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (s *Store) GetGateInfo(ctx context.Context, locationID string) (*GateInfo, error) {
	info := &GateInfo{
		Rates: make([]RateSummary, 0),
	}

	err := s.pool.QueryRow(ctx, `
		SELECT name, code, COALESCE(capacity, '{}'::jsonb)
		FROM locations
		WHERE id = $1
	`, locationID).Scan(&info.Location.Name, &info.Location.Code, &info.Capacity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("get gate info location: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT vehicle_type, first_hour_rate, subsequent_hourly_rate, daily_flat_rate
		FROM location_rates
		WHERE location_id = $1
		  AND effective_from <= CURRENT_DATE
		  AND (effective_until IS NULL OR effective_until >= CURRENT_DATE)
		ORDER BY vehicle_type ASC
	`, locationID)
	if err != nil {
		return nil, fmt.Errorf("get gate info rates: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r RateSummary
		if err := rows.Scan(&r.VehicleType, &r.FirstHourRate, &r.SubsequentHourlyRate, &r.DailyFlatRate); err != nil {
			return nil, fmt.Errorf("scan rate summary: %w", err)
		}
		info.Rates = append(info.Rates, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gate info rates rows: %w", err)
	}

	return info, nil
}

package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
)

type Rate struct {
	ID                    string     `json:"id"`
	LocationID            string     `json:"location_id"`
	VehicleType           string     `json:"vehicle_type"`
	FirstHourRate         float64    `json:"first_hour_rate"`
	SubsequentHourlyRate  float64    `json:"subsequent_hourly_rate"`
	DailyFlatRate         float64    `json:"daily_flat_rate"`
	EffectiveFrom         time.Time  `json:"effective_from"`
	EffectiveUntil        *time.Time `json:"effective_until,omitempty"`
	CreatedBy             *string    `json:"created_by,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
}

type CreateRateInput struct {
	LocationID           string
	VehicleType          string
	FirstHourRate        float64
	SubsequentHourlyRate float64
	DailyFlatRate        float64
	EffectiveFrom        time.Time
	EffectiveUntil       *time.Time
	CreatedBy            string
}

type UpdateRateInput struct {
	FirstHourRate        *float64
	SubsequentHourlyRate *float64
	DailyFlatRate        *float64
	EffectiveUntil       *time.Time
}

func (s *Store) CreateRate(ctx context.Context, input CreateRateInput) (*Rate, error) {
	var rate Rate
	err := s.pool.QueryRow(ctx, `
		INSERT INTO location_rates (
			location_id, vehicle_type, first_hour_rate, subsequent_hourly_rate,
			daily_flat_rate, effective_from, effective_until, created_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, location_id, vehicle_type, first_hour_rate, subsequent_hourly_rate,
		          daily_flat_rate, effective_from, effective_until, created_by, created_at
	`, input.LocationID, input.VehicleType, input.FirstHourRate, input.SubsequentHourlyRate,
		input.DailyFlatRate, input.EffectiveFrom, input.EffectiveUntil, input.CreatedBy).Scan(
		&rate.ID, &rate.LocationID, &rate.VehicleType, &rate.FirstHourRate, &rate.SubsequentHourlyRate,
		&rate.DailyFlatRate, &rate.EffectiveFrom, &rate.EffectiveUntil, &rate.CreatedBy, &rate.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert rate: %w", err)
	}

	return &rate, nil
}

func (s *Store) GetRateByID(ctx context.Context, id string) (*Rate, error) {
	var rate Rate
	err := s.pool.QueryRow(ctx, `
		SELECT id, location_id, vehicle_type, first_hour_rate, subsequent_hourly_rate,
		       daily_flat_rate, effective_from, effective_until, created_by, created_at
		FROM location_rates
		WHERE id = $1
	`, id).Scan(
		&rate.ID, &rate.LocationID, &rate.VehicleType, &rate.FirstHourRate, &rate.SubsequentHourlyRate,
		&rate.DailyFlatRate, &rate.EffectiveFrom, &rate.EffectiveUntil, &rate.CreatedBy, &rate.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get rate by id: %w", err)
	}

	return &rate, nil
}

func (s *Store) ListRatesByLocation(ctx context.Context, locationID string) ([]Rate, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, location_id, vehicle_type, first_hour_rate, subsequent_hourly_rate,
		       daily_flat_rate, effective_from, effective_until, created_by, created_at
		FROM location_rates
		WHERE location_id = $1
		ORDER BY vehicle_type, effective_from DESC
	`, locationID)
	if err != nil {
		return nil, fmt.Errorf("list rates: %w", err)
	}
	defer rows.Close()

	var rates []Rate
	for rows.Next() {
		var rate Rate
		if err := rows.Scan(
			&rate.ID, &rate.LocationID, &rate.VehicleType, &rate.FirstHourRate, &rate.SubsequentHourlyRate,
			&rate.DailyFlatRate, &rate.EffectiveFrom, &rate.EffectiveUntil, &rate.CreatedBy, &rate.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan rate: %w", err)
		}
		rates = append(rates, rate)
	}

	return rates, rows.Err()
}

func (s *Store) UpdateRate(ctx context.Context, id string, input UpdateRateInput) (*Rate, error) {
	var rate Rate
	err := s.pool.QueryRow(ctx, `
		UPDATE location_rates
		SET
			first_hour_rate = COALESCE($2, first_hour_rate),
			subsequent_hourly_rate = COALESCE($3, subsequent_hourly_rate),
			daily_flat_rate = COALESCE($4, daily_flat_rate),
			effective_until = COALESCE($5, effective_until),
			updated_at = now()
		WHERE id = $1
		RETURNING id, location_id, vehicle_type, first_hour_rate, subsequent_hourly_rate,
		          daily_flat_rate, effective_from, effective_until, created_by, created_at
	`, id, input.FirstHourRate, input.SubsequentHourlyRate, input.DailyFlatRate, input.EffectiveUntil).Scan(
		&rate.ID, &rate.LocationID, &rate.VehicleType, &rate.FirstHourRate, &rate.SubsequentHourlyRate,
		&rate.DailyFlatRate, &rate.EffectiveFrom, &rate.EffectiveUntil, &rate.CreatedBy, &rate.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("update rate: %w", err)
	}

	return &rate, nil
}

func (s *Store) GetActiveRate(ctx context.Context, locationID, vehicleType string, checkInDate time.Time) (*Rate, error) {
	var rate Rate
	err := s.pool.QueryRow(ctx, `
		SELECT id, location_id, vehicle_type, first_hour_rate, subsequent_hourly_rate,
		       daily_flat_rate, effective_from, effective_until, created_by, created_at
		FROM location_rates
		WHERE location_id = $1
		  AND vehicle_type = $2
		  AND effective_from <= $3
		  AND (effective_until IS NULL OR effective_until >= $3)
		ORDER BY effective_from DESC
		LIMIT 1
	`, locationID, vehicleType, checkInDate).Scan(
		&rate.ID, &rate.LocationID, &rate.VehicleType, &rate.FirstHourRate, &rate.SubsequentHourlyRate,
		&rate.DailyFlatRate, &rate.EffectiveFrom, &rate.EffectiveUntil, &rate.CreatedBy, &rate.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get active rate: %w", err)
	}

	return &rate, nil
}

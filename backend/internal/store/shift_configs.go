package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
)

type LocationShiftConfig struct {
	ID           string    `json:"id"`
	LocationID   string    `json:"location_id"`
	ShiftCode    string    `json:"shift_code"`
	ShiftNumber  int       `json:"shift_number"`
	StartTime    string    `json:"start_time"` // HH:MM:SS format
	EndTime      string    `json:"end_time"`   // HH:MM:SS format
	IsOvernight  bool      `json:"is_overnight"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateLocationShiftConfigInput struct {
	LocationID  string
	ShiftCode   string
	ShiftNumber int
	StartTime   string // HH:MM:SS
	EndTime     string // HH:MM:SS
}

type UpdateLocationShiftConfigInput struct {
	ShiftCode   *string
	ShiftNumber *int
	StartTime   *string
	EndTime     *string
}

func (s *Store) CreateLocationShiftConfig(ctx context.Context, input CreateLocationShiftConfigInput) (*LocationShiftConfig, error) {
	// Parse times to check overnight
	start, err := time.Parse("15:04:05", input.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time format: %w", err)
	}
	end, err := time.Parse("15:04:05", input.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time format: %w", err)
	}
	isOvernight := end.Before(start)

	var config LocationShiftConfig
	err = s.pool.QueryRow(ctx, `
		INSERT INTO location_shift_configs (location_id, shift_code, shift_number, start_time, end_time, is_overnight)
		VALUES ($1, $2, $3, $4::time, $5::time, $6)
		RETURNING id, location_id, shift_code, shift_number, 
		          to_char(start_time, 'HH24:MI:SS'), to_char(end_time, 'HH24:MI:SS'), 
		          is_overnight, created_at, updated_at
	`, input.LocationID, input.ShiftCode, input.ShiftNumber, input.StartTime, input.EndTime, isOvernight).Scan(
		&config.ID, &config.LocationID, &config.ShiftCode, &config.ShiftNumber,
		&config.StartTime, &config.EndTime, &config.IsOvernight, &config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create shift config: %w", err)
	}
	return &config, nil
}

func (s *Store) GetLocationShiftConfigByID(ctx context.Context, id string) (*LocationShiftConfig, error) {
	var config LocationShiftConfig
	err := s.pool.QueryRow(ctx, `
		SELECT id, location_id, shift_code, shift_number, 
		       to_char(start_time, 'HH24:MI:SS'), to_char(end_time, 'HH24:MI:SS'),
		       is_overnight, created_at, updated_at
		FROM location_shift_configs
		WHERE id = $1
	`, id).Scan(
		&config.ID, &config.LocationID, &config.ShiftCode, &config.ShiftNumber,
		&config.StartTime, &config.EndTime, &config.IsOvernight, &config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get shift config: %w", err)
	}
	return &config, nil
}

func (s *Store) GetLocationShiftConfigByCode(ctx context.Context, locationID, shiftCode string) (*LocationShiftConfig, error) {
	var config LocationShiftConfig
	err := s.pool.QueryRow(ctx, `
		SELECT id, location_id, shift_code, shift_number, 
		       to_char(start_time, 'HH24:MI:SS'), to_char(end_time, 'HH24:MI:SS'),
		       is_overnight, created_at, updated_at
		FROM location_shift_configs
		WHERE location_id = $1 AND shift_code = $2
	`, locationID, shiftCode).Scan(
		&config.ID, &config.LocationID, &config.ShiftCode, &config.ShiftNumber,
		&config.StartTime, &config.EndTime, &config.IsOvernight, &config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get shift config by code: %w", err)
	}
	return &config, nil
}

func (s *Store) ListLocationShiftConfigs(ctx context.Context, locationID string) ([]LocationShiftConfig, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, location_id, shift_code, shift_number, 
		       to_char(start_time, 'HH24:MI:SS'), to_char(end_time, 'HH24:MI:SS'),
		       is_overnight, created_at, updated_at
		FROM location_shift_configs
		WHERE location_id = $1
		ORDER BY shift_number ASC
	`, locationID)
	if err != nil {
		return nil, fmt.Errorf("list shift configs: %w", err)
	}
	defer rows.Close()

	var configs []LocationShiftConfig
	for rows.Next() {
		var config LocationShiftConfig
		if err := rows.Scan(
			&config.ID, &config.LocationID, &config.ShiftCode, &config.ShiftNumber,
			&config.StartTime, &config.EndTime, &config.IsOvernight, &config.CreatedAt, &config.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan shift config: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, rows.Err()
}

func (s *Store) UpdateLocationShiftConfig(ctx context.Context, id string, input UpdateLocationShiftConfigInput) (*LocationShiftConfig, error) {
	// Build dynamic query
	setClauses := []string{}
	args := []interface{}{id}
	argIdx := 2

	if input.ShiftCode != nil {
		setClauses = append(setClauses, fmt.Sprintf("shift_code = $%d", argIdx))
		args = append(args, *input.ShiftCode)
		argIdx++
	}
	if input.ShiftNumber != nil {
		setClauses = append(setClauses, fmt.Sprintf("shift_number = $%d", argIdx))
		args = append(args, *input.ShiftNumber)
		argIdx++
	}
	if input.StartTime != nil {
		setClauses = append(setClauses, fmt.Sprintf("start_time = $%d::time", argIdx))
		args = append(args, *input.StartTime)
		argIdx++
	}
	if input.EndTime != nil {
		setClauses = append(setClauses, fmt.Sprintf("end_time = $%d::time", argIdx))
		args = append(args, *input.EndTime)
		argIdx++
	}

	// Update is_overnight if time changed
	if input.StartTime != nil || input.EndTime != nil {
		setClauses = append(setClauses, fmt.Sprintf(`is_overnight = (
			CASE 
				WHEN COALESCE($%d::time, start_time) > COALESCE($%d::time, end_time) 
				THEN true 
				ELSE false 
			END
		)`, argIdx, argIdx+1))
		if input.StartTime != nil {
			args = append(args, *input.StartTime)
		} else {
			args = append(args, nil)
		}
		if input.EndTime != nil {
			args = append(args, *input.EndTime)
		} else {
			args = append(args, nil)
		}
		argIdx += 2
	}

	if len(setClauses) == 0 {
		return s.GetLocationShiftConfigByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = now()")

	query := fmt.Sprintf(`
		UPDATE location_shift_configs
		SET %s
		WHERE id = $1
		RETURNING id, location_id, shift_code, shift_number, 
		          to_char(start_time, 'HH24:MI:SS'), to_char(end_time, 'HH24:MI:SS'),
		          is_overnight, created_at, updated_at
	`, stringJoin(setClauses, ", "))

	var config LocationShiftConfig
	err := s.pool.QueryRow(ctx, query, args...).Scan(
		&config.ID, &config.LocationID, &config.ShiftCode, &config.ShiftNumber,
		&config.StartTime, &config.EndTime, &config.IsOvernight, &config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("update shift config: %w", err)
	}
	return &config, nil
}

func (s *Store) DeleteLocationShiftConfig(ctx context.Context, id string) error {
	result, err := s.pool.Exec(ctx, `
		DELETE FROM location_shift_configs WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete shift config: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}
	return nil
}

// GetShiftConfigByTime finds the shift config that covers the given time
// Returns ErrNotFound if no matching shift config
func (s *Store) GetShiftConfigByTime(ctx context.Context, locationID string, t time.Time) (*LocationShiftConfig, error) {
	timeStr := t.Format("15:04:05")

	var config LocationShiftConfig
	err := s.pool.QueryRow(ctx, `
		SELECT id, location_id, shift_code, shift_number, 
		       to_char(start_time, 'HH24:MI:SS'), to_char(end_time, 'HH24:MI:SS'),
		       is_overnight, created_at, updated_at
		FROM location_shift_configs
		WHERE location_id = $1
		  AND (
			  -- Normal shift (not overnight)
			  (NOT is_overnight AND $2::time >= start_time AND $2::time < end_time)
			  OR
			  -- Overnight shift
			  (is_overnight AND ($2::time >= start_time OR $2::time < end_time))
		  )
		ORDER BY shift_number ASC
		LIMIT 1
	`, locationID, timeStr).Scan(
		&config.ID, &config.LocationID, &config.ShiftCode, &config.ShiftNumber,
		&config.StartTime, &config.EndTime, &config.IsOvernight, &config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get shift config by time: %w", err)
	}
	return &config, nil
}

// GetShiftConfigByTimeWithFallback finds shift config, falls back to nearest previous shift if no match
func (s *Store) GetShiftConfigByTimeWithFallback(ctx context.Context, locationID string, t time.Time) (*LocationShiftConfig, error) {
	// First try exact match
	config, err := s.GetShiftConfigByTime(ctx, locationID, t)
	if err == nil {
		return config, nil
	}
	if err != errors.ErrNotFound {
		return nil, err
	}

	// No exact match, find the nearest previous shift
	timeStr := t.Format("15:04:05")

	// Try to find the shift with the highest start_time that is <= current time
	// Or the last shift of the day if current time is before all shifts
	var fallbackConfig LocationShiftConfig
	err = s.pool.QueryRow(ctx, `
		SELECT id, location_id, shift_code, shift_number, 
		       to_char(start_time, 'HH24:MI:SS'), to_char(end_time, 'HH24:MI:SS'),
		       is_overnight, created_at, updated_at
		FROM location_shift_configs
		WHERE location_id = $1
		  AND start_time <= $2::time
		ORDER BY start_time DESC
		LIMIT 1
	`, locationID, timeStr).Scan(
		&fallbackConfig.ID, &fallbackConfig.LocationID, &fallbackConfig.ShiftCode, &fallbackConfig.ShiftNumber,
		&fallbackConfig.StartTime, &fallbackConfig.EndTime, &fallbackConfig.IsOvernight, &fallbackConfig.CreatedAt, &fallbackConfig.UpdatedAt,
	)
	if err == nil {
		return &fallbackConfig, nil
	}

	// If no shift with start_time <= current time, get the last shift of the day
	err = s.pool.QueryRow(ctx, `
		SELECT id, location_id, shift_code, shift_number, 
		       to_char(start_time, 'HH24:MI:SS'), to_char(end_time, 'HH24:MI:SS'),
		       is_overnight, created_at, updated_at
		FROM location_shift_configs
		WHERE location_id = $1
		ORDER BY start_time DESC
		LIMIT 1
	`, locationID).Scan(
		&fallbackConfig.ID, &fallbackConfig.LocationID, &fallbackConfig.ShiftCode, &fallbackConfig.ShiftNumber,
		&fallbackConfig.StartTime, &fallbackConfig.EndTime, &fallbackConfig.IsOvernight, &fallbackConfig.CreatedAt, &fallbackConfig.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get fallback shift config: %w", err)
	}
	return &fallbackConfig, nil
}

// CreateDefaultShiftConfigs creates the default 3 shifts for a location
func (s *Store) CreateDefaultShiftConfigs(ctx context.Context, locationID string) error {
	defaults := []CreateLocationShiftConfigInput{
		{LocationID: locationID, ShiftCode: "00-08", ShiftNumber: 1, StartTime: "00:00:00", EndTime: "08:00:00"},
		{LocationID: locationID, ShiftCode: "08-16", ShiftNumber: 2, StartTime: "08:00:00", EndTime: "16:00:00"},
		{LocationID: locationID, ShiftCode: "16-24", ShiftNumber: 3, StartTime: "16:00:00", EndTime: "23:59:59"},
	}

	for _, cfg := range defaults {
		_, err := s.CreateLocationShiftConfig(ctx, cfg)
		if err != nil {
			return fmt.Errorf("create default shift config %s: %w", cfg.ShiftCode, err)
		}
	}
	return nil
}

func stringJoin(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

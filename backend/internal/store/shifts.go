package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
)

// Shift represents a daily shift instance (not template)
type Shift struct {
	ID             string    `json:"id"`
	LocationID     string    `json:"location_id"`
	ShiftNumber    int       `json:"shift_number"`
	ShiftDate      time.Time `json:"shift_date"`
	VoidCount      int       `json:"void_count"`
	IncidentCount  int       `json:"incident_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// GetOrCreateShift gets existing shift or creates new one for given location, shift number, and date
func (s *Store) GetOrCreateShift(ctx context.Context, locationID string, shiftNumber int, shiftDate time.Time) (*Shift, error) {
	// Try to get existing first
	var shift Shift
	err := s.pool.QueryRow(ctx, `
		SELECT id, location_id, shift_number, shift_date, void_count, incident_count, created_at, updated_at
		FROM shifts
		WHERE location_id = $1 AND shift_number = $2 AND shift_date = $3
	`, locationID, shiftNumber, shiftDate).Scan(
		&shift.ID, &shift.LocationID, &shift.ShiftNumber, &shift.ShiftDate,
		&shift.VoidCount, &shift.IncidentCount, &shift.CreatedAt, &shift.UpdatedAt,
	)
	if err == nil {
		return &shift, nil
	}
	if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("get shift: %w", err)
	}

	// Create new shift instance
	err = s.pool.QueryRow(ctx, `
		INSERT INTO shifts (location_id, shift_number, shift_date)
		VALUES ($1, $2, $3)
		RETURNING id, location_id, shift_number, shift_date, void_count, incident_count, created_at, updated_at
	`, locationID, shiftNumber, shiftDate).Scan(
		&shift.ID, &shift.LocationID, &shift.ShiftNumber, &shift.ShiftDate,
		&shift.VoidCount, &shift.IncidentCount, &shift.CreatedAt, &shift.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create shift: %w", err)
	}
	return &shift, nil
}

func (s *Store) GetShiftByID(ctx context.Context, id string) (*Shift, error) {
	var shift Shift
	err := s.pool.QueryRow(ctx, `
		SELECT id, location_id, shift_number, shift_date, void_count, incident_count, created_at, updated_at
		FROM shifts
		WHERE id = $1
	`, id).Scan(
		&shift.ID, &shift.LocationID, &shift.ShiftNumber, &shift.ShiftDate,
		&shift.VoidCount, &shift.IncidentCount, &shift.CreatedAt, &shift.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get shift: %w", err)
	}
	return &shift, nil
}

type ListShiftsFilters struct {
	LocationID   string
	ShiftNumber  *int
	ShiftDate    *time.Time
	DateFrom     *time.Time
	DateTo       *time.Time
}

func (s *Store) ListShifts(ctx context.Context, filters ListShiftsFilters, limit, offset int) ([]Shift, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filters.LocationID != "" {
		where += fmt.Sprintf(" AND location_id = $%d", argIdx)
		args = append(args, filters.LocationID)
		argIdx++
	}
	if filters.ShiftNumber != nil {
		where += fmt.Sprintf(" AND shift_number = $%d", argIdx)
		args = append(args, *filters.ShiftNumber)
		argIdx++
	}
	if filters.ShiftDate != nil {
		where += fmt.Sprintf(" AND shift_date = $%d", argIdx)
		args = append(args, *filters.ShiftDate)
		argIdx++
	}
	if filters.DateFrom != nil {
		where += fmt.Sprintf(" AND shift_date >= $%d", argIdx)
		args = append(args, *filters.DateFrom)
		argIdx++
	}
	if filters.DateTo != nil {
		where += fmt.Sprintf(" AND shift_date <= $%d", argIdx)
		args = append(args, *filters.DateTo)
		argIdx++
	}

	countArgs := append([]interface{}{}, args...)
	var total int
	countQuery := "SELECT COUNT(*) FROM shifts " + where
	if err := s.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count shifts: %w", err)
	}

	query := "SELECT id, location_id, shift_number, shift_date, void_count, incident_count, created_at, updated_at " +
		"FROM shifts " + where +
		fmt.Sprintf(" ORDER BY shift_date DESC, shift_number ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list shifts: %w", err)
	}
	defer rows.Close()

	var shifts []Shift
	for rows.Next() {
		var shift Shift
		if err := rows.Scan(
			&shift.ID, &shift.LocationID, &shift.ShiftNumber, &shift.ShiftDate,
			&shift.VoidCount, &shift.IncidentCount, &shift.CreatedAt, &shift.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan shift: %w", err)
		}
		shifts = append(shifts, shift)
	}

	return shifts, total, rows.Err()
}

// GetCurrentShift returns the shift instance for the current time
func (s *Store) GetCurrentShift(ctx context.Context, locationID string) (*Shift, error) {
	now := time.Now()
	
	// Get shift config for current time
	config, err := s.GetShiftConfigByTimeWithFallback(ctx, locationID, now)
	if err != nil {
		return nil, err
	}

	// Determine shift date (today, unless overnight shift and current time is before end_time)
	shiftDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if config.IsOvernight {
		endTime, _ := time.Parse("15:04:05", config.EndTime)
		if now.Hour() < endTime.Hour() || (now.Hour() == endTime.Hour() && now.Minute() < endTime.Minute()) {
			// Current time is in the "next day" part of overnight shift
			shiftDate = shiftDate.AddDate(0, 0, -1)
		}
	}

	return s.GetOrCreateShift(ctx, locationID, config.ShiftNumber, shiftDate)
}

// IncrementVoidCount increments the void count for a shift
func (s *Store) IncrementVoidCount(ctx context.Context, shiftID string) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE shifts
		SET void_count = void_count + 1,
		    updated_at = now()
		WHERE id = $1
	`, shiftID)
	if err != nil {
		return fmt.Errorf("increment void count: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}
	return nil
}

// IncrementIncidentCount increments the incident count for a shift
func (s *Store) IncrementIncidentCount(ctx context.Context, shiftID string) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE shifts
		SET incident_count = incident_count + 1,
		    updated_at = now()
		WHERE id = $1
	`, shiftID)
	if err != nil {
		return fmt.Errorf("increment incident count: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}
	return nil
}

// GetShiftSummary returns summary statistics for a shift
func (s *Store) GetShiftSummary(ctx context.Context, shiftID string) (map[string]interface{}, error) {
	// Get shift details
	shift, err := s.GetShiftByID(ctx, shiftID)
	if err != nil {
		return nil, err
	}

	// Count sessions
	var sessionCount int
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM sessions WHERE shift_id = $1
	`, shiftID).Scan(&sessionCount)
	if err != nil {
		return nil, fmt.Errorf("count sessions: %w", err)
	}

	// Sum revenue from transactions
	var totalRevenue float64
	err = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(fee_amount), 0)
		FROM transactions
		WHERE shift_id = $1 AND voided = false
	`, shiftID).Scan(&totalRevenue)
	if err != nil {
		return nil, fmt.Errorf("sum revenue: %w", err)
	}

	return map[string]interface{}{
		"shift_id":       shift.ID,
		"location_id":    shift.LocationID,
		"shift_number":   shift.ShiftNumber,
		"shift_date":     shift.ShiftDate,
		"session_count":  sessionCount,
		"void_count":     shift.VoidCount,
		"incident_count": shift.IncidentCount,
		"total_revenue":  totalRevenue,
	}, nil
}

// DEPRECATED: These methods are kept for backwards compatibility during migration
// They will be removed once all code is updated

func (s *Store) GetOpenShiftForOperator(ctx context.Context, operatorID string) (*Shift, error) {
	return nil, errors.ErrNotFound
}

func (s *Store) StartShift(ctx context.Context, input interface{}) (*Shift, error) {
	return nil, fmt.Errorf("start shift is deprecated, use auto-shift detection")
}

func (s *Store) CloseShift(ctx context.Context, id string, input interface{}) (*Shift, error) {
	return nil, fmt.Errorf("close shift is deprecated")
}

func (s *Store) ForceCloseShift(ctx context.Context, id string, input interface{}) (*Shift, error) {
	return nil, fmt.Errorf("force close shift is deprecated")
}

func (s *Store) UpdateShiftExpectedCash(ctx context.Context, shiftID string, expectedCash float64) (*Shift, error) {
	return nil, fmt.Errorf("update expected cash is deprecated")
}

func (s *Store) CalculateExpectedCashForShift(ctx context.Context, shiftID string) (float64, error) {
	var total float64
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(fee_amount), 0)
		FROM transactions
		WHERE shift_id = $1 AND voided = false AND payment_method = 'CASH'
	`, shiftID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("calculate expected cash: %w", err)
	}
	return total, nil
}

func (s *Store) SumCashByShift(ctx context.Context, shiftID string) (float64, error) {
	return s.CalculateExpectedCashForShift(ctx, shiftID)
}

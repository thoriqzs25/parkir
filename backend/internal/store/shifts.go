package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
)

type Shift struct {
	ID                   string     `json:"id"`
	OperatorID           string     `json:"operator_id"`
	LocationID           string     `json:"location_id"`
	Status               string     `json:"status"`
	StartedAt            time.Time  `json:"started_at"`
	EndedAt              *time.Time `json:"ended_at,omitempty"`
	ExpectedCash         *float64   `json:"expected_cash,omitempty"`
	CashHandoverAmount   *float64   `json:"cash_handover_amount,omitempty"`
	Discrepancy          *float64   `json:"discrepancy,omitempty"`
	DiscrepancyNotes     *string    `json:"discrepancy_notes,omitempty"`
	ForceClosedBy        *string    `json:"force_closed_by,omitempty"`
	ForceClosedReason    *string    `json:"force_closed_reason,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type StartShiftInput struct {
	OperatorID string
	LocationID string
}

type EndShiftInput struct {
	CashHandoverAmount float64
	DiscrepancyNotes   *string
}

type ForceCloseShiftInput struct {
	ForceClosedBy   string
	ForceClosedReason string
}

func (s *Store) GetOpenShiftForOperator(ctx context.Context, operatorID string) (*Shift, error) {
	var shift Shift
	err := s.pool.QueryRow(ctx, `
		SELECT id, operator_id, location_id, status, started_at, ended_at,
		       expected_cash, cash_handover_amount, discrepancy, discrepancy_notes,
		       force_closed_by, force_closed_reason, created_at, updated_at
		FROM shifts
		WHERE operator_id = $1 AND status = 'OPEN'
		LIMIT 1
	`, operatorID).Scan(
		&shift.ID, &shift.OperatorID, &shift.LocationID, &shift.Status, &shift.StartedAt, &shift.EndedAt,
		&shift.ExpectedCash, &shift.CashHandoverAmount, &shift.Discrepancy, &shift.DiscrepancyNotes,
		&shift.ForceClosedBy, &shift.ForceClosedReason, &shift.CreatedAt, &shift.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get open shift: %w", err)
	}
	return &shift, nil
}

func (s *Store) CloseShift(ctx context.Context, id string, input EndShiftInput) (*Shift, error) {
	var shift Shift
	err := s.pool.QueryRow(ctx, `
		UPDATE shifts
		SET status = CASE
				WHEN $2 = expected_cash THEN 'CLOSED'
				ELSE 'FLAGGED'
			END,
		    ended_at = now(),
		    cash_handover_amount = $2,
		    discrepancy = $2 - expected_cash,
		    discrepancy_notes = $3,
		    updated_at = now()
		WHERE id = $1
		RETURNING id, operator_id, location_id, status, started_at, ended_at,
		          expected_cash, cash_handover_amount, discrepancy, discrepancy_notes,
		          force_closed_by, force_closed_reason, created_at, updated_at
	`, id, input.CashHandoverAmount, input.DiscrepancyNotes).Scan(
		&shift.ID, &shift.OperatorID, &shift.LocationID, &shift.Status, &shift.StartedAt, &shift.EndedAt,
		&shift.ExpectedCash, &shift.CashHandoverAmount, &shift.Discrepancy, &shift.DiscrepancyNotes,
		&shift.ForceClosedBy, &shift.ForceClosedReason, &shift.CreatedAt, &shift.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("close shift: %w", err)
	}
	return &shift, nil
}

func (s *Store) ForceCloseShift(ctx context.Context, id string, input ForceCloseShiftInput) (*Shift, error) {
	var shift Shift
	err := s.pool.QueryRow(ctx, `
		UPDATE shifts
		SET status = 'FORCE_CLOSED',
		    ended_at = now(),
		    force_closed_by = $2,
		    force_closed_reason = $3,
		    updated_at = now()
		WHERE id = $1
		RETURNING id, operator_id, location_id, status, started_at, ended_at,
		          expected_cash, cash_handover_amount, discrepancy, discrepancy_notes,
		          force_closed_by, force_closed_reason, created_at, updated_at
	`, id, input.ForceClosedBy, input.ForceClosedReason).Scan(
		&shift.ID, &shift.OperatorID, &shift.LocationID, &shift.Status, &shift.StartedAt, &shift.EndedAt,
		&shift.ExpectedCash, &shift.CashHandoverAmount, &shift.Discrepancy, &shift.DiscrepancyNotes,
		&shift.ForceClosedBy, &shift.ForceClosedReason, &shift.CreatedAt, &shift.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("force close shift: %w", err)
	}
	return &shift, nil
}

func (s *Store) StartShift(ctx context.Context, input StartShiftInput) (*Shift, error) {
	var shift Shift
	err := s.pool.QueryRow(ctx, `
		INSERT INTO shifts (operator_id, location_id, status, started_at)
		VALUES ($1, $2, 'OPEN', now())
		RETURNING id, operator_id, location_id, status, started_at, ended_at,
		          expected_cash, cash_handover_amount, discrepancy, discrepancy_notes,
		          force_closed_by, force_closed_reason, created_at, updated_at
	`, input.OperatorID, input.LocationID).Scan(
		&shift.ID, &shift.OperatorID, &shift.LocationID, &shift.Status, &shift.StartedAt, &shift.EndedAt,
		&shift.ExpectedCash, &shift.CashHandoverAmount, &shift.Discrepancy, &shift.DiscrepancyNotes,
		&shift.ForceClosedBy, &shift.ForceClosedReason, &shift.CreatedAt, &shift.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("start shift: %w", err)
	}
	return &shift, nil
}

func (s *Store) GetShiftByID(ctx context.Context, id string) (*Shift, error) {
	var shift Shift
	err := s.pool.QueryRow(ctx, `
		SELECT id, operator_id, location_id, status, started_at, ended_at,
		       expected_cash, cash_handover_amount, discrepancy, discrepancy_notes,
		       force_closed_by, force_closed_reason, created_at, updated_at
		FROM shifts
		WHERE id = $1
	`, id).Scan(
		&shift.ID, &shift.OperatorID, &shift.LocationID, &shift.Status, &shift.StartedAt, &shift.EndedAt,
		&shift.ExpectedCash, &shift.CashHandoverAmount, &shift.Discrepancy, &shift.DiscrepancyNotes,
		&shift.ForceClosedBy, &shift.ForceClosedReason, &shift.CreatedAt, &shift.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get shift: %w", err)
	}
	return &shift, nil
}

func (s *Store) ListShifts(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]Shift, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if locID, ok := filters["location_id"].(string); ok && locID != "" {
		where += fmt.Sprintf(" AND location_id = $%d", argIdx)
		args = append(args, locID)
		argIdx++
	}
	if opID, ok := filters["operator_id"].(string); ok && opID != "" {
		where += fmt.Sprintf(" AND operator_id = $%d", argIdx)
		args = append(args, opID)
		argIdx++
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if from, ok := filters["date_from"].(time.Time); ok {
		where += fmt.Sprintf(" AND started_at >= $%d", argIdx)
		args = append(args, from)
		argIdx++
	}
	if to, ok := filters["date_to"].(time.Time); ok {
		where += fmt.Sprintf(" AND started_at < $%d", argIdx)
		args = append(args, to)
		argIdx++
	}

	countArgs := append([]interface{}{}, args...)
	var total int
	countQuery := "SELECT COUNT(*) FROM shifts " + where
	if err := s.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count shifts: %w", err)
	}

	query := "SELECT id, operator_id, location_id, status, started_at, ended_at, " +
		"expected_cash, cash_handover_amount, discrepancy, discrepancy_notes, " +
		"force_closed_by, force_closed_reason, created_at, updated_at " +
		"FROM shifts " + where +
		fmt.Sprintf(" ORDER BY started_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
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
			&shift.ID, &shift.OperatorID, &shift.LocationID, &shift.Status, &shift.StartedAt, &shift.EndedAt,
			&shift.ExpectedCash, &shift.CashHandoverAmount, &shift.Discrepancy, &shift.DiscrepancyNotes,
			&shift.ForceClosedBy, &shift.ForceClosedReason, &shift.CreatedAt, &shift.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan shift: %w", err)
		}
		shifts = append(shifts, shift)
	}

	return shifts, total, rows.Err()
}

func (s *Store) UpdateShiftExpectedCash(ctx context.Context, shiftID string, expectedCash float64) (*Shift, error) {
	var shift Shift
	err := s.pool.QueryRow(ctx, `
		UPDATE shifts
		SET expected_cash = $2,
		    updated_at = now()
		WHERE id = $1
		RETURNING id, operator_id, location_id, status, started_at, ended_at,
		          expected_cash, cash_handover_amount, discrepancy, discrepancy_notes,
		          force_closed_by, force_closed_reason, created_at, updated_at
	`, shiftID, expectedCash).Scan(
		&shift.ID, &shift.OperatorID, &shift.LocationID, &shift.Status, &shift.StartedAt, &shift.EndedAt,
		&shift.ExpectedCash, &shift.CashHandoverAmount, &shift.Discrepancy, &shift.DiscrepancyNotes,
		&shift.ForceClosedBy, &shift.ForceClosedReason, &shift.CreatedAt, &shift.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("update expected cash: %w", err)
	}
	return &shift, nil
}

func (s *Store) CalculateExpectedCashForShift(ctx context.Context, shiftID string) (float64, error) {
	var expected *float64
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(fee_amount), 0)
		FROM transactions
		WHERE shift_id = $1 AND voided = false AND payment_method = 'CASH'
	`, shiftID).Scan(&expected)
	if err != nil {
		return 0, fmt.Errorf("calculate expected cash: %w", err)
	}
	if expected == nil {
		return 0, nil
	}
	return *expected, nil
}

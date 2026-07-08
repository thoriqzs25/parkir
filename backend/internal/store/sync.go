package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
)

// CreateOfflineSessionInput captures a session created by the desktop app while offline.
type CreateOfflineSessionInput struct {
	ID          string
	LocationID  string
	OperatorID  string
	ShiftID     string
	Plate       string
	CityCode    string
	VehicleType string
	CheckInAt   time.Time
}

// CreateOfflineSession inserts a session that was created offline. It is idempotent
// by session ID and detects duplicate active plates at the same location.
func (s *Store) CreateOfflineSession(ctx context.Context, input CreateOfflineSessionInput) (*Session, error) {
	// Idempotency: if the session already exists, return it unchanged.
	existing, err := s.GetSessionByID(ctx, input.ID)
	if err == nil {
		return existing, nil
	}
	if err != errors.ErrNotFound {
		return nil, err
	}

	// Conflict detection: another ACTIVE/PENDING_PAYMENT session with the same plate
	// at the same location causes this offline record to be flagged for review.
	conflict := false
	_, dupErr := s.FindActiveSessionByPlate(ctx, input.LocationID, input.Plate)
	if dupErr == nil {
		conflict = true
	} else if dupErr != errors.ErrNotFound {
		return nil, fmt.Errorf("check duplicate plate: %w", dupErr)
	}

	var session Session
	err = s.pool.QueryRow(ctx, `
		INSERT INTO sessions (
			id, location_id, operator_id, shift_id, plate, city_code, vehicle_type,
			state, check_in_at, offline_sync, sync_conflict
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'ACTIVE', $8, true, $9)
		RETURNING id, location_id, operator_id, shift_id, plate, city_code, vehicle_type, state,
		          check_in_at, check_out_at, fee_amount, rate_snapshot, offline_sync, sync_conflict,
		          created_at, updated_at
	`, input.ID, input.LocationID, input.OperatorID, input.ShiftID, input.Plate,
		input.CityCode, input.VehicleType, input.CheckInAt, conflict).Scan(
		&session.ID, &session.LocationID, &session.OperatorID, &session.ShiftID, &session.Plate, &session.CityCode,
		&session.VehicleType, &session.State, &session.CheckInAt, &session.CheckOutAt, &session.FeeAmount,
		&session.RateSnapshot, &session.OfflineSync, &session.SyncConflict, &session.CreatedAt, &session.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create offline session: %w", err)
	}
	return &session, nil
}

// CreateOfflineTransactionInput captures a payment recorded while offline.
type CreateOfflineTransactionInput struct {
	ID                   string
	SessionID            string
	ShiftID              string
	OperatorID           string
	DurationHours        int
	RateFirstHour        float64
	RateSubsequentHourly float64
	RateDaily            float64
	FeeAmount            float64
	PaymentMethod        string
	AmountTendered       *float64
	ChangeAmount         *float64
	PaymentReference     *string
	ReceiptNumber        string
}

// CreateOfflineTransaction creates a payment for an offline session. It expects the
// session to already be in PENDING_PAYMENT state and uses the supplied transaction ID
// for idempotency. The session is moved to CLOSED on success.
func (s *Store) CreateOfflineTransaction(ctx context.Context, input CreateOfflineTransactionInput) (*Transaction, error) {
	// Idempotency: return the existing transaction if already synced.
	existing, err := s.GetTransactionByID(ctx, input.ID)
	if err == nil {
		return existing, nil
	}
	if err != errors.ErrNotFound {
		return nil, err
	}

	session, err := s.GetSessionByID(ctx, input.SessionID)
	if err != nil {
		return nil, err
	}
	if session.State != "PENDING_PAYMENT" {
		return nil, errors.ErrInvalidState
	}

	var tx Transaction
	err = s.pool.QueryRow(ctx, `
		INSERT INTO transactions (
			id, session_id, location_id, shift_id, operator_id, vehicle_type, plate,
			check_in_at, check_out_at, duration_hours,
			rate_first_hour, rate_subsequent_hourly, rate_daily, fee_amount,
			payment_method, amount_tendered, change_amount, payment_reference, receipt_number
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING id, session_id, location_id, shift_id, operator_id, vehicle_type, plate,
		          check_in_at, check_out_at, duration_hours,
		          rate_first_hour, rate_subsequent_hourly, rate_daily, fee_amount,
		          payment_method, amount_tendered, change_amount, payment_reference, receipt_number,
		          voided, voided_at, voided_by, void_reason, created_at, updated_at
	`, input.ID, session.ID, session.LocationID, input.ShiftID, input.OperatorID,
		session.VehicleType, session.Plate, session.CheckInAt, *session.CheckOutAt, input.DurationHours,
		input.RateFirstHour, input.RateSubsequentHourly, input.RateDaily, input.FeeAmount,
		input.PaymentMethod, input.AmountTendered, input.ChangeAmount, input.PaymentReference, input.ReceiptNumber).Scan(
		&tx.ID, &tx.SessionID, &tx.LocationID, &tx.ShiftID, &tx.OperatorID, &tx.VehicleType, &tx.Plate,
		&tx.CheckInAt, &tx.CheckOutAt, &tx.DurationHours,
		&tx.RateFirstHour, &tx.RateSubsequentHourly, &tx.RateDaily, &tx.FeeAmount,
		&tx.PaymentMethod, &tx.AmountTendered, &tx.ChangeAmount, &tx.PaymentReference, &tx.ReceiptNumber,
		&tx.Voided, &tx.VoidedAt, &tx.VoidedBy, &tx.VoidReason, &tx.CreatedAt, &tx.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create offline transaction: %w", err)
	}

	_, err = s.UpdateSessionToClosed(ctx, session.ID)
	if err != nil {
		return nil, fmt.Errorf("close session after offline payment: %w", err)
	}

	return &tx, nil
}

// ListSyncConflictsFilters filters conflict review queries.
type ListSyncConflictsFilters struct {
	LocationID string
	Resolved   *bool
}

// ListSyncConflicts returns offline-synced sessions whose plate conflicts with an
// active session at the location.
func (s *Store) ListSyncConflicts(ctx context.Context, filters ListSyncConflictsFilters, limit, offset int) ([]Session, int, error) {
	where := "WHERE offline_sync = true AND sync_conflict = true"
	args := []interface{}{}
	argIdx := 1

	if filters.LocationID != "" {
		where += fmt.Sprintf(" AND location_id = $%d", argIdx)
		args = append(args, filters.LocationID)
		argIdx++
	}

	countArgs := append([]interface{}{}, args...)
	var total int
	countQuery := "SELECT COUNT(*) FROM sessions " + where
	if err := s.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count sync conflicts: %w", err)
	}

	query := "SELECT id, location_id, operator_id, shift_id, plate, city_code, vehicle_type, state, " +
		"check_in_at, check_out_at, fee_amount, rate_snapshot, offline_sync, sync_conflict, " +
		"created_at, updated_at FROM sessions " + where +
		fmt.Sprintf(" ORDER BY check_in_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list sync conflicts: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var session Session
		if err := rows.Scan(
			&session.ID, &session.LocationID, &session.OperatorID, &session.ShiftID, &session.Plate, &session.CityCode,
			&session.VehicleType, &session.State, &session.CheckInAt, &session.CheckOutAt, &session.FeeAmount,
			&session.RateSnapshot, &session.OfflineSync, &session.SyncConflict, &session.CreatedAt, &session.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan sync conflict: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, total, rows.Err()
}

// ResolveSyncConflictAction describes how a manager resolves a conflict.
type ResolveSyncConflictAction string

const (
	// ResolveConflictVoidOffline voids the offline-synced session.
	ResolveConflictVoidOffline ResolveSyncConflictAction = "VOID_OFFLINE"
	// ResolveConflictIgnore clears the sync_conflict flag without voiding.
	ResolveConflictIgnore ResolveSyncConflictAction = "IGNORE"
)

// ResolveSyncConflictInput captures a manager's resolution choice.
type ResolveSyncConflictInput struct {
	SessionID string
	Action    ResolveSyncConflictAction
	VoidReason string
	ResolvedBy string
}

// ResolveSyncConflict applies a manager's resolution to a conflicting offline session.
func (s *Store) ResolveSyncConflict(ctx context.Context, input ResolveSyncConflictInput) (*Session, error) {
	session, err := s.GetSessionByID(ctx, input.SessionID)
	if err != nil {
		return nil, err
	}
	if !session.OfflineSync || !session.SyncConflict {
		return nil, errors.ErrInvalidState
	}

	switch input.Action {
	case ResolveConflictVoidOffline:
		session, err = s.UpdateSessionToVoided(ctx, session.ID)
		if err != nil {
			return nil, err
		}
		// Also void any associated transaction.
		tx, err := s.GetTransactionBySessionID(ctx, session.ID)
		if err == nil && !tx.Voided {
			_, _ = s.VoidTransaction(ctx, tx.ID, input.ResolvedBy, input.VoidReason)
		}
	case ResolveConflictIgnore:
		var updated Session
		err = s.pool.QueryRow(ctx, `
			UPDATE sessions
			SET sync_conflict = false,
			    updated_at = now()
			WHERE id = $1
			RETURNING id, location_id, operator_id, shift_id, plate, city_code, vehicle_type, state,
			          check_in_at, check_out_at, fee_amount, rate_snapshot, offline_sync, sync_conflict,
			          created_at, updated_at
		`, session.ID).Scan(
			&updated.ID, &updated.LocationID, &updated.OperatorID, &updated.ShiftID, &updated.Plate, &updated.CityCode,
			&updated.VehicleType, &updated.State, &updated.CheckInAt, &updated.CheckOutAt, &updated.FeeAmount,
			&updated.RateSnapshot, &updated.OfflineSync, &updated.SyncConflict, &updated.CreatedAt, &updated.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ignore sync conflict: %w", err)
		}
		session = &updated
	default:
		return nil, errors.ErrInvalidInput
	}

	return session, nil
}

// MarkSessionSyncConflict sets or clears the sync_conflict flag on a session.
func (s *Store) MarkSessionSyncConflict(ctx context.Context, sessionID string, conflict bool) (*Session, error) {
	var session Session
	err := s.pool.QueryRow(ctx, `
		UPDATE sessions
		SET sync_conflict = $2,
		    updated_at = now()
		WHERE id = $1
		RETURNING id, location_id, operator_id, shift_id, plate, city_code, vehicle_type, state,
		          check_in_at, check_out_at, fee_amount, rate_snapshot, offline_sync, sync_conflict,
		          created_at, updated_at
	`, sessionID, conflict).Scan(
		&session.ID, &session.LocationID, &session.OperatorID, &session.ShiftID, &session.Plate, &session.CityCode,
		&session.VehicleType, &session.State, &session.CheckInAt, &session.CheckOutAt, &session.FeeAmount,
		&session.RateSnapshot, &session.OfflineSync, &session.SyncConflict, &session.CreatedAt, &session.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("mark sync conflict: %w", err)
	}
	return &session, nil
}

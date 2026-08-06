package store

import (
	"context"
	"fmt"
	"time"

	"github.com/thoriqzs/PARKIR/backend/internal/errors"
)

// CreateOfflineSessionInput captures a session created by the desktop app while offline.
type CreateOfflineSessionInput struct {
	ID          string
	LocationID  string
	OperatorID  string
	ShiftNumber int
	Plate       string
	CityCode    string
	VehicleType string
	CheckInAt   time.Time
}

// CreateOfflineSession inserts a session that was created offline. It is idempotent
// by session ID. The shift number is provided by the client (desktop app).
func (s *Store) CreateOfflineSession(ctx context.Context, input CreateOfflineSessionInput) (*Session, error) {
	// Idempotency: if the session already exists, return it unchanged.
	existing, err := s.GetSessionByID(ctx, input.ID)
	if err == nil {
		return existing, nil
	}
	if err != errors.ErrNotFound {
		return nil, err
	}

	var session Session
	err = s.pool.QueryRow(ctx, `
		INSERT INTO sessions (
			id, location_id, operator_id, shift_number, plate, city_code, vehicle_type,
			state, check_in_at, offline_sync, sync_conflict
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'ACTIVE', $8, true, false)
		RETURNING id, location_id, operator_id, shift_number, plate, city_code, vehicle_type, state,
		          check_in_at, check_out_at, fee_amount, rate_snapshot, offline_sync, sync_conflict,
		          created_at, updated_at
	`, input.ID, input.LocationID, input.OperatorID, input.ShiftNumber, input.Plate,
		input.CityCode, input.VehicleType, input.CheckInAt).Scan(
		&session.ID, &session.LocationID, &session.OperatorID, &session.ShiftNumber, &session.Plate, &session.CityCode,
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

	var shiftNumber int
	if session.ShiftNumber != nil {
		shiftNumber = *session.ShiftNumber
	}

	var tx Transaction
	err = s.pool.QueryRow(ctx, `
		INSERT INTO transactions (
			id, session_id, location_id, shift_number, operator_id, vehicle_type, plate,
			check_in_at, check_out_at, duration_hours,
			rate_first_hour, rate_subsequent_hourly, rate_daily, fee_amount,
			payment_method, amount_tendered, change_amount, payment_reference, receipt_number
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING id, session_id, location_id, shift_number, operator_id, vehicle_type, plate,
		          check_in_at, check_out_at, duration_hours,
		          rate_first_hour, rate_subsequent_hourly, rate_daily, fee_amount,
		          payment_method, amount_tendered, change_amount, payment_reference, receipt_number,
		          voided, voided_at, voided_by, void_reason, created_at, updated_at
	`, input.ID, session.ID, session.LocationID, shiftNumber, input.OperatorID,
		session.VehicleType, session.Plate, session.CheckInAt, *session.CheckOutAt, input.DurationHours,
		input.RateFirstHour, input.RateSubsequentHourly, input.RateDaily, input.FeeAmount,
		input.PaymentMethod, input.AmountTendered, input.ChangeAmount, input.PaymentReference, input.ReceiptNumber).Scan(
		&tx.ID, &tx.SessionID, &tx.LocationID, &tx.ShiftNumber, &tx.OperatorID, &tx.VehicleType, &tx.Plate,
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

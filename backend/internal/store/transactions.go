package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
)

type Transaction struct {
	ID                   string     `json:"id"`
	SessionID            string     `json:"session_id"`
	LocationID           string     `json:"location_id"`
	ShiftID              string     `json:"shift_id"`
	OperatorID           string     `json:"operator_id"`
	VehicleType          string     `json:"vehicle_type"`
	Plate                string     `json:"plate"`
	CheckInAt            time.Time  `json:"check_in_at"`
	CheckOutAt           time.Time  `json:"check_out_at"`
	DurationHours        int        `json:"duration_hours"`
	RateFirstHour        float64    `json:"rate_first_hour"`
	RateSubsequentHourly float64    `json:"rate_subsequent_hourly"`
	RateDaily            float64    `json:"rate_daily"`
	FeeAmount            float64    `json:"fee_amount"`
	PaymentMethod        string     `json:"payment_method"`
	AmountTendered       *float64   `json:"amount_tendered,omitempty"`
	ChangeAmount         *float64   `json:"change_amount,omitempty"`
	PaymentReference     *string    `json:"payment_reference,omitempty"`
	ReceiptNumber        string     `json:"receipt_number"`
	Voided               bool       `json:"voided"`
	VoidedAt             *time.Time `json:"voided_at,omitempty"`
	VoidedBy             *string    `json:"voided_by,omitempty"`
	VoidReason           *string    `json:"void_reason,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type CreateTransactionInput struct {
	SessionID            string
	LocationID           string
	ShiftID              string
	OperatorID           string
	VehicleType          string
	Plate                string
	CheckInAt            time.Time
	CheckOutAt           time.Time
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

type ListTransactionsFilters struct {
	LocationID string
	ShiftID    string
	Voided     *bool
	DateFrom   *time.Time
	DateTo     *time.Time
}

func (s *Store) CreateTransaction(ctx context.Context, input CreateTransactionInput) (*Transaction, error) {
	var tx Transaction
	err := s.pool.QueryRow(ctx, `
		INSERT INTO transactions (
			session_id, location_id, shift_id, operator_id, vehicle_type, plate,
			check_in_at, check_out_at, duration_hours,
			rate_first_hour, rate_subsequent_hourly, rate_daily, fee_amount,
			payment_method, amount_tendered, change_amount, payment_reference, receipt_number
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, session_id, location_id, shift_id, operator_id, vehicle_type, plate,
		          check_in_at, check_out_at, duration_hours,
		          rate_first_hour, rate_subsequent_hourly, rate_daily, fee_amount,
		          payment_method, amount_tendered, change_amount, payment_reference, receipt_number,
		          voided, voided_at, voided_by, void_reason, created_at, updated_at
	`, input.SessionID, input.LocationID, input.ShiftID, input.OperatorID, input.VehicleType, input.Plate,
		input.CheckInAt, input.CheckOutAt, input.DurationHours,
		input.RateFirstHour, input.RateSubsequentHourly, input.RateDaily, input.FeeAmount,
		input.PaymentMethod, input.AmountTendered, input.ChangeAmount, input.PaymentReference, input.ReceiptNumber).Scan(
		&tx.ID, &tx.SessionID, &tx.LocationID, &tx.ShiftID, &tx.OperatorID, &tx.VehicleType, &tx.Plate,
		&tx.CheckInAt, &tx.CheckOutAt, &tx.DurationHours,
		&tx.RateFirstHour, &tx.RateSubsequentHourly, &tx.RateDaily, &tx.FeeAmount,
		&tx.PaymentMethod, &tx.AmountTendered, &tx.ChangeAmount, &tx.PaymentReference, &tx.ReceiptNumber,
		&tx.Voided, &tx.VoidedAt, &tx.VoidedBy, &tx.VoidReason, &tx.CreatedAt, &tx.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}
	return &tx, nil
}

func (s *Store) GetTransactionByID(ctx context.Context, id string) (*Transaction, error) {
	var tx Transaction
	err := s.pool.QueryRow(ctx, `
		SELECT id, session_id, location_id, shift_id, operator_id, vehicle_type, plate,
		       check_in_at, check_out_at, duration_hours,
		       rate_first_hour, rate_subsequent_hourly, rate_daily, fee_amount,
		       payment_method, amount_tendered, change_amount, payment_reference, receipt_number,
		       voided, voided_at, voided_by, void_reason, created_at, updated_at
		FROM transactions
		WHERE id = $1
	`, id).Scan(
		&tx.ID, &tx.SessionID, &tx.LocationID, &tx.ShiftID, &tx.OperatorID, &tx.VehicleType, &tx.Plate,
		&tx.CheckInAt, &tx.CheckOutAt, &tx.DurationHours,
		&tx.RateFirstHour, &tx.RateSubsequentHourly, &tx.RateDaily, &tx.FeeAmount,
		&tx.PaymentMethod, &tx.AmountTendered, &tx.ChangeAmount, &tx.PaymentReference, &tx.ReceiptNumber,
		&tx.Voided, &tx.VoidedAt, &tx.VoidedBy, &tx.VoidReason, &tx.CreatedAt, &tx.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get transaction: %w", err)
	}
	return &tx, nil
}

func (s *Store) VoidTransaction(ctx context.Context, id, voidedBy, reason string) (*Transaction, error) {
	var tx Transaction
	err := s.pool.QueryRow(ctx, `
		UPDATE transactions
		SET voided = true,
		    voided_at = now(),
		    voided_by = $2,
		    void_reason = $3,
		    updated_at = now()
		WHERE id = $1 AND voided = false
		RETURNING id, session_id, location_id, shift_id, operator_id, vehicle_type, plate,
		          check_in_at, check_out_at, duration_hours,
		          rate_first_hour, rate_subsequent_hourly, rate_daily, fee_amount,
		          payment_method, amount_tendered, change_amount, payment_reference, receipt_number,
		          voided, voided_at, voided_by, void_reason, created_at, updated_at
	`, id, voidedBy, reason).Scan(
		&tx.ID, &tx.SessionID, &tx.LocationID, &tx.ShiftID, &tx.OperatorID, &tx.VehicleType, &tx.Plate,
		&tx.CheckInAt, &tx.CheckOutAt, &tx.DurationHours,
		&tx.RateFirstHour, &tx.RateSubsequentHourly, &tx.RateDaily, &tx.FeeAmount,
		&tx.PaymentMethod, &tx.AmountTendered, &tx.ChangeAmount, &tx.PaymentReference, &tx.ReceiptNumber,
		&tx.Voided, &tx.VoidedAt, &tx.VoidedBy, &tx.VoidReason, &tx.CreatedAt, &tx.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("void transaction: %w", err)
	}
	return &tx, nil
}

func (s *Store) ListTransactions(ctx context.Context, filters ListTransactionsFilters, limit, offset int) ([]Transaction, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filters.LocationID != "" {
		where += fmt.Sprintf(" AND location_id = $%d", argIdx)
		args = append(args, filters.LocationID)
		argIdx++
	}
	if filters.ShiftID != "" {
		where += fmt.Sprintf(" AND shift_id = $%d", argIdx)
		args = append(args, filters.ShiftID)
		argIdx++
	}
	if filters.Voided != nil {
		where += fmt.Sprintf(" AND voided = $%d", argIdx)
		args = append(args, *filters.Voided)
		argIdx++
	}
	if filters.DateFrom != nil {
		where += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, *filters.DateFrom)
		argIdx++
	}
	if filters.DateTo != nil {
		where += fmt.Sprintf(" AND created_at < $%d", argIdx)
		args = append(args, *filters.DateTo)
		argIdx++
	}

	countArgs := append([]interface{}{}, args...)
	var total int
	countQuery := "SELECT COUNT(*) FROM transactions " + where
	if err := s.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count transactions: %w", err)
	}

	query := "SELECT id, session_id, location_id, shift_id, operator_id, vehicle_type, plate, " +
		"check_in_at, check_out_at, duration_hours, " +
		"rate_first_hour, rate_subsequent_hourly, rate_daily, fee_amount, " +
		"payment_method, amount_tendered, change_amount, payment_reference, receipt_number, " +
		"voided, voided_at, voided_by, void_reason, created_at, updated_at " +
		"FROM transactions " + where +
		fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(
			&tx.ID, &tx.SessionID, &tx.LocationID, &tx.ShiftID, &tx.OperatorID, &tx.VehicleType, &tx.Plate,
			&tx.CheckInAt, &tx.CheckOutAt, &tx.DurationHours,
			&tx.RateFirstHour, &tx.RateSubsequentHourly, &tx.RateDaily, &tx.FeeAmount,
			&tx.PaymentMethod, &tx.AmountTendered, &tx.ChangeAmount, &tx.PaymentReference, &tx.ReceiptNumber,
			&tx.Voided, &tx.VoidedAt, &tx.VoidedBy, &tx.VoidReason, &tx.CreatedAt, &tx.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, total, rows.Err()
}

func (s *Store) GetTransactionBySessionID(ctx context.Context, sessionID string) (*Transaction, error) {
	var tx Transaction
	err := s.pool.QueryRow(ctx, `
		SELECT id, session_id, location_id, shift_id, operator_id, vehicle_type, plate,
		       check_in_at, check_out_at, duration_hours,
		       rate_first_hour, rate_subsequent_hourly, rate_daily, fee_amount,
		       payment_method, amount_tendered, change_amount, payment_reference, receipt_number,
		       voided, voided_at, voided_by, void_reason, created_at, updated_at
		FROM transactions
		WHERE session_id = $1
	`, sessionID).Scan(
		&tx.ID, &tx.SessionID, &tx.LocationID, &tx.ShiftID, &tx.OperatorID, &tx.VehicleType, &tx.Plate,
		&tx.CheckInAt, &tx.CheckOutAt, &tx.DurationHours,
		&tx.RateFirstHour, &tx.RateSubsequentHourly, &tx.RateDaily, &tx.FeeAmount,
		&tx.PaymentMethod, &tx.AmountTendered, &tx.ChangeAmount, &tx.PaymentReference, &tx.ReceiptNumber,
		&tx.Voided, &tx.VoidedAt, &tx.VoidedBy, &tx.VoidReason, &tx.CreatedAt, &tx.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get transaction by session: %w", err)
	}
	return &tx, nil
}

func (s *Store) SumCashByShift(ctx context.Context, shiftID string) (float64, error) {
	var sum *float64
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(fee_amount), 0)
		FROM transactions
		WHERE shift_id = $1 AND voided = false AND payment_method = 'CASH'
	`, shiftID).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("sum cash by shift: %w", err)
	}
	if sum == nil {
		return 0, nil
	}
	return *sum, nil
}

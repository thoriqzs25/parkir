package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
)

type Session struct {
	ID            string                 `json:"id"`
	LocationID    string                 `json:"location_id"`
	OperatorID    string                 `json:"operator_id"`
	ShiftID       *string                `json:"shift_id,omitempty"`
	Plate         string                 `json:"plate"`
	CityCode      string                 `json:"city_code"`
	VehicleType   string                 `json:"vehicle_type"`
	State         string                 `json:"state"`
	CheckInAt     time.Time              `json:"check_in_at"`
	CheckOutAt    *time.Time             `json:"check_out_at,omitempty"`
	FeeAmount     *float64               `json:"fee_amount,omitempty"`
	RateSnapshot  map[string]interface{} `json:"rate_snapshot,omitempty"`
	OfflineSync   bool                   `json:"offline_sync"`
	SyncConflict  bool                   `json:"sync_conflict"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type CreateSessionInput struct {
	LocationID  string
	OperatorID  string
	ShiftID     string
	Plate       string
	CityCode    string
	VehicleType string
}

type CheckOutSessionInput struct {
	CheckOutAt   time.Time
	FeeAmount    *float64
	RateSnapshot map[string]interface{}
}

type ListSessionsFilters struct {
	LocationID string
	State      string
	Plate      string
	OperatorID string
}

func (s *Store) CreateSession(ctx context.Context, input CreateSessionInput) (*Session, error) {
	var session Session
	err := s.pool.QueryRow(ctx, `
		INSERT INTO sessions (location_id, operator_id, shift_id, plate, city_code, vehicle_type, state, check_in_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'ACTIVE', now())
		RETURNING id, location_id, operator_id, shift_id, plate, city_code, vehicle_type, state,
		          check_in_at, check_out_at, fee_amount, rate_snapshot, offline_sync, sync_conflict,
		          created_at, updated_at
	`, input.LocationID, input.OperatorID, input.ShiftID, input.Plate, input.CityCode, input.VehicleType).Scan(
		&session.ID, &session.LocationID, &session.OperatorID, &session.ShiftID, &session.Plate, &session.CityCode,
		&session.VehicleType, &session.State, &session.CheckInAt, &session.CheckOutAt, &session.FeeAmount,
		&session.RateSnapshot, &session.OfflineSync, &session.SyncConflict, &session.CreatedAt, &session.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return &session, nil
}

func (s *Store) GetSessionByID(ctx context.Context, id string) (*Session, error) {
	var session Session
	err := s.pool.QueryRow(ctx, `
		SELECT id, location_id, operator_id, shift_id, plate, city_code, vehicle_type, state,
		       check_in_at, check_out_at, fee_amount, rate_snapshot, offline_sync, sync_conflict,
		       created_at, updated_at
		FROM sessions
		WHERE id = $1
	`, id).Scan(
		&session.ID, &session.LocationID, &session.OperatorID, &session.ShiftID, &session.Plate, &session.CityCode,
		&session.VehicleType, &session.State, &session.CheckInAt, &session.CheckOutAt, &session.FeeAmount,
		&session.RateSnapshot, &session.OfflineSync, &session.SyncConflict, &session.CreatedAt, &session.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &session, nil
}

func (s *Store) UpdateSessionToPendingPayment(ctx context.Context, id string, input CheckOutSessionInput) (*Session, error) {
	var session Session
	var rateSnapshotBytes []byte
	if input.RateSnapshot != nil {
		var err error
		rateSnapshotBytes, err = json.Marshal(input.RateSnapshot)
		if err != nil {
			return nil, fmt.Errorf("marshal rate snapshot: %w", err)
		}
	}

	err := s.pool.QueryRow(ctx, `
		UPDATE sessions
		SET state = 'PENDING_PAYMENT',
		    check_out_at = $2,
		    fee_amount = $3,
		    rate_snapshot = $4,
		    updated_at = now()
		WHERE id = $1 AND state = 'ACTIVE'
		RETURNING id, location_id, operator_id, shift_id, plate, city_code, vehicle_type, state,
		          check_in_at, check_out_at, fee_amount, rate_snapshot, offline_sync, sync_conflict,
		          created_at, updated_at
	`, id, input.CheckOutAt, input.FeeAmount, rateSnapshotBytes).Scan(
		&session.ID, &session.LocationID, &session.OperatorID, &session.ShiftID, &session.Plate, &session.CityCode,
		&session.VehicleType, &session.State, &session.CheckInAt, &session.CheckOutAt, &session.FeeAmount,
		&session.RateSnapshot, &session.OfflineSync, &session.SyncConflict, &session.CreatedAt, &session.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("checkout session: %w", err)
	}
	return &session, nil
}

func (s *Store) UpdateSessionToClosed(ctx context.Context, id string) (*Session, error) {
	var session Session
	err := s.pool.QueryRow(ctx, `
		UPDATE sessions
		SET state = 'CLOSED',
		    updated_at = now()
		WHERE id = $1 AND state = 'PENDING_PAYMENT'
		RETURNING id, location_id, operator_id, shift_id, plate, city_code, vehicle_type, state,
		          check_in_at, check_out_at, fee_amount, rate_snapshot, offline_sync, sync_conflict,
		          created_at, updated_at
	`, id).Scan(
		&session.ID, &session.LocationID, &session.OperatorID, &session.ShiftID, &session.Plate, &session.CityCode,
		&session.VehicleType, &session.State, &session.CheckInAt, &session.CheckOutAt, &session.FeeAmount,
		&session.RateSnapshot, &session.OfflineSync, &session.SyncConflict, &session.CreatedAt, &session.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("close session: %w", err)
	}
	return &session, nil
}

func (s *Store) UpdateSessionToVoided(ctx context.Context, id string) (*Session, error) {
	var session Session
	err := s.pool.QueryRow(ctx, `
		UPDATE sessions
		SET state = 'VOIDED',
		    updated_at = now()
		WHERE id = $1
		RETURNING id, location_id, operator_id, shift_id, plate, city_code, vehicle_type, state,
		          check_in_at, check_out_at, fee_amount, rate_snapshot, offline_sync, sync_conflict,
		          created_at, updated_at
	`, id).Scan(
		&session.ID, &session.LocationID, &session.OperatorID, &session.ShiftID, &session.Plate, &session.CityCode,
		&session.VehicleType, &session.State, &session.CheckInAt, &session.CheckOutAt, &session.FeeAmount,
		&session.RateSnapshot, &session.OfflineSync, &session.SyncConflict, &session.CreatedAt, &session.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("void session: %w", err)
	}
	return &session, nil
}

func (s *Store) FindActiveSessionByPlate(ctx context.Context, locationID, plate string) (*Session, error) {
	var session Session
	err := s.pool.QueryRow(ctx, `
		SELECT id, location_id, operator_id, shift_id, plate, city_code, vehicle_type, state,
		       check_in_at, check_out_at, fee_amount, rate_snapshot, offline_sync, sync_conflict,
		       created_at, updated_at
		FROM sessions
		WHERE location_id = $1 AND plate = $2 AND state IN ('ACTIVE', 'PENDING_PAYMENT')
		LIMIT 1
	`, locationID, plate).Scan(
		&session.ID, &session.LocationID, &session.OperatorID, &session.ShiftID, &session.Plate, &session.CityCode,
		&session.VehicleType, &session.State, &session.CheckInAt, &session.CheckOutAt, &session.FeeAmount,
		&session.RateSnapshot, &session.OfflineSync, &session.SyncConflict, &session.CreatedAt, &session.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("find active session: %w", err)
	}
	return &session, nil
}

func (s *Store) ListSessions(ctx context.Context, filters ListSessionsFilters, limit, offset int) ([]Session, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filters.LocationID != "" {
		where += fmt.Sprintf(" AND location_id = $%d", argIdx)
		args = append(args, filters.LocationID)
		argIdx++
	}
	if filters.State != "" {
		states := strings.Split(filters.State, ",")
		placeholders := make([]string, len(states))
		for i := range states {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, states[i])
			argIdx++
		}
		where += fmt.Sprintf(" AND state IN (%s)", strings.Join(placeholders, ", "))
	}
	if filters.Plate != "" {
		where += fmt.Sprintf(" AND plate ILIKE $%d", argIdx)
		args = append(args, "%"+filters.Plate+"%")
		argIdx++
	}
	if filters.OperatorID != "" {
		where += fmt.Sprintf(" AND operator_id = $%d", argIdx)
		args = append(args, filters.OperatorID)
		argIdx++
	}

	countArgs := append([]interface{}{}, args...)
	var total int
	countQuery := "SELECT COUNT(*) FROM sessions " + where
	if err := s.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count sessions: %w", err)
	}

	query := "SELECT id, location_id, operator_id, shift_id, plate, city_code, vehicle_type, state, " +
		"check_in_at, check_out_at, fee_amount, rate_snapshot, offline_sync, sync_conflict, " +
		"created_at, updated_at FROM sessions " + where +
		fmt.Sprintf(" ORDER BY check_in_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list sessions: %w", err)
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
			return nil, 0, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, total, rows.Err()
}

package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
)

type Incident struct {
	ID                string     `json:"id"`
	LocationID        string     `json:"location_id"`
	Type              string     `json:"type"`
	State             string     `json:"state"`
	SessionID         *string    `json:"session_id,omitempty"`
	ReportedBy        string     `json:"reported_by"`
	ReportedAt        time.Time  `json:"reported_at"`
	Description       string     `json:"description"`
	ResolvedBy        *string    `json:"resolved_by,omitempty"`
	ResolvedAt        *time.Time `json:"resolved_at,omitempty"`
	ResolutionNotes   *string    `json:"resolution_notes,omitempty"`
	AdjustmentAction  *string    `json:"adjustment_action,omitempty"`
	AdjustmentEntityID *string   `json:"adjustment_entity_id,omitempty"`
	OfflineSync       bool       `json:"offline_sync"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type IncidentNote struct {
	ID         string    `json:"id"`
	IncidentID string    `json:"incident_id"`
	AuthorID   string    `json:"author_id"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}

type ListIncidentsFilters struct {
	LocationID string
	Type       string
	State      string
	DateFrom   *time.Time
	DateTo     *time.Time
}

func (s *Store) CreateIncident(ctx context.Context, input struct {
	LocationID  string
	Type        string
	SessionID   *string
	ReportedBy  string
	Description string
	OfflineSync bool
}) (*Incident, error) {
	var inc Incident
	err := s.pool.QueryRow(ctx, `
		INSERT INTO incidents (location_id, type, session_id, reported_by, description, offline_sync)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, location_id, type, state, session_id, reported_by, reported_at, description,
		          resolved_by, resolved_at, resolution_notes, adjustment_action, adjustment_entity_id,
		          offline_sync, created_at, updated_at
	`, input.LocationID, input.Type, input.SessionID, input.ReportedBy, input.Description, input.OfflineSync).Scan(
		&inc.ID, &inc.LocationID, &inc.Type, &inc.State, &inc.SessionID, &inc.ReportedBy, &inc.ReportedAt,
		&inc.Description, &inc.ResolvedBy, &inc.ResolvedAt, &inc.ResolutionNotes,
		&inc.AdjustmentAction, &inc.AdjustmentEntityID, &inc.OfflineSync, &inc.CreatedAt, &inc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create incident: %w", err)
	}
	return &inc, nil
}

func (s *Store) GetIncidentByID(ctx context.Context, id string) (*Incident, error) {
	var inc Incident
	err := s.pool.QueryRow(ctx, `
		SELECT id, location_id, type, state, session_id, reported_by, reported_at, description,
		       resolved_by, resolved_at, resolution_notes, adjustment_action, adjustment_entity_id,
		       offline_sync, created_at, updated_at
		FROM incidents
		WHERE id = $1
	`, id).Scan(
		&inc.ID, &inc.LocationID, &inc.Type, &inc.State, &inc.SessionID, &inc.ReportedBy, &inc.ReportedAt,
		&inc.Description, &inc.ResolvedBy, &inc.ResolvedAt, &inc.ResolutionNotes,
		&inc.AdjustmentAction, &inc.AdjustmentEntityID, &inc.OfflineSync, &inc.CreatedAt, &inc.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get incident: %w", err)
	}
	return &inc, nil
}

func (s *Store) ListIncidents(ctx context.Context, filters ListIncidentsFilters, limit, offset int) ([]Incident, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filters.LocationID != "" {
		where += fmt.Sprintf(" AND location_id = $%d", argIdx)
		args = append(args, filters.LocationID)
		argIdx++
	}
	if filters.Type != "" {
		where += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, filters.Type)
		argIdx++
	}
	if filters.State != "" {
		where += fmt.Sprintf(" AND state = $%d", argIdx)
		args = append(args, filters.State)
		argIdx++
	}
	if filters.DateFrom != nil {
		where += fmt.Sprintf(" AND reported_at >= $%d", argIdx)
		args = append(args, *filters.DateFrom)
		argIdx++
	}
	if filters.DateTo != nil {
		where += fmt.Sprintf(" AND reported_at < $%d", argIdx)
		args = append(args, *filters.DateTo)
		argIdx++
	}

	countArgs := append([]interface{}{}, args...)
	var total int
	countQuery := "SELECT COUNT(*) FROM incidents " + where
	if err := s.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count incidents: %w", err)
	}

	query := "SELECT id, location_id, type, state, session_id, reported_by, reported_at, description, " +
		"resolved_by, resolved_at, resolution_notes, adjustment_action, adjustment_entity_id, " +
		"offline_sync, created_at, updated_at FROM incidents " + where +
		fmt.Sprintf(" ORDER BY reported_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list incidents: %w", err)
	}
	defer rows.Close()

	var incidents []Incident
	for rows.Next() {
		var inc Incident
		if err := rows.Scan(
			&inc.ID, &inc.LocationID, &inc.Type, &inc.State, &inc.SessionID, &inc.ReportedBy, &inc.ReportedAt,
			&inc.Description, &inc.ResolvedBy, &inc.ResolvedAt, &inc.ResolutionNotes,
			&inc.AdjustmentAction, &inc.AdjustmentEntityID, &inc.OfflineSync, &inc.CreatedAt, &inc.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan incident: %w", err)
		}
		incidents = append(incidents, inc)
	}

	return incidents, total, rows.Err()
}

func (s *Store) ResolveIncident(ctx context.Context, id, resolvedBy, resolutionNotes string, adjustmentAction, adjustmentEntityID *string) (*Incident, error) {
	var inc Incident
	err := s.pool.QueryRow(ctx, `
		UPDATE incidents
		SET state = 'RESOLVED',
		    resolved_by = $2,
		    resolved_at = now(),
		    resolution_notes = $3,
		    adjustment_action = $4,
		    adjustment_entity_id = $5,
		    updated_at = now()
		WHERE id = $1 AND state IN ('OPEN', 'IN_PROGRESS')
		RETURNING id, location_id, type, state, session_id, reported_by, reported_at, description,
		          resolved_by, resolved_at, resolution_notes, adjustment_action, adjustment_entity_id,
		          offline_sync, created_at, updated_at
	`, id, resolvedBy, resolutionNotes, adjustmentAction, adjustmentEntityID).Scan(
		&inc.ID, &inc.LocationID, &inc.Type, &inc.State, &inc.SessionID, &inc.ReportedBy, &inc.ReportedAt,
		&inc.Description, &inc.ResolvedBy, &inc.ResolvedAt, &inc.ResolutionNotes,
		&inc.AdjustmentAction, &inc.AdjustmentEntityID, &inc.OfflineSync, &inc.CreatedAt, &inc.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("resolve incident: %w", err)
	}
	return &inc, nil
}

func (s *Store) CreateIncidentNote(ctx context.Context, incidentID, authorID, note string) (*IncidentNote, error) {
	var n IncidentNote
	err := s.pool.QueryRow(ctx, `
		INSERT INTO incident_notes (incident_id, author_id, note)
		VALUES ($1, $2, $3)
		RETURNING id, incident_id, author_id, note, created_at
	`, incidentID, authorID, note).Scan(
		&n.ID, &n.IncidentID, &n.AuthorID, &n.Note, &n.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create incident note: %w", err)
	}
	return &n, nil
}

func (s *Store) ListIncidentNotes(ctx context.Context, incidentID string) ([]IncidentNote, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, incident_id, author_id, note, created_at
		FROM incident_notes
		WHERE incident_id = $1
		ORDER BY created_at ASC
	`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("list incident notes: %w", err)
	}
	defer rows.Close()

	var notes []IncidentNote
	for rows.Next() {
		var n IncidentNote
		if err := rows.Scan(&n.ID, &n.IncidentID, &n.AuthorID, &n.Note, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan incident note: %w", err)
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (s *Store) ReassignSession(ctx context.Context, sessionID, newOperatorID, newShiftID string) (*Session, error) {
	var session Session
	err := s.pool.QueryRow(ctx, `
		UPDATE sessions
		SET operator_id = $2,
		    shift_id = $3,
		    updated_at = now()
		WHERE id = $1
		RETURNING id, location_id, operator_id, shift_id, plate, city_code, vehicle_type, state,
		          check_in_at, check_out_at, fee_amount, rate_snapshot, offline_sync, sync_conflict,
		          created_at, updated_at
	`, sessionID, newOperatorID, newShiftID).Scan(
		&session.ID, &session.LocationID, &session.OperatorID, &session.ShiftID, &session.Plate, &session.CityCode,
		&session.VehicleType, &session.State, &session.CheckInAt, &session.CheckOutAt, &session.FeeAmount,
		&session.RateSnapshot, &session.OfflineSync, &session.SyncConflict, &session.CreatedAt, &session.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("reassign session: %w", err)
	}
	return &session, nil
}

func (s *Store) ReassignTransactionShift(ctx context.Context, transactionID, newOperatorID, newShiftID string) (*Transaction, error) {
	var tx Transaction
	err := s.pool.QueryRow(ctx, `
		UPDATE transactions
		SET operator_id = $2,
		    shift_id = $3,
		    updated_at = now()
		WHERE id = $1
		RETURNING id, session_id, location_id, shift_id, operator_id, vehicle_type, plate,
		          check_in_at, check_out_at, duration_hours,
		          rate_first_hour, rate_subsequent_hourly, rate_daily, fee_amount,
		          payment_method, amount_tendered, change_amount, payment_reference, receipt_number,
		          voided, voided_at, voided_by, void_reason, created_at, updated_at
	`, transactionID, newOperatorID, newShiftID).Scan(
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
		return nil, fmt.Errorf("reassign transaction shift: %w", err)
	}
	return &tx, nil
}
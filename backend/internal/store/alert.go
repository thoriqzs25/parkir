package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thoriqzs/PARKIR/backend/internal/errors"
)

type Alert struct {
	ID               string                 `json:"id"`
	Code             string                 `json:"code"`
	LocationID       *string                `json:"location_id,omitempty"`
	State            string                 `json:"state"`
	EntityType       *string                `json:"entity_type,omitempty"`
	EntityID         *string                `json:"entity_id,omitempty"`
	TriggeredAt      time.Time              `json:"triggered_at"`
	AcknowledgedBy   *string                `json:"acknowledged_by,omitempty"`
	AcknowledgedAt   *time.Time             `json:"acknowledged_at,omitempty"`
	ResolvedBy       *string                `json:"resolved_by,omitempty"`
	ResolvedAt       *time.Time             `json:"resolved_at,omitempty"`
	ResolutionNotes  *string                `json:"resolution_notes,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
}

type AlertConfig struct {
	ID         string                 `json:"id"`
	LocationID *string                `json:"location_id,omitempty"`
	Code       string                 `json:"code"`
	Enabled    bool                   `json:"enabled"`
	Threshold  map[string]interface{} `json:"threshold,omitempty"`
	UpdatedBy  *string                `json:"updated_by,omitempty"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

type ListAlertsFilters struct {
	LocationID string
	State      string
	Code       string
}

func (s *Store) CreateAlert(ctx context.Context, input struct {
	Code       string
	LocationID *string
	EntityType *string
	EntityID   *string
	Metadata   map[string]interface{}
}) (*Alert, error) {
	var a Alert
	err := s.pool.QueryRow(ctx, `
		INSERT INTO alerts (code, location_id, entity_type, entity_id, metadata)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, code, location_id, state, entity_type, entity_id, triggered_at,
		          acknowledged_by, acknowledged_at, resolved_by, resolved_at, resolution_notes,
		          metadata, created_at
	`, input.Code, input.LocationID, input.EntityType, input.EntityID, input.Metadata).Scan(
		&a.ID, &a.Code, &a.LocationID, &a.State, &a.EntityType, &a.EntityID, &a.TriggeredAt,
		&a.AcknowledgedBy, &a.AcknowledgedAt, &a.ResolvedBy, &a.ResolvedAt, &a.ResolutionNotes,
		&a.Metadata, &a.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create alert: %w", err)
	}
	return &a, nil
}

func (s *Store) ListAlerts(ctx context.Context, filters ListAlertsFilters, limit, offset int) ([]Alert, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filters.LocationID != "" {
		where += fmt.Sprintf(" AND location_id = $%d", argIdx)
		args = append(args, filters.LocationID)
		argIdx++
	}
	if filters.State != "" {
		where += fmt.Sprintf(" AND state = $%d", argIdx)
		args = append(args, filters.State)
		argIdx++
	}
	if filters.Code != "" {
		where += fmt.Sprintf(" AND code = $%d", argIdx)
		args = append(args, filters.Code)
		argIdx++
	}

	countArgs := append([]interface{}{}, args...)
	var total int
	countQuery := "SELECT COUNT(*) FROM alerts " + where
	if err := s.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count alerts: %w", err)
	}

	query := "SELECT id, code, location_id, state, entity_type, entity_id, triggered_at, " +
		"acknowledged_by, acknowledged_at, resolved_by, resolved_at, resolution_notes, " +
		"metadata, created_at FROM alerts " + where +
		fmt.Sprintf(" ORDER BY triggered_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list alerts: %w", err)
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var a Alert
		if err := rows.Scan(
			&a.ID, &a.Code, &a.LocationID, &a.State, &a.EntityType, &a.EntityID, &a.TriggeredAt,
			&a.AcknowledgedBy, &a.AcknowledgedAt, &a.ResolvedBy, &a.ResolvedAt, &a.ResolutionNotes,
			&a.Metadata, &a.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan alert: %w", err)
		}
		alerts = append(alerts, a)
	}

	return alerts, total, rows.Err()
}

func (s *Store) GetAlertByID(ctx context.Context, id string) (*Alert, error) {
	var a Alert
	err := s.pool.QueryRow(ctx, `
		SELECT id, code, location_id, state, entity_type, entity_id, triggered_at,
		       acknowledged_by, acknowledged_at, resolved_by, resolved_at, resolution_notes,
		       metadata, created_at
		FROM alerts
		WHERE id = $1
	`, id).Scan(
		&a.ID, &a.Code, &a.LocationID, &a.State, &a.EntityType, &a.EntityID, &a.TriggeredAt,
		&a.AcknowledgedBy, &a.AcknowledgedAt, &a.ResolvedBy, &a.ResolvedAt, &a.ResolutionNotes,
		&a.Metadata, &a.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get alert: %w", err)
	}
	return &a, nil
}

func (s *Store) AcknowledgeAlert(ctx context.Context, id, acknowledgedBy string) (*Alert, error) {
	var a Alert
	err := s.pool.QueryRow(ctx, `
		UPDATE alerts
		SET state = 'ACKNOWLEDGED',
		    acknowledged_by = $2,
		    acknowledged_at = now()
		WHERE id = $1 AND state = 'TRIGGERED'
		RETURNING id, code, location_id, state, entity_type, entity_id, triggered_at,
		          acknowledged_by, acknowledged_at, resolved_by, resolved_at, resolution_notes,
		          metadata, created_at
	`, id, acknowledgedBy).Scan(
		&a.ID, &a.Code, &a.LocationID, &a.State, &a.EntityType, &a.EntityID, &a.TriggeredAt,
		&a.AcknowledgedBy, &a.AcknowledgedAt, &a.ResolvedBy, &a.ResolvedAt, &a.ResolutionNotes,
		&a.Metadata, &a.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("acknowledge alert: %w", err)
	}
	return &a, nil
}

func (s *Store) ResolveAlert(ctx context.Context, id, resolvedBy, resolutionNotes string) (*Alert, error) {
	var a Alert
	err := s.pool.QueryRow(ctx, `
		UPDATE alerts
		SET state = 'RESOLVED',
		    resolved_by = $2,
		    resolved_at = now(),
		    resolution_notes = $3
		WHERE id = $1 AND state IN ('TRIGGERED', 'ACKNOWLEDGED')
		RETURNING id, code, location_id, state, entity_type, entity_id, triggered_at,
		          acknowledged_by, acknowledged_at, resolved_by, resolved_at, resolution_notes,
		          metadata, created_at
	`, id, resolvedBy, resolutionNotes).Scan(
		&a.ID, &a.Code, &a.LocationID, &a.State, &a.EntityType, &a.EntityID, &a.TriggeredAt,
		&a.AcknowledgedBy, &a.AcknowledgedAt, &a.ResolvedBy, &a.ResolvedAt, &a.ResolutionNotes,
		&a.Metadata, &a.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("resolve alert: %w", err)
	}
	return &a, nil
}

func (s *Store) ListAlertConfigs(ctx context.Context, locationID string) ([]AlertConfig, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ac.id, ac.location_id, ac.code, ac.enabled, ac.threshold, u.name as updated_by, ac.updated_at
		FROM alert_configs ac
		LEFT JOIN users u ON u.id = ac.updated_by
		WHERE ac.location_id = $1 OR ac.location_id IS NULL
		ORDER BY ac.code
	`, locationID)
	if err != nil {
		return nil, fmt.Errorf("list alert configs: %w", err)
	}
	defer rows.Close()

	var configs []AlertConfig
	for rows.Next() {
		var c AlertConfig
		var updatedByName *string
		if err := rows.Scan(&c.ID, &c.LocationID, &c.Code, &c.Enabled, &c.Threshold, &updatedByName, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan alert config: %w", err)
		}
		if updatedByName != nil {
			c.UpdatedBy = updatedByName
		}
		configs = append(configs, c)
	}
	return configs, rows.Err()
}

func (s *Store) GetAlertConfigByID(ctx context.Context, id string) (*AlertConfig, error) {
	var c AlertConfig
	err := s.pool.QueryRow(ctx, `
		SELECT id, location_id, code, enabled, threshold, updated_by, updated_at
		FROM alert_configs
		WHERE id = $1
	`, id).Scan(
		&c.ID, &c.LocationID, &c.Code, &c.Enabled, &c.Threshold, &c.UpdatedBy, &c.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("get alert config: %w", err)
	}
	return &c, nil
}

func (s *Store) UpdateAlertConfig(ctx context.Context, id, updatedBy string, enabled *bool, threshold map[string]interface{}) (*AlertConfig, error) {
	var c AlertConfig
	err := s.pool.QueryRow(ctx, `
		UPDATE alert_configs
		SET enabled = COALESCE($2, enabled),
		    threshold = COALESCE($3, threshold),
		    updated_by = $4,
		    updated_at = now()
		WHERE id = $1
		RETURNING id, location_id, code, enabled, threshold, updated_by, updated_at
	`, id, enabled, threshold, updatedBy).Scan(
		&c.ID, &c.LocationID, &c.Code, &c.Enabled, &c.Threshold, &c.UpdatedBy, &c.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("update alert config: %w", err)
	}
	return &c, nil
}

func (s *Store) CountTriggeredAlerts(ctx context.Context) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM alerts WHERE state = 'TRIGGERED'`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count triggered alerts: %w", err)
	}
	return count, nil
}

func (s *Store) HasAlertForEntity(ctx context.Context, code, entityType, entityID string) (bool, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM alerts
		WHERE code = $1 AND entity_type = $2 AND entity_id = $3 AND state IN ('TRIGGERED', 'ACKNOWLEDGED')
	`, code, entityType, entityID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check existing alert: %w", err)
	}
	return count > 0, nil
}
package store

import (
	"context"
	"fmt"
	"time"
)

type AuditLog struct {
	ID         string                 `json:"id"`
	Action     string                 `json:"action"`
	ActorID    *string                `json:"actor_id,omitempty"`
	ActorRole  *string                `json:"actor_role,omitempty"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	LocationID *string                `json:"location_id,omitempty"`
	IPAddress  *string                `json:"ip_address,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

type ListAuditLogsFilters struct {
	Action     string
	ActorID    string
	EntityType string
	EntityID   string
	LocationID string
	DateFrom   *time.Time
	DateTo     *time.Time
}

func (s *Store) ListAuditLogs(ctx context.Context, filters ListAuditLogsFilters, limit, offset int) ([]AuditLog, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filters.Action != "" {
		where += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, filters.Action)
		argIdx++
	}
	if filters.ActorID != "" {
		where += fmt.Sprintf(" AND actor_id = $%d", argIdx)
		args = append(args, filters.ActorID)
		argIdx++
	}
	if filters.EntityType != "" {
		where += fmt.Sprintf(" AND entity_type = $%d", argIdx)
		args = append(args, filters.EntityType)
		argIdx++
	}
	if filters.EntityID != "" {
		where += fmt.Sprintf(" AND entity_id = $%d", argIdx)
		args = append(args, filters.EntityID)
		argIdx++
	}
	if filters.LocationID != "" {
		where += fmt.Sprintf(" AND location_id = $%d", argIdx)
		args = append(args, filters.LocationID)
		argIdx++
	}
	if filters.DateFrom != nil {
		where += fmt.Sprintf(" AND timestamp >= $%d", argIdx)
		args = append(args, *filters.DateFrom)
		argIdx++
	}
	if filters.DateTo != nil {
		where += fmt.Sprintf(" AND timestamp < $%d", argIdx)
		args = append(args, *filters.DateTo)
		argIdx++
	}

	countArgs := append([]interface{}{}, args...)
	var total int
	countQuery := "SELECT COUNT(*) FROM audit_logs " + where
	if err := s.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}

	query := "SELECT id, action, actor_id, actor_role, entity_type, entity_id, location_id, ip_address::text, metadata, timestamp " +
		"FROM audit_logs " + where +
		fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(
			&l.ID, &l.Action, &l.ActorID, &l.ActorRole, &l.EntityType, &l.EntityID,
			&l.LocationID, &l.IPAddress, &l.Metadata, &l.Timestamp,
		); err != nil {
			return nil, 0, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, l)
	}

	return logs, total, rows.Err()
}

func (s *Store) ListAuditLogsAll(ctx context.Context, filters ListAuditLogsFilters) ([]AuditLog, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filters.Action != "" {
		where += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, filters.Action)
		argIdx++
	}
	if filters.ActorID != "" {
		where += fmt.Sprintf(" AND actor_id = $%d", argIdx)
		args = append(args, filters.ActorID)
		argIdx++
	}
	if filters.EntityType != "" {
		where += fmt.Sprintf(" AND entity_type = $%d", argIdx)
		args = append(args, filters.EntityType)
		argIdx++
	}
	if filters.EntityID != "" {
		where += fmt.Sprintf(" AND entity_id = $%d", argIdx)
		args = append(args, filters.EntityID)
		argIdx++
	}
	if filters.LocationID != "" {
		where += fmt.Sprintf(" AND location_id = $%d", argIdx)
		args = append(args, filters.LocationID)
		argIdx++
	}
	if filters.DateFrom != nil {
		where += fmt.Sprintf(" AND timestamp >= $%d", argIdx)
		args = append(args, *filters.DateFrom)
		argIdx++
	}
	if filters.DateTo != nil {
		where += fmt.Sprintf(" AND timestamp < $%d", argIdx)
		args = append(args, *filters.DateTo)
		argIdx++
	}

	query := "SELECT id, action, actor_id, actor_role, entity_type, entity_id, location_id, ip_address::text, metadata, timestamp " +
		"FROM audit_logs " + where + " ORDER BY timestamp DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list all audit logs: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(
			&l.ID, &l.Action, &l.ActorID, &l.ActorRole, &l.EntityType, &l.EntityID,
			&l.LocationID, &l.IPAddress, &l.Metadata, &l.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, l)
	}

	return logs, rows.Err()
}
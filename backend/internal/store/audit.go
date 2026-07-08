package store

import (
	"context"
	"fmt"
)

type AuditLogEntry struct {
	Action     string                 `json:"action"`
	ActorID    *string                `json:"actor_id,omitempty"`
	ActorRole  *string                `json:"actor_role,omitempty"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	LocationID *string                `json:"location_id,omitempty"`
	IPAddress  *string                `json:"ip_address,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

func (s *Store) CreateAuditLog(ctx context.Context, entry AuditLogEntry) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO audit_logs (action, actor_id, actor_role, entity_type, entity_id, location_id, ip_address, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, entry.Action, entry.ActorID, entry.ActorRole, entry.EntityType, entry.EntityID,
		entry.LocationID, entry.IPAddress, entry.Metadata)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

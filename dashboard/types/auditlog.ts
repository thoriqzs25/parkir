export interface AuditLog {
  id: string;
  action: string;
  actor_id?: string;
  actor_role?: string;
  entity_type: string;
  entity_id: string;
  location_id?: string;
  ip_address?: string;
  metadata?: Record<string, unknown>;
  timestamp: string;
}
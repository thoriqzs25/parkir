export interface Incident {
  id: string;
  location_id: string;
  type: "STUCK_AT_GATE" | "PAYMENT_DISPUTE" | "OPERATOR_ERROR" | "SYSTEM_DOWNTIME";
  state: "OPEN" | "IN_PROGRESS" | "RESOLVED";
  session_id?: string;
  reported_by: string;
  reported_at: string;
  description: string;
  resolved_by?: string;
  resolved_at?: string;
  resolution_notes?: string;
  adjustment_action?: string;
  adjustment_entity_id?: string;
  offline_sync: boolean;
  created_at: string;
  updated_at: string;
}

export interface IncidentNote {
  id: string;
  incident_id: string;
  author_id: string;
  note: string;
  created_at: string;
}

export interface CreateIncidentInput {
  location_id: string;
  type: Incident["type"];
  session_id?: string;
  description: string;
}

export interface ResolveIncidentInput {
  resolution_notes: string;
  adjustment_action?: string;
  adjustment_entity_id?: string;
  manager_pin?: string;
}

export interface AddNoteInput {
  note: string;
}
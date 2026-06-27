export interface Alert {
  id: string;
  code: string;
  location_id?: string;
  state: "TRIGGERED" | "ACKNOWLEDGED" | "RESOLVED";
  entity_type?: string;
  entity_id?: string;
  triggered_at: string;
  acknowledged_by?: string;
  acknowledged_at?: string;
  resolved_by?: string;
  resolved_at?: string;
  resolution_notes?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface AlertConfig {
  id: string;
  location_id?: string;
  code: string;
  enabled: boolean;
  threshold?: Record<string, unknown>;
  updated_by?: string;
  updated_at: string;
}

export interface HealthComponents {
  status: string;
  components: {
    api: { status: string; uptime_seconds: number };
    database: { status: string };
  };
  last_check: string;
}
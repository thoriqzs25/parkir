export interface Session {
  id: string;
  location_id: string;
  operator_id: string;
  shift_id?: string;
  plate: string;
  city_code: string;
  vehicle_type: string;
  state: "ACTIVE" | "PENDING_PAYMENT" | "CLOSED" | "VOIDED";
  check_in_at: string;
  check_out_at?: string;
  fee_amount?: number;
  rate_snapshot?: Record<string, unknown>;
  offline_sync: boolean;
  sync_conflict: boolean;
  created_at: string;
  updated_at: string;
}

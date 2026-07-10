export interface Location {
  id: string;
  name: string;
  code: string;
  address?: string;
  city?: string;
  status: string;
  capacity?: Record<string, number>;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: string;
  name: string;
  email: string;
  role_id: string;
  role_name?: string;
  status: string;
  location_ids?: string[];
  created_at: string;
  updated_at: string;
}

export interface MeResponse {
  user: User;
  permissions: string[];
}

export interface Shift {
  id: string;
  operator_id: string;
  location_id: string;
  status: string;
  started_at: string;
  ended_at?: string;
  expected_cash?: number;
  cash_handover_amount?: number;
  discrepancy?: number;
  discrepancy_notes?: string;
  force_closed_by?: string;
  force_closed_reason?: string;
  created_at: string;
  updated_at: string;
}

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

export interface Transaction {
  id: string;
  session_id: string;
  location_id: string;
  shift_id: string;
  operator_id: string;
  vehicle_type: string;
  plate: string;
  check_in_at: string;
  check_out_at: string;
  duration_hours: number;
  rate_first_hour: number;
  rate_subsequent_hourly: number;
  rate_daily: number;
  fee_amount: number;
  payment_method: "CASH" | "DIGITAL";
  amount_tendered?: number;
  change_amount?: number;
  payment_reference?: string;
  receipt_number: string;
  voided: boolean;
  voided_at?: string;
  voided_by?: string;
  void_reason?: string;
  created_at: string;
  updated_at: string;
}

export interface Rate {
  id: string;
  location_id: string;
  vehicle_type: string;
  first_hour_rate: number;
  subsequent_hourly_rate: number;
  daily_flat_rate: number;
  effective_from: string;
  effective_until?: string | null;
  created_by?: string;
  created_at: string;
  updated_at: string;
}

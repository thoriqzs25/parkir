export interface Shift {
  id: string;
  operator_id: string;
  location_id: string;
  status: "OPEN" | "CLOSED" | "FLAGGED" | "FORCE_CLOSED";
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

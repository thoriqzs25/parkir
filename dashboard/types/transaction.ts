export interface Transaction {
  id: string;
  session_id: string;
  location_id: string;
  shift_id: string;
  operator_id: string;
  vehicle_type: "CAR" | "MOTO" | "TRUCK";
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

export interface VoidTransactionInput {
  manager_pin: string;
  void_reason: string;
}

export interface DailyRevenueRow {
  date: string;
  total_revenue: number;
  transaction_count: number;
  average_fee: number;
  voided_count: number;
  voided_amount: number;
}

export interface OccupancyRow {
  bucket: string;
  count: number;
}

export interface VehicleBreakdownRow {
  vehicle_type: string;
  count: number;
  total_revenue: number;
}

export interface OperatorActivityRow {
  operator_id: string;
  operator_name: string;
  session_count: number;
  total_revenue: number;
  shift_hours: number;
}
export interface Rate {
  id: string;
  location_id: string;
  vehicle_type: string;
  first_hour_rate: number;
  subsequent_hourly_rate: number;
  daily_flat_rate: number;
  effective_from: string;
  effective_until?: string;
  created_by?: string;
  created_at: string;
}

export interface CreateRateInput {
  vehicle_type: string;
  first_hour_rate: number;
  subsequent_hourly_rate: number;
  daily_flat_rate: number;
  effective_from: string;
  effective_until?: string;
}

export interface UpdateRateInput {
  first_hour_rate?: number;
  subsequent_hourly_rate?: number;
  daily_flat_rate?: number;
  effective_until?: string;
}

export interface Gate {
  id: string;
  device_id: string;
  name: string;
  location_id?: string;
  ip_address: string;
  last_seen_at?: string;
  registered_at: string;
  created_at: string;
  updated_at: string;
}

export interface RegisterGateInput {
  device_id: string;
  name?: string;
  location_id?: string;
  ip_address?: string;
}

export interface UpdateGateInput {
  name?: string;
  location_id?: string;
  ip_address?: string;
}

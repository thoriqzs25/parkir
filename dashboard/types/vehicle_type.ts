export interface VehicleType {
  name: string;
  display_name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface CreateVehicleTypeInput {
  name: string;
  display_name: string;
  description?: string;
}

export interface UpdateVehicleTypeInput {
  display_name?: string;
  description?: string;
}

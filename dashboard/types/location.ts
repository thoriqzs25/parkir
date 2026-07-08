export interface Location {
  id: string;
  name: string;
  code: string;
  address?: string;
  city?: string;
  status: "ACTIVE" | "INACTIVE";
  capacity?: Record<string, number>;
  created_at: string;
  updated_at: string;
}

export interface CreateLocationInput {
  name: string;
  code: string;
  address?: string;
  city?: string;
  capacity?: Record<string, number>;
}

export interface UpdateLocationInput {
  name?: string;
  address?: string;
  city?: string;
  status?: "ACTIVE" | "INACTIVE";
  capacity?: Record<string, number>;
}

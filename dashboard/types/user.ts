export interface User {
  id: string;
  name: string;
  email: string;
  role_id: string;
  role_name?: string;
  status: "ACTIVE" | "DEACTIVATED";
  location_ids?: string[];
  created_at: string;
  updated_at: string;
}

export interface CreateUserInput {
  name: string;
  email: string;
  password: string;
  role_id: string;
  location_ids?: string[];
}

export interface UpdateUserInput {
  name?: string;
  email?: string;
  role_id?: string;
  location_ids?: string[];
  status?: "ACTIVE" | "DEACTIVATED";
}

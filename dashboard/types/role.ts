export interface Role {
  id: string;
  name: string;
  permissions: string[];
  created_at: string;
  updated_at: string;
}

export interface CreateRoleInput {
  name: string;
  permissions: string[];
}

export interface UpdateRoleInput {
  name?: string;
  permissions?: string[];
}

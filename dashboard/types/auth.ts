import { User } from "./user";

export interface LoginInput {
  email: string;
  password: string;
}

export interface MeResponse {
  user: User;
  permissions: string[];
}

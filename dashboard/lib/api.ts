import { ApiResponse, PaginatedItems } from "@/types/api";
import { LoginInput, MeResponse } from "@/types/auth";
import { Location, CreateLocationInput, UpdateLocationInput } from "@/types/location";
import { Rate, CreateRateInput, UpdateRateInput } from "@/types/rate";
import { Role, CreateRoleInput, UpdateRoleInput } from "@/types/role";
import { Session } from "@/types/session";
import { Shift } from "@/types/shift";
import { Transaction, VoidTransactionInput } from "@/types/transaction";
import { CreateUserInput, UpdateUserInput, User } from "@/types/user";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export class ApiError extends Error {
  constructor(
    public code: string,
    message: string,
    public status: number,
    public field?: string
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_BASE_URL}${path}`;
  const res = await fetch(url, {
    ...options,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });

  let body: ApiResponse<T> | undefined;
  try {
    body = await res.json();
  } catch {
    // non-JSON response
  }

  if (!res.ok) {
    const error = body?.error;
    throw new ApiError(
      error?.code || "UNKNOWN_ERROR",
      error?.message || `Request failed: ${res.statusText}`,
      res.status,
      error?.field
    );
  }

  return body?.data as T;
}

async function refreshToken(): Promise<boolean> {
  try {
    await request<{ token: string }>("/auth/refresh", { method: "POST" });
    return true;
  } catch {
    return false;
  }
}

export async function apiRequest<T>(
  path: string,
  options: RequestInit = {},
  retry = true
): Promise<T> {
  try {
    return await request<T>(path, options);
  } catch (err) {
    if (err instanceof ApiError && err.status === 401 && retry) {
      const refreshed = await refreshToken();
      if (refreshed) {
        return request<T>(path, options);
      }
    }
    throw err;
  }
}

export function getHealth() {
  return apiRequest<{ status: string }>("/health", { cache: "no-store" });
}

// Auth
export function login(input: LoginInput) {
  return apiRequest<{ user: User; token: string }>("/auth/login", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function logout() {
  return apiRequest<{ message: string }>("/auth/logout", { method: "POST" });
}

export function getMe() {
  return apiRequest<MeResponse>("/api/v1/auth/me", { cache: "no-store" });
}

// Users
export function listUsers(params?: Record<string, string>) {
  const qs = params ? "?" + new URLSearchParams(params).toString() : "";
  return apiRequest<PaginatedItems<User>>(`/api/v1/users${qs}`);
}

export function getUser(id: string) {
  return apiRequest<User>(`/api/v1/users/${id}`);
}

export function createUser(input: CreateUserInput) {
  return apiRequest<User>("/api/v1/users", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateUser(id: string, input: UpdateUserInput) {
  return apiRequest<User>(`/api/v1/users/${id}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export function deactivateUser(id: string) {
  return apiRequest<void>(`/api/v1/users/${id}/deactivate`, { method: "POST" });
}

export function resetPassword(id: string, newPassword: string) {
  return apiRequest<void>(`/api/v1/users/${id}/reset-password`, {
    method: "POST",
    body: JSON.stringify({ new_password: newPassword }),
  });
}

export function resetPIN(id: string, newPIN: string) {
  return apiRequest<void>(`/api/v1/users/${id}/reset-pin`, {
    method: "POST",
    body: JSON.stringify({ new_pin: newPIN }),
  });
}

// Roles
export function listRoles() {
  return apiRequest<PaginatedItems<Role>>("/api/v1/roles");
}

export function getRole(id: string) {
  return apiRequest<Role>(`/api/v1/roles/${id}`);
}

export function createRole(input: CreateRoleInput) {
  return apiRequest<Role>("/api/v1/roles", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateRole(id: string, input: UpdateRoleInput) {
  return apiRequest<Role>(`/api/v1/roles/${id}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export function deleteRole(id: string) {
  return apiRequest<void>(`/api/v1/roles/${id}`, { method: "DELETE" });
}

// Locations
export function listLocations() {
  return apiRequest<PaginatedItems<Location>>("/api/v1/locations");
}

export function getLocation(id: string) {
  return apiRequest<Location>(`/api/v1/locations/${id}`);
}

export function createLocation(input: CreateLocationInput) {
  return apiRequest<Location>("/api/v1/locations", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateLocation(id: string, input: UpdateLocationInput) {
  return apiRequest<Location>(`/api/v1/locations/${id}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export function deactivateLocation(id: string) {
  return apiRequest<Location>(`/api/v1/locations/${id}/deactivate`, {
    method: "POST",
  });
}

export function assignOperator(locationId: string, userId: string) {
  return apiRequest<void>(`/api/v1/locations/${locationId}/assign-operator`, {
    method: "POST",
    body: JSON.stringify({ user_id: userId }),
  });
}

export function removeOperator(locationId: string, userId: string) {
  return apiRequest<void>(`/api/v1/locations/${locationId}/remove-operator`, {
    method: "POST",
    body: JSON.stringify({ user_id: userId }),
  });
}

// Rates
export function listRates(locationId: string) {
  return apiRequest<Rate[]>(`/api/v1/locations/${locationId}/rates`);
}

export function createRate(locationId: string, input: CreateRateInput) {
  return apiRequest<Rate>(`/api/v1/locations/${locationId}/rates`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateRate(rateId: string, input: UpdateRateInput) {
  return apiRequest<Rate>(`/api/v1/rates/${rateId}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

// Sessions
export function listSessions(params?: Record<string, string>) {
  const qs = params ? "?" + new URLSearchParams(params).toString() : "";
  return apiRequest<PaginatedItems<Session>>(`/api/v1/sessions${qs}`);
}

export function getSession(id: string, include?: "transaction") {
  const qs = include ? `?include=${include}` : "";
  return apiRequest<{ session: Session; transaction?: Transaction }>(`/api/v1/sessions/${id}${qs}`);
}

// Transactions
export function listTransactions(params?: Record<string, string>) {
  const qs = params ? "?" + new URLSearchParams(params).toString() : "";
  return apiRequest<PaginatedItems<Transaction>>(`/api/v1/transactions${qs}`);
}

export function getTransaction(id: string) {
  return apiRequest<Transaction>(`/api/v1/transactions/${id}`);
}

export function voidTransaction(id: string, input: VoidTransactionInput) {
  return apiRequest<Transaction>(`/api/v1/transactions/${id}/void`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

// Shifts
export function listShifts(params?: Record<string, string>) {
  const qs = params ? "?" + new URLSearchParams(params).toString() : "";
  return apiRequest<PaginatedItems<Shift>>(`/api/v1/shifts${qs}`);
}

export function getShift(id: string, include?: "transactions") {
  const qs = include ? `?include=${include}` : "";
  return apiRequest<{
    shift: Shift;
    transactions?: Transaction[];
    summary?: { transaction_count: number; expected_cash: number };
  }>(`/api/v1/shifts/${id}${qs}`);
}

// Sync conflicts
export function listSyncConflicts(params?: Record<string, string>) {
  const qs = params ? "?" + new URLSearchParams(params).toString() : "";
  return apiRequest<PaginatedItems<Session>>(`/api/v1/sync/conflicts${qs}`);
}

export interface ResolveSyncConflictInput {
  action: "VOID_OFFLINE" | "IGNORE";
  void_reason?: string;
}

export function resolveSyncConflict(id: string, input: ResolveSyncConflictInput) {
  return apiRequest<Session>(`/api/v1/sync/conflicts/${id}/resolve`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

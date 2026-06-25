import type { Location, MeResponse, Session, Shift, Transaction, User } from "../types";

const API_BASE_URL = "http://localhost:8080/api/v1";

interface APIError {
  code: string;
  message: string;
  details?: unknown;
}

interface Envelope<T> {
  data: T;
  error?: APIError;
  meta?: unknown;
}

let bearerToken: string | null = null;

async function getStoredToken(): Promise<string | null> {
  try {
    if (window.electronAPI && window.electronAPI.getToken) {
      return (await window.electronAPI.getToken()) || null;
    }
  } catch {
    // fall back to in-memory token
  }
  return bearerToken;
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown
): Promise<T> {
  const token = await getStoredToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE_URL}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  let envelope: Envelope<T> | null = null;
  const contentType = res.headers.get("content-type") || "";
  if (contentType.includes("application/json")) {
    envelope = (await res.json()) as Envelope<T>;
  }

  if (!res.ok) {
    const message = envelope?.error?.message || `HTTP ${res.status} error`;
    const error: Error & { code?: string; status?: number } = new Error(message);
    error.code = envelope?.error?.code || "HTTP_ERROR";
    error.status = res.status;
    throw error;
  }

  if (!envelope) {
    throw new Error("invalid response from server");
  }

  return envelope.data;
}

export async function login(email: string, password: string): Promise<{ user: User; token: string }> {
  const res = await request<{ user: User; token: string }>("POST", "/auth/login", { email, password });
  if (window.electronAPI && window.electronAPI.setToken) {
    await window.electronAPI.setToken(res.token);
  } else {
    bearerToken = res.token;
  }
  return res;
}

export async function logout(): Promise<void> {
  try {
    await request<void>("POST", "/auth/logout");
  } catch {
    // ignore logout errors
  }
  if (window.electronAPI && window.electronAPI.clearToken) {
    await window.electronAPI.clearToken();
  } else {
    bearerToken = null;
  }
}

export async function me(): Promise<MeResponse> {
  return request<MeResponse>("GET", "/auth/me");
}

export async function listLocations(): Promise<{ items: Location[]; meta?: { total: number } }> {
  return request<{ items: Location[]; meta?: { total: number } }>("GET", "/locations");
}

export async function getMyOpenShift(): Promise<Shift | null> {
  try {
    return await request<Shift>("GET", "/shifts/me/open");
  } catch (err) {
    const error = err as Error & { status?: number };
    if (error.status === 404) {
      return null;
    }
    throw err;
  }
}

export interface StartShiftInput {
  location_id: string;
}

export async function startShift(input: StartShiftInput): Promise<Shift> {
  return request<Shift>("POST", "/shifts/start", input);
}

export interface EndShiftInput {
  cash_handover_amount: number;
  discrepancy_notes?: string;
}

export async function endShift(id: string, input: EndShiftInput): Promise<Shift> {
  return request<Shift>("POST", `/shifts/${id}/end`, input);
}

export interface CheckInInput {
  location_id: string;
  plate: string;
  city_code: string;
  vehicle_type: "CAR" | "MOTO" | "TRUCK";
}

export interface CheckInResponse {
  session: Session;
  duplicate_plate_warning: boolean;
}

export async function checkIn(input: CheckInInput): Promise<CheckInResponse> {
  return request<CheckInResponse>("POST", "/sessions/check-in", input);
}

export async function checkOut(
  id: string,
  fee_amount?: number
): Promise<Session> {
  return request<Session>("POST", `/sessions/${id}/check-out`, fee_amount !== undefined ? { fee_amount } : {});
}

export interface ListSessionsFilters {
  location_id?: string;
  state?: string;
  plate?: string;
  operator_id?: string;
  limit?: number;
  offset?: number;
}

export async function listSessions(filters: ListSessionsFilters): Promise<{ items: Session[]; meta?: { total: number } }> {
  const params = new URLSearchParams();
  if (filters.location_id) params.set("location_id", filters.location_id);
  if (filters.state) params.set("state", filters.state);
  if (filters.plate) params.set("plate", filters.plate);
  if (filters.operator_id) params.set("operator_id", filters.operator_id);
  if (filters.limit !== undefined) params.set("limit", String(filters.limit));
  if (filters.offset !== undefined) params.set("offset", String(filters.offset));
  const query = params.toString();
  return request<{ items: Session[]; meta?: { total: number } }>("GET", `/sessions?${query}`);
}

export interface PaymentCashInput {
  session_id: string;
  amount_tendered: number;
}

export interface PaymentDigitalInput {
  session_id: string;
  payment_reference?: string;
}

export async function payCash(input: PaymentCashInput): Promise<Transaction> {
  return request<Transaction>("POST", "/payments/cash", input);
}

export async function payDigital(input: PaymentDigitalInput): Promise<Transaction> {
  return request<Transaction>("POST", "/payments/digital", input);
}

export async function getTransaction(id: string): Promise<Transaction> {
  return request<Transaction>("GET", `/transactions/${id}`);
}

export interface ListTransactionsFilters {
  location_id?: string;
  shift_id?: string;
  voided?: boolean;
  date_from?: string;
  date_to?: string;
  limit?: number;
  offset?: number;
}

export async function listTransactions(filters: ListTransactionsFilters): Promise<{ items: Transaction[]; meta?: { total: number } }> {
  const params = new URLSearchParams();
  if (filters.location_id) params.set("location_id", filters.location_id);
  if (filters.shift_id) params.set("shift_id", filters.shift_id);
  if (filters.voided !== undefined) params.set("voided", String(filters.voided));
  if (filters.date_from) params.set("date_from", filters.date_from);
  if (filters.date_to) params.set("date_to", filters.date_to);
  if (filters.limit !== undefined) params.set("limit", String(filters.limit));
  if (filters.offset !== undefined) params.set("offset", String(filters.offset));
  const query = params.toString();
  return request<{ items: Transaction[]; meta?: { total: number } }>("GET", `/transactions?${query}`);
}

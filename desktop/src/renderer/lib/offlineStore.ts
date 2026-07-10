import type { Session } from "../types";

const PREFIX = "parkir_desktop";

export interface LocalSession extends Session {
  pendingSync: boolean;
  offlineReceiptNumber?: string;
  transactionId?: string;
  paymentMethod?: "CASH" | "DIGITAL";
  amountTendered?: number;
  changeAmount?: number;
  paymentReference?: string;
}

export interface OfflineCheckInData {
  id: string;
  location_id: string;
  operator_id: string;
  shift_id: string;
  plate: string;
  city_code: string;
  vehicle_type: string;
  check_in_at: string;
}

export interface OfflineCheckOutData {
  session_id: string;
  check_out_at: string;
  fee_amount: number;
  rate_snapshot?: Record<string, unknown>;
}

export interface OfflinePaymentData {
  transaction_id: string;
  session_id: string;
  shift_id: string;
  operator_id: string;
  location_id: string;
  duration_hours: number;
  rate_first_hour: number;
  rate_subsequent_hourly: number;
  rate_daily: number;
  fee_amount: number;
  payment_method: "CASH" | "DIGITAL";
  amount_tendered?: number;
  change_amount?: number;
  payment_reference?: string;
  offline_receipt_number: string;
}

export type PendingItem =
  | { type: "check_in"; session: OfflineCheckInData }
  | { type: "check_out"; data: OfflineCheckOutData }
  | { type: "payment"; data: OfflinePaymentData };

interface StoredData {
  sessions: LocalSession[];
  pending: PendingItem[];
  rates: RateCache;
  lastSyncAt?: string;
}

export interface RateCache {
  locationId: string;
  rates: Array<{
    id: string;
    vehicle_type: string;
    first_hour_rate: number;
    subsequent_hourly_rate: number;
    daily_flat_rate: number;
    effective_from: string;
    effective_until?: string | null;
  }>;
  fetchedAt: string;
}

function read<T>(key: string, defaultValue: T): T {
  try {
    const raw = localStorage.getItem(`${PREFIX}:${key}`);
    if (!raw) return defaultValue;
    return JSON.parse(raw) as T;
  } catch {
    return defaultValue;
  }
}

function write<T>(key: string, value: T) {
  localStorage.setItem(`${PREFIX}:${key}`, JSON.stringify(value));
}

export function getStoredSessions(): LocalSession[] {
  return read<LocalSession[]>("sessions", []);
}

export function getPendingItems(): PendingItem[] {
  return read<PendingItem[]>("pending", []);
}

export function getRateCache(): RateCache | null {
  return read<RateCache | null>("rates", null);
}

export function getLastSyncAt(): string | null {
  return read<string | null>("lastSyncAt", null);
}

function setSessions(sessions: LocalSession[]) {
  write("sessions", sessions);
}

function setPendingItems(items: PendingItem[]) {
  write("pending", items);
}

export function setRateCache(cache: RateCache) {
  write("rates", cache);
}

export function setLastSyncAt(at: string) {
  write("lastSyncAt", at);
}

export function saveOfflineCheckIn(session: LocalSession, item: PendingItem) {
  const sessions = getStoredSessions();
  sessions.push(session);
  setSessions(sessions);

  const pending = getPendingItems();
  pending.push(item);
  setPendingItems(pending);
}

export function updateLocalSession(sessionId: string, updates: Partial<LocalSession>) {
  const sessions = getStoredSessions();
  const idx = sessions.findIndex((s) => s.id === sessionId);
  if (idx === -1) return;
  sessions[idx] = { ...sessions[idx], ...updates };
  setSessions(sessions);
}

export function saveOfflineCheckOut(item: PendingItem, updates: Partial<LocalSession>) {
  const data = (item as { type: "check_out"; data: OfflineCheckOutData }).data;
  updateLocalSession(data.session_id, updates);

  const pending = getPendingItems();
  pending.push(item);
  setPendingItems(pending);
}

export function saveOfflinePayment(item: PendingItem, updates: Partial<LocalSession>) {
  const data = (item as { type: "payment"; data: OfflinePaymentData }).data;
  updateLocalSession(data.session_id, updates);

  const pending = getPendingItems();
  pending.push(item);
  setPendingItems(pending);
}

export function clearPendingItems() {
  setPendingItems([]);
}

export function removePendingItems(predicate: (item: PendingItem, index: number) => boolean) {
  const pending = getPendingItems().filter((item, idx) => !predicate(item, idx));
  setPendingItems(pending);
}

export function markSessionsSynced(syncedSessionIds: Set<string>) {
  const sessions = getStoredSessions().map((s) => {
    if (syncedSessionIds.has(s.id)) {
      return { ...s, pendingSync: false };
    }
    return s;
  });
  setSessions(sessions);
}

export function pruneOldSessions(maxAgeHours = 48) {
  const cutoff = new Date(Date.now() - maxAgeHours * 60 * 60 * 1000).toISOString();
  const sessions = getStoredSessions().filter(
    (s) => s.pendingSync || s.check_in_at > cutoff || s.state !== "CLOSED"
  );
  setSessions(sessions);
}

export function clearAllLocalData() {
  localStorage.removeItem(`${PREFIX}:sessions`);
  localStorage.removeItem(`${PREFIX}:pending`);
  localStorage.removeItem(`${PREFIX}:rates`);
  localStorage.removeItem(`${PREFIX}:lastSyncAt`);
}

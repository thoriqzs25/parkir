import {
  getPendingItems,
  getStoredSessions,
  markSessionsSynced,
  removePendingItems,
  setLastSyncAt,
  type PendingItem,
} from "./offlineStore";
import { request } from "./api";

export interface SyncResult {
  type: string;
  session_id?: string;
  transaction_id?: string;
  receipt_number?: string;
  sync_conflict?: boolean;
  error?: string;
}

export interface BatchSyncResponse {
  results: SyncResult[];
}

function buildBatchPayload(items: PendingItem[]) {
  return {
    items: items.map((item) => {
      switch (item.type) {
        case "check_in":
          return {
            type: "check_in",
            session: item.session,
          };
        case "check_out":
          return {
            type: "check_out",
            session_id: item.data.session_id,
            check_out_at: item.data.check_out_at,
            fee_amount: item.data.fee_amount,
            rate_snapshot: item.data.rate_snapshot,
          };
        case "payment":
          return {
            type: "payment",
            transaction_id: item.data.transaction_id,
            session_id: item.data.session_id,
            shift_id: item.data.shift_id,
            operator_id: item.data.operator_id,
            location_id: item.data.location_id,
            duration_hours: item.data.duration_hours,
            rate_first_hour: item.data.rate_first_hour,
            rate_subsequent_hourly: item.data.rate_subsequent_hourly,
            rate_daily: item.data.rate_daily,
            fee_amount: item.data.fee_amount,
            payment_method: item.data.payment_method,
            amount_tendered: item.data.amount_tendered,
            change_amount: item.data.change_amount,
            payment_reference: item.data.payment_reference,
          };
        default:
          return null;
      }
    }).filter(Boolean),
  };
}

export async function syncPendingItems(): Promise<{
  success: boolean;
  results: SyncResult[];
  reprints: Array<{ sessionId: string; receiptNumber: string; plate: string; fee: number }>;
}> {
  const items = getPendingItems();
  if (items.length === 0) {
    setLastSyncAt(new Date().toISOString());
    return { success: true, results: [], reprints: [] };
  }

  const payload = buildBatchPayload(items);
  const response = await request<BatchSyncResponse>("POST", "/sync/batch", payload);

  const syncedSessionIds = new Set<string>();
  const reprints: Array<{ sessionId: string; receiptNumber: string; plate: string; fee: number }> = [];

  // Build a map of local sessions for reprint lookup.
  const sessions = getStoredSessions();
  const sessionMap = new Map(sessions.map((s) => [s.id, s]));

  for (const result of response.results) {
    if (result.type === "payment" && result.session_id && !result.error) {
      syncedSessionIds.add(result.session_id);
      if (result.receipt_number) {
        const local = sessionMap.get(result.session_id);
        if (local) {
          reprints.push({
            sessionId: result.session_id,
            receiptNumber: result.receipt_number,
            plate: local.plate,
            fee: local.fee_amount || 0,
          });
        }
      }
    }
  }

  // Only remove items that succeeded. Failed items remain for retry.
  const succeededIndices = new Set<number>();
  response.results.forEach((r, idx) => {
    if (!r.error) {
      succeededIndices.add(idx);
    }
  });

  removePendingItems((_, idx) => succeededIndices.has(idx));
  markSessionsSynced(syncedSessionIds);
  setLastSyncAt(new Date().toISOString());

  return { success: true, results: response.results, reprints };
}

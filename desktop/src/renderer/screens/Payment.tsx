import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { payCash, payDigital } from "../lib/api";
import { getRateCache, getStoredSessions, saveOfflinePayment, type LocalSession } from "../lib/offlineStore";
import { useAuth } from "../contexts/AuthContext";
import { useOnlineStatus } from "../hooks/useOnlineStatus";

function calculateDurationHours(checkInAt: string, checkOutAt: string): number {
  const diff = new Date(checkOutAt).getTime() - new Date(checkInAt).getTime();
  const hours = Math.ceil(diff / 3600000);
  return Math.max(1, hours);
}

function getRateValues(locationId: string, vehicleType: string, checkInAt: string) {
  const cache = getRateCache();
  if (!cache || cache.locationId !== locationId) return null;

  const checkInDate = new Date(checkInAt).toISOString().split("T")[0];
  const rate = cache.rates.find((r) => {
    if (r.vehicle_type !== vehicleType) return false;
    const from = r.effective_from;
    const until = r.effective_until;
    if (from && checkInDate < from) return false;
    if (until && checkInDate > until) return false;
    return true;
  });
  if (!rate) return null;

  return {
    firstHour: rate.first_hour_rate,
    subsequentHourly: rate.subsequent_hourly_rate,
    daily: rate.daily_flat_rate,
  };
}

function generateTempReceiptNumber(locationCode: string, seq: number) {
  return `${locationCode}-OFFLINE-${String(seq).padStart(5, "0")}`;
}

export function Payment() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { currentLocation, openShift } = useAuth();
  const online = useOnlineStatus();
  const sessionId = searchParams.get("sessionId") || "";
  const fee = Number(searchParams.get("fee") || "0");

  const [method, setMethod] = useState<"CASH" | "DIGITAL">("CASH");
  const [amount, setAmount] = useState<string>(fee > 0 ? String(Math.ceil(fee)) : "");
  const [reference, setReference] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!sessionId) {
      navigate("/check-out");
    }
  }, [sessionId, navigate]);

  const change = method === "CASH" && amount ? Math.max(0, Number(amount) - fee) : 0;

  const printOfflineReceipt = (localSession: LocalSession, receiptNumber: string) => {
    if (!currentLocation) return;
    const html = `
      <div style="padding: 24px; max-width: 320px; margin: 0 auto; text-align: center;">
        <h2 style="margin: 0 0 8px;">${currentLocation.name}</h2>
        <p style="margin: 0 0 16px;">Receipt</p>
        <hr />
        <p style="margin: 8px 0; font-size: 16px;"><strong>Plate:</strong> ${localSession.plate}</p>
        <p style="margin: 8px 0; font-size: 16px;"><strong>Type:</strong> ${localSession.vehicle_type}</p>
        <p style="margin: 8px 0; font-size: 14px;"><strong>In:</strong> ${new Date(localSession.check_in_at).toLocaleString("id-ID", { timeZone: "Asia/Jakarta" })}</p>
        <p style="margin: 8px 0; font-size: 14px;"><strong>Out:</strong> ${new Date(localSession.check_out_at || new Date().toISOString()).toLocaleString("id-ID", { timeZone: "Asia/Jakarta" })}</p>
        <hr />
        <p style="margin: 8px 0; font-size: 18px;"><strong>Total:</strong> Rp ${fee.toLocaleString("id-ID")}</p>
        ${method === "CASH" ? `<p style="margin: 8px 0; font-size: 14px;"><strong>Tendered:</strong> Rp ${Number(amount).toLocaleString("id-ID")}</p><p style="margin: 8px 0; font-size: 14px;"><strong>Change:</strong> Rp ${change.toLocaleString("id-ID")}</p>` : ""}
        <p style="margin: 8px 0; font-size: 14px;"><strong>Receipt:</strong> ${receiptNumber}</p>
        <hr />
        <p style="font-size: 12px; color: #666;">OFFLINE - Will sync when online</p>
      </div>
    `;
    window.electronAPI.printHtml(html).catch(() => {
      alert("Failed to print receipt.");
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      if (online) {
        let transaction;
        if (method === "CASH") {
          transaction = await payCash({
            session_id: sessionId,
            amount_tendered: Number(amount),
          });
        } else {
          transaction = await payDigital({
            session_id: sessionId,
            payment_reference: reference,
          });
        }
        navigate(`/success?transactionId=${transaction.id}`);
        return;
      }

      // Offline payment flow.
      if (!openShift || !currentLocation) {
        setError("Shift/location not available");
        setLoading(false);
        return;
      }

      const sessions = getStoredSessions();
      const localSession = sessions.find((s) => s.id === sessionId);
      if (!localSession) {
        setError("Local session not found");
        setLoading(false);
        return;
      }
      if (localSession.state !== "PENDING_PAYMENT" && localSession.state !== "ACTIVE") {
        setError("Session is not available for payment");
        setLoading(false);
        return;
      }
      if (!localSession.check_out_at) {
        setError("Session has not been checked out");
        setLoading(false);
        return;
      }

      const rateValues = getRateValues(localSession.location_id, localSession.vehicle_type, localSession.check_in_at);
      if (!rateValues) {
        setError("No cached rate available. Cannot record offline payment.");
        setLoading(false);
        return;
      }

      const transactionId = crypto.randomUUID();
      const durationHours = calculateDurationHours(localSession.check_in_at, localSession.check_out_at);
      const offlineReceiptNumber = generateTempReceiptNumber(currentLocation.code, sessions.length + 1);
      const paymentReference = reference || undefined;
      const amountTendered = method === "CASH" ? Number(amount) : undefined;
      const changeAmount = method === "CASH" ? change : undefined;

      saveOfflinePayment(
        {
          type: "payment",
          data: {
            transaction_id: transactionId,
            session_id: localSession.id,
            shift_id: openShift.id,
            operator_id: openShift.operator_id,
            location_id: currentLocation.id,
            duration_hours: durationHours,
            rate_first_hour: rateValues.firstHour,
            rate_subsequent_hourly: rateValues.subsequentHourly,
            rate_daily: rateValues.daily,
            fee_amount: fee,
            payment_method: method,
            amount_tendered: amountTendered,
            change_amount: changeAmount,
            payment_reference: paymentReference,
            offline_receipt_number: offlineReceiptNumber,
          },
        },
        {
          state: "CLOSED",
          pendingSync: true,
          transactionId,
          paymentMethod: method,
          amountTendered,
          changeAmount,
          paymentReference,
          offlineReceiptNumber,
        }
      );

      printOfflineReceipt(localSession, offlineReceiptNumber);
      navigate(`/success?localSessionId=${localSession.id}&offline=true`);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="screen">
      <button className="button secondary back" onClick={() => navigate("/check-out")}>
        &larr; Back
      </button>
      <h2>Payment {!online && <span className="offline-badge">OFFLINE</span>}</h2>
      <div className="card">
        <div className="fee-display">
          <span>Fee</span>
          <strong>Rp {fee.toLocaleString("id-ID")}</strong>
        </div>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Payment Method</label>
            <div className="segmented-control">
              <button
                type="button"
                className={method === "CASH" ? "active" : ""}
                onClick={() => setMethod("CASH")}
              >
                Cash
              </button>
              <button
                type="button"
                className={method === "DIGITAL" ? "active" : ""}
                onClick={() => setMethod("DIGITAL")}
              >
                Digital
              </button>
            </div>
          </div>

          {method === "CASH" ? (
            <div className="form-group">
              <label>Amount Tendered</label>
              <input
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                type="number"
                min={0}
                required
                autoFocus
              />
              {change > 0 && (
                <p className="info-message">Change: Rp {change.toLocaleString("id-ID")}</p>
              )}
            </div>
          ) : (
            <div className="form-group">
              <label>Reference (optional)</label>
              <input
                value={reference}
                onChange={(e) => setReference(e.target.value)}
                placeholder="Payment reference"
                autoFocus
              />
            </div>
          )}

          {error && <p className="error-message">{error}</p>}

          <button className="button primary full" type="submit" disabled={loading}>
            {loading ? "Processing..." : online ? "Pay" : "Queue Payment"}
          </button>
        </form>
      </div>
    </div>
  );
}

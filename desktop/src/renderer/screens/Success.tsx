import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { getTransaction } from "../lib/api";
import { getStoredSessions, type LocalSession } from "../lib/offlineStore";
import { useAuth } from "../contexts/AuthContext";
import type { Transaction } from "../types";

function formatWIB(date: string) {
  return new Date(date).toLocaleString("id-ID", { timeZone: "Asia/Jakarta" });
}

export function Success() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { currentLocation } = useAuth();
  const transactionId = searchParams.get("transactionId") || "";
  const localSessionId = searchParams.get("localSessionId") || "";
  const isOffline = searchParams.get("offline") === "true";

  const [transaction, setTransaction] = useState<Transaction | null>(null);
  const [localSession, setLocalSession] = useState<LocalSession | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (transactionId) {
      getTransaction(transactionId)
        .then((tx) => setTransaction(tx))
        .catch((err) => setError((err as Error).message));
    }
  }, [transactionId]);

  useEffect(() => {
    if (localSessionId) {
      const found = getStoredSessions().find((s) => s.id === localSessionId);
      if (found) setLocalSession(found);
    }
  }, [localSessionId]);

  const printReceipt = async () => {
    if (!currentLocation) return;

    let html = "";
    if (transaction) {
      html = `
        <div style="padding: 24px; max-width: 320px; margin: 0 auto; text-align: center;">
          <h2 style="margin: 0 0 8px;">${currentLocation.name}</h2>
          <p style="margin: 0 0 16px;">${transaction.receipt_number}</p>
          <hr />
          <p style="margin: 8px 0; font-size: 16px;"><strong>Plate:</strong> ${transaction.plate}</p>
          <p style="margin: 8px 0; font-size: 16px;"><strong>Type:</strong> ${transaction.vehicle_type}</p>
          <p style="margin: 8px 0; font-size: 14px;"><strong>Check-in:</strong> ${formatWIB(transaction.check_in_at)}</p>
          <p style="margin: 8px 0; font-size: 14px;"><strong>Check-out:</strong> ${formatWIB(transaction.check_out_at)}</p>
          <p style="margin: 8px 0; font-size: 14px;"><strong>Duration:</strong> ${transaction.duration_hours} hour(s)</p>
          <hr />
          <p style="margin: 8px 0; font-size: 18px;"><strong>Total: Rp ${transaction.fee_amount.toLocaleString("id-ID")}</strong></p>
          ${transaction.payment_method === "CASH" ? `
            <p style="margin: 4px 0; font-size: 14px;">Tendered: Rp ${(transaction.amount_tendered || 0).toLocaleString("id-ID")}</p>
            <p style="margin: 4px 0; font-size: 14px;">Change: Rp ${(transaction.change_amount || 0).toLocaleString("id-ID")}</p>
          ` : `
            <p style="margin: 4px 0; font-size: 14px;">Digital${transaction.payment_reference ? ` - ${transaction.payment_reference}` : ""}</p>
          `}
          <hr />
          <p style="font-size: 12px; color: #666;">Thank you</p>
        </div>
      `;
    } else if (localSession) {
      html = `
        <div style="padding: 24px; max-width: 320px; margin: 0 auto; text-align: center;">
          <h2 style="margin: 0 0 8px;">${currentLocation.name}</h2>
          <p style="margin: 0 0 16px;">${localSession.offlineReceiptNumber || "OFFLINE RECEIPT"}</p>
          <hr />
          <p style="margin: 8px 0; font-size: 16px;"><strong>Plate:</strong> ${localSession.plate}</p>
          <p style="margin: 8px 0; font-size: 16px;"><strong>Type:</strong> ${localSession.vehicle_type}</p>
          <p style="margin: 8px 0; font-size: 14px;"><strong>Check-in:</strong> ${formatWIB(localSession.check_in_at)}</p>
          <p style="margin: 8px 0; font-size: 14px;"><strong>Check-out:</strong> ${formatWIB(localSession.check_out_at || new Date().toISOString())}</p>
          <hr />
          <p style="margin: 8px 0; font-size: 18px;"><strong>Total: Rp ${(localSession.fee_amount || 0).toLocaleString("id-ID")}</strong></p>
          ${localSession.paymentMethod === "CASH" ? `
            <p style="margin: 4px 0; font-size: 14px;">Tendered: Rp ${(localSession.amountTendered || 0).toLocaleString("id-ID")}</p>
            <p style="margin: 4px 0; font-size: 14px;">Change: Rp ${(localSession.changeAmount || 0).toLocaleString("id-ID")}</p>
          ` : `
            <p style="margin: 4px 0; font-size: 14px;">Digital${localSession.paymentReference ? ` - ${localSession.paymentReference}` : ""}</p>
          `}
          <hr />
          <p style="font-size: 12px; color: #666;">Will sync when online</p>
        </div>
      `;
    }

    if (!html) return;

    try {
      await window.electronAPI.printHtml(html);
    } catch {
      alert("Failed to print receipt. You can reprint from session history.");
    }
  };

  useEffect(() => {
    if (transaction || localSession) {
      printReceipt();
    }
  }, [transaction, localSession]);

  if (error) {
    return (
      <div className="screen">
        <p className="error-message">{error}</p>
        <button className="button primary" onClick={() => navigate("/dashboard")}>
          Done
        </button>
      </div>
    );
  }

  if (!transaction && !localSession) {
    return <div className="screen">Loading receipt...</div>;
  }

  const display = transaction || {
    plate: localSession!.plate,
    vehicle_type: localSession!.vehicle_type,
    fee_amount: localSession!.fee_amount || 0,
    payment_method: localSession!.paymentMethod || "CASH",
    receipt_number: localSession!.offlineReceiptNumber,
  };

  return (
    <div className="screen">
      <div className="card success-card">
        <h2>Payment Successful</h2>
        {isOffline && <p className="offline-badge">OFFLINE - Pending sync</p>}
        <p className="receipt-number">{(display as any).receipt_number}</p>
        <p>
          <strong>{(display as any).plate}</strong> ({(display as any).vehicle_type})
        </p>
        <p>Total: Rp {(display as any).fee_amount.toLocaleString("id-ID")}</p>
        <p>Method: {(display as any).payment_method}</p>
        <div className="actions">
          <button className="button primary" onClick={printReceipt}>
            Reprint Receipt
          </button>
          <button className="button secondary" onClick={() => navigate("/dashboard")}>
            Back to Dashboard
          </button>
        </div>
      </div>
    </div>
  );
}

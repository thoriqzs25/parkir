import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { payCash, payDigital } from "../lib/api";

export function Payment() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
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
      <h2>Payment</h2>
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
            {loading ? "Processing..." : "Pay"}
          </button>
        </form>
      </div>
    </div>
  );
}

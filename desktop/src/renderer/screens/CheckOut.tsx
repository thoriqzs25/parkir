import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { checkOut as serverCheckOut, listSessions } from "../lib/api";
import { useAuth } from "../contexts/AuthContext";
import { useOnlineStatus } from "../hooks/useOnlineStatus";
import { getRateCache, getStoredSessions, saveOfflineCheckOut, type LocalSession } from "../lib/offlineStore";
import type { Session } from "../types";

function useDebounce<T>(value: T, delay: number): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(id);
  }, [value, delay]);
  return debounced;
}

function formatWIB(date: string) {
  return new Date(date).toLocaleString("id-ID", { timeZone: "Asia/Jakarta" });
}

function calculateDurationHours(checkInAt: string, checkOutAt: string): number {
  const diff = new Date(checkOutAt).getTime() - new Date(checkInAt).getTime();
  const hours = Math.ceil(diff / 3600000);
  return Math.max(1, hours);
}

function calculateFeeFromCache(locationId: string, vehicleType: string, checkInAt: string, checkOutAt: string): number | null {
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

  const durationHours = calculateDurationHours(checkInAt, checkOutAt);
  let raw = rate.first_hour_rate;
  if (durationHours > 1) {
    raw += (durationHours - 1) * rate.subsequent_hourly_rate;
  }
  return Math.min(raw, rate.daily_flat_rate);
}

export function CheckOut() {
  const { currentLocation, openShift } = useAuth();
  const online = useOnlineStatus();
  const navigate = useNavigate();
  const [query, setQuery] = useState("");
  const [serverSessions, setServerSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const debouncedQuery = useDebounce(query, 400);

  const localSessions = useMemo(() => {
    if (!currentLocation) return [];
    return getStoredSessions().filter(
      (s) =>
        s.location_id === currentLocation.id &&
        (s.state === "ACTIVE" || s.state === "PENDING_PAYMENT") &&
        s.plate.toUpperCase().includes(debouncedQuery.toUpperCase())
    );
  }, [debouncedQuery, currentLocation]);

  useEffect(() => {
    if (!currentLocation || !online) {
      setServerSessions([]);
      return;
    }
    if (debouncedQuery.length < 2) {
      setServerSessions([]);
      return;
    }

    setLoading(true);
    listSessions({
      location_id: currentLocation.id,
      state: "ACTIVE",
      plate: debouncedQuery,
      limit: 20,
    })
      .then((res) => setServerSessions(res.items))
      .catch((err) => setError((err as Error).message))
      .finally(() => setLoading(false));
  }, [debouncedQuery, currentLocation, online]);

  const handleCheckOut = async (session: Session) => {
    setError(null);
    try {
      if (online) {
        const updated = await serverCheckOut(session.id);
        navigate(`/payment?sessionId=${updated.id}&fee=${updated.fee_amount || 0}`);
        return;
      }

      if (!openShift) {
        setError("No open shift");
        return;
      }

      const checkOutAt = new Date().toISOString();
      const fee =
        session.fee_amount ??
        calculateFeeFromCache(session.location_id, session.vehicle_type, session.check_in_at, checkOutAt);

      if (fee === null) {
        setError("No cached rate available. Cannot check out offline.");
        return;
      }

      saveOfflineCheckOut(
        {
          type: "check_out",
          data: {
            session_id: session.id,
            check_out_at: checkOutAt,
            fee_amount: fee,
          },
        },
        {
          state: "PENDING_PAYMENT",
          check_out_at: checkOutAt,
          fee_amount: fee,
        }
      );

      navigate(`/payment?sessionId=${session.id}&fee=${fee}`);
    } catch (err) {
      setError((err as Error).message);
    }
  };

  if (!currentLocation || !openShift) {
    return <div className="screen">No active shift. Please select a location.</div>;
  }

  return (
    <div className="screen">
      <button className="button secondary back" onClick={() => navigate("/dashboard")}>
        &larr; Back
      </button>
      <h2>Check Out {!online && <span className="offline-badge">OFFLINE</span>}</h2>
      <div className="card">
        <div className="form-group">
          <label>Search Plate</label>
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value.toUpperCase())}
            placeholder={online ? "Type at least 2 characters..." : "Search local sessions..."}
            autoFocus
          />
        </div>
        {error && <p className="error-message">{error}</p>}
        {loading && <p>Searching...</p>}
        <div className="session-list">
          {localSessions.map((session) => (
            <div key={`local-${session.id}`} className="session-row offline-row">
              <div>
                <strong>{session.plate}</strong> ({session.vehicle_type}) {session.pendingSync && <span>[LOCAL]</span>}
              </div>
              <div>In: {formatWIB(session.check_in_at)}</div>
              <button className="button primary" onClick={() => handleCheckOut(session)}>
                Check Out
              </button>
            </div>
          ))}
          {online &&
            serverSessions.map((session) => (
              <div key={`server-${session.id}`} className="session-row">
                <div>
                  <strong>{session.plate}</strong> ({session.vehicle_type})
                </div>
                <div>In: {formatWIB(session.check_in_at)}</div>
                <button className="button primary" onClick={() => handleCheckOut(session)}>
                  Check Out
                </button>
              </div>
            ))}
          {!loading && debouncedQuery.length >= 2 && localSessions.length === 0 && serverSessions.length === 0 && (
            <p>No active sessions found.</p>
          )}
        </div>
      </div>
    </div>
  );
}

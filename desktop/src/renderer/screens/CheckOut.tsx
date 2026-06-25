import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { checkOut as serverCheckOut, listSessions } from "../lib/api";
import { useAuth } from "../contexts/AuthContext";
import type { Session } from "../types";

function useDebounce<T>(value: T, delay: number): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(id);
  }, [value, delay]);
  return debounced;
}

export function CheckOut() {
  const { currentLocation, openShift } = useAuth();
  const navigate = useNavigate();
  const [query, setQuery] = useState("");
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const debouncedQuery = useDebounce(query, 400);

  useEffect(() => {
    if (!currentLocation) return;
    if (debouncedQuery.length < 2) {
      setSessions([]);
      return;
    }

    setLoading(true);
    listSessions({
      location_id: currentLocation.id,
      state: "ACTIVE",
      plate: debouncedQuery,
      limit: 20,
    })
      .then((res) => setSessions(res.items))
      .catch((err) => setError((err as Error).message))
      .finally(() => setLoading(false));
  }, [debouncedQuery, currentLocation]);

  const handleCheckOut = async (session: Session) => {
    setError(null);
    try {
      const updated = await serverCheckOut(session.id);
      navigate(`/payment?sessionId=${updated.id}&fee=${updated.fee_amount || 0}`);
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
      <h2>Check Out</h2>
      <div className="card">
        <div className="form-group">
          <label>Search Plate</label>
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value.toUpperCase())}
            placeholder="Type at least 2 characters..."
            autoFocus
          />
        </div>
        {error && <p className="error-message">{error}</p>}
        {loading && <p>Searching...</p>}
        <div className="session-list">
          {sessions.map((session) => (
            <div key={session.id} className="session-row">
              <div>
                <strong>{session.plate}</strong> ({session.vehicle_type})
              </div>
              <div>
                In: {new Date(session.check_in_at).toLocaleString("id-ID", { timeZone: "Asia/Jakarta" })}
              </div>
              <button className="button primary" onClick={() => handleCheckOut(session)}>
                Check Out
              </button>
            </div>
          ))}
          {!loading && debouncedQuery.length >= 2 && sessions.length === 0 && (
            <p>No active sessions found.</p>
          )}
        </div>
      </div>
    </div>
  );
}

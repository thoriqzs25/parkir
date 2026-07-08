import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { listSessions } from "../lib/api";
import { useAuth } from "../contexts/AuthContext";
import type { Session } from "../types";

export function History() {
  const { currentLocation } = useAuth();
  const navigate = useNavigate();
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!currentLocation) return;
    setLoading(true);
    listSessions({
      location_id: currentLocation.id,
      state: "CLOSED,VOIDED",
      limit: 50,
    })
      .then((res) => setSessions(res.items))
      .finally(() => setLoading(false));
  }, [currentLocation]);

  return (
    <div className="screen">
      <button className="button secondary back" onClick={() => navigate("/dashboard")}>
        &larr; Back
      </button>
      <h2>Session History</h2>
      {loading ? (
        <p>Loading...</p>
      ) : (
        <div className="card">
          <table className="data-table">
            <thead>
              <tr>
                <th>Plate</th>
                <th>Type</th>
                <th>State</th>
                <th>In</th>
                <th>Out</th>
              </tr>
            </thead>
            <tbody>
              {sessions.map((session) => (
                <tr key={session.id}>
                  <td>{session.plate}</td>
                  <td>{session.vehicle_type}</td>
                  <td>{session.state}</td>
                  <td>
                    {new Date(session.check_in_at).toLocaleString("id-ID", { timeZone: "Asia/Jakarta" })}
                  </td>
                  <td>
                    {session.check_out_at
                      ? new Date(session.check_out_at).toLocaleString("id-ID", { timeZone: "Asia/Jakarta" })
                      : "-"}
                  </td>
                </tr>
              ))}
              {sessions.length === 0 && (
                <tr>
                  <td colSpan={5} className="text-center">
                    No sessions found.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

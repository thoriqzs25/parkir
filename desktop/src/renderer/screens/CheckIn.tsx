import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { checkIn } from "../lib/api";
import { saveOfflineCheckIn, type LocalSession } from "../lib/offlineStore";
import { useAuth } from "../contexts/AuthContext";
import { useOnlineStatus } from "../hooks/useOnlineStatus";
import type { Session } from "../types";

const VEHICLE_TYPES: Array<"CAR" | "MOTO" | "TRUCK"> = ["CAR", "MOTO", "TRUCK"];

function formatWIB(date: string) {
  return new Date(date).toLocaleString("id-ID", { timeZone: "Asia/Jakarta" });
}

export function CheckIn() {
  const { currentLocation, openShift } = useAuth();
  const online = useOnlineStatus();
  const navigate = useNavigate();
  const [vehicleType, setVehicleType] = useState<"CAR" | "MOTO" | "TRUCK">("CAR");
  const [plate, setPlate] = useState("");
  const [cityCode, setCityCode] = useState("B");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastSession, setLastSession] = useState<Session | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!currentLocation || !openShift) return;

    setLoading(true);
    setError(null);

    const normalizedPlate = plate.toUpperCase().trim();
    const normalizedCity = cityCode.toUpperCase().trim();
    const now = new Date().toISOString();

    try {
      if (online) {
        const res = await checkIn({
          location_id: currentLocation.id,
          plate: normalizedPlate,
          city_code: normalizedCity,
          vehicle_type: vehicleType,
        });
        setLastSession(res.session);
      } else {
        const sessionId = crypto.randomUUID();
        const localSession: LocalSession = {
          id: sessionId,
          location_id: currentLocation.id,
          operator_id: openShift.operator_id,
          shift_id: openShift.id,
          plate: normalizedPlate,
          city_code: normalizedCity || "UNKNOWN",
          vehicle_type: vehicleType,
          state: "ACTIVE",
          check_in_at: now,
          offline_sync: true,
          sync_conflict: false,
          pendingSync: true,
          created_at: now,
          updated_at: now,
        };
        saveOfflineCheckIn(localSession, {
          type: "check_in",
          session: {
            id: sessionId,
            location_id: currentLocation.id,
            operator_id: openShift.operator_id,
            shift_id: openShift.id,
            plate: normalizedPlate,
            city_code: normalizedCity || "UNKNOWN",
            vehicle_type: vehicleType,
            check_in_at: now,
          },
        });
        setLastSession(localSession);
      }
      setPlate("");
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  const printTicket = () => {
    if (!lastSession || !currentLocation) return;
    const html = `
      <div style="padding: 24px; max-width: 320px; margin: 0 auto; text-align: center;">
        <h2 style="margin: 0 0 8px;">${currentLocation.name}</h2>
        <p style="margin: 0 0 16px;">Check-in Ticket</p>
        <hr />
        <p style="margin: 8px 0; font-size: 16px;"><strong>Plate:</strong> ${lastSession.plate}</p>
        <p style="margin: 8px 0; font-size: 16px;"><strong>Type:</strong> ${lastSession.vehicle_type}</p>
        <p style="margin: 8px 0; font-size: 14px;"><strong>Time:</strong> ${formatWIB(lastSession.check_in_at)}</p>
        <hr />
        <p style="font-size: 12px; color: #666;">${online ? "" : "OFFLINE - Keep this ticket"}</p>
      </div>
    `;
    window.electronAPI.printHtml(html).catch(() => {
      alert("Failed to print ticket, but the session was created.");
    });
  };

  if (!currentLocation || !openShift) {
    return <div className="screen">No active shift. Please select a location.</div>;
  }

  return (
    <div className="screen">
      <button className="button secondary back" onClick={() => navigate("/dashboard")}>
        &larr; Back
      </button>
      <h2>Check In {!online && <span className="offline-badge">OFFLINE</span>}</h2>
      <div className="card">
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Vehicle Type</label>
            <div className="segmented-control">
              {VEHICLE_TYPES.map((t) => (
                <button
                  type="button"
                  key={t}
                  className={vehicleType === t ? "active" : ""}
                  onClick={() => setVehicleType(t)}
                >
                  {t}
                </button>
              ))}
            </div>
          </div>
          <div className="form-row">
            <div className="form-group" style={{ flex: 0.3 }}>
              <label>City</label>
              <input
                value={cityCode}
                onChange={(e) => setCityCode(e.target.value)}
                maxLength={3}
                required
              />
            </div>
            <div className="form-group" style={{ flex: 1 }}>
              <label>Plate Number</label>
              <input
                value={plate}
                onChange={(e) => setPlate(e.target.value.toUpperCase())}
                placeholder="B1234XYZ"
                required
                autoFocus
              />
            </div>
          </div>
          {error && <p className="error-message">{error}</p>}
          <button className="button primary full" type="submit" disabled={loading}>
            {loading ? "Processing..." : online ? "Check In" : "Queue Check In"}
          </button>
        </form>
      </div>

      {lastSession && (
        <div className="card result-card">
          <h3>Ticket Created</h3>
          <p>
            <strong>{lastSession.plate}</strong> ({lastSession.vehicle_type})
          </p>
          <p>{formatWIB(lastSession.check_in_at)}</p>
          <button className="button primary" onClick={printTicket}>
            Print Ticket
          </button>
        </div>
      )}
    </div>
  );
}

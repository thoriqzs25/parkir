import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import { listLocations } from "../lib/api";
import type { Location } from "../types";

const API_BASE_URL = "http://localhost:8080";

interface GateInfo {
  device_id: string;
  ip: string;
  hostname: string;
  registered: boolean;
}

export function GateSetup() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [gateIP, setGateIP] = useState("");
  const [gateInfo, setGateInfo] = useState<GateInfo | null>(null);
  const [locations, setLocations] = useState<Location[]>([]);
  const [selectedLocationId, setSelectedLocationId] = useState("");
  const [gateName, setGateName] = useState("");
  const [verifying, setVerifying] = useState(false);
  const [registering, setRegistering] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [verifyError, setVerifyError] = useState<string | null>(null);

  useEffect(() => {
    if (!user) {
      navigate("/login");
      return;
    }
    listLocations()
      .then((res) => setLocations(res.items || []))
      .catch(() => {});
  }, [user, navigate]);

  const handleVerify = async () => {
    if (!gateIP.trim()) return;
    setVerifying(true);
    setVerifyError(null);
    setGateInfo(null);
    try {
      const res = await fetch(`http://${gateIP.trim()}:9800/info`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const info: GateInfo = await res.json();
      setGateInfo(info);
      setGateName(`Gate ${info.device_id.slice(0, 8)}`);
    } catch (err) {
      setVerifyError(err instanceof Error ? err.message : "Failed to connect to gate");
    } finally {
      setVerifying(false);
    }
  };

  const handleRegister = async () => {
    if (!gateInfo || !selectedLocationId) return;
    setRegistering(true);
    setError(null);
    try {
      const token = await window.electronAPI.getToken();

      // Register in backend
      const backendRes = await fetch(`${API_BASE_URL}/api/v1/gates`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          device_id: gateInfo.device_id,
          name: gateName,
          location_id: selectedLocationId,
          ip_address: gateInfo.ip,
        }),
      });
      if (!backendRes.ok) {
        const body = await backendRes.json().catch(() => ({}));
        throw new Error(body?.error?.message || `Backend error: ${backendRes.status}`);
      }

      // Tell gate display to configure itself
      const gateRes = await fetch(`http://${gateInfo.ip}:9800/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          location_id: selectedLocationId,
          api_url: API_BASE_URL,
        }),
      });
      if (!gateRes.ok) throw new Error("Gate registration failed");

      navigate("/dashboard");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
    } finally {
      setRegistering(false);
    }
  };

  return (
    <div className="screen">
      <button className="button secondary back" onClick={() => navigate("/dashboard")}>
        &larr; Back
      </button>
      <h2>Gate Setup</h2>

      <div className="card" style={{ padding: 16 }}>
        <h3 style={{ marginBottom: 12 }}>1. Find Gate</h3>
        <p style={{ fontSize: 13, color: "#888", marginBottom: 12 }}>
          Enter the IP address shown on the gate display's registration screen.
        </p>
        <div className="form-row">
          <input
            value={gateIP}
            onChange={(e) => setGateIP(e.target.value)}
            placeholder="e.g. 192.168.1.100"
            style={{ flex: 1 }}
          />
          <button className="button primary" onClick={handleVerify} disabled={verifying || !gateIP.trim()}>
            {verifying ? "Verifying..." : "Verify"}
          </button>
        </div>
        {verifyError && <p className="error-message">{verifyError}</p>}
        {gateInfo && (
          <div style={{ marginTop: 12, fontSize: 14 }}>
            <p><strong>Device ID:</strong> <span className="mono">{gateInfo.device_id}</span></p>
            <p><strong>Hostname:</strong> {gateInfo.hostname}</p>
            <p><strong>Status:</strong> {gateInfo.registered ? "Already registered" : "New gate"}</p>
          </div>
        )}
      </div>

      {gateInfo && (
        <div className="card" style={{ padding: 16, marginTop: 16 }}>
          <h3 style={{ marginBottom: 12 }}>2. Register Gate</h3>
          <div className="form-group">
            <label>Gate Name</label>
            <input
              value={gateName}
              onChange={(e) => setGateName(e.target.value)}
              placeholder="e.g. Gate A - Entry 1"
            />
          </div>
          <div className="form-group">
            <label>Location</label>
            <select value={selectedLocationId} onChange={(e) => setSelectedLocationId(e.target.value)}>
              <option value="">— Select Location —</option>
              {locations.map((loc) => (
                <option key={loc.id} value={loc.id}>{loc.name}</option>
              ))}
            </select>
          </div>
          {error && <p className="error-message">{error}</p>}
          <button
            className="button primary full"
            onClick={handleRegister}
            disabled={!selectedLocationId || registering}
          >
            {registering ? "Registering..." : "Register Gate"}
          </button>
        </div>
      )}
    </div>
  );
}

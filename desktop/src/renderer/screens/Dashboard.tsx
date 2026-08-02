import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

export function Dashboard() {
  const { user, currentLocation } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (!user) {
      navigate("/login");
      return;
    }
    if (!currentLocation) {
      navigate("/locations");
      return;
    }
  }, [user, currentLocation, navigate]);

  if (!currentLocation) {
    return <div className="screen">Loading...</div>;
  }

  return (
    <div className="screen dashboard-screen">
      <h2>Main Menu</h2>
      <div className="shift-info card">
        <p>
          <strong>Location:</strong> {currentLocation.name} ({currentLocation.code})
        </p>
      </div>
      <div className="menu-grid">
        <button className="menu-card" onClick={() => navigate("/check-in")}>
          <h3>Check In</h3>
          <p>Register a new vehicle entry</p>
        </button>
        <button className="menu-card" onClick={() => navigate("/check-out")}>
          <h3>Check Out</h3>
          <p>Search and close a session</p>
        </button>
        <button className="menu-card" onClick={() => navigate("/history")}>
          <h3>History</h3>
          <p>View closed sessions</p>
        </button>
        <button className="menu-card" onClick={() => navigate("/incident")}>
          <h3>Report Incident</h3>
          <p>File an operational issue</p>
        </button>
        <button className="menu-card" onClick={() => navigate("/gate-setup")}>
          <h3>Gate Setup</h3>
          <p>Register entry/exit gates</p>
        </button>
      </div>
    </div>
  );
}

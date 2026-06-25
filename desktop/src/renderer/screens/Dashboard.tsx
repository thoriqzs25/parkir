import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

export function Dashboard() {
  const { user, currentLocation, openShift, loading, endShift } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (!user) {
      navigate("/login");
      return;
    }
    if (!currentLocation || !openShift) {
      navigate("/locations");
      return;
    }
  }, [user, currentLocation, openShift, navigate]);

  const handleEndShift = async () => {
    const amount = window.prompt("Enter cash handover amount:", "0");
    if (amount === null) return;
    const notes = window.prompt("Discrepancy notes (optional):");
    try {
      await endShift(Number(amount), notes || undefined);
      navigate("/locations");
    } catch (err) {
      alert("Failed to end shift.");
    }
  };

  if (loading || !currentLocation || !openShift) {
    return <div className="screen">Loading...</div>;
  }

  return (
    <div className="screen dashboard-screen">
      <h2>Main Menu</h2>
      <div className="shift-info card">
        <p>
          <strong>Shift:</strong> {new Date(openShift.started_at).toLocaleString()}
        </p>
        <p>
          <strong>Location:</strong> {currentLocation.name} ({currentLocation.code})
        </p>
        <button className="button danger" onClick={handleEndShift}>
          End Shift
        </button>
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
      </div>
    </div>
  );
}

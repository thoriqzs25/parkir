import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import { EndShiftDialog } from "../components/EndShiftDialog";

export function Dashboard() {
  const { user, currentLocation, openShift, loading, endShift } = useAuth();
  const navigate = useNavigate();
  const [showEndShift, setShowEndShift] = useState(false);

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

  const handleEndShift = async (amount: number, notes?: string) => {
    setShowEndShift(false);
    try {
      await endShift(amount, notes);
      navigate("/locations");
    } catch {
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
        <button className="button danger" onClick={() => setShowEndShift(true)}>
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
        <button className="menu-card" onClick={() => navigate("/incident")}>
          <h3>Report Incident</h3>
          <p>File an operational issue</p>
        </button>
      </div>
      <EndShiftDialog
        open={showEndShift}
        onConfirm={handleEndShift}
        onCancel={() => setShowEndShift(false)}
      />
    </div>
  );
}

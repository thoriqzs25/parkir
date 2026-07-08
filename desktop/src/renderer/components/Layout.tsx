import { useState } from "react";
import { Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

export function Layout() {
  const { user, currentLocation, openShift, logout, endShift } = useAuth();
  const navigate = useNavigate();
  const [showLogoutConfirm, setShowLogoutConfirm] = useState(false);

  const handleLogout = async () => {
    setShowLogoutConfirm(false);
    if (openShift) {
      try {
        await endShift(0, "Logged out");
      } catch {
        // Proceed with logout regardless
      }
    }
    await logout();
    navigate("/login");
  };

  return (
    <div className="app">
      <header className="header">
        <div className="header-left">
          <h1 className="logo">PARKIR Desktop</h1>
          {currentLocation && (
            <span className="location-badge">
              {currentLocation.name} ({currentLocation.code})
            </span>
          )}
          {openShift && <span className="shift-badge">Shift Open</span>}
        </div>
        <div className="header-right">
          {user && (
            <span className="user-name">{user.name} ({user.role_name || user.role_id})</span>
          )}
          <button className="button secondary" onClick={() => setShowLogoutConfirm(true)}>
            Logout
          </button>
        </div>
      </header>
      <main className="main">
        <Outlet />
      </main>

      {showLogoutConfirm && (
        <div className="overlay" onClick={() => setShowLogoutConfirm(false)}>
          <div className="dialog card" onClick={(e) => e.stopPropagation()}>
            <h2>Logout</h2>
            {openShift ? (
              <p>You have an open shift. Logging out will mark your shift as ended. Are you sure?</p>
            ) : (
              <p>Are you sure you want to logout?</p>
            )}
            <div className="dialog-actions">
              <button className="button secondary" onClick={() => setShowLogoutConfirm(false)}>
                Cancel
              </button>
              <button className="button danger" onClick={handleLogout}>
                Logout
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

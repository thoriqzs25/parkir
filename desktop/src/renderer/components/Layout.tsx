import { useState } from "react";
import { Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

export function Layout() {
  const { user, currentLocation, logout } = useAuth();
  const navigate = useNavigate();
  const [showLogoutConfirm, setShowLogoutConfirm] = useState(false);

  const handleLogout = async () => {
    setShowLogoutConfirm(false);
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
            <p>Are you sure you want to logout?</p>
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

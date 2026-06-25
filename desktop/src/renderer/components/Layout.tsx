import { Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

export function Layout() {
  const { user, currentLocation, openShift, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
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
          <button className="button secondary" onClick={handleLogout}>
            Logout
          </button>
        </div>
      </header>
      <main className="main">
        <Outlet />
      </main>
    </div>
  );
}

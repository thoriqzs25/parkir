import React, { useEffect, useState } from "react";
import { MemoryRouter, Navigate, Route, Routes, useLocation } from "react-router-dom";
import { AuthProvider, useAuth } from "./contexts/AuthContext";
import { Layout } from "./components/Layout";
import { Login } from "./screens/Login";
import { LocationSelect } from "./screens/LocationSelect";
import { Dashboard } from "./screens/Dashboard";
import { CheckIn } from "./screens/CheckIn";
import { CheckOut } from "./screens/CheckOut";
import { Payment } from "./screens/Payment";
import { Success } from "./screens/Success";
import { History } from "./screens/History";
import { IncidentReport } from "./screens/IncidentReport";
import { useOnlineStatus } from "./hooks/useOnlineStatus";
import { getPendingItems } from "./lib/offlineStore";
import { syncPendingItems } from "./lib/sync";

function NetworkStatus() {
  const isOnline = useOnlineStatus();
  const pendingCount = getPendingItems().length;
  if (isOnline) {
    if (pendingCount > 0) {
      return (
        <div className="network-status sync-pending">Online — {pendingCount} records waiting to sync</div>
      );
    }
    return null;
  }
  return <div className="network-status">Offline — {pendingCount} records queued locally</div>;
}

function SyncManager() {
  const isOnline = useOnlineStatus();
  const { currentLocation } = useAuth();
  const [syncing, setSyncing] = useState(false);
  const [lastError, setLastError] = useState<string | null>(null);

  useEffect(() => {
    if (!isOnline) return;

    const pendingCount = getPendingItems().length;
    if (pendingCount === 0) return;

    setSyncing(true);
    syncPendingItems()
      .then(({ reprints }) => {
        if (reprints.length > 0 && currentLocation) {
          reprints.forEach(({ receiptNumber, plate, fee }) => {
            const html = `
              <div style="padding: 24px; max-width: 320px; margin: 0 auto; text-align: center;">
                <h2 style="margin: 0 0 8px;">${currentLocation.name}</h2>
                <p style="margin: 0 0 16px;">${receiptNumber}</p>
                <hr />
                <p style="margin: 8px 0; font-size: 16px;"><strong>Plate:</strong> ${plate}</p>
                <p style="margin: 8px 0; font-size: 18px;"><strong>Total: Rp ${fee.toLocaleString("id-ID")}</strong></p>
                <hr />
                <p style="font-size: 12px; color: #666;">Official receipt issued after sync</p>
              </div>
            `;
            window.electronAPI.printHtml(html).catch(() => {
              // eslint-disable-next-line no-alert
              alert(`Official receipt for ${plate}: ${receiptNumber}`);
            });
          });
        }
        setLastError(null);
      })
      .catch((err) => {
        setLastError(err instanceof Error ? err.message : "Sync failed");
      })
      .finally(() => setSyncing(false));
  }, [isOnline, currentLocation]);

  if (!isOnline) return null;
  if (syncing) return <div className="network-status sync-active">Syncing offline records...</div>;
  if (lastError) return <div className="network-status sync-error">Sync error: {lastError}</div>;
  return null;
}

function RequireUser({ children }: { children: React.ReactNode }) {
  const { user } = useAuth();
  const location = useLocation();
  if (!user) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }
  return <>{children}</>;
}

function RequireShift({ children }: { children: React.ReactNode }) {
  const { openShift } = useAuth();
  if (!openShift) {
    return <Navigate to="/locations" replace />;
  }
  return <>{children}</>;
}

function AuthenticatedLayout() {
  return (
    <RequireUser>
      <Layout />
    </RequireUser>
  );
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route element={<AuthenticatedLayout />}>
        <Route
          path="/locations"
          element={
            <RequireUser>
              <LocationSelect />
            </RequireUser>
          }
        />
        <Route
          path="/dashboard"
          element={
            <RequireShift>
              <Dashboard />
            </RequireShift>
          }
        />
        <Route
          path="/check-in"
          element={
            <RequireShift>
              <CheckIn />
            </RequireShift>
          }
        />
        <Route
          path="/check-out"
          element={
            <RequireShift>
              <CheckOut />
            </RequireShift>
          }
        />
        <Route
          path="/payment"
          element={
            <RequireShift>
              <Payment />
            </RequireShift>
          }
        />
        <Route
          path="/success"
          element={
            <RequireShift>
              <Success />
            </RequireShift>
          }
        />
        <Route
          path="/history"
          element={
            <RequireShift>
              <History />
            </RequireShift>
          }
        />
        <Route
          path="/incident"
          element={
            <RequireShift>
              <IncidentReport />
            </RequireShift>
          }
        />
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Route>
    </Routes>
  );
}

export function App() {
  return (
    <AuthProvider>
      <MemoryRouter>
        <NetworkStatus />
        <SyncManager />
        <AppRoutes />
      </MemoryRouter>
    </AuthProvider>
  );
}

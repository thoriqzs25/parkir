import React from "react";
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
import { useOnlineStatus } from "./hooks/useOnlineStatus";
import "./App.css";

function NetworkStatus() {
  const isOnline = useOnlineStatus();
  if (isOnline) return null;
  return <div className="network-status">Offline — online features may not work</div>;
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
        <AppRoutes />
      </MemoryRouter>
    </AuthProvider>
  );
}

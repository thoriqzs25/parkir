import React, { createContext, useContext, useEffect, useState } from "react";
import type { Location, Shift, User } from "../types";
import * as api from "../lib/api";
import { getRateCache, setRateCache } from "../lib/offlineStore";

interface AuthContextValue {
  user: User | null;
  locations: Location[];
  currentLocation: Location | null;
  openShift: Shift | null;
  permissions: string[];
  loading: boolean;
  error: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  setCurrentLocation: (location: Location | null) => void;
  refreshOpenShift: () => Promise<void>;
  startShift: (location: Location) => Promise<void>;
  endShift: (cashHandoverAmount: number, notes?: string) => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

const LOCAL_STORAGE_EMAIL_KEY = "parkir_desktop_last_email";

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [locations, setLocations] = useState<Location[]>([]);
  const [currentLocation, setCurrentLocation] = useState<Location | null>(null);
  const [openShift, setOpenShift] = useState<Shift | null>(null);
  const [permissions, setPermissions] = useState<string[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const rememberedEmail = localStorage.getItem(LOCAL_STORAGE_EMAIL_KEY);
    if (!rememberedEmail) return;

    async function restore() {
      setLoading(true);
      try {
        const { user: meUser, permissions: perms } = await api.me();
        const allLocations = await api.listLocations();
        const userLocationIds = meUser.location_ids || [];
        const userLocations = allLocations.items.filter((loc) =>
          userLocationIds.includes(loc.id)
        );
        setUser(meUser);
        setPermissions(perms);
        setLocations(userLocations);
        if (userLocations.length === 1) {
          setCurrentLocation(userLocations[0]);
          const shift = await api.getMyOpenShift().catch(() => null);
          setOpenShift(shift);
        }
      } catch (err) {
        setError("Session expired. Please log in again.");
      } finally {
        setLoading(false);
      }
    }

    restore().catch(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (!currentLocation || !navigator.onLine) return;

    const cache = getRateCache();
    const cacheTTL = 24 * 60 * 60 * 1000;
    if (cache && cache.locationId === currentLocation.id && new Date(cache.fetchedAt).getTime() > Date.now() - cacheTTL) {
      return;
    }

    api.listRates(currentLocation.id)
      .then((rates) => {
        setRateCache({
          locationId: currentLocation.id,
          rates,
          fetchedAt: new Date().toISOString(),
        });
      })
      .catch(() => {
        // Silently use stale cache if the request fails.
      });
  }, [currentLocation]);

  const login = async (email: string, password: string) => {
    setLoading(true);
    setError(null);
    try {
      await api.login(email, password);
      const { user: meUser, permissions: perms } = await api.me();
      const allLocations = await api.listLocations();
      const userLocationIds = meUser.location_ids || [];
      const userLocations = allLocations.items.filter((loc) =>
        userLocationIds.includes(loc.id)
      );
      setUser(meUser);
      setPermissions(perms);
      setLocations(userLocations);
      localStorage.setItem(LOCAL_STORAGE_EMAIL_KEY, email);
      const shift = await api.getMyOpenShift().catch(() => null);
      setOpenShift(shift);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Login failed";
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const logout = async () => {
    setLoading(true);
    try {
      await api.logout();
    } finally {
      setUser(null);
      setLocations([]);
      setCurrentLocation(null);
      setOpenShift(null);
      setPermissions([]);
      setError(null);
      setLoading(false);
    }
  };

  const refreshOpenShift = async () => {
    if (!currentLocation) {
      setOpenShift(null);
      return;
    }
    try {
      const shift = await api.getMyOpenShift();
      setOpenShift(shift);
    } catch {
      setOpenShift(null);
    }
  };

  const startShift = async (location: Location) => {
    const shift = await api.startShift({ location_id: location.id });
    setOpenShift(shift);
  };

  const endShift = async (cashHandoverAmount: number, notes?: string) => {
    if (!openShift) return;
    await api.endShift(openShift.id, {
      cash_handover_amount: cashHandoverAmount,
      discrepancy_notes: notes,
    });
    setOpenShift(null);
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        locations,
        currentLocation,
        openShift,
        permissions,
        loading,
        error,
        login,
        logout,
        setCurrentLocation,
        refreshOpenShift,
        startShift,
        endShift,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return ctx;
}

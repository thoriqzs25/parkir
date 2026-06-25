"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  ReactNode,
} from "react";
import { useParams, useRouter, usePathname } from "next/navigation";
import { listLocations } from "@/lib/api";
import { Location } from "@/types/location";
import { useAuth } from "@/hooks/useAuth";

interface LocationContextValue {
  locations: Location[];
  currentLocation: Location | null;
  setCurrentLocation: (location: Location) => void;
  loading: boolean;
}

const LocationContext = createContext<LocationContextValue | undefined>(
  undefined
);

export function LocationProvider({ children }: { children: ReactNode }) {
  const [locations, setLocations] = useState<Location[]>([]);
  const [currentLocation, setCurrentLocation] = useState<Location | null>(null);
  const [loading, setLoading] = useState(true);
  const params = useParams();
  const router = useRouter();
  const pathname = usePathname();
  const { user, loading: authLoading } = useAuth();

  useEffect(() => {
    if (authLoading) return;
    if (!user) {
      setLoading(false);
      return;
    }

    listLocations()
      .then((res) => {
        const items = res.items || [];
        setLocations(items);

        const locationId = params.locationId as string | undefined;
        if (locationId) {
          const found = items.find((l) => l.id === locationId);
          if (found) {
            setCurrentLocation(found);
            setLoading(false);
            return;
          }
        }

        // Default to first location
        if (items.length > 0) {
          setCurrentLocation(items[0]);
          const rest = pathname.replace(/^\/[^\/]+/, "");
          router.replace(`/${items[0].id}${rest || "/sessions/active"}`);
        }
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, [user, authLoading, params.locationId, router, pathname]);

  const handleSetLocation = (location: Location) => {
    setCurrentLocation(location);
    const rest = pathname.replace(/^\/[^\/]+/, "");
    router.push(`/${location.id}${rest || "/sessions/active"}`);
  };

  return (
    <LocationContext.Provider
      value={{
        locations,
        currentLocation,
        setCurrentLocation: handleSetLocation,
        loading,
      }}
    >
      {children}
    </LocationContext.Provider>
  );
}

export function useLocation() {
  const ctx = useContext(LocationContext);
  if (!ctx) {
    throw new Error("useLocation must be used within LocationProvider");
  }
  return ctx;
}

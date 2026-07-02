import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import type { Location } from "../types";

export function LocationSelect() {
  const { user, locations, currentLocation, setCurrentLocation, startShift, openShift } = useAuth();
  const navigate = useNavigate();
  const [selectedLocation, setSelectedLocation] = useState<Location | null>(null);
  const [starting, setStarting] = useState(false);

  useEffect(() => {
    if (!user) {
      navigate("/login");
    } else if (currentLocation && openShift) {
      navigate("/dashboard");
    }
  }, [user, currentLocation, openShift, navigate]);

  const handleSelect = (location: Location) => {
    setSelectedLocation(location);
    setCurrentLocation(location);
  };

  const handleStartShift = async () => {
    if (!selectedLocation) return;
    setStarting(true);
    try {
      await startShift(selectedLocation);
      navigate("/dashboard");
    } catch {
      alert("Could not start shift.");
    } finally {
      setStarting(false);
    }
  };

  const handleContinue = () => {
    navigate("/dashboard");
  };

  if (!user) return null;

  return (
    <div className="screen">
      <div className="card">
        <h2>Select Location</h2>
        <p>Choose the location where you will operate today.</p>
        <div className="location-grid">
          {locations.map((location) => (
            <button
              key={location.id}
              className={`location-card${selectedLocation?.id === location.id ? " selected" : ""}`}
              onClick={() => handleSelect(location)}
            >
              <h3>{location.name}</h3>
              <p>{location.code}</p>
              <p>{location.city}</p>
            </button>
          ))}
        </div>
        {selectedLocation && (
          <div className="location-actions">
            <button className="button primary" onClick={handleStartShift} disabled={starting}>
              {starting ? "Starting..." : "Start Shift"}
            </button>
            <button className="button secondary" onClick={handleContinue}>
              Continue without shift
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

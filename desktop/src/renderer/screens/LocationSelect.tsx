import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import type { Location } from "../types";

export function LocationSelect() {
  const { user, locations, currentLocation, setCurrentLocation, startShift, openShift } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (!user) {
      navigate("/login");
    } else if (currentLocation && openShift) {
      navigate("/dashboard");
    }
  }, [user, currentLocation, openShift, navigate]);

  const handleSelect = async (location: Location) => {
    setCurrentLocation(location);
    try {
      await startShift(location);
      navigate("/dashboard");
    } catch (err) {
      alert("Could not start shift.");
    }
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
              className="location-card"
              onClick={() => handleSelect(location)}
            >
              <h3>{location.name}</h3>
              <p>{location.code}</p>
              <p>{location.city}</p>
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}

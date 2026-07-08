import { useEffect, useState } from "react";

export function useOnlineStatus(): boolean {
  const [online, setOnline] = useState(true);

  useEffect(() => {
    setOnline(window.electronAPI ? window.electronAPI.isOnline() : navigator.onLine);

    if (window.electronAPI && window.electronAPI.onOnlineStatusChange) {
      window.electronAPI.onOnlineStatusChange((isOnline) => setOnline(isOnline));
    } else {
      const goOnline = () => setOnline(true);
      const goOffline = () => setOnline(false);
      window.addEventListener("online", goOnline);
      window.addEventListener("offline", goOffline);
      return () => {
        window.removeEventListener("online", goOnline);
        window.removeEventListener("offline", goOffline);
      };
    }
  }, []);

  return online;
}

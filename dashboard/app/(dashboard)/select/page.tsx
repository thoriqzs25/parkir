"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useLocation } from "@/components/layout/LocationProvider";

export default function SelectLocationPage() {
  const router = useRouter();
  const { currentLocation, loading } = useLocation();

  useEffect(() => {
    if (!loading && currentLocation) {
      router.replace(`/${currentLocation.id}/sessions/active`);
    }
  }, [loading, currentLocation, router]);

  return (
    <div className="flex min-h-screen items-center justify-center">
      <p className="text-gray-500">Selecting location...</p>
    </div>
  );
}

"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

export default function LocationHomePage() {
  const router = useRouter();

  useEffect(() => {
    router.replace("sessions/active");
  }, [router]);

  return null;
}

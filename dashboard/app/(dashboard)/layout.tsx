"use client";

import { AuthGuard } from "@/components/layout/AuthGuard";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { LocationProvider } from "@/components/layout/LocationProvider";

export default function DashboardRootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <AuthGuard>
      <LocationProvider>
        <DashboardLayout>{children}</DashboardLayout>
      </LocationProvider>
    </AuthGuard>
  );
}

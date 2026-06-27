"use client";

import { useParams, useRouter } from "next/navigation";
import { useAuth } from "@/hooks/useAuth";
import { hasPermission } from "@/lib/permissions";
import { Card } from "@/components/ui/Card";

export default function ReportsPage() {
  const { permissions } = useAuth();
  const params = useParams();
  const router = useRouter();
  const locationId = params.locationId as string;

  const reports = [
    { href: `/${locationId}/reports/daily-revenue`, label: "Daily Revenue", desc: "Revenue, transaction count, and averages per day", perm: "reports:view_revenue" },
    { href: `/${locationId}/reports/occupancy`, label: "Occupancy", desc: "Check-in volume over time (hourly/daily)", perm: "reports:view_occupancy" },
    { href: `/${locationId}/reports/vehicle-breakdown`, label: "Vehicle Breakdown", desc: "Count and revenue by vehicle type", perm: "reports:view_revenue" },
    { href: `/${locationId}/reports/operator-activity`, label: "Operator Activity", desc: "Sessions, revenue, and hours per operator", perm: "reports:view_operators" },
  ].filter((r) => hasPermission(permissions, r.perm));

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold">Reports</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
      {reports.map((r) => (
          <div key={r.href} className="cursor-pointer" onClick={() => router.push(r.href)}>
            <Card className="hover:shadow-md transition-shadow">
              <h3 className="text-lg font-semibold text-blue-600">{r.label}</h3>
              <p className="text-sm text-gray-500 mt-1">{r.desc}</p>
            </Card>
          </div>
        ))}
      </div>
    </div>
  );
}
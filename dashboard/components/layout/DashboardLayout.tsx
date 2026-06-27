"use client";

import Link from "next/link";
import { usePathname, useParams } from "next/navigation";
import {
  LayoutDashboard,
  Users,
  Shield,
  MapPin,
  Banknote,
  Car,
  Receipt,
  Clock,
  LogOut,
  AlertTriangle,
  AlertCircle,
  ScrollText,
  Activity,
  Bell,
  BarChart3,
} from "lucide-react";
import { clsx } from "clsx";
import { useAuth } from "@/hooks/useAuth";
import { useLocation } from "./LocationProvider";
import { Button } from "@/components/ui/Button";
import { hasPermission } from "@/lib/permissions";

export function DashboardLayout({ children }: { children: React.ReactNode }) {
  const { user, permissions, logout } = useAuth();
  const { locations, currentLocation, setCurrentLocation } = useLocation();
  const pathname = usePathname();
  const params = useParams();
  const locationId = (params.locationId as string) || currentLocation?.id || "";

  const nav = [
    { href: `/${locationId}/sessions/active`, label: "Active Sessions", icon: Car, perm: "sessions:view" },
    { href: `/${locationId}/sessions/history`, label: "History", icon: Clock, perm: "sessions:view" },
    { href: `/${locationId}/transactions`, label: "Transactions", icon: Receipt, perm: "sessions:view" },
    { href: `/${locationId}/sync-conflicts`, label: "Sync Conflicts", icon: AlertTriangle, perm: "sessions:view" },
    { href: `/${locationId}/shifts`, label: "Shifts", icon: Clock, perm: "shifts:view" },
    { href: `/${locationId}/incidents`, label: "Incidents", icon: AlertCircle, perm: "incidents:view" },
    { href: `/${locationId}/audit-logs`, label: "Audit Logs", icon: ScrollText, perm: "observability:view_audit" },
    { href: `/${locationId}/health`, label: "Health", icon: Activity, perm: "observability:view_health" },
    { href: `/${locationId}/alerts`, label: "Alerts", icon: Bell, perm: "observability:view_alerts" },
    { href: `/${locationId}/reports`, label: "Reports", icon: BarChart3, perm: "reports:view_revenue" },
    { href: `/${locationId}/locations`, label: "Locations", icon: MapPin, perm: "locations:view" },
    { href: `/${locationId}/rates`, label: "Rates", icon: Banknote, perm: "rates:view" },
    { href: `/${locationId}/users`, label: "Users", icon: Users, perm: "users:view" },
    { href: `/${locationId}/roles`, label: "Roles", icon: Shield, perm: "users:view" },
  ];

  return (
    <div className="flex min-h-screen bg-gray-50">
      <aside className="w-64 flex-shrink-0 border-r border-gray-200 bg-white">
        <div className="flex h-16 items-center px-6 border-b border-gray-200">
          <LayoutDashboard className="h-6 w-6 text-blue-600 mr-2" />
          <span className="text-lg font-bold">PARKIR</span>
        </div>

        <div className="p-4 border-b border-gray-200">
          <label className="block text-xs font-medium text-gray-500 mb-1">
            Location
          </label>
          <select
            value={currentLocation?.id || ""}
            onChange={(e) => {
              const loc = locations.find((l) => l.id === e.target.value);
              if (loc) setCurrentLocation(loc);
            }}
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
          >
            {locations.map((loc) => (
              <option key={loc.id} value={loc.id}>
                {loc.name}
              </option>
            ))}
          </select>
        </div>

        <nav className="p-4 space-y-1">
          {nav
            .filter((item) => hasPermission(permissions, item.perm))
            .map((item) => {
              const active = pathname.startsWith(item.href);
              const Icon = item.icon;
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={clsx(
                    "flex items-center rounded-md px-3 py-2 text-sm font-medium transition-colors",
                    active
                      ? "bg-blue-50 text-blue-700"
                      : "text-gray-700 hover:bg-gray-100"
                  )}
                >
                  <Icon className="mr-3 h-4 w-4" />
                  {item.label}
                </Link>
              );
            })}
        </nav>
      </aside>

      <div className="flex flex-1 flex-col">
        <header className="flex h-16 items-center justify-between border-b border-gray-200 bg-white px-6">
          <h1 className="text-lg font-semibold text-gray-900">
            {currentLocation?.name || "Dashboard"}
          </h1>
          <div className="flex items-center gap-4">
            <span className="text-sm text-gray-600">{user?.name}</span>
            <Button variant="ghost" size="sm" onClick={logout}>
              <LogOut className="mr-2 h-4 w-4" />
              Logout
            </Button>
          </div>
        </header>

        <main className="flex-1 p-6 overflow-auto">{children}</main>
      </div>
    </div>
  );
}

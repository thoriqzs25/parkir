"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { getHealthComponents } from "@/lib/api";
import { HealthComponents } from "@/types/alert";
import { Badge } from "@/components/ui/Badge";
import { Card, CardTitle } from "@/components/ui/Card";

export default function HealthPage() {
  const [health, setHealth] = useState<HealthComponents | null>(null);

  useEffect(() => {
    const fetchHealth = async () => {
      try {
        const data = await getHealthComponents();
        setHealth(data);
      } catch {
        setHealth({
          status: "error",
          components: {
            api: { status: "down", uptime_seconds: 0 },
            database: { status: "disconnected" },
          },
          last_check: new Date().toISOString(),
        });
      }
    };

    fetchHealth();
    const interval = setInterval(fetchHealth, 60000);
    return () => clearInterval(interval);
  }, []);

  if (!health) return <p className="text-gray-500">Loading health status...</p>;

  const apiUp = health.components.api.status === "up";
  const dbUp = health.components.database.status === "connected";
  const uptime = health.components.api.uptime_seconds;
  const uptimeDisplay = uptime > 86400
    ? `${Math.floor(uptime / 86400)}d ${Math.floor((uptime % 86400) / 3600)}h`
    : uptime > 3600
    ? `${Math.floor(uptime / 3600)}h ${Math.floor((uptime % 3600) / 60)}m`
    : `${Math.floor(uptime / 60)}m ${uptime % 60}s`;

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold">System Health</h2>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card>
          <CardTitle>API Server</CardTitle>
          <div className="mt-2 space-y-2">
            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-500">Status:</span>
              <Badge variant={apiUp ? "success" : "danger"}>{apiUp ? "Up" : "Down"}</Badge>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-500">Uptime:</span>
              <span className="text-sm font-mono">{uptimeDisplay}</span>
            </div>
          </div>
        </Card>

        <Card>
          <CardTitle>Database</CardTitle>
          <div className="mt-2 space-y-2">
            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-500">Status:</span>
              <Badge variant={dbUp ? "success" : "danger"}>{dbUp ? "Connected" : "Disconnected"}</Badge>
            </div>
          </div>
        </Card>
      </div>

      <p className="text-xs text-gray-400">
        Last checked: {new Date(health.last_check).toLocaleString("id-ID")}
        <span className="ml-2">(auto-refreshes every 60s)</span>
      </p>
    </div>
  );
}
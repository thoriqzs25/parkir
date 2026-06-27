"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { listAlerts, acknowledgeAlert, resolveAlert } from "@/lib/api";
import { Alert } from "@/types/alert";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { Dialog } from "@/components/ui/Dialog";
import { formatWIBDateTime } from "@/lib/time";

const codeLabels: Record<string, string> = {
  LONG_SESSION: "Long Session",
  UNPAID_EXIT: "Unpaid Exit",
  HIGH_VOID_RATE: "High Void Rate",
  SYNC_FAILURE: "Sync Failure",
  GATEWAY_FAILURE: "Gateway Failure",
};

const stateVariants: Record<string, "danger" | "warning" | "success"> = {
  TRIGGERED: "danger",
  ACKNOWLEDGED: "warning",
  RESOLVED: "success",
};

export default function AlertsPage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [offset, setOffset] = useState(0);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [stateFilter, setStateFilter] = useState("TRIGGERED");
  const [resolveDialog, setResolveDialog] = useState<string | null>(null);
  const [resolutionNotes, setResolutionNotes] = useState("");
  const limit = 20;

  const load = async (newOffset = 0) => {
    setLoading(true);
    try {
      const q: Record<string, string> = {
        location_id: locationId,
        limit: String(limit),
        offset: String(newOffset),
      };
      if (stateFilter) q.state = stateFilter;
      const res = await listAlerts(q);
      setAlerts(newOffset === 0 ? res.items || [] : [...alerts, ...(res.items || [])]);
      setTotal(res.meta?.total || 0);
      setOffset(newOffset);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load alerts");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load(0);
  }, [locationId, stateFilter]);

  const handleAcknowledge = async (id: string) => {
    try {
      await acknowledgeAlert(id);
      toast.success("Alert acknowledged");
      load(0);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to acknowledge alert");
    }
  };

  const handleResolve = async () => {
    if (!resolveDialog || !resolutionNotes) return;
    try {
      await resolveAlert(resolveDialog, resolutionNotes);
      toast.success("Alert resolved");
      setResolveDialog(null);
      setResolutionNotes("");
      load(0);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to resolve alert");
    }
  };

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold">Alerts</h2>

      <div className="flex gap-2">
        {["TRIGGERED", "ACKNOWLEDGED", "RESOLVED", ""].map((s) => (
          <Button
            key={s}
            variant={stateFilter === s ? "primary" : "secondary"}
            size="sm"
            onClick={() => setStateFilter(s)}
          >
            {s || "All"}
          </Button>
        ))}
      </div>

      {alerts.length === 0 && !loading ? (
        <p className="text-gray-500">No alerts found.</p>
      ) : (
        <Table>
          <Thead>
            <tr>
              <Th>Code</Th>
              <Th>State</Th>
              <Th>Triggered</Th>
              <Th>Entity</Th>
              <Th>Actions</Th>
            </tr>
          </Thead>
          <Tbody>
            {alerts.map((a) => (
              <tr key={a.id}>
                <Td><code className="text-xs bg-gray-100 px-1 rounded">{codeLabels[a.code] || a.code}</code></Td>
                <Td><Badge variant={stateVariants[a.state]}>{a.state}</Badge></Td>
                <Td>{formatWIBDateTime(a.triggered_at)}</Td>
                <Td>{a.entity_type ? `${a.entity_type}:${a.entity_id?.substring(0, 8)}` : "-"}</Td>
                <Td>
                  <div className="flex gap-1">
                    {a.state === "TRIGGERED" && (
                      <Button variant="ghost" size="sm" onClick={() => handleAcknowledge(a.id)}>
                        Acknowledge
                      </Button>
                    )}
                    {a.state !== "RESOLVED" && (
                      <Button variant="ghost" size="sm" onClick={() => setResolveDialog(a.id)}>
                        Resolve
                      </Button>
                    )}
                  </div>
                </Td>
              </tr>
            ))}
          </Tbody>
        </Table>
      )}

      {alerts.length < total && (
        <Button variant="secondary" onClick={() => load(offset + limit)} disabled={loading}>
          {loading ? "Loading..." : "Load more"}
        </Button>
      )}

      <Dialog open={!!resolveDialog} onClose={() => setResolveDialog(null)} title="Resolve Alert">
        <div className="space-y-4">
          <Input
            label="Resolution Notes"
            placeholder="Describe how the alert was resolved..."
            value={resolutionNotes}
            onChange={(e) => setResolutionNotes(e.target.value)}
          />
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => setResolveDialog(null)}>Cancel</Button>
            <Button onClick={handleResolve} disabled={!resolutionNotes}>Resolve</Button>
          </div>
        </div>
      </Dialog>
    </div>
  );
}
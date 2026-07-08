"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { listSyncConflicts, resolveSyncConflict } from "@/lib/api";
import { Session } from "@/types/session";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { formatWIBDateTime } from "@/lib/time";

export default function SyncConflictsPage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const [conflicts, setConflicts] = useState<Session[]>([]);
  const [loading, setLoading] = useState(false);
  const [resolving, setResolving] = useState<Record<string, boolean>>({});

  const load = async () => {
    setLoading(true);
    try {
      const res = await listSyncConflicts({
        location_id: locationId,
        limit: "50",
        offset: "0",
      });
      setConflicts(res.items || []);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load sync conflicts");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, [locationId]);

  const handleResolve = async (id: string, action: "VOID_OFFLINE" | "IGNORE") => {
    setResolving((prev) => ({ ...prev, [id]: true }));
    try {
      await resolveSyncConflict(id, {
        action,
        void_reason: action === "VOID_OFFLINE" ? "Duplicate active plate" : undefined,
      });
      toast.success("Conflict resolved");
      await load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to resolve conflict");
    } finally {
      setResolving((prev) => ({ ...prev, [id]: false }));
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Sync Conflicts</h2>
        <Button variant="secondary" onClick={load} disabled={loading}>
          {loading ? "Refreshing..." : "Refresh"}
        </Button>
      </div>

      <p className="text-sm text-gray-600">
        These offline check-ins conflict with another active session at the same location.
        Choose to void the offline record or ignore the conflict if it is harmless.
      </p>

      {conflicts.length === 0 && !loading ? (
        <p className="text-gray-500">No sync conflicts.</p>
      ) : (
        <Table>
          <Thead>
            <tr>
              <Th>Plate</Th>
              <Th>Type</Th>
              <Th>Check In</Th>
              <Th>State</Th>
              <Th>Actions</Th>
            </tr>
          </Thead>
          <Tbody>
            {conflicts.map((c) => (
              <tr key={c.id}>
                <Td>
                  <span className="font-medium">{c.plate}</span>
                </Td>
                <Td>{c.vehicle_type}</Td>
                <Td>{formatWIBDateTime(c.check_in_at)}</Td>
                <Td>
                  <Badge variant="danger">CONFLICT</Badge>
                </Td>
                <Td>
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      variant="danger"
                      onClick={() => handleResolve(c.id, "VOID_OFFLINE")}
                      disabled={resolving[c.id]}
                    >
                      Void Offline
                    </Button>
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={() => handleResolve(c.id, "IGNORE")}
                      disabled={resolving[c.id]}
                    >
                      Ignore
                    </Button>
                  </div>
                </Td>
              </tr>
            ))}
          </Tbody>
        </Table>
      )}
    </div>
  );
}

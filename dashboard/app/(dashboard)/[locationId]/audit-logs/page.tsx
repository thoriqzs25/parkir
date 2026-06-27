"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { listAuditLogs, exportAuditLogs } from "@/lib/api";
import { AuditLog } from "@/types/auditlog";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { formatWIBDateTime } from "@/lib/time";

export default function AuditLogsPage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [offset, setOffset] = useState(0);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [actionFilter, setActionFilter] = useState("");
  const limit = 50;

  const load = async (newOffset = 0) => {
    setLoading(true);
    try {
      const q: Record<string, string> = {
        location_id: locationId,
        limit: String(limit),
        offset: String(newOffset),
      };
      if (actionFilter) q.action = actionFilter;
      const res = await listAuditLogs(q);
      setLogs(newOffset === 0 ? res.items || [] : [...logs, ...(res.items || [])]);
      setTotal(res.meta?.total || 0);
      setOffset(newOffset);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load audit logs");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load(0);
  }, [locationId, actionFilter]);

  const handleExportCSV = () => {
    const q: Record<string, string> = { location_id: locationId };
    if (actionFilter) q.action = actionFilter;
    const url = exportAuditLogs(q);
    window.open(url, "_blank");
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Audit Logs</h2>
        <Button variant="secondary" size="sm" onClick={handleExportCSV}>
          Export CSV
        </Button>
      </div>

      <div className="flex gap-2">
        <Input
          placeholder="Filter by action..."
          value={actionFilter}
          onChange={(e) => setActionFilter(e.target.value)}
        />
      </div>

      {logs.length === 0 && !loading ? (
        <p className="text-gray-500">No audit logs found.</p>
      ) : (
        <Table>
          <Thead>
            <tr>
              <Th>Timestamp</Th>
              <Th>Action</Th>
              <Th>Actor</Th>
              <Th>Role</Th>
              <Th>Entity</Th>
              <Th>Entity ID</Th>
            </tr>
          </Thead>
          <Tbody>
            {logs.map((l) => (
              <tr key={l.id}>
                <Td>{formatWIBDateTime(l.timestamp)}</Td>
                <Td><code className="text-xs bg-gray-100 px-1 rounded">{l.action}</code></Td>
                <Td>{l.actor_id ? l.actor_id.substring(0, 8) : "-"}</Td>
                <Td>{l.actor_role || "-"}</Td>
                <Td>{l.entity_type}</Td>
                <Td className="max-w-[120px] truncate">{l.entity_id.substring(0, 8)}</Td>
              </tr>
            ))}
          </Tbody>
        </Table>
      )}

      {logs.length < total && (
        <Button variant="secondary" onClick={() => load(offset + limit)} disabled={loading}>
          {loading ? "Loading..." : "Load more"}
        </Button>
      )}
    </div>
  );
}
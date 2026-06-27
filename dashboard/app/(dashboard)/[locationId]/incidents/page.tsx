"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { toast } from "sonner";
import { listIncidents } from "@/lib/api";
import { Incident } from "@/types/incident";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { formatWIBDateTime } from "@/lib/time";

const typeLabels: Record<string, string> = {
  STUCK_AT_GATE: "Stuck at Gate",
  PAYMENT_DISPUTE: "Payment Dispute",
  OPERATOR_ERROR: "Operator Error",
  SYSTEM_DOWNTIME: "System Downtime",
};

const stateVariants: Record<string, "danger" | "warning" | "success"> = {
  OPEN: "danger",
  IN_PROGRESS: "warning",
  RESOLVED: "success",
};

export default function IncidentsPage() {
  const params = useParams();
  const router = useRouter();
  const locationId = params.locationId as string;
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [offset, setOffset] = useState(0);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [stateFilter, setStateFilter] = useState("OPEN");
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
      const res = await listIncidents(q);
      setIncidents(newOffset === 0 ? res.items || [] : [...incidents, ...(res.items || [])]);
      setTotal(res.meta?.total || 0);
      setOffset(newOffset);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load incidents");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load(0);
  }, [locationId, stateFilter]);

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold">Incidents</h2>

      <div className="flex gap-2">
        {["OPEN", "IN_PROGRESS", "RESOLVED", ""].map((s) => (
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

      {incidents.length === 0 && !loading ? (
        <p className="text-gray-500">No incidents found.</p>
      ) : (
        <Table>
          <Thead>
            <tr>
              <Th>Type</Th>
              <Th>State</Th>
              <Th>Description</Th>
              <Th>Reported</Th>
              <Th>Actions</Th>
            </tr>
          </Thead>
          <Tbody>
            {incidents.map((inc) => (
              <tr key={inc.id}>
                <Td>{typeLabels[inc.type] || inc.type}</Td>
                <Td><Badge variant={stateVariants[inc.state]}>{inc.state}</Badge></Td>
                <Td className="max-w-xs truncate">{inc.description}</Td>
                <Td>{formatWIBDateTime(inc.reported_at)}</Td>
                <Td>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => router.push(`/${locationId}/incidents/${inc.id}`)}
                  >
                    View
                  </Button>
                </Td>
              </tr>
            ))}
          </Tbody>
        </Table>
      )}

      {incidents.length < total && (
        <Button variant="secondary" onClick={() => load(offset + limit)} disabled={loading}>
          {loading ? "Loading..." : "Load more"}
        </Button>
      )}
    </div>
  );
}
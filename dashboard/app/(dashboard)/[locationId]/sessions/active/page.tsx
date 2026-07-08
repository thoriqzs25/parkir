"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { toast } from "sonner";
import { RefreshCw } from "lucide-react";
import { listSessions } from "@/lib/api";
import { Session } from "@/types/session";
import { Button } from "@/components/ui/Button";
import { Badge } from "@/components/ui/Badge";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { formatWIBDateTime } from "@/lib/time";

export default function ActiveSessionsPage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(true);

  const load = async () => {
    setLoading(true);
    try {
      const res = await listSessions({
        location_id: locationId,
        state: "ACTIVE,PENDING_PAYMENT",
        limit: "50",
      });
      setSessions(res.items || []);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load sessions");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, [locationId]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Active Sessions</h2>
        <Button variant="secondary" size="sm" onClick={load} disabled={loading}>
          <RefreshCw className="mr-2 h-4 w-4" />
          Refresh
        </Button>
      </div>

      {loading && sessions.length === 0 ? (
        <p className="text-gray-500">Loading...</p>
      ) : sessions.length === 0 ? (
        <p className="text-gray-500">No active sessions.</p>
      ) : (
        <Table>
          <Thead>
            <tr>
              <Th>Plate</Th>
              <Th>Type</Th>
              <Th>State</Th>
              <Th>Check In</Th>
              <Th>Fee</Th>
            </tr>
          </Thead>
          <Tbody>
            {sessions.map((s) => (
              <tr key={s.id} className="cursor-pointer hover:bg-gray-50">
                <Td>
                  <Link href={`/${locationId}/sessions/${s.id}`} className="font-medium text-blue-600 hover:underline">
                    {s.plate}
                  </Link>
                </Td>
                <Td>{s.vehicle_type}</Td>
                <Td>
                  <Badge
                    variant={
                      s.state === "ACTIVE"
                        ? "success"
                        : s.state === "PENDING_PAYMENT"
                        ? "warning"
                        : "default"
                    }
                  >
                    {s.state}
                  </Badge>
                </Td>
                <Td>{formatWIBDateTime(s.check_in_at)}</Td>
                <Td>
                  {s.fee_amount !== undefined
                    ? `Rp ${s.fee_amount.toLocaleString("id-ID")}`
                    : "-"}
                </Td>
              </tr>
            ))}
          </Tbody>
        </Table>
      )}
    </div>
  );
}

"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { toast } from "sonner";
import { listSessions } from "@/lib/api";
import { Session } from "@/types/session";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { formatWIBDateTime } from "@/lib/time";

export default function SessionHistoryPage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const [sessions, setSessions] = useState<Session[]>([]);
  const [plate, setPlate] = useState("");
  const [offset, setOffset] = useState(0);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const limit = 20;

  const load = async (newOffset = 0) => {
    setLoading(true);
    try {
      const q: Record<string, string> = {
        location_id: locationId,
        state: "CLOSED,VOIDED",
        limit: String(limit),
        offset: String(newOffset),
      };
      if (plate) q.plate = plate;
      const res = await listSessions(q);
      setSessions(newOffset === 0 ? res.items || [] : [...sessions, ...(res.items || [])]);
      setTotal(res.meta?.total || 0);
      setOffset(newOffset);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load sessions");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load(0);
  }, [locationId, plate]);

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold">Session History</h2>

      <div className="flex max-w-md gap-2">
        <Input
          placeholder="Search plate..."
          value={plate}
          onChange={(e) => setPlate(e.target.value)}
        />
      </div>

      {sessions.length === 0 && !loading ? (
        <p className="text-gray-500">No sessions found.</p>
      ) : (
        <Table>
          <Thead>
            <tr>
              <Th>Plate</Th>
              <Th>Type</Th>
              <Th>State</Th>
              <Th>Check In</Th>
              <Th>Check Out</Th>
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
                  <Badge variant={s.state === "VOIDED" ? "danger" : "default"}>
                    {s.state}
                  </Badge>
                </Td>
                <Td>{formatWIBDateTime(s.check_in_at)}</Td>
                <Td>{s.check_out_at ? formatWIBDateTime(s.check_out_at) : "-"}</Td>
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

      {sessions.length < total && (
        <Button
          variant="secondary"
          onClick={() => load(offset + limit)}
          disabled={loading}
        >
          {loading ? "Loading..." : "Load more"}
        </Button>
      )}
    </div>
  );
}

"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { toast } from "sonner";
import { listShifts } from "@/lib/api";
import { Shift } from "@/types/shift";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { formatWIBDateTime } from "@/lib/time";

export default function ShiftsPage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const [shifts, setShifts] = useState<Shift[]>([]);
  const [offset, setOffset] = useState(0);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const limit = 20;

  const load = async (newOffset = 0) => {
    setLoading(true);
    try {
      const res = await listShifts({
        location_id: locationId,
        limit: String(limit),
        offset: String(newOffset),
      });
      setShifts(newOffset === 0 ? res.items || [] : [...shifts, ...(res.items || [])]);
      setTotal(res.meta?.total || 0);
      setOffset(newOffset);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load shifts");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load(0);
  }, [locationId]);

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold">Shifts</h2>

      {shifts.length === 0 && !loading ? (
        <p className="text-gray-500">No shifts found.</p>
      ) : (
        <Table>
          <Thead>
            <tr>
              <Th>Status</Th>
              <Th>Started</Th>
              <Th>Ended</Th>
              <Th>Expected Cash</Th>
              <Th>Handover</Th>
              <Th>Discrepancy</Th>
            </tr>
          </Thead>
          <Tbody>
            {shifts.map((s) => (
              <tr key={s.id} className="cursor-pointer hover:bg-gray-50">
                <Td>
                  <Link href={`/${locationId}/shifts/${s.id}`} className="font-medium text-blue-600 hover:underline">
                    <Badge
                      variant={
                        s.status === "OPEN"
                          ? "success"
                          : s.status === "FORCE_CLOSED"
                          ? "danger"
                          : s.status === "FLAGGED"
                          ? "warning"
                          : "default"
                      }
                    >
                      {s.status}
                    </Badge>
                  </Link>
                </Td>
                <Td>{formatWIBDateTime(s.started_at)}</Td>
                <Td>{s.ended_at ? formatWIBDateTime(s.ended_at) : "-"}</Td>
                <Td>
                  {s.expected_cash !== undefined
                    ? `Rp ${s.expected_cash.toLocaleString("id-ID")}`
                    : "-"}
                </Td>
                <Td>
                  {s.cash_handover_amount !== undefined
                    ? `Rp ${s.cash_handover_amount.toLocaleString("id-ID")}`
                    : "-"}
                </Td>
                <Td>
                  {s.discrepancy !== undefined
                    ? `Rp ${s.discrepancy.toLocaleString("id-ID")}`
                    : "-"}
                </Td>
              </tr>
            ))}
          </Tbody>
        </Table>
      )}

      {shifts.length < total && (
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

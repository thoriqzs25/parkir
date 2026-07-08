"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { getOperatorActivity, getOperatorActivityCSVUrl } from "@/lib/api";
import { OperatorActivityRow } from "@/types/report";
import { Card, CardTitle } from "@/components/ui/Card";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
} from "recharts";

export default function OperatorActivityPage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const [rows, setRows] = useState<OperatorActivityRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [dateFrom, setDateFrom] = useState(() => {
    const d = new Date(); d.setDate(d.getDate() - 7);
    return d.toISOString().split("T")[0];
  });
  const [dateTo, setDateTo] = useState(() => new Date().toISOString().split("T")[0]);
  const [operatorFilter, setOperatorFilter] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const p: Record<string, string> = {
        location_id: locationId,
        date_from: dateFrom,
        date_to: dateTo,
      };
      if (operatorFilter) p.operator_id = operatorFilter;
      const data = await getOperatorActivity(p);
      setRows(data || []);
    } catch (err) {
      toast.error("Failed to load operator activity");
    } finally {
      setLoading(false);
    }
  }, [locationId, dateFrom, dateTo, operatorFilter]);

  useEffect(() => { load(); }, [load]);

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold">Operator Activity</h2>

      <div className="flex flex-wrap gap-2 items-end">
        <Input label="From" type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} />
        <Input label="To" type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)} />
        <Input label="Operator ID (optional)" placeholder="Filter by operator" value={operatorFilter} onChange={(e) => setOperatorFilter(e.target.value)} />
        <Button variant="secondary" size="sm" onClick={load}>Refresh</Button>
        <Button variant="secondary" size="sm" onClick={() => window.open(getOperatorActivityCSVUrl({ location_id: locationId, date_from: dateFrom, date_to: dateTo }), "_blank")}>
          Export CSV
        </Button>
        <Button variant="secondary" size="sm" onClick={() => window.print()}>Print PDF</Button>
      </div>

      {rows.length > 0 && (
        <div className="bg-white rounded-lg border p-4">
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={rows.map(r => ({ ...r, name: r.operator_name.length > 15 ? r.operator_name.slice(0, 15) + "…" : r.operator_name }))} layout="vertical">
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis type="number" tick={{ fontSize: 12 }} />
              <YAxis dataKey="name" type="category" tick={{ fontSize: 12 }} width={120} />
              <Tooltip />
              <Bar dataKey="total_revenue" fill="#8b5cf6" name="Revenue" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}

      {loading ? <p className="text-gray-500">Loading...</p> : (
        <Table>
          <Thead><tr><Th>Operator</Th><Th>Sessions</Th><Th>Revenue</Th><Th>Shift Hours</Th></tr></Thead>
          <Tbody>
            {rows.map((r) => (
              <tr key={r.operator_id}>
                <Td className="font-medium">{r.operator_name}</Td>
                <Td>{r.session_count}</Td>
                <Td>Rp {r.total_revenue.toLocaleString("id-ID")}</Td>
                <Td>{r.shift_hours.toFixed(1)}h</Td>
              </tr>
            ))}
          </Tbody>
        </Table>
      )}
    </div>
  );
}
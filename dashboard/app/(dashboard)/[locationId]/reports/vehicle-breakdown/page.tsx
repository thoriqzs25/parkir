"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { getVehicleBreakdown, getVehicleBreakdownCSVUrl } from "@/lib/api";
import { VehicleBreakdownRow } from "@/types/report";
import { Card, CardTitle } from "@/components/ui/Card";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import {
  PieChart, Pie, Cell, Tooltip, ResponsiveContainer, Legend,
} from "recharts";

const COLORS = ["#3b82f6", "#10b981", "#f59e0b"];

export default function VehicleBreakdownPage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const [rows, setRows] = useState<VehicleBreakdownRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [dateFrom, setDateFrom] = useState(() => {
    const d = new Date(); d.setDate(d.getDate() - 7);
    return d.toISOString().split("T")[0];
  });
  const [dateTo, setDateTo] = useState(() => new Date().toISOString().split("T")[0]);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await getVehicleBreakdown({
        location_id: locationId,
        date_from: dateFrom,
        date_to: dateTo,
      });
      setRows(data || []);
    } catch (err) {
      toast.error("Failed to load vehicle breakdown");
    } finally {
      setLoading(false);
    }
  }, [locationId, dateFrom, dateTo]);

  useEffect(() => { load(); }, [load]);

  const totalRevenue = rows.reduce((s, r) => s + r.total_revenue, 0);
  const totalCount = rows.reduce((s, r) => s + r.count, 0);

  const pieData = rows.map((r) => ({
    name: r.vehicle_type,
    value: r.total_revenue,
  }));

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold">Vehicle Breakdown</h2>

      <div className="flex flex-wrap gap-2 items-end">
        <Input label="From" type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} />
        <Input label="To" type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)} />
        <Button variant="secondary" size="sm" onClick={load}>Refresh</Button>
        <Button variant="secondary" size="sm" onClick={() => window.open(getVehicleBreakdownCSVUrl({ location_id: locationId, date_from: dateFrom, date_to: dateTo }), "_blank")}>
          Export CSV
        </Button>
        <Button variant="secondary" size="sm" onClick={() => window.print()}>Print PDF</Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card><CardTitle>Total Revenue</CardTitle><p className="text-2xl font-bold">Rp {totalRevenue.toLocaleString("id-ID")}</p></Card>
        <Card><CardTitle>Total Vehicles</CardTitle><p className="text-2xl font-bold">{totalCount}</p></Card>
      </div>

      {rows.length > 0 && (
        <div className="bg-white rounded-lg border p-4">
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie data={pieData} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={100} label>
                {pieData.map((_, idx) => <Cell key={idx} fill={COLORS[idx % COLORS.length]} />)}
              </Pie>
              <Tooltip />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
        </div>
      )}

      {loading ? <p className="text-gray-500">Loading...</p> : (
        <Table>
          <Thead><tr><Th>Vehicle Type</Th><Th>Count</Th><Th>Revenue</Th><Th>% of Total</Th></tr></Thead>
          <Tbody>
            {rows.map((r) => (
              <tr key={r.vehicle_type}>
                <Td className="font-medium">{r.vehicle_type}</Td>
                <Td>{r.count}</Td>
                <Td>Rp {r.total_revenue.toLocaleString("id-ID")}</Td>
                <Td>{totalRevenue > 0 ? ((r.total_revenue / totalRevenue) * 100).toFixed(1) : 0}%</Td>
              </tr>
            ))}
          </Tbody>
        </Table>
      )}
    </div>
  );
}
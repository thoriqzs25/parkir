"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { getDailyRevenue, getDailyRevenueCSVUrl } from "@/lib/api";
import { DailyRevenueRow } from "@/types/report";
import { Card, CardTitle } from "@/components/ui/Card";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
} from "recharts";

export default function DailyRevenuePage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const [rows, setRows] = useState<DailyRevenueRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [dateFrom, setDateFrom] = useState(() => {
    const d = new Date(); d.setDate(d.getDate() - 7);
    return d.toISOString().split("T")[0];
  });
  const [dateTo, setDateTo] = useState(() => new Date().toISOString().split("T")[0]);
  const [includeVoided, setIncludeVoided] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await getDailyRevenue({
        location_id: locationId,
        date_from: dateFrom,
        date_to: dateTo,
        include_voided: String(includeVoided),
      });
      setRows(data || []);
    } catch (err) {
      toast.error("Failed to load revenue report");
    } finally {
      setLoading(false);
    }
  }, [locationId, dateFrom, dateTo, includeVoided]);

  useEffect(() => { load(); }, [load]);

  const totalRevenue = rows.reduce((s, r) => s + r.total_revenue, 0);
  const totalTransactions = rows.reduce((s, r) => s + r.transaction_count, 0);

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold">Daily Revenue</h2>

      <div className="flex flex-wrap gap-2 items-end">
        <Input label="From" type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} />
        <Input label="To" type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)} />
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={includeVoided} onChange={(e) => setIncludeVoided(e.target.checked)} />
          Show voided
        </label>
        <Button variant="secondary" size="sm" onClick={load}>Refresh</Button>
        <Button variant="secondary" size="sm" onClick={() => window.open(getDailyRevenueCSVUrl({ location_id: locationId, date_from: dateFrom, date_to: dateTo }), "_blank")}>
          Export CSV
        </Button>
        <Button variant="secondary" size="sm" onClick={() => window.print()}>
          Print PDF
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card><CardTitle>Total Revenue</CardTitle><p className="text-2xl font-bold">Rp {totalRevenue.toLocaleString("id-ID")}</p></Card>
        <Card><CardTitle>Transactions</CardTitle><p className="text-2xl font-bold">{totalTransactions.toLocaleString("id-ID")}</p></Card>
        <Card><CardTitle>Avg per Transaction</CardTitle><p className="text-2xl font-bold">Rp {(totalTransactions > 0 ? totalRevenue / totalTransactions : 0).toLocaleString("id-ID", { maximumFractionDigits: 0 })}</p></Card>
      </div>

      <div className="bg-white rounded-lg border p-4">
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={rows.map(r => ({ ...r, date: r.date.slice(5) }))}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="date" tick={{ fontSize: 12 }} />
            <YAxis tick={{ fontSize: 12 }} />
            <Tooltip />
            <Bar dataKey="total_revenue" fill="#3b82f6" name="Revenue" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      {loading ? <p className="text-gray-500">Loading...</p> : (
        <Table>
          <Thead><tr><Th>Date</Th><Th>Revenue</Th><Th>Transactions</Th><Th>Avg Fee</Th>{includeVoided && <><Th>Voided Count</Th><Th>Voided Amount</Th></>}</tr></Thead>
          <Tbody>
            {rows.map((r) => (
              <tr key={r.date}>
                <Td>{r.date}</Td>
                <Td>Rp {r.total_revenue.toLocaleString("id-ID")}</Td>
                <Td>{r.transaction_count}</Td>
                <Td>Rp {r.average_fee.toLocaleString("id-ID", { maximumFractionDigits: 0 })}</Td>
                {includeVoided && <><Td>{r.voided_count}</Td><Td>Rp {r.voided_amount.toLocaleString("id-ID")}</Td></>}
              </tr>
            ))}
          </Tbody>
        </Table>
      )}
    </div>
  );
}
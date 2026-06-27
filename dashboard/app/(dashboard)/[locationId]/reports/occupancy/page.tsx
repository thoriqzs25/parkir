"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { getOccupancy, getOccupancyCSVUrl } from "@/lib/api";
import { OccupancyRow } from "@/types/report";
import { Card, CardTitle } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
} from "recharts";
import { formatWIBDate } from "@/lib/time";

export default function OccupancyPage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const [rows, setRows] = useState<OccupancyRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [dateFrom, setDateFrom] = useState(() => {
    const d = new Date(); d.setDate(d.getDate() - 7);
    return d.toISOString().split("T")[0];
  });
  const [dateTo, setDateTo] = useState(() => new Date().toISOString().split("T")[0]);
  const [granularity, setGranularity] = useState<"day" | "hour">("day");

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await getOccupancy({
        location_id: locationId,
        date_from: dateFrom,
        date_to: dateTo,
        granularity,
      });
      setRows(data || []);
    } catch (err) {
      toast.error("Failed to load occupancy report");
    } finally {
      setLoading(false);
    }
  }, [locationId, dateFrom, dateTo, granularity]);

  useEffect(() => { load(); }, [load]);

  const totalCheckIns = rows.reduce((s, r) => s + r.count, 0);

  const chartData = rows.map((r) => ({
    label: granularity === "hour"
      ? new Date(r.bucket).toLocaleString("id-ID", { hour: "2-digit", day: "numeric", month: "short" })
      : formatWIBDate(r.bucket),
    count: r.count,
  }));

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold">Occupancy</h2>

      <div className="flex flex-wrap gap-2 items-end">
        <Input label="From" type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} />
        <Input label="To" type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)} />
        <div className="flex gap-1">
          <Button variant={granularity === "day" ? "primary" : "secondary"} size="sm" onClick={() => setGranularity("day")}>Daily</Button>
          <Button variant={granularity === "hour" ? "primary" : "secondary"} size="sm" onClick={() => setGranularity("hour")}>Hourly</Button>
        </div>
        <Button variant="secondary" size="sm" onClick={load}>Refresh</Button>
        <Button variant="secondary" size="sm" onClick={() => window.open(getOccupancyCSVUrl({ location_id: locationId, date_from: dateFrom, date_to: dateTo, granularity }), "_blank")}>
          Export CSV
        </Button>
        <Button variant="secondary" size="sm" onClick={() => window.print()}>Print PDF</Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card><CardTitle>Total Check-Ins</CardTitle><p className="text-2xl font-bold">{totalCheckIns.toLocaleString("id-ID")}</p></Card>
        <Card><CardTitle>Period</CardTitle><p className="text-lg font-medium">{dateFrom} — {dateTo}</p></Card>
        <Card><CardTitle>Granularity</CardTitle><p className="text-lg font-medium capitalize">{granularity}</p></Card>
      </div>

      <div className="bg-white rounded-lg border p-4">
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="label" tick={{ fontSize: 11 }} interval="preserveStartEnd" />
            <YAxis tick={{ fontSize: 12 }} />
            <Tooltip />
            <Bar dataKey="count" fill="#10b981" name="Check-ins" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      {loading ? <p className="text-gray-500">Loading...</p> : (
        <div className="grid grid-cols-7 gap-1">
          {rows.map((r) => {
            const d = new Date(r.bucket);
            const max = Math.max(...rows.map(x => x.count), 1);
            const intensity = Math.round((r.count / max) * 100);
            return (
              <div key={r.bucket} className="text-center p-1 rounded text-xs"
                style={{ backgroundColor: `rgba(16, 185, 129, ${0.1 + (intensity / 100) * 0.7})` }}
                title={`${r.count} check-ins`}
              >
                <div>{granularity === "hour" ? d.getHours() + ":00" : d.getDate()}</div>
                <div className="font-bold">{r.count}</div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
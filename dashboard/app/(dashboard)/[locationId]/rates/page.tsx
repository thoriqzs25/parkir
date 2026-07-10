"use client";

import { useEffect, useState, useMemo, useCallback } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { listRates, createRate, updateRate, listVehicleTypes } from "@/lib/api";
import { Rate } from "@/types/rate";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Select } from "@/components/ui/Select";
import { Dialog } from "@/components/ui/Dialog";
import { Badge } from "@/components/ui/Badge";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { hasPermission } from "@/lib/permissions";
import { useAuth } from "@/hooks/useAuth";
import { formatWIBDate } from "@/lib/time";

type FormData = {
  vehicle_type: string;
  first_hour_rate: string;
  subsequent_hourly_rate: string;
  daily_flat_rate: string;
  effective_from: string;
  effective_until?: string;
};
type RatePayload = {
  vehicle_type: string;
  first_hour_rate: number;
  subsequent_hourly_rate: number;
  daily_flat_rate: number;
  effective_from: string;
  effective_until?: string;
};

export default function RatesPage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const { permissions } = useAuth();
  const [rates, setRates] = useState<Rate[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<Rate | null>(null);
  const [vehicleTypes, setVehicleTypes] = useState<string[]>([]);

  const schema = useMemo(() => z.object({
    vehicle_type: z.string().refine((v) => vehicleTypes.includes(v), "Invalid vehicle type"),
    first_hour_rate: z.string().min(1, "Required"),
    subsequent_hourly_rate: z.string().min(1, "Required"),
    daily_flat_rate: z.string().min(1, "Required"),
    effective_from: z.string().min(1, "Required"),
    effective_until: z.string().optional(),
  }), [vehicleTypes]);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    mode: "onBlur",
  });

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [ratesRes, vts] = await Promise.all([
        listRates(locationId),
        listVehicleTypes(),
      ]);
      setRates(ratesRes || []);
      setVehicleTypes(vts.map((vt) => vt.name));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load data");
    } finally {
      setLoading(false);
    }
  }, [locationId]);

  useEffect(() => {
    load();
  }, [load]);

  const openCreate = () => {
    setEditing(null);
    reset({
      vehicle_type: "CAR",
      first_hour_rate: "0",
      subsequent_hourly_rate: "0",
      daily_flat_rate: "0",
      effective_from: new Date().toISOString().split("T")[0],
    });
    setDialogOpen(true);
  };

  const openEdit = (rate: Rate) => {
    setEditing(rate);
    reset({
      vehicle_type: rate.vehicle_type,
      first_hour_rate: String(rate.first_hour_rate),
      subsequent_hourly_rate: String(rate.subsequent_hourly_rate),
      daily_flat_rate: String(rate.daily_flat_rate),
      effective_from: rate.effective_from.split("T")[0],
      effective_until: rate.effective_until?.split("T")[0] || "",
    });
    setDialogOpen(true);
  };

  const onSubmit = async (data: FormData) => {
    try {
      const payload: RatePayload = {
        vehicle_type: data.vehicle_type,
        first_hour_rate: Number(data.first_hour_rate),
        subsequent_hourly_rate: Number(data.subsequent_hourly_rate),
        daily_flat_rate: Number(data.daily_flat_rate),
        effective_from: data.effective_from,
        effective_until: data.effective_until || undefined,
      };
      if (editing) {
        await updateRate(editing.id, payload);
        toast.success("Rate updated");
      } else {
        await createRate(locationId, payload);
        toast.success("Rate created");
      }
      setDialogOpen(false);
      load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to save rate");
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Rates</h2>
        {hasPermission(permissions, "rates:create") && (
          <Button size="sm" onClick={openCreate}>
            Add Rate
          </Button>
        )}
      </div>

      {loading ? (
        <p className="text-gray-500">Loading...</p>
      ) : (
        <Table>
          <Thead>
            <tr>
              <Th>Vehicle</Th>
              <Th>First Hour</Th>
              <Th>Subsequent</Th>
              <Th>Daily Flat</Th>
              <Th>Effective From</Th>
              <Th>Effective Until</Th>
              <Th>Actions</Th>
            </tr>
          </Thead>
          <Tbody>
            {rates.map((r) => (
              <tr key={r.id}>
                <Td>
                  <Badge variant="info">{r.vehicle_type}</Badge>
                </Td>
                <Td>Rp {r.first_hour_rate.toLocaleString("id-ID")}</Td>
                <Td>Rp {r.subsequent_hourly_rate.toLocaleString("id-ID")}</Td>
                <Td>Rp {r.daily_flat_rate.toLocaleString("id-ID")}</Td>
                <Td>{formatWIBDate(r.effective_from)}</Td>
                <Td>{r.effective_until ? formatWIBDate(r.effective_until) : "-"}</Td>
                <Td>
                  {hasPermission(permissions, "rates:edit") && (
                    <Button variant="ghost" size="sm" onClick={() => openEdit(r)}>
                      Edit
                    </Button>
                  )}
                </Td>
              </tr>
            ))}
          </Tbody>
        </Table>
      )}

      <Dialog
        open={dialogOpen}
        onClose={() => setDialogOpen(false)}
        title={editing ? "Edit Rate" : "Create Rate"}
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          {!editing && (
            <Select
              label="Vehicle Type"
              error={errors.vehicle_type?.message}
              {...register("vehicle_type")}
            >
              {vehicleTypes.map((t) => (
                <option key={t} value={t}>
                  {t}
                </option>
              ))}
            </Select>
          )}
          <Input
            label="First Hour Rate"
            type="number"
            error={errors.first_hour_rate?.message}
            {...register("first_hour_rate")}
          />
          <Input
            label="Subsequent Hourly Rate"
            type="number"
            error={errors.subsequent_hourly_rate?.message}
            {...register("subsequent_hourly_rate")}
          />
          <Input
            label="Daily Flat Rate"
            type="number"
            error={errors.daily_flat_rate?.message}
            {...register("daily_flat_rate")}
          />
          <Input
            label="Effective From"
            type="date"
            error={errors.effective_from?.message}
            {...register("effective_from")}
          />
          <Input label="Effective Until" type="date" {...register("effective_until")} />
          <div className="flex justify-end gap-2">
            <Button variant="secondary" type="button" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button type="submit">{editing ? "Update" : "Create"}</Button>
          </div>
        </form>
      </Dialog>
    </div>
  );
}

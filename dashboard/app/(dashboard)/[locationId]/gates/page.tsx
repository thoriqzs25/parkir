"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { listGates, registerGate, updateGate, deleteGate } from "@/lib/api";
import type { Gate } from "@/types/gate";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Dialog } from "@/components/ui/Dialog";
import { Badge } from "@/components/ui/Badge";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { hasPermission } from "@/lib/permissions";
import { useAuth } from "@/hooks/useAuth";
import { formatWIBDate } from "@/lib/time";

const schema = z.object({
  device_id: z.string().min(1, "Device ID is required"),
  name: z.string().max(100).optional(),
  ip_address: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

export default function GatesPage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const { permissions } = useAuth();
  const [gates, setGates] = useState<Gate[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<Gate | null>(null);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    mode: "onBlur",
    defaultValues: { device_id: "", name: "", ip_address: "" },
  });

  const load = async () => {
    setLoading(true);
    try {
      const res = await listGates(locationId);
      setGates(res || []);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load gates");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, [locationId]);

  const openCreate = () => {
    setEditing(null);
    reset({ device_id: "", name: "", ip_address: "" });
    setDialogOpen(true);
  };

  const openEdit = (gate: Gate) => {
    setEditing(gate);
    reset({
      device_id: gate.device_id,
      name: gate.name,
      ip_address: gate.ip_address,
    });
    setDialogOpen(true);
  };

  const onSubmit = async (data: FormData) => {
    try {
      if (editing) {
        await updateGate(editing.id, {
          name: data.name || undefined,
          ip_address: data.ip_address || undefined,
        });
        toast.success("Gate updated");
      } else {
        await registerGate({
          device_id: data.device_id,
          name: data.name || undefined,
          ip_address: data.ip_address || "",
          location_id: locationId,
        });
        toast.success("Gate registered");
      }
      setDialogOpen(false);
      load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to save gate");
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this gate?")) return;
    try {
      await deleteGate(id);
      toast.success("Gate deleted");
      load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to delete gate");
    }
  };

  const canManage = hasPermission(permissions, "gates:register");
  const canEdit = hasPermission(permissions, "gates:edit");
  const canDelete = hasPermission(permissions, "gates:delete");

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Gates</h2>
        {canManage && (
          <Button size="sm" onClick={openCreate}>Register Gate</Button>
        )}
      </div>

      {loading ? (
        <p className="text-gray-500">Loading...</p>
      ) : (
        <Table>
          <Thead>
            <tr>
              <Th>Name</Th>
              <Th>Device ID</Th>
              <Th>IP Address</Th>
              <Th>Last Seen</Th>
              <Th>Registered</Th>
              <Th>Actions</Th>
            </tr>
          </Thead>
          <Tbody>
            {gates.map((g) => (
              <tr key={g.id}>
                <Td>{g.name || "-"}</Td>
                <Td><Badge variant="info">{g.device_id.slice(0, 16)}&hellip;</Badge></Td>
                <Td className="font-mono">{g.ip_address || "-"}</Td>
                <Td>{g.last_seen_at ? formatWIBDate(g.last_seen_at) : "-"}</Td>
                <Td>{formatWIBDate(g.registered_at)}</Td>
                <Td>
                  <div className="flex gap-2">
                    {canEdit && (
                      <Button variant="ghost" size="sm" onClick={() => openEdit(g)}>Edit</Button>
                    )}
                    {canDelete && (
                      <Button variant="danger" size="sm" onClick={() => handleDelete(g.id)}>Delete</Button>
                    )}
                  </div>
                </Td>
              </tr>
            ))}
          </Tbody>
        </Table>
      )}

      <Dialog
        open={dialogOpen}
        onClose={() => setDialogOpen(false)}
        title={editing ? "Edit Gate" : "Register Gate"}
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input
            label="Device ID"
            error={errors.device_id?.message}
            disabled={!!editing}
            {...register("device_id")}
          />
          <Input
            label="Name"
            error={errors.name?.message}
            {...register("name")}
          />
          <Input
            label="IP Address"
            error={errors.ip_address?.message}
            {...register("ip_address")}
          />
          <div className="flex justify-end gap-2">
            <Button variant="secondary" type="button" onClick={() => setDialogOpen(false)}>Cancel</Button>
            <Button type="submit">{editing ? "Update" : "Register"}</Button>
          </div>
        </form>
      </Dialog>
    </div>
  );
}

"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { listVehicleTypes, createVehicleType, updateVehicleType, deleteVehicleType } from "@/lib/api";
import type { VehicleType } from "@/types/vehicle_type";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Dialog } from "@/components/ui/Dialog";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { hasPermission } from "@/lib/permissions";
import { useAuth } from "@/hooks/useAuth";
import { formatWIBDate } from "@/lib/time";

const schema = z.object({
  name: z.string().min(1, "Name is required").max(20, "Max 20 characters"),
  display_name: z.string().min(1, "Display name is required").max(100),
  description: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

export default function VehicleTypesPage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const { permissions } = useAuth();
  const [types, setTypes] = useState<VehicleType[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<VehicleType | null>(null);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    mode: "onBlur",
    defaultValues: { name: "", display_name: "", description: "" },
  });

  const load = async () => {
    setLoading(true);
    try {
      const res = await listVehicleTypes();
      setTypes(res || []);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load vehicle types");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, [locationId]);

  const openCreate = () => {
    setEditing(null);
    reset({ name: "", display_name: "", description: "" });
    setDialogOpen(true);
  };

  const openEdit = (vt: VehicleType) => {
    setEditing(vt);
    reset({ name: vt.name, display_name: vt.display_name, description: vt.description });
    setDialogOpen(true);
  };

  const onSubmit = async (data: FormData) => {
    try {
      if (editing) {
        await updateVehicleType(editing.name, {
          display_name: data.display_name,
          description: data.description || undefined,
        });
        toast.success("Vehicle type updated");
      } else {
        await createVehicleType({
          name: data.name,
          display_name: data.display_name,
          description: data.description || undefined,
        });
        toast.success("Vehicle type created");
      }
      setDialogOpen(false);
      load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to save vehicle type");
    }
  };

  const handleDelete = async (name: string) => {
    if (!confirm(`Delete vehicle type "${name}"? This cannot be undone if it is in use.`)) return;
    try {
      await deleteVehicleType(name);
      toast.success("Vehicle type deleted");
      load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to delete vehicle type");
    }
  };

  const canManage = hasPermission(permissions, "vehicle-types:create");
  const canEdit = hasPermission(permissions, "vehicle-types:edit");
  const canDelete = hasPermission(permissions, "vehicle-types:delete");

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Vehicle Types</h2>
        {canManage && (
          <Button size="sm" onClick={openCreate}>
            Add Vehicle Type
          </Button>
        )}
      </div>

      {loading ? (
        <p className="text-gray-500">Loading...</p>
      ) : (
        <Table>
          <Thead>
            <tr>
              <Th>Name</Th>
              <Th>Display Name</Th>
              <Th>Description</Th>
              <Th>Created At</Th>
              <Th>Actions</Th>
            </tr>
          </Thead>
          <Tbody>
            {types.map((vt) => (
              <tr key={vt.name}>
                <Td className="font-mono">{vt.name}</Td>
                <Td>{vt.display_name}</Td>
                <Td>{vt.description || "-"}</Td>
                <Td>{formatWIBDate(vt.created_at)}</Td>
                <Td>
                  <div className="flex gap-2">
                    {canEdit && (
                      <Button variant="ghost" size="sm" onClick={() => openEdit(vt)}>
                        Edit
                      </Button>
                    )}
                    {canDelete && (
                      <Button variant="danger" size="sm" onClick={() => handleDelete(vt.name)}>
                        Delete
                      </Button>
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
        title={editing ? "Edit Vehicle Type" : "Create Vehicle Type"}
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input
            label="Name"
            error={errors.name?.message}
            disabled={!!editing}
            {...register("name")}
          />
          <Input
            label="Display Name"
            error={errors.display_name?.message}
            {...register("display_name")}
          />
          <Input
            label="Description"
            error={errors.description?.message}
            {...register("description")}
          />
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

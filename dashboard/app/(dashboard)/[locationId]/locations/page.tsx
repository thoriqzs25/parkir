"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  listLocations,
  createLocation,
  updateLocation,
  deactivateLocation,
  listUsers,
  assignOperator,
  removeOperator,
} from "@/lib/api";
import { Location } from "@/types/location";
import { User } from "@/types/user";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Dialog } from "@/components/ui/Dialog";
import { Badge } from "@/components/ui/Badge";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { hasPermission } from "@/lib/permissions";
import { useAuth } from "@/hooks/useAuth";

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  code: z.string().min(1, "Code is required"),
  address: z.string().optional(),
  city: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

export default function LocationsPage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const { permissions } = useAuth();
  const [locations, setLocations] = useState<Location[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<Location | null>(null);
  const [managingLocation, setManagingLocation] = useState<Location | null>(null);
  const [users, setUsers] = useState<User[]>([]);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    mode: "onBlur",
  });

  const load = async () => {
    setLoading(true);
    try {
      const res = await listLocations();
      setLocations(res.items || []);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load locations");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, [locationId]);

  const openCreate = () => {
    setEditing(null);
    reset({ name: "", code: "", address: "", city: "" });
    setDialogOpen(true);
  };

  const openEdit = (loc: Location) => {
    setEditing(loc);
    reset({
      name: loc.name,
      code: loc.code,
      address: loc.address || "",
      city: loc.city || "",
    });
    setDialogOpen(true);
  };

  const onSubmit = async (data: FormData) => {
    try {
      if (editing) {
        await updateLocation(editing.id, data);
        toast.success("Location updated");
      } else {
        await createLocation(data);
        toast.success("Location created");
      }
      setDialogOpen(false);
      load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to save location");
    }
  };

  const handleDeactivate = async (id: string) => {
    if (!confirm("Deactivate this location?")) return;
    try {
      await deactivateLocation(id);
      toast.success("Location deactivated");
      load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to deactivate");
    }
  };

  const openOperators = async (loc: Location) => {
    setManagingLocation(loc);
    try {
      const res = await listUsers();
      setUsers(res.items || []);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load users");
    }
  };

  const handleAssign = async (userId: string) => {
    if (!managingLocation) return;
    try {
      await assignOperator(managingLocation.id, userId);
      toast.success("Operator assigned");
      const res = await listUsers();
      setUsers(res.items || []);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to assign operator");
    }
  };

  const handleRemove = async (userId: string) => {
    if (!managingLocation) return;
    try {
      await removeOperator(managingLocation.id, userId);
      toast.success("Operator removed");
      const res = await listUsers();
      setUsers(res.items || []);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to remove operator");
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Locations</h2>
        {hasPermission(permissions, "locations:create") && (
          <Button size="sm" onClick={openCreate}>
            Add Location
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
              <Th>Code</Th>
              <Th>City</Th>
              <Th>Status</Th>
              <Th>Actions</Th>
            </tr>
          </Thead>
          <Tbody>
            {locations.map((loc) => (
              <tr key={loc.id}>
                <Td>{loc.name}</Td>
                <Td>{loc.code}</Td>
                <Td>{loc.city || "-"}</Td>
                <Td>
                  <Badge variant={loc.status === "ACTIVE" ? "success" : "danger"}>
                    {loc.status}
                  </Badge>
                </Td>
                <Td>
                  <div className="flex gap-2">
                    {hasPermission(permissions, "locations:edit") && (
                      <Button variant="ghost" size="sm" onClick={() => openEdit(loc)}>
                        Edit
                      </Button>
                    )}
                    {hasPermission(permissions, "locations:assign_operators") && (
                      <Button variant="secondary" size="sm" onClick={() => openOperators(loc)}>
                        Operators
                      </Button>
                    )}
                    {hasPermission(permissions, "locations:deactivate") &&
                      loc.status === "ACTIVE" && (
                        <Button
                          variant="danger"
                          size="sm"
                          onClick={() => handleDeactivate(loc.id)}
                        >
                          Deactivate
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
        title={editing ? "Edit Location" : "Create Location"}
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input label="Name" error={errors.name?.message} {...register("name")} />
          <Input label="Code" error={errors.code?.message} {...register("code")} />
          <Input label="Address" {...register("address")} />
          <Input label="City" {...register("city")} />
          <div className="flex justify-end gap-2">
            <Button variant="secondary" type="button" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button type="submit">{editing ? "Update" : "Create"}</Button>
          </div>
        </form>
      </Dialog>

      <Dialog
        open={!!managingLocation}
        onClose={() => setManagingLocation(null)}
        title={managingLocation ? `Operators for ${managingLocation.name}` : "Operators"}
      >
        <div className="max-h-96 overflow-auto">
          {users.length === 0 ? (
            <p className="text-gray-500">No users found.</p>
          ) : (
            <div className="space-y-2">
              {users.map((u) => {
                const assigned = managingLocation
                  ? (u.location_ids || []).includes(managingLocation.id)
                  : false;
                return (
                  <div
                    key={u.id}
                    className="flex items-center justify-between rounded border border-gray-200 p-3"
                  >
                    <div>
                      <p className="font-medium">{u.name}</p>
                      <p className="text-sm text-gray-500">{u.email}</p>
                    </div>
                    {assigned ? (
                      <Button variant="danger" size="sm" onClick={() => handleRemove(u.id)}>
                        Remove
                      </Button>
                    ) : (
                      <Button variant="secondary" size="sm" onClick={() => handleAssign(u.id)}>
                        Assign
                      </Button>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </div>
        <div className="mt-4 flex justify-end">
          <Button variant="secondary" onClick={() => setManagingLocation(null)}>
            Close
          </Button>
        </div>
      </Dialog>
    </div>
  );
}

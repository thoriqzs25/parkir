"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { listRoles, createRole, updateRole, deleteRole } from "@/lib/api";
import { Role } from "@/types/role";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Dialog } from "@/components/ui/Dialog";
import { Badge } from "@/components/ui/Badge";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { hasPermission } from "@/lib/permissions";
import { useAuth } from "@/hooks/useAuth";

const ALL_PERMISSIONS = [
  "sessions:view",
  "sessions:create",
  "sessions:close",
  "sessions:void",
  "payments:collect_cash",
  "payments:collect_digital",
  "payments:void",
  "shifts:start",
  "shifts:end",
  "shifts:view",
  "shifts:force_close",
  "shifts:resolve_discrepancy",
  "locations:view",
  "locations:create",
  "locations:edit",
  "locations:deactivate",
  "locations:assign_operators",
  "rates:view",
  "rates:create",
  "rates:edit",
  "users:view",
  "users:create",
  "users:edit",
  "users:deactivate",
  "reports:view_revenue",
  "reports:view_occupancy",
  "reports:view_operators",
  "incidents:view",
  "incidents:create",
  "incidents:resolve",
  "adjustments:void_transaction",
  "adjustments:reassign_session",
  "finance:view_transactions",
  "finance:view_revenue_summary",
  "finance:view_cash_handover",
  "finance:view_all_locations",
  "finance:export",
  "observability:view_health",
  "observability:view_audit",
  "observability:view_alerts",
  "observability:manage_alerts",
];

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  permissions: z.array(z.string()).min(1, "At least one permission is required"),
});

type FormData = z.infer<typeof schema>;

export default function RolesPage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const { permissions: userPermissions } = useAuth();
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<Role | null>(null);

  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    mode: "onBlur",
    defaultValues: { name: "", permissions: [] },
  });

  const selectedPermissions = watch("permissions") || [];

  const load = async () => {
    setLoading(true);
    try {
      const res = await listRoles();
      setRoles(res.items || []);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load roles");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, [locationId]);

  const openCreate = () => {
    setEditing(null);
    reset({ name: "", permissions: [] });
    setDialogOpen(true);
  };

  const openEdit = (role: Role) => {
    setEditing(role);
    reset({ name: role.name, permissions: role.permissions });
    setDialogOpen(true);
  };

  const togglePermission = (perm: string) => {
    const next = selectedPermissions.includes(perm)
      ? selectedPermissions.filter((p) => p !== perm)
      : [...selectedPermissions, perm];
    setValue("permissions", next, { shouldValidate: true });
  };

  const onSubmit = async (data: FormData) => {
    try {
      if (editing) {
        await updateRole(editing.id, data);
        toast.success("Role updated");
      } else {
        await createRole(data);
        toast.success("Role created");
      }
      setDialogOpen(false);
      load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to save role");
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this role?")) return;
    try {
      await deleteRole(id);
      toast.success("Role deleted");
      load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to delete role");
    }
  };

  const canManage = hasPermission(userPermissions, "users:create");

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Roles</h2>
        {canManage && (
          <Button size="sm" onClick={openCreate}>
            Add Role
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
              <Th>Permissions</Th>
              <Th>Actions</Th>
            </tr>
          </Thead>
          <Tbody>
            {roles.map((r) => (
              <tr key={r.id}>
                <Td>{r.name}</Td>
                <Td>
                  <div className="flex flex-wrap gap-1">
                    {r.permissions.slice(0, 5).map((p) => (
                      <Badge key={p} variant="info">
                        {p}
                      </Badge>
                    ))}
                    {r.permissions.length > 5 && (
                      <Badge variant="default">+{r.permissions.length - 5}</Badge>
                    )}
                  </div>
                </Td>
                <Td>
                  <div className="flex gap-2">
                    {canManage && (
                      <>
                        <Button variant="ghost" size="sm" onClick={() => openEdit(r)}>
                          Edit
                        </Button>
                        <Button variant="danger" size="sm" onClick={() => handleDelete(r.id)}>
                          Delete
                        </Button>
                      </>
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
        title={editing ? "Edit Role" : "Create Role"}
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input label="Name" error={errors.name?.message} {...register("name")} />
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Permissions
            </label>
            <div className="max-h-60 overflow-y-auto rounded-md border border-gray-300 p-2 space-y-1">
              {ALL_PERMISSIONS.map((perm) => (
                <label key={perm} className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={selectedPermissions.includes(perm)}
                    onChange={() => togglePermission(perm)}
                  />
                  {perm}
                </label>
              ))}
            </div>
            {errors.permissions && (
              <p className="mt-1 text-xs text-red-600">{errors.permissions.message}</p>
            )}
          </div>
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

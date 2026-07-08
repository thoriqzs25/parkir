"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  listUsers,
  createUser,
  updateUser,
  deactivateUser,
  resetPassword,
  resetPIN,
  listRoles,
  listLocations,
} from "@/lib/api";
import { CreateUserInput, UpdateUserInput, User } from "@/types/user";
import { Role } from "@/types/role";
import { Location } from "@/types/location";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Select } from "@/components/ui/Select";
import { Dialog } from "@/components/ui/Dialog";
import { Badge } from "@/components/ui/Badge";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { hasPermission } from "@/lib/permissions";
import { useAuth } from "@/hooks/useAuth";

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  email: z.string().email("Invalid email"),
  password: z.string().optional(),
  role_id: z.string().min(1, "Role is required"),
  location_ids: z.array(z.string()),
});

type FormData = z.infer<typeof schema>;

export default function UsersPage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const { permissions } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [locations, setLocations] = useState<Location[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<User | null>(null);
  const [resetUser, setResetUser] = useState<User | null>(null);
  const [newPassword, setNewPassword] = useState("");
  const [newPIN, setNewPIN] = useState("");
  const [pinUser, setPinUser] = useState<User | null>(null);

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
      const [u, r, l] = await Promise.all([
        listUsers(),
        listRoles(),
        listLocations(),
      ]);
      setUsers(u.items || []);
      setRoles(r.items || []);
      setLocations(l.items || []);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load data");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, [locationId]);

  const openCreate = () => {
    setEditing(null);
    reset({ name: "", email: "", password: "", role_id: "", location_ids: [] });
    setDialogOpen(true);
  };

  const openEdit = (user: User) => {
    setEditing(user);
    reset({
      name: user.name,
      email: user.email,
      role_id: user.role_id,
      location_ids: user.location_ids || [],
    });
    setDialogOpen(true);
  };

  const onSubmit = async (data: FormData) => {
    try {
      if (editing) {
        const payload: UpdateUserInput = {
          name: data.name,
          email: data.email,
          role_id: data.role_id,
          location_ids: data.location_ids,
        };
        await updateUser(editing.id, payload);
        toast.success("User updated");
      } else {
        if (!data.password || data.password.length < 8) {
          toast.error("Password must be at least 8 characters");
          return;
        }
        const payload: CreateUserInput = {
          name: data.name,
          email: data.email,
          password: data.password,
          role_id: data.role_id,
          location_ids: data.location_ids,
        };
        await createUser(payload);
        toast.success("User created");
      }
      setDialogOpen(false);
      load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to save user");
    }
  };

  const handleDeactivate = async (id: string) => {
    if (!confirm("Deactivate this user?")) return;
    try {
      await deactivateUser(id);
      toast.success("User deactivated");
      load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to deactivate");
    }
  };

  const handleResetPassword = async () => {
    if (!resetUser) return;
    if (newPassword.length < 8) {
      toast.error("Password must be at least 8 characters");
      return;
    }
    try {
      await resetPassword(resetUser.id, newPassword);
      toast.success("Password reset successfully");
      setResetUser(null);
      setNewPassword("");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to reset password");
    }
  };

  const handleResetPIN = async () => {
    if (!pinUser) return;
    if (newPIN.length !== 6) {
      toast.error("PIN must be 6 digits");
      return;
    }
    try {
      await resetPIN(pinUser.id, newPIN);
      toast.success("PIN reset successfully");
      setPinUser(null);
      setNewPIN("");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to reset PIN");
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Users</h2>
        {hasPermission(permissions, "users:create") && (
          <Button size="sm" onClick={openCreate}>
            Add User
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
              <Th>Email</Th>
              <Th>Role</Th>
              <Th>Status</Th>
              <Th>Actions</Th>
            </tr>
          </Thead>
          <Tbody>
            {users.map((u) => (
              <tr key={u.id}>
                <Td>{u.name}</Td>
                <Td>{u.email}</Td>
                <Td>{u.role_name || "-"}</Td>
                <Td>
                  <Badge variant={u.status === "ACTIVE" ? "success" : "danger"}>
                    {u.status}
                  </Badge>
                </Td>
                <Td>
                  <div className="flex gap-2">
                    {hasPermission(permissions, "users:edit") && (
                      <Button variant="ghost" size="sm" onClick={() => openEdit(u)}>
                        Edit
                      </Button>
                    )}
                    {hasPermission(permissions, "users:edit") && (
                      <Button
                        variant="secondary"
                        size="sm"
                        onClick={() => {
                          setResetUser(u);
                          setNewPassword("");
                        }}
                      >
                        Reset Password
                      </Button>
                    )}
                    {hasPermission(permissions, "users:edit") && (
                      <Button
                        variant="secondary"
                        size="sm"
                        onClick={() => {
                          setPinUser(u);
                          setNewPIN("");
                        }}
                      >
                        Reset PIN
                      </Button>
                    )}
                    {hasPermission(permissions, "users:deactivate") &&
                      u.status === "ACTIVE" && (
                        <Button
                          variant="danger"
                          size="sm"
                          onClick={() => handleDeactivate(u.id)}
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
        title={editing ? "Edit User" : "Create User"}
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input label="Name" error={errors.name?.message} {...register("name")} />
          <Input label="Email" error={errors.email?.message} {...register("email")} />
          {!editing && (
            <Input
              label="Password"
              type="password"
              error={errors.password?.message}
              {...register("password")}
            />
          )}
          <Select label="Role" error={errors.role_id?.message} {...register("role_id")}>
            <option value="">Select role</option>
            {roles.map((r) => (
              <option key={r.id} value={r.id}>
                {r.name}
              </option>
            ))}
          </Select>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Locations
            </label>
            <select
              multiple
              {...register("location_ids")}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm h-32"
            >
              {locations.map((l) => (
                <option key={l.id} value={l.id}>
                  {l.name}
                </option>
              ))}
            </select>
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="secondary" type="button" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button type="submit">{editing ? "Update" : "Create"}</Button>
          </div>
        </form>
      </Dialog>

      <Dialog
        open={!!resetUser}
        onClose={() => setResetUser(null)}
        title={resetUser ? `Reset password for ${resetUser.name}` : "Reset Password"}
      >
        <div className="space-y-4">
          <Input
            label="New Password"
            type="password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
          />
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => setResetUser(null)}>
              Cancel
            </Button>
            <Button onClick={handleResetPassword}>Reset</Button>
          </div>
        </div>
      </Dialog>

      <Dialog
        open={!!pinUser}
        onClose={() => setPinUser(null)}
        title={pinUser ? `Reset PIN for ${pinUser.name}` : "Reset PIN"}
      >
        <div className="space-y-4">
          <Input
            label="New PIN (6 digits)"
            type="text"
            inputMode="numeric"
            maxLength={6}
            value={newPIN}
            onChange={(e) => setNewPIN(e.target.value.replace(/\D/g, ""))}
          />
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => setPinUser(null)}>
              Cancel
            </Button>
            <Button onClick={handleResetPIN}>Reset</Button>
          </div>
        </div>
      </Dialog>
    </div>
  );
}

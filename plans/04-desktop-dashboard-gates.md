# Milestone 4: Desktop & Dashboard — Gate Management UI

## Objective

Add gate registration to the desktop app (LAN discovery) and gate management to the dashboard.

## Desktop files

| # | File | Action |
|---|------|--------|
| 1 | `desktop/src/renderer/screens/GateSetup.tsx` | **New** — LAN discovery + registration |
| 2 | `desktop/src/renderer/App.tsx` | **Edit** — add `/gate-setup` route |
| 3 | `desktop/src/renderer/components/Layout.tsx` | **Edit** — add nav link |

## Dashboard files

| # | File | Action |
|---|------|--------|
| 4 | `dashboard/types/gate.ts` | **New** |
| 5 | `dashboard/lib/api.ts` | **Edit** — add gate API functions |
| 6 | `dashboard/app/(dashboard)/[locationId]/gates/page.tsx` | **New** |
| 7 | `dashboard/components/layout/DashboardLayout.tsx` | **Edit** — add nav link |

## 4.1 Desktop — GateSetup screen

New file at `desktop/src/renderer/screens/GateSetup.tsx`.

### Discovery mechanism

The desktop app queries mDNS for `_parkir-gate._tcp` services on the LAN. Uses the `multicast-dns` npm package.

Requires:
```bash
cd desktop && npm install multicast-dns
cd desktop && npm install --save-dev @types/multicast-dns
```

### Component

```tsx
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { listLocations } from '../lib/api'
import type { Location } from '../types'
import { toast } from 'sonner' // if available, else inline

interface DiscoveredGate {
  deviceId: string
  ip: string
  hostname: string
}

export default function GateSetup() {
  const navigate = useNavigate()
  const [gates, setGates] = useState<DiscoveredGate[]>([])
  const [locations, setLocations] = useState<Location[]>([])
  const [selectedDeviceId, setSelectedDeviceId] = useState<string | null>(null)
  const [selectedLocationId, setSelectedLocationId] = useState('')
  const [gateName, setGateName] = useState('')
  const [scanning, setScanning] = useState(true)
  const [registering, setRegistering] = useState(false)

  useEffect(() => {
    // Fetch locations
    listLocations().then((res) => {
      setLocations(res.items || [])
    }).catch(() => {})

    // mDNS discovery
    let mdns: any
    try {
      const multicastdns = require('multicast-dns')
      mdns = multicastdns()

      mdns.on('response', (response: any) => {
        const srvRecords = response.answers?.filter((a: any) =>
          a.type === 'SRV' && a.name === '_parkir-gate._tcp.local'
        ) || []
        const txtRecords = response.answers?.filter((a: any) =>
          a.type === 'TXT' && a.name === '_parkir-gate._tcp.local'
        ) || []
        const aRecords = response.answers?.filter((a: any) =>
          a.type === 'A'
        ) || []

        srvRecords.forEach((srv: any) => {
          const deviceId = txtRecords[0]?.data?.toString() || 'unknown'
          const ip = aRecords[0]?.data || 'unknown'
          setGates((prev) => {
            if (prev.some((g) => g.deviceId === deviceId)) return prev
            return [...prev, { deviceId, ip, hostname: srv.data?.target || '' }]
          })
        })
      })

      // Send query every 5s
      const query = () => {
        mdns.query({ questions: [{ name: '_parkir-gate._tcp.local', type: 'SRV' }] })
      }
      query()
      const interval = setInterval(query, 5000)
      setScanning(false)

      return () => {
        clearInterval(interval)
        mdns.destroy()
      }
    } catch {
      setScanning(false)
    }
  }, [])

  const handleRegister = async () => {
    if (!selectedDeviceId || !selectedLocationId) return
    setRegistering(true)

    try {
      const gate = gates.find((g) => g.deviceId === selectedDeviceId)
      if (!gate) throw new Error('Gate not found')

      // Register in backend
      const res = await fetch('/api/v1/gates', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          device_id: gate.deviceId,
          name: gateName || `Gate ${gate.deviceId.slice(0, 8)}`,
          location_id: selectedLocationId,
          ip_address: gate.ip,
        }),
      })
      if (!res.ok) throw new Error('Backend registration failed')
      const gateData = await res.json()

      // Tell gate display to configure itself
      const gateRes = await fetch(`http://${gate.ip}:9800/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          location_id: selectedLocationId,
          api_url: window.location.origin || 'http://localhost:8080',
        }),
      })
      if (!gateRes.ok) throw new Error('Gate registration failed')

      toast.success('Gate registered successfully')
      navigate('/dashboard')
    } catch (err: any) {
      toast.error(err.message || 'Registration failed')
    } finally {
      setRegistering(false)
    }
  }

  return (
    <div className="screen">
      <button className="button secondary back" onClick={() => navigate('/dashboard')}>
        &larr; Back
      </button>
      <h2>Gate Setup</h2>

      {scanning && <p>Scanning network for gates...</p>}

      {gates.length === 0 && !scanning && (
        <div className="card" style={{ padding: 24, textAlign: 'center' }}>
          <p>No gates found on the network.</p>
          <p style={{ color: '#888', fontSize: 13, marginTop: 8 }}>
            Make sure the gate display app is running and on the same LAN.
          </p>
        </div>
      )}

      {gates.length > 0 && (
        <div className="card" style={{ padding: 16 }}>
          <table className="data-table">
            <thead>
              <tr>
                <th style={{ width: 40 }}></th>
                <th>Device ID</th>
                <th>IP Address</th>
                <th>Hostname</th>
              </tr>
            </thead>
            <tbody>
              {gates.map((g) => (
                <tr key={g.deviceId} className={selectedDeviceId === g.deviceId ? 'selected' : ''}>
                  <td>
                    <input
                      type="radio"
                      name="gate"
                      checked={selectedDeviceId === g.deviceId}
                      onChange={() => setSelectedDeviceId(g.deviceId)}
                    />
                  </td>
                  <td className="mono">{g.deviceId.slice(0, 16)}…</td>
                  <td className="mono">{g.ip}</td>
                  <td>{g.hostname}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {selectedDeviceId && (
        <div className="card" style={{ padding: 16, marginTop: 16 }}>
          <h3 style={{ marginBottom: 12 }}>Register Gate</h3>
          <div className="form-group">
            <label>Gate Name</label>
            <input
              value={gateName}
              onChange={(e) => setGateName(e.target.value)}
              placeholder="e.g. Gate A - Entry 1"
            />
          </div>
          <div className="form-group">
            <label>Location</label>
            <select value={selectedLocationId} onChange={(e) => setSelectedLocationId(e.target.value)}>
              <option value="">— Select Location —</option>
              {locations.map((loc) => (
                <option key={loc.id} value={loc.id}>{loc.name}</option>
              ))}
            </select>
          </div>
          <button
            className="button primary full"
            onClick={handleRegister}
            disabled={!selectedLocationId || registering}
          >
            {registering ? 'Registering...' : 'Register Gate'}
          </button>
        </div>
      )}
    </div>
  )
}
```

### 4.1a Route — `desktop/src/renderer/App.tsx`

Add import:
```tsx
import GateSetup from './screens/GateSetup'
```

Add route inside the `AuthenticatedLayout` route group:
```tsx
<Route path="/gate-setup" element={<RequireUser><GateSetup /></RequireUser>} />
```

### 4.1b Nav link — `desktop/src/renderer/components/Layout.tsx`

Add `Wifi` to lucide-react imports (check if `Wifi` icon exists in lucide-react; if not, use `Radio` or `Server`).

```tsx
import { ..., Wifi } from 'lucide-react'
```

Add to nav items (after History, before Logout or in an admin section):
```tsx
{ href: '/gate-setup', label: 'Gate Setup', icon: Wifi, perm: 'gates:register' }
```

Note: The desktop app doesn't have a permission check per nav item currently (Layout.tsx doesn't filter by perm like the dashboard does). Either add permission filtering or just show the link to all users for now.

## 4.2 Dashboard — Gates page

### 4.2a Types — `dashboard/types/gate.ts`

```typescript
export interface Gate {
  id: string
  device_id: string
  name: string
  location_id?: string
  ip_address: string
  last_seen_at?: string
  registered_at: string
  created_at: string
  updated_at: string
}

export interface RegisterGateInput {
  device_id: string
  name?: string
  location_id?: string
  ip_address?: string
}

export interface UpdateGateInput {
  name?: string
  location_id?: string
  ip_address?: string
}
```

### 4.2b API functions — `dashboard/lib/api.ts`

Add import:
```typescript
import { Gate, RegisterGateInput, UpdateGateInput } from '@/types/gate'
```

Add functions:
```typescript
export function listGates(locationId?: string) {
  const params = locationId ? `?location_id=${encodeURIComponent(locationId)}` : ''
  return apiRequest<Gate[]>(`/api/v1/gates${params}`)
}

export function getGate(id: string) {
  return apiRequest<Gate>(`/api/v1/gates/${encodeURIComponent(id)}`)
}

export function registerGate(input: RegisterGateInput) {
  return apiRequest<Gate>('/api/v1/gates', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function updateGate(id: string, input: UpdateGateInput) {
  return apiRequest<Gate>(`/api/v1/gates/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    body: JSON.stringify(input),
  })
}

export function deleteGate(id: string) {
  return apiRequest<void>(`/api/v1/gates/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
}
```

### 4.2c Page — `dashboard/app/(dashboard)/[locationId]/gates/page.tsx`

Follows the pattern of `roles/page.tsx`:

```tsx
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
  name: z.string().max(100).optional(),
  device_id: z.string().min(1, "Device ID is required"),
  ip_address: z.string().optional(),
  location_id: z.string().optional(),
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
    defaultValues: { name: "", device_id: "", ip_address: "" },
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
    reset({ name: "", device_id: "", ip_address: "", location_id: locationId });
    setDialogOpen(true);
  };

  const openEdit = (gate: Gate) => {
    setEditing(gate);
    reset({
      name: gate.name,
      device_id: gate.device_id,
      ip_address: gate.ip_address,
      location_id: gate.location_id || "",
    });
    setDialogOpen(true);
  };

  const onSubmit = async (data: FormData) => {
    try {
      if (editing) {
        await updateGate(editing.id, {
          name: data.name || undefined,
          ip_address: data.ip_address || undefined,
          location_id: data.location_id || undefined,
        });
        toast.success("Gate updated");
      } else {
        await registerGate({
          device_id: data.device_id,
          name: data.name || undefined,
          ip_address: data.ip_address || "",
          location_id: data.location_id || locationId,
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
                <Td><Badge variant="info">{g.device_id.slice(0, 16)}…</Badge></Td>
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
          <Input label="Device ID" error={errors.device_id?.message} disabled={!!editing} {...register("device_id")} />
          <Input label="Name" error={errors.name?.message} {...register("name")} />
          <Input label="IP Address" error={errors.ip_address?.message} {...register("ip_address")} />
          <div className="flex justify-end gap-2">
            <Button variant="secondary" type="button" onClick={() => setDialogOpen(false)}>Cancel</Button>
            <Button type="submit">{editing ? "Update" : "Register"}</Button>
          </div>
        </form>
      </Dialog>
    </div>
  );
}
```

### 4.2d Nav link — `dashboard/components/layout/DashboardLayout.tsx`

Add `Wifi` to imports from `lucide-react`:
```tsx
import { ..., Wifi } from "lucide-react";
```

Add nav item (before Backups or after Health):
```tsx
{ href: `/${locationId}/gates`, label: "Gates", icon: Wifi, perm: "gates:view" },
```

## 4.3 Manual verification

### Desktop
```bash
cd desktop && npm run dev
```
1. Login → navigate to Gate Setup
2. If gate display is running on LAN, it appears in the list
3. Select gate, pick location, enter name, Register
4. Gate display restarts into display mode
5. Backend has gate record

### Dashboard
```bash
cd dashboard && npm run dev
```
1. Login → navigate to a location → Gates nav link visible
2. Table shows registered gates for this location
3. Register / Edit / Delete work
4. TypeScript compiles: `npx tsc --noEmit` → no errors

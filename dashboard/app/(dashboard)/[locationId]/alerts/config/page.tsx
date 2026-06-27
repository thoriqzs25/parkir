"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { listAlertConfigs, updateAlertConfig } from "@/lib/api";
import { AlertConfig } from "@/types/alert";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";

const codeLabels: Record<string, string> = {
  LONG_SESSION: "Long Session",
  UNPAID_EXIT: "Unpaid Exit",
  HIGH_VOID_RATE: "High Void Rate",
  SYNC_FAILURE: "Sync Failure",
  GATEWAY_FAILURE: "Gateway Failure",
};

export default function AlertConfigsPage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const [configs, setConfigs] = useState<AlertConfig[]>([]);
  const [loading, setLoading] = useState(true);

  const load = async () => {
    setLoading(true);
    try {
      const data = await listAlertConfigs(locationId);
      setConfigs(data || []);
    } catch (err) {
      toast.error("Failed to load alert configs");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, [locationId]);

  const handleToggle = async (config: AlertConfig) => {
    try {
      await updateAlertConfig(config.id, { enabled: !config.enabled });
      toast.success(`${codeLabels[config.code] || config.code} ${config.enabled ? "disabled" : "enabled"}`);
      load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to update config");
    }
  };

  if (loading) return <p className="text-gray-500">Loading...</p>;

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold">Alert Configurations</h2>
      <p className="text-sm text-gray-500">Manage alert thresholds and enable/disable alert rules for this location.</p>

      <Table>
        <Thead>
          <tr>
            <Th>Alert Rule</Th>
            <Th>Scope</Th>
            <Th>Status</Th>
            <Th>Updated</Th>
            <Th>Actions</Th>
          </tr>
        </Thead>
        <Tbody>
          {configs.map((c) => (
            <tr key={c.id}>
              <Td><code className="text-xs bg-gray-100 px-1 rounded">{codeLabels[c.code] || c.code}</code></Td>
              <Td>{c.location_id ? "Per-Location" : "Global"}</Td>
              <Td><Badge variant={c.enabled ? "success" : "danger"}>{c.enabled ? "Enabled" : "Disabled"}</Badge></Td>
              <Td className="text-sm text-gray-500">{c.updated_by || "system"}</Td>
              <Td>
                <Button variant="ghost" size="sm" onClick={() => handleToggle(c)}>
                  {c.enabled ? "Disable" : "Enable"}
                </Button>
              </Td>
            </tr>
          ))}
        </Tbody>
      </Table>
    </div>
  );
}
"use client";

import { useEffect, useState, useCallback } from "react";
import { toast } from "sonner";
import { listBackups, triggerBackup, BackupListResponse } from "@/lib/api";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Card, CardTitle } from "@/components/ui/Card";

export default function BackupsPage() {
  const [data, setData] = useState<BackupListResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [triggering, setTriggering] = useState(false);

  const fetchBackups = useCallback(async () => {
    try {
      const result = await listBackups();
      setData(result);
    } catch {
      toast.error("Failed to load backups");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchBackups();
  }, [fetchBackups]);

  const handleTrigger = async () => {
    setTriggering(true);
    try {
      const result = await triggerBackup();
      setData(result);
      toast.success("Backup started");
    } catch {
      toast.error("Failed to trigger backup");
    } finally {
      setTriggering(false);
    }
  };

  const formatSize = (bytes: number) => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const statusVariant = (status: string) => {
    switch (status) {
      case "success": return "success" as const;
      case "failed": return "danger" as const;
      case "running": return "warning" as const;
      default: return "default" as const;
    }
  };

  if (loading) return <p className="text-gray-500">Loading backups...</p>;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Database Backups</h2>
        <Button onClick={handleTrigger} disabled={triggering || data?.status === "running"}>
          {triggering ? "Starting..." : "Run Backup Now"}
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardTitle>Status</CardTitle>
          <Badge variant={statusVariant(data?.status || "idle")}>
            {data?.status || "idle"}
          </Badge>
        </Card>

        <Card>
          <CardTitle>Last Run</CardTitle>
          <p className="text-sm text-gray-700">
            {data?.last_run_at
              ? new Date(data.last_run_at).toLocaleString("id-ID")
              : "Never"}
          </p>
        </Card>

        <Card>
          <CardTitle>Last Status</CardTitle>
          {data?.last_status ? (
            <Badge variant={statusVariant(data.last_status)}>
              {data.last_status}
            </Badge>
          ) : (
            <span className="text-sm text-gray-400">N/A</span>
          )}
        </Card>
      </div>

      <Card>
        <CardTitle>Backup History</CardTitle>
        {data?.items && data.items.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full text-sm">
              <thead>
                <tr className="border-b border-gray-200">
                  <th className="text-left py-2 px-3 font-medium text-gray-500">Filename</th>
                  <th className="text-left py-2 px-3 font-medium text-gray-500">Size</th>
                  <th className="text-left py-2 px-3 font-medium text-gray-500">Created At</th>
                  <th className="text-left py-2 px-3 font-medium text-gray-500">Status</th>
                </tr>
              </thead>
              <tbody>
                {data.items.map((file) => (
                  <tr key={file.filename} className="border-b border-gray-100 hover:bg-gray-50">
                    <td className="py-2 px-3 font-mono text-xs">{file.filename}</td>
                    <td className="py-2 px-3">{formatSize(file.size_bytes)}</td>
                    <td className="py-2 px-3">
                      {new Date(file.created_at).toLocaleString("id-ID")}
                    </td>
                    <td className="py-2 px-3">
                      <Badge variant={statusVariant(file.status)}>{file.status}</Badge>
                      {file.error && (
                        <span className="ml-2 text-xs text-red-500" title={file.error}>
                          (!)
                        </span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-sm text-gray-400">No backups yet</p>
        )}
      </Card>
    </div>
  );
}
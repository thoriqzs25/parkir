"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { toast } from "sonner";
import { getIncident, resolveIncident, getIncidentNotes, createIncidentNote } from "@/lib/api";
import { Incident, IncidentNote } from "@/types/incident";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Select } from "@/components/ui/Select";
import { Card, CardTitle } from "@/components/ui/Card";
import { formatWIBDateTime } from "@/lib/time";
import { useAuth } from "@/hooks/useAuth";
import { hasPermission } from "@/lib/permissions";

const typeLabels: Record<string, string> = {
  STUCK_AT_GATE: "Stuck at Gate",
  PAYMENT_DISPUTE: "Payment Dispute",
  OPERATOR_ERROR: "Operator Error",
  SYSTEM_DOWNTIME: "System Downtime",
};

const stateVariants: Record<string, "danger" | "warning" | "success"> = {
  OPEN: "danger",
  IN_PROGRESS: "warning",
  RESOLVED: "success",
};

export default function IncidentDetailPage() {
  const params = useParams();
  const router = useRouter();
  const locationId = params.locationId as string;
  const incidentId = params.id as string;
  const { permissions } = useAuth();

  const [incident, setIncident] = useState<Incident | null>(null);
  const [notes, setNotes] = useState<IncidentNote[]>([]);
  const [note, setNote] = useState("");
  const [resolutionNotes, setResolutionNotes] = useState("");
  const [adjustmentAction, setAdjustmentAction] = useState("");
  const [adjustmentEntityId, setAdjustmentEntityId] = useState("");
  const [managerPin, setManagerPin] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const canResolve = hasPermission(permissions, "incidents:resolve");

  useEffect(() => {
    load();
  }, [incidentId]);

  const load = async () => {
    try {
      const [inc, incNotes] = await Promise.all([
        getIncident(incidentId),
        getIncidentNotes(incidentId),
      ]);
      setIncident(inc);
      setNotes(incNotes || []);
    } catch (err) {
      toast.error("Failed to load incident");
      router.push(`/${locationId}/incidents`);
    }
  };

  const handleResolve = async () => {
    if (!resolutionNotes) {
      toast.error("Resolution notes are required");
      return;
    }
    if (adjustmentAction && !managerPin) {
      toast.error("Manager PIN is required for adjustments");
      return;
    }

    setSubmitting(true);
    try {
      await resolveIncident(incidentId, {
        resolution_notes: resolutionNotes,
        adjustment_action: adjustmentAction || undefined,
        adjustment_entity_id: adjustmentAction ? adjustmentEntityId : undefined,
        manager_pin: adjustmentAction ? managerPin : undefined,
      });
      toast.success("Incident resolved");
      setResolutionNotes("");
      setAdjustmentAction("");
      setAdjustmentEntityId("");
      setManagerPin("");
      load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to resolve incident");
    } finally {
      setSubmitting(false);
    }
  };

  const handleAddNote = async () => {
    if (!note.trim()) return;
    try {
      await createIncidentNote(incidentId, { note });
      setNote("");
      const updatedNotes = await getIncidentNotes(incidentId);
      setNotes(updatedNotes || []);
    } catch (err) {
      toast.error("Failed to add note");
    }
  };

  if (!incident) return <p className="text-gray-500">Loading...</p>;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Incident Detail</h2>
        <Button variant="ghost" size="sm" onClick={() => router.push(`/${locationId}/incidents`)}>
          Back to list
        </Button>
      </div>

      <Card>
        <div className="space-y-3 p-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-sm text-gray-500">Type</p>
              <p className="font-medium">{typeLabels[incident.type]}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500">State</p>
              <Badge variant={stateVariants[incident.state]}>{incident.state}</Badge>
            </div>
            <div>
              <p className="text-sm text-gray-500">Reported At</p>
              <p className="font-medium">{formatWIBDateTime(incident.reported_at)}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500">Session</p>
              <p className="font-medium">{incident.session_id || "-"}</p>
            </div>
          </div>
          <div>
            <p className="text-sm text-gray-500">Description</p>
            <p className="mt-1 whitespace-pre-wrap">{incident.description}</p>
          </div>
          {incident.resolved_at && (
            <>
              <div>
                <p className="text-sm text-gray-500">Resolution Notes</p>
                <p className="mt-1 whitespace-pre-wrap">{incident.resolution_notes}</p>
              </div>
              {incident.adjustment_action && (
                <div>
                  <p className="text-sm text-gray-500">Adjustment</p>
                  <p className="font-medium">
                    {incident.adjustment_action} — {incident.adjustment_entity_id}
                  </p>
                </div>
              )}
            </>
          )}
        </div>
      </Card>

      {incident.state !== "RESOLVED" && canResolve && (
        <Card>
          <CardTitle>Resolve Incident</CardTitle>
          <div className="space-y-4 p-4">
            <Input
              label="Resolution Notes"
              placeholder="Describe the resolution..."
              value={resolutionNotes}
              onChange={(e) => setResolutionNotes(e.target.value)}
            />
            <Select
              label="Adjustment Action (optional)"
              value={adjustmentAction}
              onChange={(e) => setAdjustmentAction(e.target.value)}
            >
              <option value="">None</option>
              <option value="VOID_TRANSACTION">Void Transaction</option>
              <option value="REASSIGN_SESSION">Reassign Session</option>
            </Select>
            {adjustmentAction && (
              <>
                <Input
                  label="Adjustment Entity ID (session or transaction ID)"
                  placeholder="UUID"
                  value={adjustmentEntityId}
                  onChange={(e) => setAdjustmentEntityId(e.target.value)}
                />
                <Input
                  label="Manager PIN"
                  type="password"
                  placeholder="Enter your 6-digit PIN"
                  value={managerPin}
                  onChange={(e) => setManagerPin(e.target.value)}
                />
              </>
            )}
            <Button onClick={handleResolve} disabled={submitting}>
              {submitting ? "Resolving..." : "Resolve"}
            </Button>
          </div>
        </Card>
      )}

      <Card>
        <CardTitle>Notes</CardTitle>
        <div className="space-y-4 p-4">
          <div className="flex gap-2">
            <Input
              placeholder="Add a note..."
              value={note}
              onChange={(e) => setNote(e.target.value)}
            />
            <Button onClick={handleAddNote}>Add</Button>
          </div>
          {notes.length === 0 ? (
            <p className="text-sm text-gray-500">No notes yet.</p>
          ) : (
            <div className="space-y-3">
              {notes.map((n) => (
                <div key={n.id} className="rounded-lg border p-3">
                  <p className="text-sm whitespace-pre-wrap">{n.note}</p>
                  <p className="mt-1 text-xs text-gray-500">{formatWIBDateTime(n.created_at)}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      </Card>
    </div>
  );
}
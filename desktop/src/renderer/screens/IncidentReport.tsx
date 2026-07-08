import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

export function IncidentReport() {
  const { currentLocation } = useAuth();
  const navigate = useNavigate();
  const [type, setType] = useState("OPERATOR_ERROR");
  const [description, setDescription] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async () => {
    if (!currentLocation || !description.trim()) return;
    setSubmitting(true);
    try {
      const { request } = await import("../lib/api");
      await request("POST", "/incidents", {
        location_id: currentLocation.id,
        type,
        description,
      });
      alert("Incident reported successfully.");
      navigate("/dashboard");
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to report incident");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="screen">
      <button className="button secondary back" onClick={() => navigate("/dashboard")}>
        &larr; Back
      </button>
      <h2>Report Incident</h2>
      <div className="card form">
        <label>Type</label>
        <select value={type} onChange={(e) => setType(e.target.value)}>
          <option value="STUCK_AT_GATE">Stuck at Gate</option>
          <option value="PAYMENT_DISPUTE">Payment Dispute</option>
          <option value="OPERATOR_ERROR">Operator Error</option>
          <option value="SYSTEM_DOWNTIME">System Downtime</option>
        </select>
        <label>Description</label>
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={4}
          placeholder="Describe what happened..."
        />
        <button
          className="button primary"
          onClick={handleSubmit}
          disabled={submitting || !description.trim()}
        >
          {submitting ? "Submitting..." : "Submit"}
        </button>
      </div>
    </div>
  );
}
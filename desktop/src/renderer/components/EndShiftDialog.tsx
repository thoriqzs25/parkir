import { useState } from "react";

interface EndShiftDialogProps {
  open: boolean;
  onConfirm: (amount: number, notes?: string) => void;
  onCancel: () => void;
}

export function EndShiftDialog({ open, onConfirm, onCancel }: EndShiftDialogProps) {
  const [amount, setAmount] = useState("0");
  const [notes, setNotes] = useState("");

  if (!open) return null;

  return (
    <div className="overlay" onClick={onCancel}>
      <div className="dialog card" onClick={(e) => e.stopPropagation()}>
        <h2>End Shift</h2>
        <div className="form-group">
          <label>Cash Handover Amount</label>
          <input
            type="number"
            min="0"
            step="any"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            autoFocus
          />
        </div>
        <div className="form-group">
          <label>Discrepancy Notes (optional)</label>
          <input
            type="text"
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="e.g. shortage, overage, etc."
          />
        </div>
        <div className="dialog-actions">
          <button className="button secondary" onClick={onCancel}>
            Cancel
          </button>
          <button className="button danger" onClick={() => onConfirm(Number(amount), notes || undefined)}>
            End Shift
          </button>
        </div>
      </div>
    </div>
  );
}

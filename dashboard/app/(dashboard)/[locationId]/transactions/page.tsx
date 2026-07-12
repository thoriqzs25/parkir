"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { listTransactions, voidTransaction } from "@/lib/api";
import { Transaction } from "@/types/transaction";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Dialog } from "@/components/ui/Dialog";
import { Input } from "@/components/ui/Input";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { formatWIBDateTime } from "@/lib/time";

export default function TransactionsPage() {
  const params = useParams();
  const locationId = params.locationId as string;
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [offset, setOffset] = useState(0);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [filter, setFilter] = useState<"all" | "voided" | "non-voided">("all");
  const [voidDialogOpen, setVoidDialogOpen] = useState(false);
  const [selectedTransaction, setSelectedTransaction] = useState<Transaction | null>(null);
  const [managerPin, setManagerPin] = useState("");
  const [voidReason, setVoidReason] = useState("");
  const [voiding, setVoiding] = useState(false);
  const limit = 20;

  const load = async (newOffset = 0) => {
    setLoading(true);
    try {
      const q: Record<string, string> = {
        location_id: locationId,
        limit: String(limit),
        offset: String(newOffset),
      };
      if (filter === "voided") q.voided = "true";
      if (filter === "non-voided") q.voided = "false";
      const res = await listTransactions(q);
      setTransactions(
        newOffset === 0 ? res.items || [] : [...transactions, ...(res.items || [])]
      );
      setTotal(res.meta?.total || 0);
      setOffset(newOffset);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load transactions");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load(0);
  }, [locationId, filter]);

  const handleVoidOpen = (transaction: Transaction) => {
    setSelectedTransaction(transaction);
    setManagerPin("");
    setVoidReason("");
    setVoidDialogOpen(true);
  };

  const handleVoidClose = () => {
    setVoidDialogOpen(false);
    setSelectedTransaction(null);
    setManagerPin("");
    setVoidReason("");
  };

  const handleVoidSubmit = async () => {
    if (!selectedTransaction) return;
    if (!managerPin || managerPin.length !== 6) {
      toast.error("Manager PIN must be 6 digits");
      return;
    }
    if (!voidReason.trim()) {
      toast.error("Void reason is required");
      return;
    }

    setVoiding(true);
    try {
      await voidTransaction(selectedTransaction.id, {
        manager_pin: managerPin,
        void_reason: voidReason,
      });
      toast.success("Transaction voided successfully");
      handleVoidClose();
      load(0);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to void transaction");
    } finally {
      setVoiding(false);
    }
  };

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold">Transactions</h2>

      <div className="flex gap-2">
        {(["all", "voided", "non-voided"] as const).map((f) => (
          <Button
            key={f}
            variant={filter === f ? "primary" : "secondary"}
            size="sm"
            onClick={() => setFilter(f)}
          >
            {f === "non-voided" ? "Active" : f.charAt(0).toUpperCase() + f.slice(1)}
          </Button>
        ))}
      </div>

      {transactions.length === 0 && !loading ? (
        <p className="text-gray-500">No transactions found.</p>
      ) : (
          <Table>
          <Thead>
            <tr>
              <Th>Receipt</Th>
              <Th>Plate</Th>
              <Th>Method</Th>
              <Th>Amount</Th>
              <Th>Created</Th>
              <Th>Status</Th>
              <Th>Actions</Th>
            </tr>
          </Thead>
          <Tbody>
            {transactions.map((t) => (
              <tr key={t.id}>
                <Td>{t.receipt_number}</Td>
                <Td>{t.plate}</Td>
                <Td>{t.payment_method}</Td>
                <Td>Rp {t.fee_amount.toLocaleString("id-ID")}</Td>
                <Td>{formatWIBDateTime(t.created_at)}</Td>
                <Td>
                  {t.voided ? (
                    <Badge variant="danger">VOIDED</Badge>
                  ) : (
                    <Badge variant="success">OK</Badge>
                  )}
                </Td>
                <Td>
                  {!t.voided && (
                    <Button
                      variant="danger"
                      size="sm"
                      onClick={() => handleVoidOpen(t)}
                    >
                      Void
                    </Button>
                  )}
                </Td>
              </tr>
            ))}
          </Tbody>
        </Table>
      )}

      {transactions.length < total && (
        <Button
          variant="secondary"
          onClick={() => load(offset + limit)}
          disabled={loading}
        >
          {loading ? "Loading..." : "Load more"}
        </Button>
      )}

      <Dialog
        open={voidDialogOpen}
        onClose={handleVoidClose}
        title={`Void Transaction ${selectedTransaction?.receipt_number || ""}`}
      >
        <div className="space-y-4">
          <p className="text-sm text-gray-600">
            Plate: {selectedTransaction?.plate} | Amount: Rp {selectedTransaction?.fee_amount.toLocaleString("id-ID")}
          </p>
          <Input
            label="Manager PIN (6 digits)"
            type="password"
            inputMode="numeric"
            maxLength={6}
            value={managerPin}
            onChange={(e) => setManagerPin(e.target.value.replace(/\D/g, ""))}
            placeholder="123456"
          />
          <div className="w-full">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Void Reason
            </label>
            <textarea
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              rows={3}
              value={voidReason}
              onChange={(e) => setVoidReason(e.target.value)}
              placeholder="Enter reason for voiding this transaction"
            />
          </div>
          <div className="flex gap-2 justify-end">
            <Button variant="secondary" onClick={handleVoidClose} disabled={voiding}>
              Cancel
            </Button>
            <Button variant="danger" onClick={handleVoidSubmit} disabled={voiding}>
              {voiding ? "Voiding..." : "Confirm Void"}
            </Button>
          </div>
        </div>
      </Dialog>
    </div>
  );
}

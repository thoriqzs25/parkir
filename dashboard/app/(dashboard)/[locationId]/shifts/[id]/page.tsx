"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { toast } from "sonner";
import { ArrowLeft } from "lucide-react";
import { getShift } from "@/lib/api";
import { Shift } from "@/types/shift";
import { Transaction } from "@/types/transaction";
import { Button } from "@/components/ui/Button";
import { Badge } from "@/components/ui/Badge";
import { Card, CardTitle } from "@/components/ui/Card";
import { Table, Thead, Tbody, Th, Td } from "@/components/ui/Table";
import { formatWIBDateTime } from "@/lib/time";

function formatCurrency(n?: number) {
  if (n === undefined || n === null) return "-";
  return `Rp ${n.toLocaleString("id-ID")}`;
}

export default function ShiftDetailPage() {
  const params = useParams();
  const router = useRouter();
  const locationId = params.locationId as string;
  const shiftId = params.id as string;

  const [shift, setShift] = useState<Shift | null>(null);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [summary, setSummary] = useState<{ transaction_count: number; expected_cash: number } | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    getShift(shiftId, "transactions")
      .then((res) => {
        setShift(res.shift);
        setTransactions(res.transactions || []);
        setSummary(res.summary || null);
      })
      .catch((err) => {
        toast.error(err instanceof Error ? err.message : "Failed to load shift");
      })
      .finally(() => setLoading(false));
  }, [shiftId]);

  if (loading) {
    return <p className="text-gray-500">Loading shift...</p>;
  }

  if (!shift) {
    return (
      <div className="space-y-4">
        <p className="text-gray-500">Shift not found.</p>
        <Button variant="secondary" onClick={() => router.push(`/${locationId}/shifts`)}>
          Back to shifts
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="secondary" size="sm" asChild>
          <Link href={`/${locationId}/shifts`}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back
          </Link>
        </Button>
        <h2 className="text-xl font-semibold">Shift Detail</h2>
      </div>

      <Card>
        <CardTitle>Shift Summary</CardTitle>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <p className="text-sm text-gray-500">Status</p>
            <Badge
              variant={
                shift.status === "OPEN"
                  ? "success"
                  : shift.status === "FORCE_CLOSED"
                  ? "danger"
                  : shift.status === "FLAGGED"
                  ? "warning"
                  : "default"
              }
            >
              {shift.status}
            </Badge>
          </div>
          <div>
            <p className="text-sm text-gray-500">Started</p>
            <p className="font-medium">{formatWIBDateTime(shift.started_at)}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Ended</p>
            <p className="font-medium">{shift.ended_at ? formatWIBDateTime(shift.ended_at) : "-"}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Expected Cash</p>
            <p className="font-medium">{formatCurrency(shift.expected_cash)}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Cash Handover</p>
            <p className="font-medium">{formatCurrency(shift.cash_handover_amount)}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Discrepancy</p>
            <p className="font-medium">{formatCurrency(shift.discrepancy)}</p>
          </div>
          {shift.discrepancy_notes && (
            <div className="sm:col-span-2">
              <p className="text-sm text-gray-500">Discrepancy Notes</p>
              <p className="font-medium">{shift.discrepancy_notes}</p>
            </div>
          )}
          {shift.force_closed_reason && (
            <div className="sm:col-span-2">
              <p className="text-sm text-gray-500">Force Close Reason</p>
              <p className="font-medium">{shift.force_closed_reason}</p>
            </div>
          )}
        </div>
      </Card>

      <Card>
        <CardTitle>Cash Summary</CardTitle>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <p className="text-sm text-gray-500">Transactions</p>
            <p className="font-medium">{summary?.transaction_count ?? 0}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Expected Cash (from transactions)</p>
            <p className="font-medium">{formatCurrency(summary?.expected_cash)}</p>
          </div>
        </div>
      </Card>

      <Card>
        <CardTitle>Transactions</CardTitle>
        {transactions.length === 0 ? (
          <p className="text-gray-500">No transactions for this shift.</p>
        ) : (
          <Table>
            <Thead>
              <tr>
                <Th>Receipt</Th>
                <Th>Plate</Th>
                <Th>Method</Th>
                <Th>Amount</Th>
                <Th>Status</Th>
              </tr>
            </Thead>
            <Tbody>
              {transactions.map((tx) => (
                <tr key={tx.id}>
                  <Td>{tx.receipt_number}</Td>
                  <Td>{tx.plate}</Td>
                  <Td>{tx.payment_method}</Td>
                  <Td>{formatCurrency(tx.fee_amount)}</Td>
                  <Td>
                    {tx.voided ? <Badge variant="danger">Voided</Badge> : <Badge variant="success">Paid</Badge>}
                  </Td>
                </tr>
              ))}
            </Tbody>
          </Table>
        )}
      </Card>
    </div>
  );
}

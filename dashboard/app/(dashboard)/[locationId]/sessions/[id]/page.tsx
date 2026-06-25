"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { toast } from "sonner";
import { ArrowLeft } from "lucide-react";
import { getSession } from "@/lib/api";
import { Session } from "@/types/session";
import { Transaction } from "@/types/transaction";
import { Button } from "@/components/ui/Button";
import { Badge } from "@/components/ui/Badge";
import { Card, CardTitle } from "@/components/ui/Card";
import { formatWIBDateTime } from "@/lib/time";

function formatCurrency(n?: number) {
  if (n === undefined || n === null) return "-";
  return `Rp ${n.toLocaleString("id-ID")}`;
}

function calculateDuration(checkIn: string, checkOut?: string) {
  const start = new Date(checkIn).getTime();
  const end = checkOut ? new Date(checkOut).getTime() : Date.now();
  const hours = Math.max(1, Math.ceil((end - start) / (1000 * 60 * 60)));
  return `${hours} hour${hours > 1 ? "s" : ""}`;
}

export default function SessionDetailPage() {
  const params = useParams();
  const router = useRouter();
  const locationId = params.locationId as string;
  const sessionId = params.id as string;

  const [session, setSession] = useState<Session | null>(null);
  const [transaction, setTransaction] = useState<Transaction | undefined>(undefined);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    getSession(sessionId, "transaction")
      .then((res) => {
        setSession(res.session);
        setTransaction(res.transaction);
      })
      .catch((err) => {
        toast.error(err instanceof Error ? err.message : "Failed to load session");
      })
      .finally(() => setLoading(false));
  }, [sessionId]);

  if (loading) {
    return <p className="text-gray-500">Loading session...</p>;
  }

  if (!session) {
    return (
      <div className="space-y-4">
        <p className="text-gray-500">Session not found.</p>
        <Button variant="secondary" onClick={() => router.push(`/${locationId}/sessions/active`)}>
          Back to sessions
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="secondary" size="sm" asChild>
          <Link href={`/${locationId}/sessions/history`}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back
          </Link>
        </Button>
        <h2 className="text-xl font-semibold">Session Detail</h2>
      </div>

      <Card>
        <CardTitle>{session.plate}</CardTitle>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <p className="text-sm text-gray-500">Vehicle Type</p>
            <p className="font-medium">{session.vehicle_type}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">State</p>
            <Badge
              variant={
                session.state === "ACTIVE"
                  ? "success"
                  : session.state === "PENDING_PAYMENT"
                  ? "warning"
                  : session.state === "VOIDED"
                  ? "danger"
                  : "default"
              }
            >
              {session.state}
            </Badge>
          </div>
          <div>
            <p className="text-sm text-gray-500">Check In</p>
            <p className="font-medium">{formatWIBDateTime(session.check_in_at)}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Check Out</p>
            <p className="font-medium">
              {session.check_out_at ? formatWIBDateTime(session.check_out_at) : "-"}
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Duration</p>
            <p className="font-medium">{calculateDuration(session.check_in_at, session.check_out_at)}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Fee</p>
            <p className="font-medium">{formatCurrency(session.fee_amount)}</p>
          </div>
          {session.rate_snapshot && Object.keys(session.rate_snapshot).length > 0 && (
            <div className="sm:col-span-2">
              <p className="text-sm text-gray-500">Rate Snapshot</p>
              <pre className="mt-1 max-h-40 overflow-auto rounded bg-gray-50 p-2 text-xs text-gray-700">
                {JSON.stringify(session.rate_snapshot, null, 2)}
              </pre>
            </div>
          )}
        </div>
      </Card>

      {transaction && (
        <Card>
          <CardTitle>Transaction</CardTitle>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <p className="text-sm text-gray-500">Receipt Number</p>
              <p className="font-medium">{transaction.receipt_number}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500">Payment Method</p>
              <p className="font-medium">{transaction.payment_method}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500">Amount</p>
              <p className="font-medium">{formatCurrency(transaction.fee_amount)}</p>
            </div>
            {transaction.payment_method === "CASH" && (
              <>
                <div>
                  <p className="text-sm text-gray-500">Tendered</p>
                  <p className="font-medium">{formatCurrency(transaction.amount_tendered ?? undefined)}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-500">Change</p>
                  <p className="font-medium">{formatCurrency(transaction.change_amount ?? undefined)}</p>
                </div>
              </>
            )}
            {transaction.payment_reference && (
              <div>
                <p className="text-sm text-gray-500">Reference</p>
                <p className="font-medium">{transaction.payment_reference}</p>
              </div>
            )}
            <div>
              <p className="text-sm text-gray-500">Paid At</p>
              <p className="font-medium">{formatWIBDateTime(transaction.created_at)}</p>
            </div>
            {transaction.voided && (
              <div className="sm:col-span-2">
                <Badge variant="danger">Voided</Badge>
                <p className="mt-1 text-sm text-gray-500">
                  Reason: {transaction.void_reason || "-"}
                </p>
                <p className="text-sm text-gray-500">
                  Voided at: {transaction.voided_at ? formatWIBDateTime(transaction.voided_at) : "-"}
                </p>
              </div>
            )}
          </div>
        </Card>
      )}
    </div>
  );
}

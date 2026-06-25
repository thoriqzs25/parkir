"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { listTransactions } from "@/lib/api";
import { Transaction } from "@/types/transaction";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
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
    </div>
  );
}

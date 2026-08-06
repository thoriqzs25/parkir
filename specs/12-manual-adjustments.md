# Chapter 12 — Manual Adjustments

## 12.1 Overview

Manual adjustments are privileged operations that modify or correct committed records. They require elevated permissions and **manager authorization** for every action. Every adjustment is fully audit-logged with actor, reason, and timestamp.

In v2, the primary adjustment is **voiding transactions** (no operator reassignment since there are no operators).

| Adjustment | Permission Required | Applicable States |
|-----------|-------------------|------------------|
| Void / Cancel a Transaction | `payments:void` | `CLOSED`, `PENDING_PAYMENT` |

---

## 12.2 Authorization Model

All manual adjustments require:
1. The acting user to have the relevant permission (`payments:void`).
2. **Manager PIN confirmation** — a second-factor step where the authorizing manager enters their PIN before the action is committed.

**Manager PIN:**
- A 6-digit numeric PIN set separately from the manager's login password.
- Configured in the manager's profile settings.
- Not the same as the login password.

---

## 12.3 Void / Cancel a Transaction

### Purpose
Cancel a completed or pending payment, removing it from revenue totals. Used when:
- Session was created in error (hardware issue).
- Payment dispute resolved in driver's favor.
- Duplicate transaction detected.
- Offline SOP transaction needs correction.

### Applicable States
- `CLOSED` — transaction already recorded.
- `PENDING_PAYMENT` — check-out initiated but payment not yet confirmed.

### Procedure

**From Web Dashboard (Manager):**

1. Navigate to **Transactions**.
2. Search for the transaction by: transaction ID, session ID, or gate ID.
3. Click **Void** button.
4. System displays transaction details:
   - Location, gate, vehicle type.
   - Check-in/check-out times.
   - Amount, payment method.
   - Shift number.
5. Manager enters **void reason** (required):
   - Text field (min 10 characters).
   - Examples: "Hardware error - duplicate transaction", "Payment dispute resolved", "Offline SOP correction".
6. Manager enters **PIN** to authorize.
7. System validates PIN.
8. System marks transaction as `voided = true`, records `voided_at`, `voided_by`, `void_reason`.
9. System updates session state to `VOIDED`.
10. Audit log records void action.
11. Transaction excluded from revenue reports.

### Void Constraints
- Cannot void already-voided transaction.
- Cannot void transaction older than 30 days (configurable).
- Void requires manager permission + PIN.
- Void reason is mandatory.

---

## 12.4 Void from Server Room App (Staff Request)

### Purpose
Staff can request transaction void from server room app (requires manager approval via dashboard).

### Procedure

**Staff Request (Server Room App):**

1. Staff identifies transaction to void (e.g., offline SOP correction).
2. Staff opens transaction detail on server room app.
3. Staff clicks **Request Void**.
4. Staff enters reason.
5. System creates void request (pending manager approval).
6. Request visible on dashboard (pending approvals).

**Manager Approval (Dashboard):**

1. Manager navigates to **Pending Approvals**.
2. Manager reviews void request (transaction details, reason, staff note).
3. Manager enters PIN to approve.
4. System voids transaction (same as direct void).
5. Audit log records void action (manager as authorizer, staff as requester).

---

## 12.5 Adjustment Audit Trail

Every adjustment is logged in `audit_logs`:

```json
{
  "id": "uuid",
  "user_id": "manager-uuid",
  "action": "TRANSACTION_VOIDED",
  "entity_type": "transaction",
  "entity_id": "transaction-uuid",
  "location_id": "location-uuid",
  "metadata": {
    "transaction_id": "transaction-uuid",
    "session_id": "session-uuid",
    "amount": 1100000,
    "payment_method": "EMONEY",
    "void_reason": "Hardware error - duplicate transaction",
    "requested_by": "staff-uuid"  // null if manager initiated
  },
  "ip_address": "192.168.1.100",
  "created_at": "2026-08-06T14:32:00Z"
}
```

---

## 12.6 Adjustment Reporting

### Adjustment Log (Dashboard)

| Column | Description |
|--------|-------------|
| Date | Adjustment date |
| Time | Adjustment time |
| Location | Location name |
| Transaction ID | Transaction ID |
| Amount | Voided amount |
| Reason | Void reason |
| Authorized By | Manager name |
| Requested By | Staff name (if applicable) |

### Adjustment Summary

| Metric | Description |
|--------|-------------|
| Total Voids (Today) | Count of voided transactions today |
| Total Voided Amount (Today) | Sum of voided amounts today |
| Void Rate | Voided transactions / Total transactions (%) |
| Top Void Reasons | Most common void reasons |

---

## 12.7 Design Decisions

**Why only transaction voiding (no operator reassignment)?**
- No operators in v2 (fully automated).
- Sessions/transactions are system-generated.
- No manual data entry to correct.

**Why manager PIN required?**
- Prevents unauthorized voids.
- Audit trail (who authorized).
- Compliance (financial controls).

**Why void reason mandatory?**
- Audit trail (why voided).
- Prevents abuse.
- Compliance (documentation).

**Why 30-day void limit?**
- Prevents old transaction modifications.
- Accounting period closure.
- Configurable if needed.

**Why staff can request but not approve?**
- Separation of duties.
- Manager oversight.
- Prevents staff from voiding own transactions.

---

*End of Chapter 12 — Manual Adjustments (v2)*

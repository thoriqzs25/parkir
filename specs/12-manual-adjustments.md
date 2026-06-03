# Chapter 12 — Manual Adjustment Procedures

## 12.1 Overview

Manual adjustments are privileged operations that modify or correct committed records. They require elevated permissions and **manager authorization** for every action. Every adjustment is fully audit-logged with actor, reason, and timestamp.

Two adjustment types are supported in v1:

| Adjustment | Permission Required | Applicable States |
|-----------|-------------------|------------------|
| Void / Cancel a Transaction | `adjustments:void_transaction` | `CLOSED`, `PENDING_PAYMENT` |
| Reassign Session to Another Operator | `adjustments:reassign_session` | `ACTIVE`, `PENDING_PAYMENT`, `CLOSED` |

---

## 12.2 Authorization Model

All manual adjustments require:
1. The acting user to have the relevant `adjustments:*` permission.
2. **Manager PIN confirmation** — a second-factor step where the authorizing manager enters their PIN before the action is committed.

If the operator and manager are different people:
- Operator initiates the adjustment request from the desktop app or web dashboard.
- Manager reviews and authorizes from the web dashboard (or by entering PIN on the operator's terminal).

If the manager is performing the adjustment themselves:
- They enter their own PIN to confirm.

**PIN:**
- A 6-digit numeric PIN set separately from the manager's login password.
- Configured in the manager's profile settings.
- Not the same as the login password.

---

## 12.3 Void / Cancel a Transaction

### Purpose
Cancel a completed or pending payment, removing it from revenue totals. Used when a session was created in error, a dispute is resolved in the driver's favor, or an operator error requires correction.

### Applicable States
- `CLOSED` — transaction already recorded
- `PENDING_PAYMENT` — check-out initiated but payment not yet confirmed

### Procedure

**From Web Dashboard (Manager):**

1. Navigate to **Adjustments → Void Transaction**.
2. Search for the transaction by: receipt number, session ID, or plate number.
3. Transaction details are displayed: plate, vehicle type, check-in/out times, fee, payment method.
4. Select **Void Transaction**.
5. Enter void reason (free text, required, min 10 characters).
6. Enter manager PIN to authorize.
7. System confirms: transaction is voided.

**From Desktop App (Operator initiates, Manager authorizes):**

1. Operator navigates to session detail screen.
2. Selects **Request Void**.
3. Enters description of the problem (pre-fills incident if one is linked).
4. Manager is notified (in-app alert on dashboard).
5. Manager reviews and authorizes from dashboard (steps 4–7 above).

### Effects
| Target | Change |
|--------|--------|
| `transactions.voided` | Set to `true` |
| `transactions.voided_at` | Set to current timestamp |
| `transactions.voided_by` | Set to authorizing manager's user ID |
| `transactions.void_reason` | Set to entered reason |
| `sessions.state` | Set to `VOIDED` |
| Revenue reports | Transaction excluded from all totals |
| Audit log | Entry created: `TRANSACTION_VOIDED` |

### Constraints
- A transaction can only be voided once.
- Voiding is irreversible — there is no "un-void" in v1.
- A voided session's plate becomes available for new check-in at the same location.

---

## 12.4 Reassign Session to a Different Operator

### Purpose
Correct the operator attributed to a session. Used when a session was accidentally opened under the wrong operator's account (e.g. a colleague was still logged in on a shared terminal).

### Applicable States
- `ACTIVE`
- `PENDING_PAYMENT`
- `CLOSED`

Not applicable to `VOIDED` sessions.

### Procedure

**From Web Dashboard (Manager):**

1. Navigate to **Adjustments → Reassign Session**.
2. Search for the session by: session ID, plate number, or operator name.
3. Session details are displayed: plate, vehicle type, current operator, check-in time, state.
4. Select the **Target Operator** from a dropdown (must be an active operator at the same location).
5. Enter reassignment reason (free text, required).
6. Enter manager PIN to authorize.
7. System confirms: session operator updated.

### Effects

| Target | Change |
|--------|--------|
| `sessions.operator_id` | Updated to target operator's ID |
| `transactions.operator_id` | Updated if a closed transaction exists |
| Original operator's activity log | Session marked as reassigned away; visible in log |
| Target operator's activity log | Session appears with `[REASSIGNED]` tag |
| Audit log | Entry created: `SESSION_REASSIGNED` with both operator IDs and reason |

### Constraints
- The target operator must be assigned to the same location as the session.
- Reassignment does not change the session's state, timestamps, or fee.
- The original operator attribution is preserved in the audit log — it is never erased.

---

## 12.5 Audit Log Entries for Adjustments

Every adjustment produces an immutable audit log entry. See Chapter 13 for full audit log spec.

### TRANSACTION_VOIDED Entry
```json
{
  "action": "TRANSACTION_VOIDED",
  "actor_id": "<manager_user_id>",
  "entity_type": "transaction",
  "entity_id": "<transaction_id>",
  "location_id": "<location_id>",
  "metadata": {
    "session_id": "<session_id>",
    "plate": "B 1234 XYZ",
    "fee_amount": 20000,
    "void_reason": "Driver was overcharged due to operator error"
  },
  "timestamp": "2025-03-15T14:32:00Z"
}
```

### SESSION_REASSIGNED Entry
```json
{
  "action": "SESSION_REASSIGNED",
  "actor_id": "<manager_user_id>",
  "entity_type": "session",
  "entity_id": "<session_id>",
  "location_id": "<location_id>",
  "metadata": {
    "original_operator_id": "<operator_a_id>",
    "new_operator_id": "<operator_b_id>",
    "reason": "Session opened under wrong account"
  },
  "timestamp": "2025-03-15T14:35:00Z"
}
```

---

## 12.6 Adjustment History View

Managers can view a history of all adjustments from the web dashboard:

**Filters:**
- Date range
- Adjustment type (void / reassign)
- Location
- Authorizing manager

**Table columns:**

| Column | Description |
|--------|-------------|
| Timestamp | When the adjustment was made |
| Type | Void / Reassign |
| Session / Receipt | Linked record |
| Plate | Vehicle plate |
| Reason | Entered reason |
| Authorized By | Manager name |

---

## 12.7 Out of Scope (v1)

- In-place editing of plate number or vehicle type (requires void + re-create instead)
- Fee override / discount without full void + re-create
- Multi-level approval (single manager authorization is sufficient)
- Adjustment requests queue (v1 uses synchronous authorization)

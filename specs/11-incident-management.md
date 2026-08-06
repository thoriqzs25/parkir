# Chapter 11 — Incident Management

## 11.1 Overview

The incident management module provides a structured workflow for handling hardware failures, payment issues, and other exceptions in the automated parking system. Incidents are triggered by alerts (driver presses button, hardware error detected) and handled by server room staff.

Every incident is recorded with full context and remains in the system permanently for audit purposes.

---

## 11.2 Incident Types

| Code | Name | Typical Trigger |
|------|------|----------------|
| `HARDWARE_FAILURE` | Hardware Failure | Gate motor, printer, scanner, or sensor malfunction |
| `PAYMENT_FAILURE` | Payment Failure | Payment terminal error, insufficient balance, card read error |
| `QR_UNREADABLE` | QR Unreadable | QR code damaged, dirty, or unreadable by scanner |
| `PRINTER_JAM` | Printer Jam | Thermal printer paper jam or empty |
| `GATE_MOTOR_FAILURE` | Gate Motor Failure | Gate barrier won't open or close |
| `VEHICLE_STUCK` | Vehicle Stuck | Vehicle stuck at gate (barrier came down, slow-moving) |
| `SYSTEM_DOWNTIME` | System Downtime | Server room app or cloud backend unreachable |
| `OTHER` | Other | Any other incident not covered above |

---

## 11.3 Incident Lifecycle

```
Alert triggered (driver presses button, hardware error)
      │
      ▼
  INCIDENT_CREATED
  - Incident record created
  - State: OPEN
  - Audio alert played in server room
      │
      │  Staff walks to gate, investigates
      │
      ▼
  INCIDENT_IN_PROGRESS
  - Staff working on resolution
  - State: IN_PROGRESS
      │
      │  Staff resolves issue
      │
      ▼
  INCIDENT_RESOLVED
  - Resolution noted
  - State: RESOLVED
  - Gate returns to normal operation
```

---

## 11.4 Incident Creation

### Automatic Creation (Hardware Errors)

Gate app detects hardware error → sends alert to server room app → server room app creates incident.

**Example:**
```json
{
  "incident_type": "PRINTER_JAM",
  "gate_id": "GATE-ENTRY-01",
  "description": "Printer paper jam detected",
  "state": "OPEN",
  "created_at": "2026-08-06T10:30:00Z"
}
```

### Manual Creation (Driver Alert)

Driver presses alert button → gate app sends alert to server room app → server room app creates incident.

**Example:**
```json
{
  "incident_type": "QR_UNREADABLE",
  "gate_id": "GATE-EXIT-01",
  "description": "Driver pressed alert button - QR unreadable",
  "state": "OPEN",
  "created_at": "2026-08-06T14:32:00Z"
}
```

### Audio Alert

Server room app plays audio alert when incident created:
```
"Gate {gate_id} needs assistance"
```

Example:
```
"Gate GATE-EXIT-01 needs assistance"
```

---

## 11.5 Incident Resolution

### Staff Workflow

1. Audio alert plays in server room.
2. Staff walks to gate.
3. Staff investigates issue.
4. Staff resolves issue (or runs offline SOP).
5. Staff marks incident as `IN_PROGRESS` (on server room app).
6. Staff resolves issue.
7. Staff marks incident as `RESOLVED` (on server room app).
8. Gate returns to normal operation.

### Resolution Actions

| Incident Type | Typical Resolution |
|---------------|-------------------|
| `HARDWARE_FAILURE` | Restart gate app, check hardware connections, call technician |
| `PAYMENT_FAILURE` | Driver tops up card and retries, or staff handles via offline SOP |
| `QR_UNREADABLE` | Staff manually enters session ID (if available), or runs offline SOP |
| `PRINTER_JAM` | Staff clears jam, refills paper, resets printer |
| `GATE_MOTOR_FAILURE` | Staff manually lifts barrier, calls technician |
| `VEHICLE_STUCK` | Staff manually operates gate, assists driver |
| `SYSTEM_DOWNTIME` | Wait for system to recover, run offline SOP if needed |
| `OTHER` | Staff handles based on situation |

### Offline SOP

If system cannot be fixed immediately, staff runs offline SOP:
- Manual gate operation (if authorized).
- Paper receipt (if needed).
- Record transaction later (when system restored).
- Incident marked as `RESOLVED` with note: "Offline SOP executed".

---

## 11.6 Incident Notes

Staff can add notes to incidents (for audit trail).

**Example:**
```json
{
  "incident_id": "uuid",
  "note": "Printer jam cleared. Refilled paper. Gate operational.",
  "created_by": "staff-uuid",
  "created_at": "2026-08-06T10:45:00Z"
}
```

---

## 11.7 Incident Reporting

### Incident List View (Dashboard)

| Column | Description |
|--------|-------------|
| Date | Incident date |
| Time | Incident time |
| Location | Location name |
| Gate | Gate ID |
| Type | Incident type |
| State | OPEN / IN_PROGRESS / RESOLVED |
| Duration | Time from creation to resolution |
| Actions | View details |

### Incident Detail View

- Incident info (type, gate, location, state, timestamps).
- Description.
- Notes (timeline).
- Resolution details.

### Incident Reports

**Incident Summary (by type):**
| Type | Count | Avg Resolution Time |
|------|-------|---------------------|
| PRINTER_JAM | 15 | 5 minutes |
| QR_UNREADABLE | 8 | 3 minutes |
| GATE_MOTOR_FAILURE | 2 | 30 minutes |

**Incident Trend (over time):**
- Line chart: incidents per day/week/month.
- Breakdown by type.

---

## 11.8 Design Decisions

**Why automatic incident creation?**
- Faster response (no manual reporting).
- Accurate (system detects errors).
- Audit trail (all incidents logged).

**Why audio alerts?**
- Immediate awareness (staff on-site).
- Faster response time.
- No need to check dashboard.

**Why incident notes?**
- Audit trail (what happened, how resolved).
- Knowledge sharing (staff learn from past incidents).
- Compliance (documentation).

**Why offline SOP reference?**
- System can't fix everything immediately.
- Staff need fallback procedure.
- Keeps parking operational during outages.

---

*End of Chapter 11 — Incident Management (v2)*

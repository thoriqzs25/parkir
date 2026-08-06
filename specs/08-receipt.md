# Chapter 8 — Receipt

## 8.1 Overview

In v2, there are two types of printed documents:

1. **Entry Ticket** — Automatic thermal receipt dispensed on entry (QR code + session info).
2. **Exit Receipt** — Optional thermal receipt printed on exit (driver presses button).

Both are printed by the gate app via thermal printer (Epson). No operator involvement.

---

## 8.2 Entry Ticket

### Print Trigger

| Trigger | Behavior |
|---------|----------|
| Session created (state: ACTIVE) | Entry ticket printed automatically |

- Printed by gate app (entry gate).
- Uses Epson thermal printer (ESC/POS commands).
- If printer unavailable: gate stays closed, alert triggered.

### Ticket Layout

Thermal receipt width: **58mm** (standard).

```
================================
        [LOCATION NAME]
      [Location Address]
================================
Ticket No  : AMB-CENTRAL-20260806-001
Date       : 06 Aug 2026  10:30
Vehicle    : Motorcycle
--------------------------------

      [QR CODE - 200x200px]

--------------------------------
Kunci kendaraan anda dengan rapat.
Jangan tinggalkan karcis parkir di
dalam kendaraan Anda.
================================
```

### QR Code Content

QR code encodes JSON:

```json
{
  "session_id": "uuid-v4",
  "location_id": "uuid-v4",
  "timestamp": "2026-08-06T10:30:00Z",
  "vehicle_type": "MOTO"
}
```

**QR Code Specs:**
- Size: 200x200 pixels (printable on 58mm paper).
- Error correction: Medium (25%).
- Encoding: UTF-8.

### Ticket Number Format

```
{LOCATION_CODE}-{DATE}-{SEQUENCE}

Example: AMB-CENTRAL-20260806-001
```

- Location code: from location config.
- Date: YYYYMMDD.
- Sequence: daily sequence number (001, 002, ...).

---

## 8.3 Exit Receipt (Optional)

### Print Trigger

| Trigger | Behavior |
|---------|----------|
| Driver presses receipt button | Exit receipt printed on demand |

- Printed by gate app (exit gate).
- Optional (driver can skip).
- Available for 30 seconds after gate opens.
- If printer unavailable: receipt not printed, but session still closes.

### Receipt Layout

Thermal receipt width: **58mm** (standard).

```
================================
        [LOCATION NAME]
      [Location Address]
================================
Receipt No : AMB-CENTRAL-20260806-001
Date       : 06 Aug 2026  14:32
--------------------------------
Vehicle    : Motorcycle
--------------------------------
Check-in   : 06 Aug 2026  10:30
Check-out  : 06 Aug 2026  14:32
Duration   : 4 hours 2 minutes
--------------------------------
Fee        : Rp 8,000
Payment    : E-Money
Shift      : Shift #15
--------------------------------
Transaction: uuid-v4
Config     : rate_v42
================================
   Terima kasih atas kunjungan Anda
================================
```

### Receipt Fields

| Field | Description |
|-------|-------------|
| Receipt No | Same as ticket number (entry ticket) |
| Date | Check-out date/time |
| Vehicle | Vehicle type (from gate config) |
| Check-in | Check-in date/time |
| Check-out | Check-out date/time |
| Duration | Parking duration (hours, minutes) |
| Fee | Total fee (in Rupiah) |
| Payment | Payment method (E-Money, Flazz, etc.) |
| Shift | Shift number (continuous) |
| Transaction | Transaction ID (UUID) |
| Config | Rate config version (audit trail) |

---

## 8.4 Re-Print

### Entry Ticket
- Not re-printable (driver must exit to get receipt).
- If ticket lost: staff handles via offline SOP.

### Exit Receipt
- Re-print available for 30 seconds after gate opens.
- Driver presses receipt button again.
- Re-printed receipts marked `REPRINT` on document.
- After 30 seconds: receipt no longer available.

---

## 8.5 Printer Configuration

### Printer Type
- Epson thermal printer (existing AMB hardware).
- Communication: USB or serial (via hub).
- Paper width: 58mm (standard).

### Printer Commands
- ESC/POS command set.
- Gate app sends commands via HAL (hardware abstraction layer).

### Printer Status
- Gate app monitors printer status (online/offline, paper status).
- If printer offline: alert triggered, gate stays closed (entry) or receipt not printed (exit).

---

## 8.6 Receipt Number Sequence

### Entry Ticket
- Sequence per location per day.
- Format: `{LOCATION_CODE}-{YYYYMMDD}-{SEQ}`.
- Sequence resets daily (001, 002, ...).
- Stored in server room app (local DB).

### Exit Receipt
- Same number as entry ticket (no separate sequence).
- Links exit receipt to entry ticket.

---

## 8.7 Design Decisions

**Why QR code on entry ticket?**
- Encodes session data (no manual entry).
- Fast scanning at exit (< 1 second).
- Tamper-resistant (encrypted or signed).
- Standard practice (matches AMB's current system).

**Why optional exit receipt?**
- Reduces paper waste (not all drivers want receipt).
- Faster exit (no waiting for receipt).
- Driver can press button if needed.

**Why no operator info on receipt?**
- No operators (fully automated).
- Gate ID could be added (for audit).

**Why config version on receipt?**
- Audit trail (which rate was used).
- Debugging (if fee calculation issue).
- Compliance (proof of rate used).

**Why shift number on receipt?**
- Audit trail (which shift session was in).
- Reporting (aggregate by shift).
- Reconciliation (shift-based reporting).

---

*End of Chapter 8 — Receipt (v2)*

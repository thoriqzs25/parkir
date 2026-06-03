# Chapter 8 — Receipt

## 8.1 Overview

A thermal receipt is automatically printed when a parking session is successfully closed. The receipt serves as the driver's proof of payment and the operator's confirmation that the transaction was recorded. Receipts are printed from the operator's desktop app to a thermal printer connected to the operator's workstation.

---

## 8.2 Print Trigger

| Trigger | Behavior |
|---------|----------|
| Session moves to `CLOSED` | Receipt print job sent automatically |
| Operator manually re-prints | Available from session detail view; requires `sessions:view` permission |

- If the printer is unavailable at the time of session close, the system shows an error but **does not block the session from closing**. The operator can retry printing from the session detail screen.
- Re-print is allowed any number of times; re-printed receipts are marked `REPRINT` on the document.

---

## 8.3 Receipt Layout

Thermal receipt width: **58mm** (standard) or **80mm** (wider format) — configurable per terminal.

```
================================
        [LOCATION NAME]
      [Location Address]
================================
Receipt No : GMP01-20250315-00042
Date       : 15 Mar 2025  14:32
Operator   : Budi Santoso
--------------------------------
Plate      : B 1234 XYZ
Vehicle    : Car
--------------------------------
Check-in   : 15 Mar 2025  11:00
Check-out  : 15 Mar 2025  14:32
Duration   : 4 hours
--------------------------------
Rate       : Rp 5,000 / hour
Fee        : Rp 20,000
--------------------------------
Payment    : CASH
Tendered   : Rp 25,000
Change     : Rp  5,000
================================
     Thank you for parking!
        [LOCATION NAME]
================================
```

For digital payments, the `Tendered` and `Change` lines are replaced with:

```
Payment    : DIGITAL (QRIS)
Ref        : TRX-20250315-88291
```

---

## 8.4 Receipt Fields

| Field | Source | Notes |
|-------|--------|-------|
| Location name | `locations.name` | |
| Location address | `locations.address` | |
| Receipt number | `transactions.receipt_number` | Format: `[CODE]-[YYYYMMDD]-[SEQ]` |
| Date & time | `transactions.check_out_at` | Formatted for local timezone |
| Operator name | `users.name` | Operator who closed the session |
| Plate number | `transactions.plate` | |
| Vehicle type | `transactions.vehicle_type` | Human-readable (e.g. "Car") |
| Check-in time | `transactions.check_in_at` | |
| Check-out time | `transactions.check_out_at` | |
| Duration | `transactions.duration_hours` | e.g. "4 hours" |
| Rate | `transactions.rate_hourly` | e.g. "Rp 5,000 / hour" |
| Fee | `transactions.fee_amount` | |
| Payment method | `transactions.payment_method` | |
| Amount tendered | `transactions.amount_tendered` | Cash only |
| Change | `transactions.change_amount` | Cash only |
| Payment reference | `transactions.payment_reference` | Digital only |
| Reprint indicator | Runtime flag | Added if this is a re-print |

---

## 8.5 Printer Integration

- Printer communication uses **ESC/POS** protocol (industry standard for thermal printers).
- Printer is configured per terminal (operator workstation) in the desktop app settings.
- Configuration:
  - Connection type: USB / Serial / Network (TCP/IP)
  - Paper width: 58mm / 80mm
  - Character encoding: UTF-8
- Printer status is polled before printing; if unavailable, operator is warned.

### Supported Printer Models (Recommended)
Any ESC/POS-compatible thermal printer. Common examples:
- Epson TM-T82
- Epson TM-T88
- Xprinter XP-58 / XP-80
- BIXOLON SRP-350

---

## 8.6 Voided Transaction Receipt

- If a transaction is voided after a receipt was printed, **no automatic reprint** is triggered.
- A manager or operator can manually print a **VOID NOTICE** from the transaction detail view.
- The void notice uses the same receipt format but with a `** VOID **` header and the void reason printed.

---

## 8.7 Offline Receipt Printing

When the desktop app is in offline mode:
- Receipts can still be printed using locally cached rate and session data.
- Receipt numbers in offline mode use a temporary format: `[LOCATION_CODE]-OFFLINE-[LOCAL_SEQ]`.
- Once the session syncs to the backend, the receipt number is updated to the official format.
- If the driver needs an official receipt for a synced transaction, the operator can re-print from the session detail view.

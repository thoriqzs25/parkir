# Chapter 15 — Out of Scope (v1)

## 15.1 Purpose

This chapter explicitly documents features and capabilities that are **not** part of the v1 implementation. These boundaries exist to keep the initial build focused and deliverable. Items listed here are candidates for future versions.

---

## 15.2 Out of Scope Items

### 15.2.1 Parking Slot-Level Tracking

**What it means:** Tracking which specific numbered bay or slot a vehicle occupies (e.g. "Vehicle B 1234 XYZ is in Slot B-12").

**Why excluded:** Adds significant complexity in data modeling, UI, and operational workflow. The system tracks occupancy counts per vehicle type — not per individual slot.

**Future path:** Can be introduced as a slot assignment feature in v2, potentially with QR-code-based slot identification.

---

### 15.2.2 Monthly Subscription / Pass-Based Billing

**What it means:** Recurring monthly fees for registered subscribers who can park for a flat monthly rate.

**Why excluded:** Requires a subscriber management module, recurring billing logic, pass validation at check-in, and integration with a payment processing gateway for scheduled charges.

**Future path:** Design as a separate billing plan type in the rate configuration module, with a linked subscriber registry.

---

### 15.2.3 Driver-Facing Portal or Mobile App

**What it means:** An interface for drivers themselves — to check occupancy before arriving, view their parking history, pay digitally via their own phone, or receive digital receipts.

**Why excluded:** Different UX, different security model, and a much larger scope than the admin/operator system.

**Future path:** A separate driver-facing app or web portal, integrated with the same backend API.

---

### 15.2.4 Access Control Hardware Integration

**What it means:** Automated gate/barrier control — the system sending a signal to open or close physical entry/exit barriers based on check-in or payment status.

**Why excluded:** Requires hardware integration layer (serial/TCP), vendor-specific protocols, and real-time reliability guarantees beyond typical web API behavior.

**Future path:** Introduce a hardware integration service as a separate module, communicating via webhook or direct TCP connection to gate controllers.

---

### 15.2.5 Push Notifications via Email, SMS, or WhatsApp

**What it means:** Sending incident alerts, revenue summaries, or system health notifications to users via external channels.

**Why excluded:** Requires third-party integrations (SMTP, Twilio, WhatsApp Business API), additional user preference management, and delivery reliability infrastructure.

**Future path:** Add a notification preferences table and integrate with a notification delivery service (e.g. Firebase, Twilio, or a custom SMTP setup).

---

### 15.2.6 Mixed Payment (Partial Cash + Partial Digital)

**What it means:** Splitting a single parking fee between cash and digital payment (e.g. pay Rp 15,000 in cash and Rp 5,000 via QRIS).

**Why excluded:** Complicates the transaction model, receipt layout, and reconciliation logic. Low priority for initial operational needs.

**Future path:** Extend the `transactions` table to support multiple payment line items.

---

### 15.2.7 Multi-Currency Support

**What it means:** Operating in multiple currencies across different locations.

**Why excluded:** Single-currency operation is assumed for v1. All amounts are in one currency (configurable at system setup, but not changeable per transaction).

**Future path:** Add a `currency` field to locations and transactions; update all financial calculations and reporting accordingly.

---

### 15.2.8 EV Charging Slot Management

**What it means:** Tracking charging stations, managing EV-specific fees (e.g. per kWh or per minute of charging), or reserving charging slots.

**Why excluded:** Requires a separate resource (charger) management layer beyond parking sessions.

**Future path:** Introduce as a resource type alongside parking slots, with its own fee model.

---

### 15.2.9 Self-Service Payment Kiosks

**What it means:** Standalone kiosk machines that allow drivers to pay without an operator present.

**Why excluded:** Requires hardware integration, kiosk-specific UI, and change-dispensing logic.

**Future path:** A kiosk mode could be built on the same backend API with a simplified, touch-friendly frontend.

---

### 15.2.10 Automated Digital Payment Gateway Callbacks

**What it means:** The system automatically confirming a digital payment when the gateway sends a webhook (instead of requiring the operator to manually confirm).

**Why excluded:** Requires a registered webhook endpoint, gateway API credentials, signature verification, and idempotent payment handling.

**Future path:** Add a `/webhooks/payment` endpoint per gateway, process callbacks asynchronously, and auto-advance sessions from `PENDING_PAYMENT` to `CLOSED`.

---

## 15.3 Deferral vs. Never

| Item | Classification |
|------|---------------|
| Parking slot-level tracking | Deferred to v2 |
| Monthly subscription billing | Deferred to v2 |
| Driver-facing portal | Deferred to v2 |
| Hardware gate integration | Deferred to v2 |
| Push notifications | Deferred to v2 |
| Mixed payment | Deferred to v2 |
| Multi-currency | Deferred, low priority |
| EV charging | Deferred, low priority |
| Self-service kiosks | Deferred, dependent on hardware |
| Automated payment callbacks | Deferred to v1.1 (high value, moderate complexity) |

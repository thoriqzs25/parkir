# Chapter 15 — Out of Scope (v2)

## 15.1 Purpose

This chapter explicitly documents features and capabilities that are **not** part of the v2 implementation. These boundaries exist to keep the build focused and deliverable. Items listed here are candidates for future versions.

---

## 15.2 Out of Scope Items

### 15.2.1 License Plate Recognition (LPR/ANPR)

**What it means:** Automatic license plate recognition using cameras at entry/exit gates.

**Why excluded:** Adds significant hardware cost (LPR cameras), complexity (plate recognition algorithms), and integration work. QR tickets are sufficient for session tracking.

**Future path:** Can be added in v3 for security/audit purposes. Would require LPR camera integration, plate validation logic, and plate-to-session matching.

---

### 15.2.2 Mobile App for Drivers

**What it means:** Driver-facing mobile app for reservations, payments, digital receipts, parking history.

**Why excluded:** Different UX, different security model, and much larger scope. v2 focuses on automated gate system for AMB.

**Future path:** Separate driver-facing app or web portal in v3, integrated with same backend API.

---

### 15.2.3 Monthly Subscription / Pass-Based Billing

**What it means:** Recurring monthly fees for registered subscribers who can park for a flat monthly rate.

**Why excluded:** Requires subscriber management module, recurring billing logic, pass validation at entry, and integration with payment gateway for scheduled charges. AMB's current model is pay-per-use.

**Future path:** Design as separate billing plan type in rate configuration module, with linked subscriber registry in v3.

---

### 15.2.4 Parking Slot-Level Tracking

**What it means:** Tracking which specific numbered bay or slot a vehicle occupies (e.g., "Vehicle is in Slot B-12").

**Why excluded:** Adds complexity in data modeling, UI, and operational workflow. System tracks occupancy counts per vehicle type — not per individual slot. No operators to manage slots.

**Future path:** Can be introduced in v3 with slot assignment feature, potentially with QR-code-based slot identification.

---

### 15.2.5 Multi-Tenant Support

**What it means:** Supporting multiple parking companies (tenants) on same system instance.

**Why excluded:** v2 is built for AMB (single tenant). Multi-tenancy requires tenant isolation, separate billing per tenant, tenant-specific configs, and more complex data model.

**Future path:** Add tenant_id to all entities, implement tenant isolation, build tenant management dashboard in v3.

---

### 15.2.6 Cash Payments

**What it means:** Accepting cash payments at automated gates.

**Why excluded:** Fully automated gates cannot handle cash (no operator to collect/change). Cash requires human intervention. v2 is cashless (e-money/Flazz only).

**Future path:** Could add cash acceptor hardware in v3 (complex, requires change dispensing, security).

---

### 15.2.7 Mixed Payment (Partial E-Money + Partial Other)

**What it means:** Splitting a single parking fee between multiple payment methods.

**Why excluded:** Complicates transaction model, receipt layout, and reconciliation logic. Low priority for automated system.

**Future path:** Extend transactions table to support multiple payment line items in v3.

---

### 15.2.8 Multi-Currency Support

**What it means:** Operating in multiple currencies across different locations.

**Why excluded:** Single-currency operation (IDR) assumed for v2. All amounts in Indonesian Rupiah.

**Future path:** Add currency field to locations and transactions in v3; update all financial calculations and reporting.

---

### 15.2.9 EV Charging Slot Management

**What it means:** Tracking charging stations, managing EV-specific fees (per kWh or per minute), reserving charging slots.

**Why excluded:** Requires separate resource (charger) management layer beyond parking sessions. Not needed for AMB.

**Future path:** Introduce as resource type alongside parking sessions in v3, with its own fee model.

---

### 15.2.10 Push Notifications via Email, SMS, or WhatsApp

**What it means:** Sending incident alerts, revenue summaries, or system health notifications to users via external channels.

**Why excluded:** v2 uses internal alerting (audio for staff, email/Telegram for developers). No driver notifications needed (automated system).

**Future path:** Add notification preferences table and integrate with notification delivery service in v3.

---

## 15.3 In Scope for v2 (Was Out of Scope in v1)

The following were out of scope in v1 but are **in scope** for v2:

| Feature | v1 Status | v2 Status |
|---------|-----------|-----------|
| Hardware gate integration | Out of scope | **In scope** (fully automated) |
| Self-service kiosks | Out of scope | **In scope** (automated gates) |
| Automated payment callbacks | Out of scope | **In scope** (payment vendor integration) |
| Offline operation | Limited (desktop app) | **In scope** (full local DB, server room app) |
| Multi-location management | Basic | **In scope** (20+ locations, city grouping) |

---

## 15.4 Deferral vs. Never

| Item | Classification |
|------|---------------|
| License plate recognition | Deferred to v3 |
| Mobile app for drivers | Deferred to v3 |
| Monthly subscription billing | Deferred to v3 |
| Parking slot-level tracking | Deferred to v3 |
| Multi-tenant support | Deferred to v3 |
| Cash payments | Never (automated system) |
| Mixed payment | Deferred to v3 |
| Multi-currency | Deferred, low priority |
| EV charging | Deferred, low priority |
| Push notifications | Deferred to v3 |

---

*End of Chapter 15 — Out of Scope (v2)*

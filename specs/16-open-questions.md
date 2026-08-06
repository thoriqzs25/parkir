# Chapter 16 — Open Questions (v2)

## 16.1 Purpose

This chapter documents decisions that have not yet been made and that will affect implementation scope, architecture, or design. Each question includes its impact and suggested options to consider.

---

## 16.2 Open Questions

---

### Q1 — Which payment vendor/terminal should be integrated for e-money/Flazz?

**Impact:** Payment terminal hardware, communication protocol, signal format, integration complexity.

**Context:** AMB will provide payment vendor. PARKIR needs to integrate with vendor's terminal to receive payment completion signal.

**Options:**
| Option | Notes |
|--------|-------|
| **Vendor TBD** | AMB to provide vendor details |
| **Multiple vendors** | Support multiple payment terminals (more flexibility, more integration work) |

**Decision needed:** Payment terminal model, communication protocol (serial/USB/network), signal format (HTTP callback/webhook/shared file), error handling.

**Decision needed before:** Payment module implementation (Week 4).

---

### Q2 — What is the exact hardware specification for gate components?

**Impact:** Hardware abstraction layer (HAL), printer drivers, scanner integration, gate motor control.

**Context:** AMB will provide hardware (Epson printer, Panda PRJ-777 scanner, gate motor, sensors). Need exact specs for integration.

**Questions to answer:**
- Epson printer model? Communication protocol (USB/serial)?
- Panda PRJ-777 scanner communication protocol (USB HID/serial)?
- Gate motor control protocol (via hub)?
- Sensor types (loop sensor, button) and communication protocol?
- Hub specifications (connects all hardware)?

**Decision needed before:** Gate app hardware integration (Week 3-4).

---

### Q3 — What is the communication protocol between gate app and server room app?

**Impact:** Network architecture, latency, reliability, auto-recovery.

**Options:**
| Option | Pros | Cons |
|--------|------|------|
| **HTTP + mDNS** | Simple, auto-discovery, easy debugging | Request/response only (no push) |
| **WebSocket** | Real-time bidirectional, persistent connection | More complex, connection management |
| **MQTT** | Pub/sub, lightweight, IoT-friendly | Overkill for 2-10 gates, broker dependency |

**Recommendation:** HTTP + mDNS (simpler, auto-recovery via mDNS, sufficient for command/response pattern).

**Decision needed before:** Gate app + server room app implementation (Week 3).

---

### Q4 — What is the exact format for QR code on entry ticket?

**Impact:** QR code generation, scanner parsing, session validation, security.

**Options:**
| Option | Pros | Cons |
|--------|------|------|
| **JSON** | Flexible, extensible, easy to parse | Larger QR code (more data) |
| **Compact binary** | Small QR code, fast scanning | Harder to debug, less flexible |
| **Encrypted JSON** | Tamper-resistant | More complex, slower |
| **Signed JSON** | Tamper-evident | More complex |

**Recommendation:** JSON (simple, flexible, sufficient for now). Can add encryption/signing in v3 if needed.

**Decision needed before:** Entry flow implementation (Week 3).

---

### Q5 — How should config versioning be implemented?

**Impact:** Audit trail, offline operation, config sync, data model.

**Options:**
| Option | Notes |
|--------|-------|
| **Auto-increment integer** | Simple (v1, v2, v3, ...) |
| **UUID** | Globally unique, no sequencing |
| **Timestamp-based** | v20260806103000 (sortable, unique) |
| **Semantic versioning** | v1.0.0, v1.1.0 (overkill for configs) |

**Recommendation:** Auto-increment integer per config type (rate_v1, rate_v2, shift_v1, shift_v2, ...).

**Decision needed before:** Config management implementation (Week 2).

---

### Q6 — What is the offline SOP (Standard Operating Procedure) for staff?

**Impact:** Incident handling, exception resolution, staff training, system design.

**Context:** When system fails (hardware, payment, internet), staff runs offline SOP. Need to understand what staff does.

**Questions to answer:**
- How does staff manually open gate (if gate motor fails)?
- How does staff record transaction (if system down)?
- How does staff handle payment (if payment terminal down)?
- How does staff reconcile offline transactions later?

**Decision needed:** Document offline SOP (AMB to provide). System should support offline transaction recording.

**Decision needed before:** Incident management implementation (Week 5).

---

### Q7 — What is the deployment strategy for 20+ locations?

**Impact:** Infrastructure, deployment automation, monitoring, support.

**Questions to answer:**
- Cloud provider (AWS, GCP, DigitalOcean, local Indonesian provider)?
- Database hosting (managed RDS or self-hosted)?
- Deployment automation (CI/CD, USB deployment)?
- Monitoring stack (Loki, Prometheus, Grafana)?
- Support model (who handles hardware issues, system issues)?

**Recommendation:** 
- Cloud: DigitalOcean or AWS (Jakarta region).
- Database: Managed PostgreSQL (RDS).
- Deployment: CI/CD for cloud, USB for server room/gate apps.
- Monitoring: Loki + Prometheus + Grafana (self-hosted).

**Decision needed before:** Infrastructure setup (Week 1-2).

---

### Q8 — What are the exact shift config requirements for AMB?

**Impact:** Shift management, reporting, reconciliation.

**Context:** v2 uses continuous shift numbers (shift_1, shift_2, ...). Need to confirm AMB's shift requirements.

**Questions to answer:**
- How many shifts per day? (Default: 3)
- What are shift time windows? (e.g., 06:00-14:00, 14:00-22:00, 22:00-06:00)
- Do shifts vary by location? (All same, or location-specific?)
- How are shifts used for reporting? (Revenue by shift, reconciliation?)

**Decision needed:** Confirm with AMB operations team.

**Decision needed before:** Shift management implementation (Week 2).

---

### Q9 — What are the performance requirements for automated entry/exit?

**Impact:** Hardware selection, network architecture, system design.

**Questions to answer:**
- Target entry flow time? (Button press to gate open)
- Target exit flow time? (QR scan to gate open)
- Target payment processing time? (Card tap to gate open)
- Peak throughput? (Vehicles per hour per gate)

**Recommendation:**
- Entry: < 5 seconds.
- Exit: < 10 seconds (including payment).
- Payment: < 3 seconds.
- Throughput: 100-200 vehicles/hour/gate.

**Decision needed:** Confirm with AMB operations team.

**Decision needed before:** Hardware integration (Week 3-4).

---

### Q10 — What are the security requirements?

**Impact:** Authentication, authorization, data encryption, audit logging.

**Questions to answer:**
- Dashboard authentication (JWT, session management)?
- Gate app authentication (none, LAN-only)?
- Server room app authentication (API key to cloud)?
- Data encryption (at rest, in transit)?
- Audit log retention (2 years minimum)?
- Compliance requirements (data privacy, financial regulations)?

**Recommendation:**
- Dashboard: JWT RS256, httpOnly cookie.
- Gate app: No auth (LAN-only, trusted).
- Server room app: API key (to cloud).
- Encryption: TLS in transit, encrypted DB at rest.
- Audit: 2-year retention, immutable.

**Decision needed:** Confirm with AMB IT/security team.

**Decision needed before:** Security implementation (Week 7-8).

---

## 16.3 Decision Log

Track answers here as decisions are made.

| # | Question | Decision | Decided By | Date |
|---|---------|---------|-----------|------|
| Q1 | Payment vendor | TBD (AMB to provide) | | |
| Q2 | Hardware specs | Epson printer, Panda PRJ-777 scanner (details TBD) | | |
| Q3 | Gate ↔ server room protocol | HTTP + mDNS | (spec author) | 2026-08-06 |
| Q4 | QR code format | JSON | (spec author) | 2026-08-06 |
| Q5 | Config versioning | Auto-increment integer (rate_v1, shift_v1, ...) | (spec author) | 2026-08-06 |
| Q6 | Offline SOP | TBD (AMB to document) | | |
| Q7 | Deployment strategy | DigitalOcean/AWS + managed PostgreSQL + USB deployment | (spec author) | 2026-08-06 |
| Q8 | Shift config | 3 shifts/day, continuous numbering (TBD with AMB) | | |
| Q9 | Performance requirements | Entry < 5s, Exit < 10s, Payment < 3s | (spec author) | 2026-08-06 |
| Q10 | Security requirements | JWT for dashboard, API key for server room, LAN-only for gate | (spec author) | 2026-08-06 |

---

*End of Chapter 16 — Open Questions (v2)*

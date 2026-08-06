# Chapter 2 — System Goals

## 2.1 Goal Summary

PARKIR v2 is built to replace AMB's expensive third-party automated parking system with a cost-effective, fully automated solution. The system is built around five primary goals.

| # | Goal | Primary Users | Key Modules |
|---|------|--------------|-------------|
| G1 | Automated Entry/Exit | Drivers (indirect) | Gate Apps, Sessions |
| G2 | Tap-to-Go Payment | Drivers (indirect) | Payments, Rates |
| G3 | Multi-Location Management | AMB Admins, Managers | Dashboard, Reports |
| G4 | Exception Handling | Server Room Staff | Incidents, Alerts |
| G5 | System Observability | Developers, Admins | Health Monitor, Audit Log, Alerts |

---

## 2.2 G1 — Automated Entry/Exit

### Intent
Provide fully automated, reliable, and fast entry/exit flows for drivers. No human operators at gates.

### Functional Requirements
- System must dispense QR-coded ticket on entry (thermal printer).
- System must read QR ticket on exit (fixed-mount scanner).
- System must calculate fee automatically (server room app).
- System must open gate automatically after payment.
- Entry flow must complete in < 5 seconds (button press to gate open).
- Exit flow must complete in < 10 seconds (QR scan to gate open).
- System must work offline (server room app has local DB).
- System must handle hardware failures gracefully (gate stays closed, alert triggered).

### Acceptance Criteria
- [ ] Vehicle enters loop → button press → ticket dispensed → gate opens (< 5s).
- [ ] Driver scans QR → fee displayed → payment processed → gate opens (< 10s).
- [ ] System continues operating without internet (offline mode).
- [ ] Hardware failure → gate stays closed → alert triggered.
- [ ] 95% uptime (hardware-dependent).

### Success Metrics
- Entry flow time: < 5 seconds (p95).
- Exit flow time: < 10 seconds (p95).
- Hardware uptime: 95%.
- Zero operator intervention for normal flows.

---

## 2.3 G2 — Tap-to-Go Payment

### Intent
Provide fast, cashless payment via e-money/Flazz cards. No cash handling.

### Functional Requirements
- Payment must be tap-to-go only (e-money, Flazz).
- Payment terminal provided by third-party vendor.
- PARKIR receives payment completion signal from vendor.
- Fee calculated using local rate config (server room app).
- Rate config versioned (audit trail).
- Transaction record created for every closed session.
- Voided transactions excluded from revenue totals.

### Acceptance Criteria
- [ ] Driver taps card → payment processed → gate opens (< 3s).
- [ ] Payment failure → gate stays closed → error displayed.
- [ ] Transaction recorded with config version (audit trail).
- [ ] Voided transactions excluded from reports.

### Success Metrics
- Payment processing time: < 3 seconds (p95).
- Payment success rate: > 99%.
- Zero cash handling.

---

## 2.4 G3 — Multi-Location Management

### Intent
Provide central management for AMB's 20+ parking locations. Grouped by city. Real-time monitoring and reporting.

### Functional Requirements
- Dashboard must support 20+ locations.
- Locations grouped by city (filterable).
- Real-time gate status monitoring (online/offline/busy).
- Revenue reports (per location + total).
- Occupancy reports (per location).
- Transaction reports (with filters).
- Export to CSV/Excel.
- Gate configuration via dashboard (vehicle type, gate type).
- Rate configuration per location (versioned).
- User management (staff, managers, admins, owners).

### Acceptance Criteria
- [ ] Dashboard loads all 20+ locations (< 2s).
- [ ] Real-time gate status updates (< 30s latency).
- [ ] Revenue reports accurate (match transaction records).
- [ ] Export to CSV works for all reports.
- [ ] Gate configuration flows to server room app → gate app.

### Success Metrics
- Dashboard page load time: < 2 seconds.
- Report generation time: < 5 seconds.
- Data accuracy: 100% (reports match transactions).
- User satisfaction: > 80%.

---

## 2.5 G4 — Exception Handling

### Intent
Enable server room staff to monitor gates and handle exceptions (hardware failures, payment issues) quickly.

### Functional Requirements
- Server room app monitors all gates at location (2-10 gates).
- Audio alert when gate needs assistance ("Gate {gate_id} needs assistance").
- Visual alert on staff monitor (gate status grid).
- Staff walks to gate, handles exception (offline SOP if needed).
- Incident logged (hardware failure, payment failure, QR unreadable).
- Incident resolution tracked.

### Acceptance Criteria
- [ ] Gate failure → audio alert within 5 seconds.
- [ ] Staff responds to alert (< 2 minutes).
- [ ] Incident logged with gate_id, type, timestamp.
- [ ] Incident resolution tracked (open → resolved).

### Success Metrics
- Alert response time: < 2 minutes.
- Incident resolution time: < 30 minutes.
- Staff satisfaction: > 80%.

---

## 2.6 G5 — System Observability

### Intent
Provide comprehensive monitoring, logging, and alerting for developers and admins.

### Functional Requirements
- Centralized logging (Loki) for all components.
- Metrics collection (Prometheus) for all components.
- Visualization (Grafana) for system health.
- Alert routing: email/Telegram for critical alerts, audio for gate alerts.
- Audit log (immutable, 2-year retention).
- Health checks: gate apps (15s), server room apps (20s), cloud backend (15s).

### Acceptance Criteria
- [ ] All logs centralized in Loki (searchable).
- [ ] Metrics visible in Grafana (dashboards).
- [ ] Critical alerts sent to email + Telegram (< 1 min).
- [ ] Gate alerts played as audio (< 5s).
- [ ] Audit log immutable (no deletes/updates).

### Success Metrics
- Log retention: 30 days (Loki).
- Alert delivery time: < 1 minute.
- System uptime: 99.5% (cloud backend).
- Mean time to detect (MTTD): < 5 minutes.

---

## 2.7 Business Goals

### Cost Reduction
- Replace expensive third-party system.
- Per-location monthly fee model.
- Target: 50% cost reduction vs current system.

### Scalability
- Support 20+ locations (AMB).
- Support 2-10 gates per location.
- Support future expansion (100+ locations).

### Reliability
- 95% uptime (hardware-dependent).
- Offline operation (gates work without internet).
- Fast recovery (replacement mini PCs via USB).

### Maintainability
- Stateless gate apps (easy to replace).
- Local DB in server room (easy to backup/restore).
- USB deployment (no complex installation).

---

## 2.8 Non-Goals (v2)

**Not in scope for v2:**
- License plate recognition (LPR/ANPR).
- Mobile app for drivers.
- Monthly subscriptions or passes.
- Individual parking bay tracking.
- Multi-tenant (multiple companies).
- Cash payments.
- Operator-assisted flows.

**May be in scope for v3:**
- License plate recognition (security/audit).
- Mobile app (driver notifications, reservations).
- Monthly subscriptions (regular parkers).
- Multi-tenant (sell to other companies).

---

*End of Chapter 2 — System Goals (v2)*

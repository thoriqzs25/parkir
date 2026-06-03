# Chapter 16 — Open Questions

## 16.1 Purpose

This chapter documents decisions that have not yet been made and that will affect implementation scope, architecture, or design. Each question includes its impact and suggested options to consider.

---

## 16.2 Open Questions

---

### Q1 — Which digital payment gateway(s) should be integrated?

**Impact:** Backend payment confirmation flow, transaction record fields, operator UX for digital payments, reconciliation reports.

**Options:**
| Option | Notes |
|--------|-------|
| **Midtrans** | Popular in Indonesia; supports QRIS, GoPay, OVO, bank transfer; good docs |
| **Xendit** | Strong API, supports QRIS, e-wallets, virtual accounts |
| **QRIS direct** | Standardized QR code; requires Bank Indonesia registration; gateway-agnostic |
| **Multiple gateways** | More flexibility but more integration work |

**Recommendation:** Start with one gateway (Midtrans or Xendit) for v1. Add automated callback handling in v1.1.

**Decision needed before:** Payment module implementation.

---

### Q2 — Should the operator desktop app be Electron-based or a native Windows app?

**Impact:** Tech stack, development effort, update delivery, offline capabilities, printer integration.

**Options:**
| Option | Pros | Cons |
|--------|------|------|
| **Electron** | Web tech stack (reuse with dashboard), easy updates, cross-platform | Higher memory usage, larger bundle |
| **Native Windows (C# / WinForms / WPF)** | Faster, tighter printer integration | Separate codebase from web |
| **Progressive Web App (PWA)** | No install needed, easy updates | Printer access limited without native bridge |

**Recommendation:** Electron — allows sharing UI components with the web dashboard and simplifies team skillset requirements.

**Decision needed before:** Desktop app architecture setup.

---

### Q3 — What is the expected peak load per location?

**Impact:** Database sizing, API concurrency design, rate limiting, infrastructure capacity planning.

**Questions to answer:**
- How many vehicles enter/exit per hour at the busiest location?
- How many operator terminals run simultaneously per location?
- How many locations will be onboarded in year 1?

**Why it matters:** A location handling 500 vehicles/hour is very different from one handling 50. This affects connection pool sizing, caching strategy, and whether a simple single-server setup suffices or whether horizontal scaling is needed from the start.

---

### Q4 — Is there an operator shift system?

**Impact:** Reports, cash reconciliation, operator activity attribution, session ownership.

**What this means:** In many parking operations, operators have defined shifts (e.g. 07:00–15:00 and 15:00–23:00). At shift end, a cash handover happens — the operator hands collected cash to a supervisor.

**If shift tracking is added:**
- Need: shift_start, shift_end, total_cash_expected, total_cash_handed_over.
- Reports become shift-scoped instead of just date-scoped.
- A discrepancy report (expected vs actual cash) becomes possible.

**Options:**
| Option | Notes |
|--------|-------|
| No shifts (v1) | Operators just log in/out; reports are date-based |
| Basic shift tracking | Log shift start/end; no cash reconciliation |
| Full shift + cash handover | Full reconciliation report per shift |

**Recommendation:** Defer to v2 unless the client's operations depend on cash shift reconciliation.

---

### Q5 — Should rate configurations support time-of-day pricing?

**Impact:** Rate data model, fee calculation logic, rate configuration UI.

**What this means:** Different rates for different times of day — e.g.:
- 06:00–22:00: Rp 5,000/hour
- 22:00–06:00 (overnight): Rp 3,000/hour

**Complexity added:**
- Rate lookup must consider time of day, not just date.
- Sessions spanning two rate windows need pro-rated calculation.
- Rate configuration UI becomes significantly more complex.

**Options:**
| Option | Notes |
|--------|-------|
| **Flat daily rate only (v1)** | Single hourly rate per vehicle type per location; simple |
| **Time-of-day rates** | More accurate billing but much higher complexity |

**Recommendation:** Defer to v2. Design the rate schema to be extensible (the `location_rates` table can gain `time_from` / `time_until` columns later without breaking v1 data).

---

### Q6 — How should sync conflicts in offline mode be resolved?

**Impact:** Offline sync logic, conflict resolution UI for managers, data integrity.

**Scenario:** Operator A checks in plate "B 1234 XYZ" offline. Meanwhile, Operator B (at the same location, online) also checks in "B 1234 XYZ". When Operator A reconnects, the system detects a conflict.

**Options:**
| Option | Notes |
|--------|-------|
| **Flag and defer to manager** | Conflict shown in dashboard; manager decides which session is correct |
| **Last-write-wins** | Whichever syncs last becomes the active session |
| **First-write-wins** | Online session takes precedence; offline session is auto-voided |

**Recommendation:** Flag and defer to manager (specified in Chapter 11). Confirm this is acceptable operationally.

---

### Q7 — What is the deployment environment?

**Impact:** Infrastructure decisions, database hosting, CI/CD, backup strategy, security requirements.

**Questions to answer:**
- Cloud (AWS / GCP / Azure) or on-premise?
- Managed database (RDS, Cloud SQL) or self-hosted PostgreSQL?
- Single region or multi-region?
- Is there an existing IT team to manage servers, or does this need to be fully managed?

**Recommendation:** Start with a single-region cloud deployment (e.g. AWS ap-southeast-1 for Indonesia). Use managed RDS for PostgreSQL to reduce operational burden.

---

## 16.3 Decision Log

Track answers here as decisions are made.

| # | Question | Decision | Decided By | Date |
|---|---------|---------|-----------|------|
| Q1 | Payment gateway | TBD | | |
| Q2 | Desktop app tech | Electron (web-based desktop) | | |
| Q3 | Peak load per location | 500–3,000 vehicles/day | | |
| Q4 | Shift system | Basic shift tracking (start/end) | | |
| Q5 | Time-of-day pricing | Tiered rate: first hour different from subsequent hours | | |
| Q6 | Offline conflict resolution | Flag + manager review | (spec author) | |
| Q7 | Deployment environment | TBD | | |

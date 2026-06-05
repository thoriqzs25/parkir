# Chapter 18 — System Architecture & Connectivity

## 18.1 Overview

This chapter defines the end-to-end system architecture, physical deployment topology, and component-level connectivity flows for the Parking Administration System (PAS). It covers every node in the data path from the vehicle arriving at the gate to the transaction being persisted in the database and the receipt being printed.

All connectivity described here is required for production deployment, including the barrier gate integration which extends the v1 scope into physical access control.

---

## 18.2 Component Catalog

### 18.2.1 Facility-Side Components (Booth / Parking Location)

| Component | Role | Platform / Protocol | Resilience Requirement |
|-----------|------|---------------------|------------------------|
| **Operator App** | Primary user interface for check-in, check-out, payment, and incident reporting. | Electron + React on Windows 10/11. | Must function in offline mode without backend connectivity. |
| **Local SQLite** | Embedded database inside the Operator App for offline data caching. | SQLite file on local disk. | Survives app restarts; syncs on reconnect. |
| **Thermal Printer** | Receipt printer connected to the Operator PC. | USB or Serial (RS-232). Direct driver call from Electron. | Works offline; no backend dependency. |
| **Gate Controller** | Dedicated edge device that physically opens and closes the barrier gate. | Raspberry Pi or industrial MCU on local LAN. HTTP or WebSocket from Operator App. | Must operate independently of the Operator PC and internet. |
| **Loop Sensor** | Vehicle detection sensor embedded in the road surface at gate entry/exit. | Wired digital input to Gate Controller. | Hardwired safety interlock; no software dependency. |
| **Gate Motor** | Electromechanical actuator that raises and lowers the barrier arm. | Relay or motor driver controlled by Gate Controller. | Failsafe close on power loss (configurable). |

### 18.2.2 Cloud / Data Center Components

| Component | Role | Platform / Protocol | Resilience Requirement |
|-----------|------|---------------------|------------------------|
| **Backend API** | Core application server exposing REST/gRPC endpoints for auth, sessions, payments, rates, and configuration. | Containerized service (e.g., Node.js / Python / Go). HTTPS externally, HTTP internally behind load balancer. | Horizontally scalable; stateless. |
| **Sync Service** | Reconciliation worker that accepts offline batches from Operator Apps and resolves conflicts before writing to PostgreSQL. | Background worker or queue consumer (e.g., Bull / Celery / custom). | Idempotent; retries with exponential backoff. |
| **PostgreSQL** | Primary transactional database. | Managed PostgreSQL with read replicas. | Automated backups; point-in-time recovery. |
| **Payment Gateway** | External digital payment provider (e.g., Midtrans, Xendit, Stripe). | HTTPS REST with webhook callbacks. | Provider SLA; idempotency keys on charges. |
| **Observability** | Aggregate subsystem for audit logging, health monitoring, and anomaly alerting. | Ingests events from Backend API; writes to PostgreSQL audit tables and alert queues. | Append-only audit logs; alerts are best-effort. |

### 18.2.3 Client Components

| Component | Role | Platform / Protocol |
|-----------|------|---------------------|
| **Manager Browser** | Web dashboard for managers and administrators. | Chrome / Firefox / Edge. React SPA. JWT authentication over HTTPS. |

---

## 18.3 Network Topology

```
┌──────────────────────────────────────────────────────────────────────────────────────────┐
│                         Parking Facility (Booth LAN)                              │
│                                                                                  │
│   ┌───────────────┐       ┌───────────────┐                                       │
│   │ Operator App    │       │ Gate Controller   │                                       │
│   │ (Windows PC)    │       │ (Pi / MCU)      │                                       │
│   ├───────────────┤       ├───────────────┤                                       │
│   │ • Electron      │       │ • HTTP/WebSocket│                                       │
│   │ • Local SQLite  │       │ • Safety Logic  │                                       │
│   │ • Thermal Print │       │ • Whitelist     │                                       │
│   └───────────────┘       └───────────────┘                                       │
│          │                       │                                               │
│          │ USB / Serial        │ Wired I/O                                       │
│          │                       │                                               │
│   ┌───────────────┐       ┌───────────────┐                                       │
│   │ Thermal Printer │       │ Loop Sensor       │                                       │
│   │ USB / Serial    │       │ Inductive / Safety│                                       │
│   └───────────────┘       └───────────────┘                                       │
│                                                                                  │
│   ┌───────────────┐                                                 LAN          │
│   │ Manager Browser │                              ┌───────────────┐       │
│   │ (React SPA)     │                              │ Router / Switch │       │
│   └───────────────┘                              └───────────────┘       │
│          │                                                    │                 │
│          └───────────────────────────────────────────────────┼──────────────────────────────────┼────────────────┘                 │
│            Internet / VPN (HTTPS)                         │                    │
│                                                           │                    │
└─────────────────────────────────────────────────────────────────────────────────┴──────────────────────────────────────────────────────────────────┘
                                                  │
                                                  │
                                            ┌───────────────┐
                                            │ Cloud / Data Center │
                                            ├───────────────┤
                                            │ • Backend API       │
                                            │ • Sync Service      │
                                            │ • PostgreSQL        │
                                            │ • Observability     │
                                            └───────────────┘
                                                  │
                                            ┌───────────────┐
                                            │ Payment Gateway     │
                                            │ (External Provider) │
                                            └───────────────┘
```

---

## 18.4 Data Flows

### 18.4.1 Flow A — Online Vehicle Entry

```
Loop Sensor ─detect─→ Gate Controller (ready state)
     Operator taps [Check In] → Operator App validates plate + vehicle type
          Operator App ─POST /sessions─→ Backend API
               Backend API ─INSERT─→ PostgreSQL
          Backend API ─session_created─→ Operator App
     Operator App ─POST /gate/open─→ Gate Controller (LAN)
          Gate Controller ─relay_on─→ Gate Motor (barrier raises)
               Loop Sensor ─vehicle_passed─→ Gate Controller
          Gate Controller ─relay_off + auto_close_timer─→ Gate Motor
     Operator App ─thermal_print─→ Thermal Printer (entry ticket, optional)
```

**Offline variant:** If the internet is down, the Operator App writes the session to Local SQLite, immediately signals the Gate Controller over LAN to open the gate, and queues the session for background sync. The thermal printer still fires.

---

### 18.4.2 Flow B — Online Vehicle Exit (Payment + Gate)

```
Operator searches plate / session ID → Operator App fetches from Backend API
     Backend API ─SELECT─→ PostgreSQL ─session─→ Backend API ─→ Operator App
          Operator App calculates fee from rate_snapshot + duration
     Operator selects payment method:
          CASH: Operator records amount tendered; change computed locally.
          DIGITAL: Operator App ─charge_request─→ Backend API ─→ Payment Gateway
               Payment Gateway ─callback─→ Backend API ─success─→ Operator App
     Operator taps [Confirm Payment]
          Operator App ─POST /transactions─→ Backend API ─INSERT─→ PostgreSQL
          Backend API ─close_session─→ PostgreSQL
     Operator App ─thermal_print─→ Thermal Printer (receipt)
     Operator App ─POST /gate/open─→ Gate Controller (LAN) ─→ Gate Motor
```

**Offline variant:** Cash payment is fully offline. Digital payment is deferred (flagged as pending). Session close and transaction are written to Local SQLite. Gate opens over LAN. On reconnect, Sync Service uploads the offline batch, processes deferred digital charges, and resolves any conflicts.

---

### 18.4.3 Flow C — Offline Sync Reconciliation

```
Operator App detects network recovery
     Operator App reads Local SQLite unsynced queue
          Operator App ─POST /sync/batch─→ Sync Service
               Sync Service validates each record:
                    • Deduplicate by client-generated UUID
                    • Check plate conflicts (duplicate active sessions)
                    • Verify rate_snapshot against current location_rates
               Sync Service ─INSERT/UPDATE─→ PostgreSQL
                    Sync Service ─POST /sync/result─→ Operator App
               Sync Service fires alerts for unresolvable conflicts
          Sync Service ─WRITE─→ Observability (audit trail)
     Operator App marks queue as synced; surfaces conflicts in UI
```

---

## 18.5 Gate Control Protocol

### 18.5.1 Interface

The Operator App communicates with the Gate Controller over the **booth LAN** using a lightweight protocol:

- **Primary:** HTTP REST (`POST /gate/{command}`) or WebSocket for persistent connection.
- **Fallback:** If the Gate Controller is configured with a local whitelist, it can validate plates and autonomously open without Operator App involvement.

### 18.5.2 Commands

| Command | Triggered By | Gate Controller Action |
|---------|-------------|----------------------|
| `open` | Operator App (after successful check-in or confirmed payment) | Energize relay for 3–5 seconds to raise barrier. Start auto-close timer. |
| `close` | Operator App (manual override) or auto-close timer expiry | De-energize relay; lower barrier. |
| `status` | Operator App (health check) | Return gate state: `OPEN`, `CLOSED`, `MOVING`, `FAULT`. |

### 18.5.3 Safety Interlocks (Controller-local)

The Gate Controller **must not** rely on the Operator App or backend for safety-critical logic:

1. **Loop Sensor Override:** If the loop sensor reports a vehicle present, the barrier must not close. The close command is held until the sensor clears.
2. **Timeout Auto-Close:** If the gate has been open for longer than the configured timeout (e.g., 30 seconds) and the loop sensor is clear, the controller closes the gate automatically.
3. **Failsafe Close:** On power loss to the controller, the gate motor defaults to a closed position (spring or gravity return, or NC relay configuration).
4. **Anti-Crush:** If a secondary safety sensor (e.g., photocell) is triggered during close, the controller must reverse and reopen immediately.

---

## 18.6 Failure Mode Matrix

| Failure | Impact | System Response |
|---------|--------|-----------------|
| Internet down (booth → cloud) | Operator App cannot reach Backend API. | App switches to offline mode. Local SQLite cache activates. Gate Controller still reachable over LAN. Cash payments accepted. |
| Backend API down | No cloud persistence; no digital payments. | Same as above. Sync queue accumulates. |
| Operator PC crash / reboot | App unavailable until restart. | Gate Controller operates autonomously via whitelist if configured. Otherwise, manual gate override key. |
| Gate Controller offline | Cannot open/close gate from app. | Operator uses physical override button or key switch. Incident reported. |
| Loop Sensor failure | Gate cannot safely detect vehicle presence. | Gate Controller enters `FAULT` state; refuses close commands. Operator manually manages traffic. Incident reported. |
| Thermal Printer failure | Cannot print receipts locally. | Operator notes printer fault. Receipt data is stored electronically (Backend or Local SQLite). Printer replaced during shift. |
| PostgreSQL outage | Backend cannot persist new data. | Backend returns 503. Operator App falls back to offline mode. Existing read-only queries may still work via cache. |
| Payment Gateway timeout | Digital payment hangs or fails. | Backend retries with idempotency key. If still failing, operator offers cash alternative or defers digital charge. |

---

## 18.7 Data Residency & Security

- **LAN traffic** (Operator App ↔ Gate Controller) is unencrypted HTTP by default. If the booth network is physically secured, this is acceptable. For hardened deployments, Gate Controller should support TLS or mTLS.
- **Internet traffic** (all cloud-facing) must use TLS 1.2+ with valid certificates.
- **Audit logs** are append-only in PostgreSQL. Application DB user has `REVOKE UPDATE, DELETE`.
- **Payment data** (tokens, references) must never be stored in Local SQLite. Only the payment reference ID and method are cached offline.

---

## 18.8 Key Concepts (Architecture-specific)

| Term | Definition |
|------|-----------|
| **Gate Controller** | Dedicated edge computer (Raspberry Pi / MCU) on the booth LAN that manages the gate motor, loop sensor, and safety interlocks independently of the Operator PC. |
| **Loop Sensor** | Inductive or magnetic vehicle detector embedded in the road surface, wired directly to the Gate Controller. |
| **Local SQLite** | Embedded database inside the Operator App used as a write-ahead cache when the backend is unreachable. |
| **Sync Service** | Background worker that ingests offline batches from Operator Apps, deduplicates, validates, and writes them to PostgreSQL. |
| **Offline Mode** | Degraded operating state where the Operator App functions using only Local SQLite and LAN-connected peripherals (printer, gate controller). |
| **Failsafe Close** | Gate motor configuration that defaults to a closed barrier position on power loss or controller failure. |
| **Whitelist** | A cached list of plates or session states stored on the Gate Controller enabling autonomous operation without the Operator App. |

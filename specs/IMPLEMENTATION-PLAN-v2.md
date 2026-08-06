# PARKIR v2 — Implementation Timeline & Steps
**Created:** 2026-08-06
**Target:** MVP for AMB (1 location, 2-4 gates)

---

## Phase 1: Foundation (Week 1-2)

### 1.1 Cloud Backend Setup
**Duration:** 3 days
**Deliverables:**
- [ ] Database schema (gates, locations, configs, sessions, transactions)
- [ ] Basic API structure (Go/Gin)
- [ ] Authentication (JWT, superadmin bootstrap)
- [ ] Location management API (CRUD)
- [ ] Config management API (rates, shifts with versioning)
- [ ] Gate registration API
- [ ] Sync API endpoints (receive data from server room apps)
- [ ] Health check endpoint

**Tech Stack:**
- Go 1.22+ with Gin framework
- PostgreSQL (cloud database)
- JWT RS256 authentication
- pgx for database access

**Steps:**
1. Set up Go project structure (cmd/api, internal/domain, internal/store)
2. Create database migrations (locations, gates, configs, sessions, transactions, users, roles)
3. Implement authentication middleware
4. Implement location CRUD handlers
5. Implement config CRUD handlers (rates, shifts) with version tracking
6. Implement gate registration handler
7. Implement sync handler (receive sessions, transactions, gate status)
8. Implement health check endpoint
9. Write API tests
10. Deploy to staging server

---

### 1.2 Server Room App Foundation
**Duration:** 4 days
**Deliverables:**
- [ ] Local database setup (PostgreSQL or SQLite)
- [ ] Config sync from cloud (poll every 1 min)
- [ ] Gate discovery via mDNS
- [ ] HTTP server for gate communication
- [ ] Gate registry (track connected gates)
- [ ] Health check for gate apps (ping every 15s)
- [ ] Status reporting to cloud (every 20s)

**Tech Stack:**
- Go or Rust (for performance + small binary)
- SQLite (local database, portable)
- mDNS library (github.com/miekg/dns or similar)
- HTTP server (net/http or framework)

**Steps:**
1. Set up project structure
2. Implement local database (SQLite) with schema (gates, configs, sessions, transactions, sync_queue)
3. Implement cloud sync client (HTTP client to poll configs, push data)
4. Implement mDNS discovery (announce server room app, discover gate apps)
5. Implement HTTP server (endpoints for gate apps to communicate)
6. Implement gate registry (track gate_id, gate_type, vehicle_type, status, last_seen)
7. Implement health check loop (ping each gate every 15s)
8. Implement status reporting to cloud (every 20s, send gate statuses)
9. Implement config version tracking
10. Write unit tests

---

### 1.3 Gate App Foundation
**Duration:** 3 days
**Deliverables:**
- [ ] HTTP server (receive commands from server room app)
- [ ] mDNS announcement (auto-discover by server room app)
- [ ] Basic hardware abstraction layer (HAL)
- [ ] Communication with server room app (HTTP client)
- [ ] Gate ID detection (hardware serial/MAC)
- [ ] Minimal config storage (server_room_url, gate_id)

**Tech Stack:**
- Go or Rust (small binary, cross-platform)
- HTTP server
- mDNS library
- Serial/USB libraries for hardware

**Steps:**
1. Set up project structure
2. Implement HTTP server (endpoints: /info, /command, /health)
3. Implement mDNS announcement (parking-gate-{gate_id}.local:8080)
4. Implement hardware detection (read serial/MAC for gate_id)
5. Implement config storage (JSON file: server_room_url, gate_id)
6. Implement HTTP client to communicate with server room app
7. Implement hardware abstraction layer (interfaces for printer, gate motor, scanner)
8. Write unit tests
9. Test on mini PC hardware

---

## Phase 2: Entry Flow (Week 3)

### 2.1 Entry Gate Hardware Integration
**Duration:** 3 days
**Deliverables:**
- [ ] Vehicle loop sensor integration (detect vehicle presence)
- [ ] Ticket button integration (receive button press)
- [ ] Epson thermal printer integration (print QR ticket)
- [ ] Gate motor integration (open/close gate)
- [ ] QR code generation (session_id + location_id + timestamp + vehicle_type)

**Steps:**
1. Interface with central hub (understand communication protocol)
2. Implement vehicle loop sensor reader (detect vehicle in loop)
3. Implement ticket button handler (receive button press event)
4. Implement button logic: only dispense if vehicle detected
5. Implement QR code generation (encode: session_id, location_id, timestamp, vehicle_type)
6. Implement Epson thermal printer driver (ESC/POS commands or USB)
7. Implement ticket printing (QR code + warning text + timestamp + vehicle type)
8. Implement gate motor control (open/close via hub)
9. Implement entry sequence: button press → create session → print ticket → open gate
10. Test with actual hardware

### 2.2 Server Room App Entry Logic
**Duration:** 2 days
**Deliverables:**
- [ ] Session creation API (for gate app to call)
- [ ] Session storage in local DB
- [ ] Shift number assignment (from shift config)
- [ ] Sync session to cloud (queue for sync)

**Steps:**
1. Implement session creation handler (receive from gate app)
2. Implement session storage (save to local DB with status ACTIVE)
3. Implement shift number assignment (determine current shift from config)
4. Implement session sync queue (mark for sync to cloud)
5. Implement cloud sync for sessions (push to cloud backend)
6. Write integration tests

### 2.3 Cloud Backend Entry Support
**Duration:** 1 day
**Deliverables:**
- [ ] Session creation API (receive from server room app)
- [ ] Session storage in cloud DB
- [ ] Session query API (for dashboard)

**Steps:**
1. Implement session creation handler (receive from server room app sync)
2. Implement session storage (save to cloud DB)
3. Implement session query API (list sessions, get by ID)
4. Write tests

---

## Phase 3: Exit Flow (Week 4)

### 3.1 Exit Gate Hardware Integration
**Duration:** 3 days
**Deliverables:**
- [ ] QR scanner integration (Panda PRJ-777, USB HID)
- [ ] Payment terminal integration (vendor API, TBD)
- [ ] Driver-facing monitor UI (display fee, check-in time, duration)
- [ ] Receipt button integration (print receipt on demand)
- [ ] Alert button integration (notify staff)

**Steps:**
1. Implement QR scanner reader (USB HID, read keyboard input)
2. Implement QR data parsing (extract session_id, location_id, timestamp, vehicle_type)
3. Implement payment terminal integration (vendor-specific, TBD)
4. Implement driver-facing monitor UI (web-based or native, show fee + instructions)
5. Implement receipt button handler (print receipt: check-in/out time, fee, vehicle type, shift)
6. Implement receipt printing (thermal printer)
7. Implement alert button handler (send alert to server room app)
8. Test with actual hardware

### 3.2 Server Room App Exit Logic
**Duration:** 3 days
**Deliverables:**
- [ ] Fee calculation engine (rate lookup + calculation)
- [ ] Session lookup by QR data
- [ ] Session update (check-out time, fee, shift number)
- [ ] Transaction creation
- [ ] Payment finalization
- [ ] Audio alert system (for gate alerts)

**Steps:**
1. Implement fee calculation engine:
   - Lookup rate config (by location_id, vehicle_type, check_in_date)
   - Calculate duration (check_out - check_in)
   - Calculate fee (first_hour + subsequent_hours * rate, capped at daily_flat_rate)
   - Handle multi-day stays (recurring 24-hour blocks)
2. Implement session lookup (find session by session_id from QR)
3. Implement session update (set check_out_at, fee_amount, shift_number, state=CLOSED)
4. Implement transaction creation (record payment: amount, method, reference)
5. Implement payment finalization (mark session CLOSED, create transaction)
6. Implement sync queue for transactions (mark for sync to cloud)
7. Implement cloud sync for transactions (push to cloud backend)
8. Implement audio alert system (play sound when gate app sends alert, announce gate_id)
9. Write integration tests

### 3.3 Cloud Backend Exit Support
**Duration:** 1 day
**Deliverables:**
- [ ] Transaction creation API (receive from server room app)
- [ ] Transaction storage in cloud DB
- [ ] Transaction query API (for dashboard)

**Steps:**
1. Implement transaction creation handler (receive from server room app sync)
2. Implement transaction storage (save to cloud DB)
3. Implement transaction query API (list transactions, filters)
4. Write tests

---

## Phase 4: Dashboard & Monitoring (Week 5)

### 4.1 Dashboard Foundation
**Duration:** 3 days
**Deliverables:**
- [ ] Next.js app setup
- [ ] Authentication (login, JWT cookie)
- [ ] Location selector (grouped by city)
- [ ] Layout with sidebar navigation

**Steps:**
1. Set up Next.js 14 app (App Router, TypeScript, Tailwind)
2. Implement authentication (login page, JWT cookie handling)
3. Implement location context (selected location, grouped by city)
4. Implement layout with sidebar (navigation links)
5. Implement responsive design

### 4.2 Dashboard - Live Monitoring
**Duration:** 2 days
**Deliverables:**
- [ ] Gate status overview (online/offline/busy)
- [ ] Active sessions view
- [ ] Revenue today (live)
- [ ] Alert counts

**Steps:**
1. Implement gate status API (fetch from cloud backend)
2. Implement gate status UI (cards showing each gate, color-coded status)
3. Implement active sessions API + UI
4. Implement revenue today API + UI
5. Implement alert counts API + UI
6. Implement auto-refresh (every 30s)

### 4.3 Dashboard - Gate Management
**Duration:** 2 days
**Deliverables:**
- [ ] Unregistered gates view (new gates detected)
- [ ] Gate configuration form (vehicle_type, gate_type, location)
- [ ] Registered gates list (status, last_seen)
- [ ] Gate health history

**Steps:**
1. Implement unregistered gates API (fetch from cloud)
2. Implement unregistered gates UI (list with "Configure" button)
3. Implement gate configuration form (modal or page)
4. Implement gate config save (send to cloud, cloud pushes to server room app)
5. Implement registered gates list UI
6. Implement gate health history UI (chart or timeline)

### 4.4 Dashboard - Configuration Management
**Duration:** 3 days
**Deliverables:**
- [ ] Rate management (per location, with versioning)
- [ ] Shift config management (per location)
- [ ] User management (create accounts, assign roles/permissions)
- [ ] Manual refresh buttons

**Steps:**
1. Implement rate management UI (list, create, edit rates per location)
2. Implement rate versioning display (show which version is active)
3. Implement shift config management UI
4. Implement user management UI (list users, create user, assign role + locations)
5. Implement role/permission management UI
6. Implement manual refresh buttons (for configs and data)

### 4.5 Dashboard - Reports
**Duration:** 3 days
**Deliverables:**
- [ ] Daily revenue report (per location + total, chart + table)
- [ ] Occupancy report (by time bucket)
- [ ] Transaction list (with filters)
- [ ] Export to CSV/Excel

**Steps:**
1. Implement daily revenue API (aggregate from cloud DB)
2. Implement daily revenue UI (chart + table, date range picker)
3. Implement occupancy API + UI
4. Implement transaction list API (with filters: date, location, vehicle_type)
5. Implement transaction list UI (table with filters)
6. Implement CSV/Excel export (all reports)

---

## Phase 5: Offline & Sync (Week 6)

### 5.1 Server Room App Offline Mode
**Duration:** 3 days
**Deliverables:**
- [ ] Offline operation (gates work without internet)
- [ ] Transaction sync queue
- [ ] Exponential backoff retry
- [ ] Sync failure alerting

**Steps:**
1. Implement offline mode detection (check internet connectivity)
2. Implement transaction sync queue (store unsynced transactions in local DB)
3. Implement sync loop (try to push queued transactions to cloud)
4. Implement exponential backoff retry (1s, 2s, 4s, 8s, max 30s)
5. Implement sync failure alerting (log error, increment failure counter)
6. Implement sync success handling (mark transaction as synced, remove from queue)
7. Test offline scenario (disconnect internet, verify gates still work, reconnect, verify sync)

### 5.2 Cloud Backend Sync Handling
**Duration:** 2 days
**Deliverables:**
- [ ] Batch sync API (receive multiple transactions)
- [ ] Conflict detection (duplicate transactions)
- [ ] Sync status tracking

**Steps:**
1. Implement batch sync API (receive array of transactions from server room app)
2. Implement duplicate detection (check transaction_id, skip if exists)
3. Implement sync status tracking (mark transaction as synced in cloud DB)
4. Implement sync response (return list of successfully synced transaction_ids)
5. Write tests

---

## Phase 6: Monitoring & Alerting (Week 7)

### 6.1 Logging Infrastructure
**Duration:** 2 days
**Deliverables:**
- [ ] Loki setup (log aggregation)
- [ ] Gate app logging (to Loki)
- [ ] Server room app logging (to Loki)
- [ ] Cloud backend logging (to Loki)
- [ ] Dashboard logging (to Loki)

**Steps:**
1. Set up Loki (log aggregation server)
2. Implement logging in gate app (error, warn, info → Loki)
3. Implement logging in server room app (→ Loki)
4. Implement logging in cloud backend (→ Loki)
5. Implement logging in dashboard (→ Loki)
6. Configure log retention policy

### 6.2 Metrics & Monitoring
**Duration:** 3 days
**Deliverables:**
- [ ] Prometheus setup (metrics collection)
- [ ] Grafana dashboards (visualization)
- [ ] Gate app metrics (uptime, response time, hardware status)
- [ ] Server room app metrics (uptime, DB size, sync queue length, transaction rate)
- [ ] Cloud backend metrics (uptime, API response time, DB connections)
- [ ] Dashboard metrics (uptime, page load time)

**Steps:**
1. Set up Prometheus (metrics collection server)
2. Set up Grafana (visualization)
3. Implement metrics in gate app (expose /metrics endpoint)
4. Implement metrics in server room app (expose /metrics endpoint)
5. Implement metrics in cloud backend (expose /metrics endpoint)
6. Implement metrics in dashboard (expose /metrics endpoint)
7. Create Grafana dashboards (one for each component)
8. Implement gate metrics collection by server room app (via ping API)

### 6.3 Alerting
**Duration:** 2 days
**Deliverables:**
- [ ] Email alerts (for developer, server room app down)
- [ ] Telegram alerts (for developer)
- [ ] Dashboard alerts (visual indicators)

**Steps:**
1. Implement email alerting (SMTP integration, send on sync failure, server room down)
2. Implement Telegram alerting (Telegram Bot API, send on critical errors)
3. Implement dashboard alerts (visual indicators for gate/server room health)
4. Test alerting (simulate failures, verify alerts sent)

---

## Phase 7: Testing & Deployment (Week 8)

### 7.1 Integration Testing
**Duration:** 3 days
**Deliverables:**
- [ ] End-to-end entry flow test (hardware + software)
- [ ] End-to-end exit flow test (hardware + software)
- [ ] Offline mode test
- [ ] Sync test
- [ ] Alert test

**Steps:**
1. Set up test environment (1 location, 2 gates: 1 entry, 1 exit)
2. Test entry flow (vehicle enters, button press, ticket dispensed, gate opens)
3. Test exit flow (QR scan, fee calculation, payment, gate opens)
4. Test offline mode (disconnect internet, verify gates work, reconnect, verify sync)
5. Test alert flow (press alert button, verify audio alarm)
6. Test payment failure (simulate failed payment, verify gate stays closed)
7. Test hardware failure (simulate printer jam, verify gate stays closed, alert triggered)
8. Fix bugs

### 7.2 Deployment Package
**Duration:** 2 days
**Deliverables:**
- [ ] Gate app USB installer
- [ ] Server room app USB installer
- [ ] Installation guide
- [ ] Configuration guide

**Steps:**
1. Create gate app installer (USB package with auto-install script)
2. Create server room app installer (USB package with auto-install script)
3. Write installation guide (step-by-step for field technician)
4. Write configuration guide (how to configure gates via dashboard)
5. Test installation process (install on fresh hardware)

### 7.3 Documentation
**Duration:** 2 days
**Deliverables:**
- [ ] API documentation
- [ ] Hardware setup guide
- [ ] Troubleshooting guide
- [ ] Staff training material

**Steps:**
1. Write API documentation (endpoints, request/response examples)
2. Write hardware setup guide (how to connect hub, sensors, printer, scanner)
3. Write troubleshooting guide (common issues + solutions)
4. Create staff training material (how to handle alerts, offline SOP)

---

## Phase 8: Pilot & Iteration (Week 9-10)

### 8.1 Pilot Deployment
**Duration:** 1 week
**Deliverables:**
- [ ] Deploy to 1 AMB location (2-4 gates)
- [ ] Monitor for 1 week
- [ ] Collect feedback
- [ ] Fix critical bugs

**Steps:**
1. Deploy to 1 AMB location (install hardware, install software)
2. Train AMB staff (how to use system, handle alerts)
3. Monitor system for 1 week (check logs, metrics, alerts)
4. Collect feedback from AMB staff
5. Fix critical bugs
6. Iterate on UI/UX based on feedback

### 8.2 Preparation for Scale
**Duration:** 1 week
**Deliverables:**
- [ ] Multi-location support verified
- [ ] Performance testing (20+ locations)
- [ ] Documentation updated
- [ ] Rollout plan for remaining locations

**Steps:**
1. Verify multi-location support (dashboard shows all locations, data aggregated correctly)
2. Performance test (simulate 20+ locations, verify cloud backend handles load)
3. Update documentation based on pilot learnings
4. Create rollout plan (timeline for deploying to remaining 19+ locations)
5. Train field technicians (how to deploy to new locations)

---

## Timeline Summary

| Phase | Duration | Key Deliverables |
|-------|----------|------------------|
| **Phase 1: Foundation** | Week 1-2 | Cloud backend, server room app, gate app (foundation) |
| **Phase 2: Entry Flow** | Week 3 | Entry gate hardware integration, session creation |
| **Phase 3: Exit Flow** | Week 4 | Exit gate hardware integration, fee calculation, payment |
| **Phase 4: Dashboard** | Week 5 | Dashboard UI, gate management, config management, reports |
| **Phase 5: Offline & Sync** | Week 6 | Offline mode, sync queue, retry logic |
| **Phase 6: Monitoring** | Week 7 | Logging (Loki), metrics (Prometheus/Grafana), alerting |
| **Phase 7: Testing & Deployment** | Week 8 | Integration testing, USB installers, documentation |
| **Phase 8: Pilot** | Week 9-10 | Deploy to 1 location, monitor, iterate, prepare for scale |

**Total:** 10 weeks to MVP (1 location)

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Hardware integration delays | Start hardware integration early (Phase 2-3), have backup hardware |
| Payment vendor integration | Vendor TBD, start communication early, have fallback (manual payment) |
| Offline sync conflicts | Use transaction_id as unique key, cloud detects duplicates |
| Performance at scale | Performance test in Phase 8, optimize before rolling to 20+ locations |
| Staff adoption | Training material (Phase 7), pilot with AMB staff (Phase 8) |

---

## Post-MVP Roadmap

**Month 3:** Deploy to 5 more locations (total 6)
**Month 4:** Deploy to remaining locations (total 20+)
**Month 5:** Advanced features (receipt printing optimization, multi-language support)
**Month 6:** Performance optimization, cost reduction

---

*End of Implementation Plan*

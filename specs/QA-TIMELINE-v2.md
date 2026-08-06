# PARKIR v2 — QA Timeline & Implementation
**Created:** 2026-08-06

---

## QA Components

### 1. Test Strategy
- Unit testing
- Integration testing
- End-to-end (E2E) testing
- Performance testing
- Security testing
- Hardware testing
- Offline testing
- User acceptance testing (UAT)

### 2. Test Environments
- Local (development)
- Staging (cloud)
- Production (cloud)
- Pilot location (AMB)

### 3. Test Automation
- API test automation
- UI test automation
- Hardware simulation
- CI/CD integration

### 4. Manual Testing
- Exploratory testing
- Usability testing
- Hardware integration testing
- Edge case testing

---

## Week 1-2: QA Foundation

### Test Strategy Document (2 days)
**Deliverables:**
- [ ] Test strategy document (scope, approach, tools, environments)
- [ ] Test levels (unit, integration, E2E, performance, security)
- [ ] Test types (functional, regression, smoke, sanity, exploratory)
- [ ] Entry/exit criteria for each test level
- [ ] Defect management process (severity, priority, workflow)
- [ ] Test environment requirements
- [ ] Test data management strategy
- [ ] Test automation strategy (what to automate, what to manual test)

### Test Environment Setup (2 days)
**Deliverables:**
- [ ] Local development environment:
  - Docker Compose (PostgreSQL, backend API)
  - Mock server room app (for gate app testing)
  - Mock gate apps (for server room app testing)
  - Mock hardware (simulated printer, scanner, gate motor)
- [ ] Staging environment (cloud):
  - Backend API (staging)
  - PostgreSQL (staging)
  - Dashboard (staging)
  - Test data seed script
- [ ] Test data management:
  - Seed script (create test locations, gates, users, configs)
  - Test data cleanup script
  - Test data documentation

### Bug Tracking Setup (1 day)
**Deliverables:**
- [ ] Choose bug tracking tool (GitHub Issues, Jira, Linear)
- [ ] Configure workflow:
  - New → Triaged → In Progress → In Review → Done
  - Severity levels: Critical, High, Medium, Low
  - Priority levels: P0 (immediate), P1 (this week), P2 (next sprint), P3 (backlog)
- [ ] Bug report template:
  - Description
  - Steps to reproduce
  - Expected vs actual result
  - Environment (OS, browser, app version)
  - Screenshots/logs
  - Severity + priority
- [ ] Integration with CI/CD (auto-create bug on test failure)

### Test Case Management (1 day)
**Deliverables:**
- [ ] Choose test case management tool (TestRail, Zephyr, or simple markdown)
- [ ] Test case template:
  - Test case ID
  - Title
  - Preconditions
  - Steps
  - Expected result
  - Postconditions
- [ ] Organize test cases by module:
  - Authentication
  - Locations
  - Gates
  - Configs (rates, shifts)
  - Sessions
  - Transactions
  - Reports
  - Users/Roles

---

## Week 3-4: Unit & Integration Testing

### Backend Unit Tests (3 days)
**Deliverables:**
- [ ] Cloud backend unit tests:
  - Auth handlers (login, logout, refresh, me)
  - Location handlers (CRUD)
  - Config handlers (rates, shifts, versioning)
  - Gate handlers (register, update, list)
  - Session handlers (create, close)
  - Transaction handlers (create, void)
  - Sync handlers (batch sync, duplicate detection)
  - Fee calculation logic (multi-day stays, daily caps)
  - Shift assignment logic
- [ ] Test coverage: 80%+ for critical paths
- [ ] Mock external dependencies (database, email, Telegram)

### Server Room App Unit Tests (2 days)
**Deliverables:**
- [ ] Unit tests:
  - Fee calculation (edge cases: multi-day, overnight, daily cap)
  - Shift assignment (determine current shift)
  - Config sync (poll cloud, version comparison)
  - Data sync (queue management, retry logic)
  - Gate management (registry, health check)
- [ ] Test coverage: 80%+
- [ ] Mock cloud backend, gate apps

### Gate App Unit Tests (2 days)
**Deliverables:**
- [ ] Unit tests:
  - Entry sequence (button press → session creation → print → open gate)
  - Exit sequence (QR scan → fee display → payment → open gate)
  - Error handling (printer jam, gate failure, payment failed)
  - Hardware abstraction layer (mock printer, scanner, gate motor)
- [ ] Test coverage: 80%+
- [ ] Mock hardware (simulated printer, scanner, gate motor)

### Backend Integration Tests (3 days)
**Deliverables:**
- [ ] Cloud backend integration tests:
  - Auth flow (login → access protected endpoint → logout)
  - Location CRUD (create → read → update → deactivate)
  - Config CRUD (create rate → verify version increment → update → verify version)
  - Gate registration (register → configure → verify status)
  - Session lifecycle (create → close → verify transaction)
  - Sync flow (server room app syncs sessions/transactions → verify in cloud DB)
  - Duplicate detection (sync same transaction twice → verify only one saved)
- [ ] Use test database (separate from dev/prod)
- [ ] Test data setup/teardown (before/after each test)

### Server Room App Integration Tests (2 days)
**Deliverables:**
- [ ] Integration tests:
  - Config sync (mock cloud → verify local DB updated)
  - Data sync (create session → sync to cloud → verify)
  - Gate discovery (mock gate app → verify registered)
  - Fee calculation (create session → calculate fee → verify)
  - Offline mode (disconnect cloud → create session → reconnect → verify sync)
- [ ] Mock cloud backend, gate apps

---

## Week 5-6: E2E Testing

### Dashboard E2E Tests (3 days)
**Deliverables:**
- [ ] E2E test framework setup (Playwright or Cypress)
- [ ] Test scenarios:
  - Authentication (login → verify redirect → logout)
  - Location switching (select location → verify URL change → verify data)
  - Gate management (view gates → configure new gate → verify status)
  - Rate management (create rate → edit rate → verify version)
  - Shift config (create shift → edit shift → verify)
  - User management (create user → edit user → deactivate user)
  - Reports (view daily revenue → filter by date → export CSV)
  - Manual refresh (click refresh → verify loading state → verify data updated)
- [ ] Test data setup (seed test data before tests)
- [ ] Test data cleanup (delete test data after tests)
- [ ] CI/CD integration (run E2E tests on staging before deployment)

### API E2E Tests (2 days)
**Deliverables:**
- [ ] API test framework setup (Postman, Bruno, or custom scripts)
- [ ] Test scenarios:
  - Auth flow (login → refresh → me → logout)
  - Full entry flow (create session → sync to cloud → verify)
  - Full exit flow (create session → close session → create transaction → sync → verify)
  - Config management (create rate → create shift → verify)
  - Gate management (register gate → configure gate → verify)
  - Sync flow (batch sync → verify duplicates rejected)
- [ ] Test environment: staging
- [ ] Test data: seed script

### Hardware Integration E2E Tests (3 days)
**Deliverables:**
- [ ] Hardware test setup (pilot location or lab):
  - 1 entry gate (with loop sensor, button, printer, gate motor)
  - 1 exit gate (with QR scanner, payment terminal, monitor, gate motor)
  - Server room PC
- [ ] Test scenarios:
  - Entry flow (vehicle in loop → button press → ticket dispensed → gate opens → vehicle enters → gate closes)
  - Exit flow (QR scan → fee displayed → payment → gate opens → vehicle exits → gate closes)
  - Receipt printing (press receipt button → receipt printed)
  - Alert flow (press alert button → audio alarm in server room)
  - Error scenarios:
    - Printer jam → gate stays closed → alert triggered
    - QR unreadable → gate stays closed → alert triggered
    - Payment failed → gate stays closed → message displayed
    - Gate motor failure → gate stays closed → alert triggered
- [ ] Test with real hardware (pilot location)
- [ ] Document hardware-specific issues

---

## Week 7-8: Performance & Security Testing

### Performance Testing (3 days)
**Deliverables:**
- [ ] Performance test framework setup (k6, JMeter, or locust)
- [ ] Load test scenarios:
  - API load test:
    - 100 concurrent users (dashboard)
    - 1000 requests/minute (API endpoints)
    - Measure: response time, error rate, throughput
  - Sync load test:
    - 20 locations syncing simultaneously
    - 100 transactions per location per hour
    - Measure: sync latency, queue length, DB performance
  - Database load test:
    - 10,000 sessions per day
    - 10,000 transactions per day
    - Measure: query time, index performance
- [ ] Performance benchmarks:
  - API response time: < 500ms (p95)
  - Sync latency: < 5 seconds
  - DB query time: < 100ms (p95)
- [ ] Identify bottlenecks, optimize
- [ ] Performance test report (document results, recommendations)

### Security Testing (3 days)
**Deliverables:**
- [ ] Security test framework setup (OWASP ZAP, Burp Suite, or manual)
- [ ] Security test scenarios:
  - Authentication:
    - Brute force attack (rate limit login endpoint)
    - JWT token manipulation (modify payload → verify rejected)
    - Session fixation (verify new token on login)
    - Cookie security (verify httpOnly, secure, sameSite flags)
  - Authorization:
    - Vertical privilege escalation (regular user → access admin endpoint → verify denied)
    - Horizontal privilege escalation (user A → access user B data → verify denied)
    - Location scoping (user assigned to location A → access location B data → verify denied)
  - Input validation:
    - SQL injection (inject SQL in input → verify rejected)
    - XSS (inject script in input → verify sanitized)
    - CSRF (verify CSRF token required for state-changing requests)
  - API security:
    - Rate limiting (send 1000 requests → verify rate limited)
    - CORS (verify only allowed origins)
    - SSL/TLS (verify TLS 1.2+, strong ciphers)
- [ ] Security test report (document vulnerabilities, remediation plan)
- [ ] Fix critical + high severity vulnerabilities

### Offline Testing (2 days)
**Deliverables:**
- [ ] Offline test scenarios:
  - Internet outage:
    - Disconnect internet at location
    - Verify gates still operate (entry/exit)
    - Verify transactions queued locally
    - Reconnect internet → verify sync completes
    - Verify no duplicate transactions
  - LAN outage (gate ↔ server room):
    - Disconnect LAN cable
    - Verify gate stops operating
    - Verify alert triggered (audio alarm)
    - Reconnect LAN → verify gate resumes
  - Server room PC failure:
    - Shutdown server room PC
    - Verify gates stop operating
    - Verify alert triggered
    - Restart server room PC → verify gates resume
  - Cloud backend failure:
    - Shutdown cloud backend
    - Verify server room apps continue operating
    - Verify transactions queued
    - Restart cloud backend → verify sync completes
- [ ] Document offline behavior, edge cases

---

## Week 9-10: User Acceptance Testing (UAT)

### UAT Planning (1 day)
**Deliverables:**
- [ ] UAT plan document:
  - Scope (what features to test)
  - Participants (AMB staff, 5-10 users)
  - Schedule (1 week)
  - Environment (pilot location)
  - Test scenarios (real-world workflows)
  - Success criteria (no critical bugs, user satisfaction > 80%)
- [ ] UAT test cases (real-world scenarios):
  - Driver enters parking (entry flow)
  - Driver exits parking (exit flow)
  - Driver payment fails (topup → retry)
  - Driver presses alert button (staff responds)
  - Staff configures new gate
  - Staff views reports
  - Staff handles offline scenario
- [ ] UAT feedback form (rating, comments, suggestions)

### UAT Execution (5 days)
**Deliverables:**
- [ ] Day 1: Training
  - Train AMB staff (how to use system, handle alerts, offline SOP)
  - Provide user manual
  - Answer questions
- [ ] Day 2-3: Guided testing
  - AMB staff uses system (supervised)
  - Observe workflows, note issues
  - Collect feedback
- [ ] Day 4-5: Independent testing
  - AMB staff uses system (unsupervised)
  - Monitor logs, metrics, alerts
  - Collect feedback
- [ ] Bug triage:
  - Triage bugs (severity, priority)
  - Fix critical + high bugs immediately
  - Document medium + low bugs for post-MVP
- [ ] UAT report:
  - Test results (pass/fail)
  - Bug summary (critical, high, medium, low)
  - User feedback summary
  - Recommendations (go/no-go for pilot)

### Usability Testing (2 days)
**Deliverables:**
- [ ] Usability test scenarios:
  - Dashboard:
    - Login → navigate to gate status → configure new gate
    - View reports → filter by date → export CSV
    - Create user → assign role + locations
  - Gate app (driver):
    - Enter parking (entry flow)
    - Exit parking (exit flow)
    - Handle payment failure
  - Server room app (staff):
    - Monitor gates → respond to alert
    - Handle offline scenario
- [ ] Usability test participants (5-7 users):
  - AMB staff (3-4)
  - Drivers (2-3)
- [ ] Measure:
  - Time on task
  - Error rate
  - User satisfaction (1-5 scale)
- [ ] Usability test report:
  - Pain points
  - Recommendations
  - UI/UX improvements

---

## Week 11-12: Release Testing & QA Handoff

### Regression Testing (2 days)
**Deliverables:**
- [ ] Full regression test suite:
  - All E2E tests (dashboard, API, hardware)
  - All integration tests
  - Critical unit tests
- [ ] Run on staging environment
- [ ] Fix any failures
- [ ] Regression test report (pass/fail)

### Smoke Testing (1 day)
**Deliverables:**
- [ ] Smoke test suite (critical path):
  - Login → view gates → configure gate
  - Entry flow (button press → ticket → gate opens)
  - Exit flow (QR scan → fee → payment → gate opens)
  - Sync (server room app → cloud)
- [ ] Run on production environment (after deployment)
- [ ] Verify all critical features working

### QA Documentation (2 days)
**Deliverables:**
- [ ] Test plan document (final version)
- [ ] Test case repository (all test cases)
- [ ] Test data documentation (seed scripts, test accounts)
- [ ] Test environment documentation (setup instructions)
- [ ] Bug report template
- [ ] Test automation documentation (how to run tests)
- [ ] QA metrics dashboard (test coverage, bug trends, pass rate)

### QA Handoff (1 day)
**Deliverables:**
- [ ] QA handoff document:
  - Test coverage summary
  - Known issues (bugs not fixed, deferred to post-MVP)
  - Risk assessment (what could go wrong)
  - Recommendations (monitoring, testing in production)
- [ ] Handoff meeting (QA → dev team → operations):
  - Review test results
  - Review known issues
  - Review monitoring plan
  - Review incident response plan

---

## QA Timeline Summary

| Week | Focus | Key Deliverables |
|------|-------|------------------|
| **1-2** | QA foundation | Test strategy, test environments (local + staging), bug tracking, test case management |
| **3-4** | Unit & integration tests | Backend unit tests (80%+ coverage), server room app unit tests, gate app unit tests, backend integration tests, server room app integration tests |
| **5-6** | E2E testing | Dashboard E2E (Playwright/Cypress), API E2E (Postman/Bruno), hardware integration E2E (pilot location) |
| **7-8** | Performance & security | Load tests (k6/JMeter, API + sync + DB), security tests (OWASP ZAP, auth, injection, CSRF), offline testing (internet/LAN/server room failure) |
| **9-10** | UAT | UAT planning, UAT execution (5 days with AMB staff), usability testing (5-7 users), UAT report |
| **11-12** | Release testing | Regression testing, smoke testing (production), QA documentation, QA handoff |

---

## QA Tools

### Test Automation
| Tool | Purpose |
|------|---------|
| **Go testing** | Unit + integration tests (backend, server room app, gate app) |
| **Playwright** or **Cypress** | E2E tests (dashboard) |
| **Postman** or **Bruno** | API tests |
| **k6** or **JMeter** | Performance/load tests |
| **OWASP ZAP** or **Burp Suite** | Security tests |

### Test Environments
| Environment | Purpose |
|-------------|---------|
| **Local** | Development, unit/integration tests |
| **Staging** | E2E tests, performance tests, UAT |
| **Production** | Smoke tests, monitoring |
| **Pilot location** | Hardware integration tests, UAT |

### Bug Tracking & Test Management
| Tool | Purpose |
|------|---------|
| **GitHub Issues** or **Jira** | Bug tracking |
| **TestRail** or **Zephyr** or **Markdown** | Test case management |
| **Grafana** | QA metrics dashboard |

---

## QA Metrics

### Test Coverage
| Metric | Target |
|--------|--------|
| Unit test coverage (backend) | 80%+ |
| Unit test coverage (server room app) | 80%+ |
| Unit test coverage (gate app) | 80%+ |
| E2E test coverage (critical paths) | 100% |
| API test coverage (all endpoints) | 100% |

### Quality Metrics
| Metric | Target |
|--------|--------|
| Critical bugs (production) | 0 |
| High bugs (production) | < 5 |
| Bug fix time (critical) | < 4 hours |
| Bug fix time (high) | < 1 day |
| Test pass rate (regression) | 95%+ |
| UAT satisfaction | 80%+ |

### Performance Metrics
| Metric | Target |
|--------|--------|
| API response time (p95) | < 500ms |
| Sync latency (p95) | < 5 seconds |
| DB query time (p95) | < 100ms |
| Dashboard page load time | < 3 seconds |
| System uptime | 95%+ |

---

## QA Risks & Mitigation

| Risk | Mitigation |
|------|------------|
| Insufficient test coverage | Enforce 80%+ unit test coverage, 100% E2E for critical paths |
| Hardware integration bugs | Test with real hardware early (Week 5-6), use mocks for unit tests |
| Offline sync bugs | Extensive offline testing (Week 7-8), test all failure scenarios |
| Performance bottlenecks | Load test early (Week 7-8), optimize before pilot |
| Security vulnerabilities | Security test (Week 7-8), fix critical + high before pilot |
| UAT failures | UAT planning (Week 9), fix critical bugs before go-live |
| Test environment issues | Document test environment setup, use Docker for local testing |
| Test data management | Seed scripts, cleanup scripts, test data documentation |

---

## Post-MVP QA

### Month 3: Iteration Testing
- Test new features (based on pilot feedback)
- Regression testing
- UAT for new features

### Month 4-6: Scale Testing
- Load test (20+ locations)
- Stress test (100+ locations)
- Longevity test (run system for 30 days, monitor for memory leaks, DB growth)

### Ongoing: Continuous QA
- Automated tests in CI/CD (run on every PR)
- Regression tests before each release
- Monitoring production metrics (error rate, performance)
- Collect user feedback, prioritize bugs

---

*End of QA Timeline*

# PARKIR v2 — Infrastructure Timeline & Implementation
**Created:** 2026-08-06

---

## Infrastructure Components

### 1. Cloud Infrastructure
- Backend servers (API)
- PostgreSQL database
- Load balancer
- CDN (static assets)
- Monitoring stack (Loki, Prometheus, Grafana)
- Backup storage

### 2. Location Infrastructure (per location)
- Server room mini PC
- Gate mini PCs (1 per gate)
- LAN network (switch, router)
- Internet connection
- Hardware (hub, sensors, printers, scanners, monitors)

### 3. Deployment & CI/CD
- Build pipeline
- Artifact storage
- Deployment scripts
- Rollback procedures

### 4. Monitoring & Alerting
- Log aggregation (Loki)
- Metrics collection (Prometheus)
- Visualization (Grafana)
- Alert routing (email, Telegram)

---

## Week 1-2: Cloud Infrastructure Setup

### Cloud Provider Selection (1 day)
**Deliverables:**
- [ ] Choose cloud provider (AWS, GCP, DigitalOcean, or local Indonesian provider)
- [ ] Account setup + billing
- [ ] Region selection (Jakarta/Singapore for low latency)
- [ ] Resource limits review

**Recommendation:** DigitalOcean or AWS (good balance of cost + features)

### Network Setup (1 day)
**Deliverables:**
- [ ] VPC (Virtual Private Cloud)
- [ ] Subnets (public, private)
- [ ] Security groups (firewall rules)
- [ ] VPN (optional, for admin access)

### Backend Server Setup (2 days)
**Deliverables:**
- [ ] Provision server(s) (2 vCPU, 4GB RAM minimum)
- [ ] OS: Ubuntu 22.04 LTS
- [ ] SSH access (key-based, disable password)
- [ ] Fail2ban (brute force protection)
- [ ] UFW firewall (allow: 22 SSH, 80 HTTP, 443 HTTPS, 5432 PostgreSQL internal only)
- [ ] Automatic security updates (unattended-upgrades)

### PostgreSQL Database Setup (2 days)
**Deliverables:**
- [ ] Install PostgreSQL 15
- [ ] Configure:
  - Listen on private IP only (not public)
  - SSL/TLS encryption
  - Connection limits
  - Performance tuning (work_mem, shared_buffers, etc.)
- [ ] Create database + user
- [ ] Automated backups (pg_dump, daily, retain 30 days)
- [ ] Backup storage (S3 or local, encrypted)
- [ ] Replication setup (optional, for high availability)

### Domain & SSL (1 day)
**Deliverables:**
- [ ] Domain registration (parkir-amb.com or similar)
- [ ] DNS setup (A record → server IP)
- [ ] SSL certificate (Let's Encrypt, auto-renew)
- [ ] Nginx reverse proxy:
  - HTTP → HTTPS redirect
  - SSL termination
  - Proxy to backend API
  - Rate limiting (optional)

### Load Balancer (Optional, 1 day)
**Deliverables:**
- [ ] Setup load balancer (if multiple backend servers)
- [ ] Health check configuration
- [ ] SSL termination at load balancer
- [ ] Sticky sessions (if needed)

**Note:** For MVP (1 location), single server is fine. Add load balancer when scaling to 10+ locations.

---

## Week 3-4: Monitoring Infrastructure

### Loki Setup (Log Aggregation) (2 days)
**Deliverables:**
- [ ] Install Loki (standalone or Docker)
- [ ] Configure retention (30 days)
- [ ] Configure storage (local disk or S3)
- [ ] Expose API endpoint (internal only)
- [ ] Authentication (basic auth or API key)

**Architecture:**
```
Applications → Promtail (agent) → Loki → Grafana (query)
```

### Promtail Setup (Log Collector) (1 day)
**Deliverables:**
- [ ] Install Promtail on backend server
- [ ] Configure log sources:
  - Backend API logs (stdout → Loki)
  - Nginx logs (access, error → Loki)
  - PostgreSQL logs (optional)
- [ ] Labels (job, level, service)
- [ ] Test log ingestion

### Prometheus Setup (Metrics Collection) (2 days)
**Deliverables:**
- [ ] Install Prometheus
- [ ] Configure scrape targets:
  - Backend API (/metrics endpoint)
  - Node exporter (system metrics)
  - PostgreSQL exporter (database metrics)
- [ ] Configure retention (90 days)
- [ ] Configure storage (local disk)
- [ ] Expose API endpoint (internal only)

### Node Exporter Setup (1 day)
**Deliverables:**
- [ ] Install node_exporter on backend server
- [ ] Expose metrics endpoint (:9100/metrics)
- [ ] Prometheus scrapes every 15s

### PostgreSQL Exporter Setup (1 day)
**Deliverables:**
- [ ] Install postgres_exporter
- [ ] Configure database connection
- [ ] Expose metrics endpoint (:9187/metrics)
- [ ] Prometheus scrapes every 15s

### Grafana Setup (Visualization) (2 days)
**Deliverables:**
- [ ] Install Grafana
- [ ] Configure data sources:
  - Loki (logs)
  - Prometheus (metrics)
- [ ] Authentication (admin user, optional SSO)
- [ ] Create dashboards:
  - System overview (CPU, memory, disk, network)
  - Backend API (request rate, latency, errors)
  - PostgreSQL (connections, queries, cache hit ratio)
  - Sync status (queue length, sync success rate)
- [ ] Dashboard sharing (read-only links for AMB admins)

### Alertmanager Setup (Alert Routing) (1 day)
**Deliverables:**
- [ ] Install Alertmanager
- [ ] Configure alert rules:
  - High CPU usage (>80% for 5 min)
  - High memory usage (>90% for 5 min)
  - Disk space low (<10% free)
  - Backend API down (health check fails)
  - PostgreSQL down
  - Sync failures (>10 failed in 1 hour)
- [ ] Configure notification channels:
  - Email (SMTP)
  - Telegram (bot API)
- [ ] Test alerts (trigger manually, verify notification)

---

## Week 5-6: Backup & Disaster Recovery

### Automated Backups (2 days)
**Deliverables:**
- [ ] PostgreSQL backup script:
  - pg_dump (full backup, daily)
  - Compress (gzip)
  - Encrypt (gpg or openssl)
  - Upload to S3 (or remote storage)
  - Retain 30 days (delete older)
- [ ] Cron job (daily at 2 AM)
- [ ] Backup verification (test restore monthly)
- [ ] Backup monitoring (alert if backup fails)

### Disaster Recovery Plan (1 day)
**Deliverables:**
- [ ] Document recovery procedures:
  - Server failure → restore from backup
  - Database corruption → restore from backup
  - Data loss → point-in-time recovery (if replication enabled)
- [ ] RTO (Recovery Time Objective): 4 hours
- [ ] RPO (Recovery Point Objective): 24 hours (daily backup)
- [ ] Test recovery procedure (quarterly)

### Monitoring Backup Success (1 day)
**Deliverables:**
- [ ] Backup script logs to Loki
- [ ] Alert if backup fails (Alertmanager rule)
- [ ] Dashboard widget: last backup timestamp, backup size

---

## Week 7-8: Security Hardening

### Application Security (2 days)
**Deliverables:**
- [ ] Backend API security:
  - Rate limiting (Nginx or application-level)
  - CORS configuration (restrict to dashboard domain)
  - Input validation (prevent SQL injection, XSS)
  - SQL injection prevention (use parameterized queries)
  - XSS prevention (sanitize user input)
  - CSRF protection (if using cookies)
- [ ] Authentication security:
  - JWT RS256 (asymmetric keys)
  - Token expiry (8 hours access, 7 days refresh)
  - Secure cookie flags (httpOnly, secure, sameSite)
  - Password hashing (bcrypt, cost 12)
  - Brute force protection (rate limit login endpoint)

### Network Security (2 days)
**Deliverables:**
- [ ] Firewall rules (UFW):
  - Allow: 22 (SSH, admin IPs only)
  - Allow: 80, 443 (HTTP/HTTPS, public)
  - Allow: 5432 (PostgreSQL, internal only)
  - Deny: all other ports
- [ ] SSH hardening:
  - Key-based auth only (disable password)
  - Disable root login
  - Change default port (optional)
  - Fail2ban (brute force protection)
- [ ] SSL/TLS:
  - TLS 1.2+ only
  - Strong cipher suites
  - HSTS header (strict transport security)
  - SSL Labs score: A+

### Secrets Management (1 day)
**Deliverables:**
- [ ] Environment variables (never commit to git)
- [ ] .env file (local development)
- [ ] Secrets storage (production):
  - Option 1: Environment variables (simple)
  - Option 2: Vault (advanced, for scaling)
- [ ] Secrets to manage:
  - DATABASE_URL
  - JWT_PRIVATE_KEY
  - JWT_PUBLIC_KEY
  - SMTP credentials (for email alerts)
  - Telegram bot token (for alerts)
  - S3 credentials (for backups)

### Audit Logging (1 day)
**Deliverables:**
- [ ] Audit log table (PostgreSQL):
  - user_id, action, entity_type, entity_id, timestamp, IP, metadata
- [ ] Log all state-changing actions:
  - Login/logout
  - Create/update/delete (locations, gates, configs, users)
  - Session creation/closure
  - Transaction creation/void
- [ ] Audit log viewer (dashboard page)
- [ ] Export audit logs (CSV)

---

## Week 9-10: Location Infrastructure (Pilot)

### Server Room Setup (2 days)
**Deliverables:**
- [ ] Mini PC installation (Ubuntu 22.04 LTS)
- [ ] Network configuration:
  - Static IP (e.g., 192.168.1.10)
  - DNS configuration
  - Internet connection (fiber or 4G backup)
- [ ] Firewall (UFW):
  - Allow: 22 (SSH, local network only)
  - Allow: 8080 (server room app, local network only)
  - Allow: 9090 (Prometheus metrics, local network only)
  - Deny: all other ports
- [ ] Install server room app (USB)
- [ ] Configure server room app:
  - Cloud backend URL
  - Sync interval
  - Gate health check interval
- [ ] Test connectivity to cloud backend

### Gate Mini PC Setup (2 days)
**Deliverables:**
- [ ] Mini PC installation (Ubuntu 22.04 LTS or lightweight Linux)
- [ ] Network configuration:
  - Static IP (e.g., 192.168.1.101, 192.168.1.102, ...)
  - DNS configuration
- [ ] Firewall (UFW):
  - Allow: 8080 (gate app, local network only)
  - Deny: all other ports
- [ ] Install gate app (USB)
- [ ] Configure gate app:
  - Auto-detect gate_id (hardware serial/MAC)
  - Announce via mDNS
- [ ] Connect hardware:
  - Hub (USB or serial)
  - Vehicle loop sensor
  - Ticket button
  - Epson thermal printer (USB or serial)
  - Gate motor (via hub)
  - QR scanner (USB)
  - Payment terminal (USB or serial, TBD)
  - Driver-facing monitor (HDMI)
  - Alert button
- [ ] Test hardware integration

### LAN Setup (1 day)
**Deliverables:**
- [ ] Network switch (connect server room PC + gate PCs)
- [ ] Router (connect to internet)
- [ ] IP address scheme:
  - Server room: 192.168.1.10
  - Gate 1 (entry): 192.168.1.101
  - Gate 2 (exit): 192.168.1.102
  - Gate 3 (entry): 192.168.1.103
  - Gate 4 (exit): 192.168.1.104
- [ ] Test LAN connectivity (ping between devices)
- [ ] Test mDNS discovery (server room app discovers gates)

### Internet Setup (1 day)
**Deliverables:**
- [ ] Internet connection (fiber, minimum 10 Mbps)
- [ ] Backup connection (4G router, optional)
- [ ] Router configuration:
  - NAT (server room PC → cloud)
  - Port forwarding (not needed, server room app initiates outbound connections)
- [ ] Test internet connectivity
- [ ] Test cloud sync (server room app → cloud backend)

### Local Monitoring (1 day)
**Deliverables:**
- [ ] Install Promtail on server room PC
- [ ] Configure log sources:
  - Server room app logs → Loki (cloud)
  - Gate app logs → Loki (cloud, via server room app)
- [ ] Install node_exporter on server room PC
- [ ] Configure Prometheus (cloud) to scrape:
  - Server room app metrics (via internet)
  - Gate app metrics (via server room app relay)

### Local Backup (1 day)
**Deliverables:**
- [ ] Server room app local DB backup:
  - SQLite dump (daily)
  - Compress (gzip)
  - Store in local directory (e.g., /backups)
  - Retain 7 days (delete older)
- [ ] Cron job (daily at 3 AM)
- [ ] Test backup + restore

---

## Week 11-12: CI/CD & Deployment

### CI/CD Pipeline Setup (3 days)
**Deliverables:**
- [ ] Choose CI/CD tool (GitHub Actions, GitLab CI, or Jenkins)
- [ ] Build pipeline:
  - On push to main:
    - Run tests (unit, integration)
    - Build backend binary
    - Build Docker image (optional)
    - Push to artifact registry (Docker Hub, ECR)
  - On push to release/*:
    - Build release binaries (Linux, macOS, Windows)
    - Upload to artifact storage (S3, GitHub Releases)
- [ ] Deploy pipeline:
  - On merge to main:
    - Deploy to staging (automatic)
    - Run smoke tests
  - Manual trigger:
    - Deploy to production (after approval)
- [ ] Rollback procedure:
  - Keep last 5 releases
  - Rollback script (restore previous version)

### Artifact Storage (1 day)
**Deliverables:**
- [ ] Choose artifact storage (S3, GitHub Releases, or Nexus)
- [ ] Upload release binaries:
  - backend (Linux amd64)
  - server-room-app (Linux amd64)
  - gate-app (Linux amd64)
  - dashboard (Docker image or static files)
- [ ] Versioning (semantic versioning, v1.0.0, v1.0.1, ...)
- [ ] Retention policy (keep last 10 releases)

### Deployment Scripts (2 days)
**Deliverables:**
- [ ] Backend deployment script:
  - SSH to server
  - Stop backend service (systemd)
  - Backup current binary
  - Download new binary from artifact storage
  - Start backend service
  - Health check (wait for /health/ready)
  - Rollback if health check fails
- [ ] Server room app deployment script:
  - Package app + dependencies (USB or download)
  - Field technician runs script on mini PC
  - Stop service, replace binary, start service
  - Health check
- [ ] Gate app deployment script:
  - Same as server room app
- [ ] Dashboard deployment script:
  - Build static files (Next.js export)
  - Upload to CDN (S3 + CloudFront) or serve from backend
  - Invalidate CDN cache

### USB Deployment Package (2 days)
**Deliverables:**
- [ ] Create USB installer for server room app:
  - Auto-install script (install dependencies, copy binary, create systemd service)
  - Configuration wizard (cloud backend URL, sync interval)
  - Uninstall script
- [ ] Create USB installer for gate app:
  - Same as server room app
- [ ] Test USB installation (fresh mini PC, plug USB, run installer)
- [ ] Documentation (installation guide for field technicians)

---

## Infrastructure Timeline Summary

| Week | Cloud Infrastructure | Location Infrastructure | CI/CD & Deployment |
|------|---------------------|------------------------|-------------------|
| **1-2** | Cloud provider, network, backend server, PostgreSQL, domain/SSL, load balancer (optional) | — | — |
| **3-4** | Monitoring (Loki, Promtail, Prometheus, node_exporter, postgres_exporter, Grafana, Alertmanager) | — | — |
| **5-6** | Backup (PostgreSQL backup script, DR plan, monitoring) | — | — |
| **7-8** | Security (app security, network security, secrets, audit logging) | — | — |
| **9-10** | Local monitoring (Promtail, node_exporter, metrics relay) | Server room setup, gate setup, LAN, internet, local backup | — |
| **11-12** | — | — | CI/CD pipeline, artifact storage, deployment scripts, USB packages |

---

## Infrastructure Tech Stack

### Cloud Infrastructure
- **Cloud Provider:** DigitalOcean or AWS
- **Server:** Ubuntu 22.04 LTS
- **Database:** PostgreSQL 15
- **Reverse Proxy:** Nginx
- **SSL:** Let's Encrypt
- **Monitoring:** Loki (logs), Prometheus (metrics), Grafana (visualization), Alertmanager (alerts)
- **Backup:** pg_dump + S3 (or local)
- **CI/CD:** GitHub Actions or GitLab CI

### Location Infrastructure
- **Server Room PC:** Mini PC, Ubuntu 22.04 LTS, 8GB RAM, 256GB SSD
- **Gate PC:** Mini PC, Ubuntu 22.04 LTS (or lightweight Linux), 4GB RAM, 128GB SSD
- **Network:** Gigabit switch, router with internet
- **Internet:** Fiber (10+ Mbps) + 4G backup (optional)

---

## Infrastructure Costs (Estimate)

### Cloud Infrastructure (Monthly)
| Item | Cost (USD) |
|------|-----------|
| Backend server (2 vCPU, 4GB RAM) | $20-40 |
| PostgreSQL (managed, optional) | $30-60 |
| Monitoring stack (Loki, Prometheus, Grafana) | $0 (self-hosted) |
| Backup storage (S3, 100GB) | $5-10 |
| Domain + SSL | $1-2 |
| **Total (cloud)** | **$56-112/month** |

### Location Infrastructure (One-time, per location)
| Item | Cost (USD) |
|------|-----------|
| Server room mini PC | $300-500 |
| Gate mini PCs (4 gates) | $800-1200 |
| Network equipment (switch, router) | $100-200 |
| Hardware (hub, sensors, printers, scanners, monitors) | $2000-4000 |
| Internet setup (fiber installation) | $200-500 |
| **Total (per location)** | **$3400-6400** |

### Location Infrastructure (Monthly, per location)
| Item | Cost (USD) |
|------|-----------|
| Internet (fiber) | $50-100 |
| 4G backup (optional) | $20-40 |
| **Total (monthly per location)** | **$70-140** |

---

## Infrastructure Risks & Mitigation

| Risk | Mitigation |
|------|------------|
| Cloud provider outage | Multi-region backup (optional), local offline operation |
| Database failure | Daily backups, replication (optional), test restore procedure |
| Internet outage at location | Local offline operation, 4G backup (optional) |
| Hardware failure at location | Spare mini PCs, offline SOP, quick replacement |
| Security breach | Firewall, SSL, secrets management, audit logging, regular updates |
| Backup failure | Monitor backup success, test restore quarterly |
| Deployment failure | Rollback procedure, keep last 5 releases, smoke tests |

---

## Scaling Plan

### Phase 1: MVP (1 location)
- Single backend server
- Single PostgreSQL instance
- Self-hosted monitoring
- Manual deployments

### Phase 2: 5 locations
- Single backend server (still sufficient)
- PostgreSQL: consider managed database (RDS or similar)
- Monitoring: still self-hosted
- CI/CD: automated deployments

### Phase 3: 20+ locations
- Multiple backend servers + load balancer
- PostgreSQL: managed database + read replicas
- Monitoring: dedicated monitoring server
- CDN for dashboard static assets
- Automated scaling (horizontal)

### Phase 4: 100+ locations
- Microservices architecture (split sync, API, reporting)
- PostgreSQL: sharding or partitioning
- Monitoring: distributed (Prometheus federation)
- Multi-region deployment
- Kubernetes (optional, for orchestration)

---

*End of Infrastructure Timeline*

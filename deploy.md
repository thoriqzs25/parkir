# PARKIR Deployment Runbook

## Overview

This document describes how to deploy PARKIR to staging and production environments on Tencent Cloud Jakarta.

### Architecture

```
Nginx (443) ──┬── /api/* ──→ Go Backend (8080)
              └── /* ──────→ Next.js Dashboard (3000)
```

- **Nginx** terminates SSL and reverse-proxies to backend and dashboard containers.
- **PostgreSQL** runs as a Docker container with persistent volume.
- **Daily backups** are written to `/var/backups/parkir` and rotated after 90 days.
- **Logs** are forwarded to Loki via promtail (placeholder — configure LOKI_URL when ready).

---

## Prerequisites

- SSH access to the VM
- Docker and Docker Compose installed
- Domain names configured (DNS A records pointing to VM IP)
- JWT key pair generated (`make generate-jwt-keys`)
- `.env` file populated with secrets

---

## Initial VM Setup

```bash
# 1. SSH into the VM
ssh user@<vm-ip>

# 2. Install Docker (if not present)
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
# Log out and back in

# 3. Install Docker Compose plugin
sudo apt-get install docker-compose-plugin

# 4. Clone the repository
git clone <repo-url> parkir
cd parkir

# 5. Copy environment config
cp env.production .env
# Edit .env with real values (DB_PASSWORD, domains, etc.)
nano .env

# 6. Create JWT keys directory and generate keys
mkdir -p keys
# Generate keys and copy them:
# jwt-private.pem and jwt-public.pem to ./keys/

# 7. Initial SSL certificate
# Replace your-domain.com with actual domain
docker compose -f docker-compose.prod.yml run --rm certbot certonly \
  --webroot --webroot-path=/var/www/certbot \
  -d api.your-domain.com -d dashboard.your-domain.com

# 8. Copy certs to nginx ssl directory
sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem nginx/ssl/cert.pem
sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem nginx/ssl/key.pem
```

---

## Deploy (Staging)

```bash
# SSH into staging VM
ssh user@<staging-vm-ip>
cd parkir

# Pull latest code
git pull

# Build and start
docker compose -f docker-compose.staging.yml build
docker compose -f docker-compose.staging.yml up -d

# Run migrations
docker compose -f docker-compose.staging.yml exec backend ./api \
  -- migrate

# Verify health
curl https://<staging-domain>/health
curl https://<staging-domain>/health/ready
```

---

## Deploy (Production)

```bash
# SSH into production VM
ssh user@<prod-vm-ip>
cd parkir

# Pull latest code
git pull

# Build and start
docker compose -f docker-compose.prod.yml build --no-cache backend
docker compose -f docker-compose.prod.yml build dashboard
docker compose -f docker-compose.prod.yml up -d

# Run migrations
docker compose -f docker-compose.prod.yml exec backend ./api -- migrate

# Verify health
curl https://<prod-domain>/health
curl https://<prod-domain>/health/ready
```

---

## Rollback

```bash
# If current version has issues, revert to previous Docker image tag:

# 1. Check what the previous image tag was (from docker images)
docker images parkir-backend

# 2. Re-tag and restart
docker tag parkir-backend:<previous-tag> parkir-backend:latest
docker compose -f docker-compose.prod.yml up -d backend

# Or rebuild from a specific git commit
git checkout <previous-tag-or-commit>
docker compose -f docker-compose.prod.yml build backend
docker compose -f docker-compose.prod.yml up -d backend
git checkout main  # restore to latest
```

---

## Database Backup & Restore

### Automatic Backups

- Backups run at 2:00 AM daily via the built-in scheduler in the backend.
- Files are stored in `/var/backups/parkir/` on the VM.
- Retention: 90 days (files older than 90 days are automatically deleted).

### Manual Backup

```bash
# Via API (requires auth)
curl -X POST https://<domain>/api/v1/backups/run

# Or via pg_dump directly
docker compose -f docker-compose.prod.yml exec postgres \
  pg_dump -U parkir parkir | gzip > backup-$(date +%Y%m%d).sql.gz
```

### List Backups

```bash
curl https://<domain>/api/v1/backups
```

### Restore

```bash
# 1. Find the backup file
ls -la /var/backups/parkir/

# 2. Restore using pg_restore
gunzip -c /var/backups/parkir/parkir-20250315-020002.sql.gz | \
  docker compose -f docker-compose.prod.yml exec -T postgres \
  psql -U parkir parkir

# Or
cat /var/backups/parkir/backup-20250315.sql.gz | gunzip | \
  docker compose -f docker-compose.prod.yml exec -T postgres \
  psql -U parkir parkir
```

---

## Desktop Build and Distribution

```bash
# Build desktop app locally
cd desktop
npm install
npm run build

# The built binary/package is in dist/
# Distribute manually (SCP, USB, etc.)
```

---

## Environment Variables Reference

| Variable | Required | Description |
|----------|----------|-------------|
| `ENV` | Yes | `production` or `staging` |
| `PORT` | No | Backend port (default: 8080) |
| `DB_PASSWORD` | Yes | PostgreSQL password |
| `DATABASE_URL` | Yes | Full PostgreSQL connection string |
| `JWT_PRIVATE_KEY_PATH` | Yes | Path to RSA private key (PEM) |
| `JWT_PUBLIC_KEY_PATH` | Yes | Path to RSA public key (PEM) |
| `FRONTEND_URL` | Yes | Dashboard URL for CORS |
| `LOG_LEVEL` | No | `debug`, `info`, `warn`, `error` |
| `BACKUP_DIR` | No | Directory for DB backups |
| `MIGRATIONS_PATH` | No | Path to migration files |
| `API_DOMAIN` | Yes | API domain for nginx config |
| `DASHBOARD_DOMAIN` | Yes | Dashboard domain for nginx config |
| `LOKI_URL` | No | Loki push endpoint (promtail) |

---

## Monitoring Checklist

After a deployment, verify:

- [ ] `GET /health` returns `{"status":"ok"}`
- [ ] `GET /health/ready` returns `{"status":"ok","database":"connected"}`
- [ ] Dashboard loads at `https://<dashboard-domain>`
- [ ] Login works with seed owner credentials
- [ ] Backup API returns list: `GET /api/v1/backups`
- [ ] Logs are visible in Loki (if configured)

---

## Incident Response

### API is down

```bash
# Check container status
docker compose -f docker-compose.prod.yml ps

# View logs
docker compose -f docker-compose.prod.yml logs --tail=50 backend

# Restart
docker compose -f docker-compose.prod.yml restart backend
```

### Database is down

```bash
# Check Postgres logs
docker compose -f docker-compose.prod.yml logs --tail=50 postgres

# Restart
docker compose -f docker-compose.prod.yml restart postgres

# If data corruption: restore from backup (see Backup & Restore section)
```

### Desktop cannot connect

- Verify the API domain is correct in the desktop config
- Check that SSL certificate is valid (`openssl s_client -connect <domain>:443`)
- Verify the desktop can reach the API (ping, curl)
- Check backend logs for auth failures

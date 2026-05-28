# Infrastructure — DigitalOcean Single Droplet

Deployment stack for **restful-template** (Go + PostgreSQL + Redis + Grafana Observability) on a single DigitalOcean Droplet.

**Stack:** Terraform · Ansible · Docker Compose · GitHub Actions · Caddy

```
infra/
├── terraform/              # Provision: Droplet, firewall, reserved IP, domain
│   ├── main.tf
│   ├── variables.tf
│   ├── outputs.tf
│   └── terraform.tfvars.example
├── ansible/                # Server setup: Docker, Caddy, UFW, fail2ban, swap
│   ├── playbook.yml
│   ├── inventory           # Fill in your server IP here
│   ├── ansible.cfg
│   ├── requirements.yml
│   └── roles/
│       ├── base/           # OS hardening: updates, swap, UFW, fail2ban, SSH
│       ├── docker/         # Docker Engine + Compose v2
│       └── caddy/          # Reverse proxy + auto HTTPS (Let's Encrypt)
├── .github/workflows/
│   ├── ci-cd.yml           # CI (lint+test) → Build (GHCR) → Deploy (SSH)
│   └── ansible.yml         # Manual trigger: run Ansible playbook via Actions
├── scripts/
│   ├── deploy.sh           # Manual deploy helper
│   └── init-db.sql         # PostgreSQL extensions init
├── docker-compose.yml      # Production stack (app + db + redis + observability)
├── .env.example            # All required environment variables
├── Makefile                # Shortcut commands
└── README.md
```

---

## Architecture

```
Internet
    │ HTTPS (443) / HTTP (80)
    ▼
 Caddy (reverse proxy + auto TLS)
    │ localhost:8080
    ▼
 restful-template (Go app, distroless)
    │                    │                │
    ▼                    ▼                ▼
 PostgreSQL           Redis          Alloy (OTLP)
 (port 5432)       (port 6379)      (port 4317)
                                        │
                              ┌─────────┼─────────┐
                              ▼         ▼         ▼
                            Tempo     Loki   Prometheus
                              └─────────┼─────────┘
                                        ▼
                                     Grafana
                                   (port 3000)
```

---

## Prerequisites

```bash
# Required tools
brew install terraform ansible

# DigitalOcean CLI (optional but useful)
brew install doctl
doctl auth init
```

---

## Step 1 — Terraform: Provision Droplet

```bash
cd infra/terraform

# Configure your credentials
cp terraform.tfvars.example terraform.tfvars
nano terraform.tfvars   # Set do_token, ssh_key_name, domain

# Preview what will be created
terraform plan

# Create resources (~1 minute)
terraform apply

# Note your server IP
terraform output reserved_ip
```

**What Terraform creates:**
- Ubuntu 24.04 Droplet (`s-1vcpu-2gb` by default)
- Reserved (static) IP address
- Firewall: SSH (22), HTTP (80), HTTPS (443) only
- DigitalOcean Project grouping
- Optional: DNS records for your domain

---

## Step 2 — Ansible: Configure Server

```bash
cd infra/ansible

# Fill in the server IP from Step 1
echo "YOUR_SERVER_IP" > inventory

# Test SSH connectivity
ansible all -i inventory -m ping

# Run full server setup (~5–10 minutes)
ansible-playbook -i inventory playbook.yml

# Run specific roles only
ansible-playbook -i inventory playbook.yml --tags docker
ansible-playbook -i inventory playbook.yml --tags caddy

# With a custom domain (for Caddy auto-HTTPS)
ansible-playbook -i inventory playbook.yml -e "domain=api.example.com"
```

**What Ansible configures:**
- System updates + unattended security upgrades
- Swap (1 GB) + tuned swappiness
- UFW firewall (ports 22, 80, 443)
- fail2ban (SSH brute-force protection)
- SSH hardening (no root login, no password auth)
- Docker Engine + Compose v2
- `deploy` user with Docker group membership
- Caddy reverse proxy with auto HTTPS
- App directory `/home/deploy/app`

---

## Step 3 — First Deploy (Manual)

```bash
# Create production .env
cp infra/.env.example infra/.env
nano infra/.env   # Fill in all secrets

# Get server IP from Terraform
export SERVER_IP=$(cd infra/terraform && terraform output -raw reserved_ip)

# Upload compose file and .env
scp infra/docker-compose.yml infra/.env deploy@$SERVER_IP:~/app/

# Copy config files for observability stack
ssh deploy@$SERVER_IP "mkdir -p ~/app/config"
scp -r config deploy@$SERVER_IP:~/app/

# SSH in and start the stack
ssh deploy@$SERVER_IP
cd ~/app
docker compose up -d

# Verify everything is running
docker compose ps
docker compose logs -f app
```

**Health checks:**
```bash
# App API
curl http://$SERVER_IP/api/v1/health

# Grafana (via SSH tunnel for security)
ssh -L 3000:localhost:3000 deploy@$SERVER_IP
# Then open: http://localhost:3000
```

---

## Step 4 — GitHub Actions: Automated CI/CD

### Required Secrets

Go to: **Settings → Secrets and variables → Actions → Secrets**

| Secret | Description |
|--------|-------------|
| `SSH_PRIVATE_KEY` | Content of `~/.ssh/id_ed25519` |
| `SERVER_IP` | Reserved IP from Terraform output |
| `DOCKERHUB_USERNAME` | Docker Hub username |
| `DOCKERHUB_TOKEN` | Docker Hub access token (not password) |

### Required Variables

Go to: **Settings → Secrets and variables → Actions → Variables**

| Variable | Example |
|----------|---------|
| `APP_URL` | `https://api.example.com` or `http://YOUR_IP` |
| `GRAFANA_ROOT_URL` | `http://localhost:3000` (keep internal) |

### Generate strong secrets:

```bash
# JWT secret (64 bytes)
openssl rand -hex 64

# Database / Redis passwords
openssl rand -base64 32
```

### CI/CD Pipeline

```
push to main
     │
     ▼
CI: golangci-lint + go test ./...
  ├── Services: postgres:18-alpine, redis:7-alpine
     │
     ▼
Build: docker build → push to docker.io/sammidev/restful-template:{sha}
     │
     ▼
Deploy:
  1. Generate .env from secrets
  2. scp docker-compose.yml + .env → server
  3. docker login docker.io
  4. docker compose pull + up -d --no-deps app
  5. docker image prune -f
     │
     ▼
Health check: curl APP_URL/api/v1/health
```

---

## Observability Stack

The production `docker-compose.yml` ships with a full Grafana OSS observability stack:

| Service | Port | Purpose |
|---------|------|---------|
| **Grafana** | `3000` | Dashboards (metrics, logs, traces) |
| **Prometheus** | `9090` | Metrics collection & storage |
| **Loki** | `3100` | Log aggregation |
| **Tempo** | `3200` | Distributed tracing |
| **Alloy** | `4317/4318` | OTLP collector (app → Tempo + Loki) |

### Data Flow

```
App (TELEMETRY_OTLP_ENDPOINT=alloy:4317)
    │
    ▼
Alloy
  ├── traces  → Tempo
  └── logs    → Loki  (via Docker socket discovery)

Prometheus → scrapes metrics from app (port 8080)
Tempo      → writes span metrics → Prometheus

Grafana → queries all four backends
```

### Accessing Grafana

All observability ports are bound to `127.0.0.1` (localhost only). Access them via SSH tunnel:

```bash
# Grafana
ssh -L 3000:localhost:3000 deploy@$SERVER_IP
open http://localhost:3000
# Login with GRAFANA_ADMIN_USER / GRAFANA_ADMIN_PASSWORD

# Prometheus
ssh -L 9090:localhost:9090 deploy@$SERVER_IP
open http://localhost:9090

# Alloy UI
ssh -L 12345:localhost:12345 deploy@$SERVER_IP
open http://localhost:12345
```

### Optional: Expose Grafana via Caddy

Add to your `Caddyfile` (or configure via Ansible `-e` variables):

```
grafana.example.com {
    reverse_proxy localhost:3000
    basicauth /* {
        admin $GRAFANA_HTPASSWD_HASH
    }
}
```

---

## Useful Commands (via Makefile)

```bash
# From infra/ directory:

make tf-plan          # Preview Terraform changes
make tf-apply         # Apply infrastructure
make tf-output        # Show Terraform outputs (IP, etc.)

make ansible-deps     # Install Ansible Galaxy collections
make ansible-ping     # Test SSH connectivity
make ansible-setup    # Run full server setup

make deploy           # Manual deploy (latest image)
make deploy-tag TAG=sha-abc123  # Deploy specific image

make ssh              # SSH into server
make logs             # Stream app logs
make logs-all         # Stream all service logs
make status           # Show container status
make restart          # Restart app container

make db-backup        # Backup database to local file
make db-shell         # Open PostgreSQL shell
```

---

## Droplet Sizing Guide

| Use case | Size | Price/mo | Notes |
|----------|------|----------|-------|
| Dev / staging | `s-1vcpu-1gb` | ~$6 | No monitoring stack |
| **Production** | `s-1vcpu-2gb` | ~$12 | ✅ Recommended baseline |
| Production + observability | `s-2vcpu-4gb` | ~$24 | Full Grafana stack |
| High traffic | `s-2vcpu-8gb` | ~$48 | Heavy workloads |

> **Note:** The full observability stack (Prometheus + Grafana + Loki + Tempo + Alloy) uses ~500MB RAM. Use `s-2vcpu-4gb` or larger when running all services on the same droplet.

---

## Troubleshooting

**Caddy cannot obtain SSL certificate:**
```bash
# Check domain DNS is pointing to server IP
dig +short api.example.com

# Ensure ports 80 and 443 are open
ufw status

# Check Caddy logs
journalctl -u caddy -f
```

**App container keeps restarting:**
```bash
# Check container logs
docker compose logs app

# Verify all env vars are set in .env
docker compose config | grep -A 30 "environment:"

# Check db/redis are healthy
docker compose ps
```

**Database connection refused:**
```bash
# Confirm postgres is running and healthy
docker compose ps db
docker compose exec db pg_isready -U todo -d todo
```

**Grafana shows "No data":**
```bash
# Verify Alloy is receiving traces
curl http://localhost:12345/metrics | grep otelcol

# Check Prometheus targets
curl http://localhost:9090/targets

# Verify app is sending to Alloy
docker compose logs alloy | tail -50
```

**SSH denied after Ansible:**
```bash
# Ansible disables root login and password auth
# Use the deploy user with your SSH key:
ssh -i ~/.ssh/id_ed25519 deploy@SERVER_IP
```

terraform {
  required_version = ">= 1.5"
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

provider "digitalocean" {
  token = var.do_token
}

# ─────────────────────────────────────────
# SSH Key
# ─────────────────────────────────────────
data "digitalocean_ssh_key" "deploy" {
  name = var.ssh_key_name
}

# ─────────────────────────────────────────
# Droplet
# ─────────────────────────────────────────
resource "digitalocean_droplet" "app" {
  name     = "${var.project_name}-${var.environment}"
  region   = var.region
  size     = var.droplet_size
  image    = "ubuntu-24-04-x64"
  ssh_keys = [data.digitalocean_ssh_key.deploy.fingerprint]

  tags = [var.project_name, var.environment]

  lifecycle {
    # Prevent accidental destruction in production
    prevent_destroy = false
  }
}

# ─────────────────────────────────────────
# Reserved IP (statis, tidak berubah saat recreate)
# ─────────────────────────────────────────
resource "digitalocean_reserved_ip" "app" {
  region = var.region
}

resource "digitalocean_reserved_ip_assignment" "app" {
  ip_address = digitalocean_reserved_ip.app.ip_address
  droplet_id = digitalocean_droplet.app.id
}

# ─────────────────────────────────────────
# Firewall — hanya buka port yang diperlukan
# ─────────────────────────────────────────
resource "digitalocean_firewall" "app" {
  name        = "${var.project_name}-${var.environment}-fw"
  droplet_ids = [digitalocean_droplet.app.id]

  # Inbound: SSH
  inbound_rule {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = var.ssh_allowed_ips
  }

  # Inbound: HTTP (Caddy redirect ke HTTPS)
  inbound_rule {
    protocol         = "tcp"
    port_range       = "80"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  # Inbound: HTTPS
  inbound_rule {
    protocol         = "tcp"
    port_range       = "443"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  # Outbound: semua diizinkan
  outbound_rule {
    protocol              = "tcp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "udp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "icmp"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}

# ─────────────────────────────────────────
# Domain & DNS (opsional)
# ─────────────────────────────────────────
resource "digitalocean_domain" "app" {
  count      = var.domain != "" ? 1 : 0
  name       = var.domain
  ip_address = digitalocean_reserved_ip.app.ip_address
}

resource "digitalocean_record" "www" {
  count  = var.domain != "" ? 1 : 0
  domain = digitalocean_domain.app[0].name
  type   = "A"
  name   = "www"
  value  = digitalocean_reserved_ip.app.ip_address
  ttl    = 300
}

# ─────────────────────────────────────────
# Project grouping di DO dashboard
# ─────────────────────────────────────────
resource "digitalocean_project" "app" {
  name        = "${var.project_name}-${var.environment}"
  description = "Infrastructure for ${var.project_name} (${var.environment})"
  purpose     = "Web Application"
  environment = title(var.environment)

  resources = [
    digitalocean_droplet.app.urn,
    digitalocean_reserved_ip.app.urn,
  ]
}

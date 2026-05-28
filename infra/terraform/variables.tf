variable "do_token" {
  description = "DigitalOcean API token"
  type        = string
  sensitive   = true
}

variable "project_name" {
  description = "Project name (used for naming all resources)"
  type        = string
  default     = "restful-template"
}

variable "environment" {
  description = "Environment: production / staging"
  type        = string
  default     = "production"

  validation {
    condition     = contains(["production", "staging", "dev"], var.environment)
    error_message = "Environment harus: production, staging, atau dev."
  }
}

variable "region" {
  description = "DigitalOcean region. Lihat: doctl compute region list"
  type        = string
  default     = "sgp1" # Singapore — terdekat untuk Indonesia

  # Opsi populer: sgp1 (Singapore), blr1 (Bangalore), nyc3, ams3
}

variable "droplet_size" {
  description = "Ukuran Droplet. Lihat: doctl compute size list"
  type        = string
  default     = "s-1vcpu-2gb" # ~$12/bln — recommended untuk production + PostgreSQL

  # s-1vcpu-1gb  = ~$6/bln  (dev/staging only)
  # s-1vcpu-2gb  = ~$12/bln (production ringan)
  # s-2vcpu-4gb  = ~$24/bln (production + monitoring)
}

variable "ssh_key_name" {
  description = "Nama SSH key yang sudah di-upload ke DigitalOcean"
  type        = string
}

variable "ssh_allowed_ips" {
  description = "IP yang diizinkan SSH. Default: semua (tidak recommended untuk production)"
  type        = list(string)
  default     = ["0.0.0.0/0", "::/0"]
  # Ganti dengan IP spesifik: ["203.0.113.0/32"]
}

variable "domain" {
  description = "Domain name (opsional). Kosongkan jika belum punya domain."
  type        = string
  default     = ""
  # Contoh: "example.com"
}

variable "create_project" {
  description = "Jika project sudah ada, set false agar Terraform tidak mencoba membuat ulang dan menyebabkan error."
  type        = bool
  default     = false
}

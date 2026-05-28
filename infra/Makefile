# Makefile — Shortcut commands untuk infra
# Usage: make <target>

.PHONY: help tf-plan tf-apply tf-destroy ansible-ping ansible-setup deploy ssh logs

# Ambil IP dari Terraform output secara otomatis
SERVER_IP := $(shell cd terraform && terraform output -raw reserved_ip 2>/dev/null || echo "")
SSH_CMD   := ssh deploy@$(SERVER_IP)

help: ## Tampilkan semua commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ─────────────────────────────────────────
# Terraform
# ─────────────────────────────────────────
tf-plan: ## Preview perubahan infrastruktur
	cd terraform && terraform plan

tf-apply: ## Apply — buat/update infrastruktur
	cd terraform && terraform apply

tf-destroy: ## HATI-HATI: Hapus semua infrastruktur
	cd terraform && terraform destroy

tf-output: ## Tampilkan output Terraform (IP, dll)
	cd terraform && terraform output

# ─────────────────────────────────────────
# Ansible
# ─────────────────────────────────────────
ansible-deps: ## Install Ansible Galaxy collections
	cd ansible && ansible-galaxy collection install -r requirements.yml

ansible-ping: ## Cek koneksi SSH ke server
	cd ansible && ansible all -i inventory -m ping

ansible-setup: ## Jalankan full server setup
	cd ansible && ansible-playbook -i inventory playbook.yml

ansible-setup-check: ## Dry-run setup (tidak ada perubahan)
	cd ansible && ansible-playbook -i inventory playbook.yml --check

# ─────────────────────────────────────────
# Deploy
# ─────────────────────────────────────────
deploy: ## Deploy manual ke production
	@if [ -z "$(SERVER_IP)" ]; then echo "❌ SERVER_IP kosong. Jalankan terraform apply dulu."; exit 1; fi
	SERVER_IP=$(SERVER_IP) bash scripts/deploy.sh

deploy-tag: ## Deploy image spesifik: make deploy-tag TAG=sha-abc123
	@if [ -z "$(TAG)" ]; then echo "❌ Tentukan TAG: make deploy-tag TAG=sha-abc123"; exit 1; fi
	SERVER_IP=$(SERVER_IP) bash scripts/deploy.sh $(TAG)

# ─────────────────────────────────────────
# Operations
# ─────────────────────────────────────────
ssh: ## SSH ke server
	$(SSH_CMD)

logs: ## Lihat logs app secara live
	$(SSH_CMD) "cd ~/app && docker compose logs -f app"

logs-all: ## Lihat semua logs (app + db + redis)
	$(SSH_CMD) "cd ~/app && docker compose logs -f"

status: ## Status containers di server
	$(SSH_CMD) "cd ~/app && docker compose ps"

restart: ## Restart app container
	$(SSH_CMD) "cd ~/app && docker compose restart app"

db-backup: ## Backup database ke lokal
	$(SSH_CMD) "docker compose -f ~/app/docker-compose.yml exec -T db \
		pg_dump -U \$$POSTGRES_USER \$$POSTGRES_DB" > backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "✅ Backup tersimpan"

db-shell: ## Buka PostgreSQL shell
	$(SSH_CMD) "cd ~/app && docker compose exec db psql -U \$$POSTGRES_USER \$$POSTGRES_DB"

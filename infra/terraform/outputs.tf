output "droplet_id" {
  description = "Droplet ID"
  value       = digitalocean_droplet.app.id
}

output "droplet_name" {
  description = "Droplet name"
  value       = digitalocean_droplet.app.name
}

output "reserved_ip" {
  description = "Reserved (static) IP address"
  value       = digitalocean_reserved_ip.app.ip_address
}

output "droplet_ip" {
  description = "Droplet IP (same as Reserved IP after assignment)"
  value       = digitalocean_reserved_ip_assignment.app.ip_address
}

output "domain" {
  description = "Domain configured (if any)"
  value       = var.domain != "" ? digitalocean_domain.app[0].name : "No domain configured"
}

output "ssh_command" {
  description = "SSH command to access the server"
  value       = "ssh deploy@${digitalocean_reserved_ip.app.ip_address}"
}

output "project_name" {
  description = "DigitalOcean project name"
  value       = digitalocean_project.app.name
}

output "next_steps" {
  description = "Next steps after terraform apply"
  value       = <<-EOT
    ✅ Terraform complete!

    Server IP: ${digitalocean_reserved_ip.app.ip_address}

    1. Update Ansible inventory:
       echo "${digitalocean_reserved_ip.app.ip_address}" > ../ansible/inventory

    2. Run Ansible server setup:
       cd ../ansible && ansible-playbook -i inventory playbook.yml

    3. Copy files to server:
       scp ../docker-compose.yml ../.env deploy@${digitalocean_reserved_ip.app.ip_address}:~/app/

    4. Copy observability config:
       scp -r ../../config deploy@${digitalocean_reserved_ip.app.ip_address}:~/app/

    5. Start the stack:
       ssh deploy@${digitalocean_reserved_ip.app.ip_address} "cd ~/app && docker compose up -d"

    6. Health check:
       curl http://${digitalocean_reserved_ip.app.ip_address}/api/v1/health
  EOT
}

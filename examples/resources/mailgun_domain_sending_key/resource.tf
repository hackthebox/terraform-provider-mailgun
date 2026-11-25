# Create a domain sending key for an application
resource "mailgun_domain_sending_key" "app" {
  domain      = "mail.example.com"
  description = "Production application key"
}

# Store the secret in Vault
resource "vault_generic_secret" "mailgun_key" {
  path = "secret/mailgun/app"

  data_json = jsonencode({
    api_key = mailgun_domain_sending_key.app.secret
  })
}

# Create a sending key with expiration
resource "mailgun_domain_sending_key" "temporary" {
  domain      = "mail.example.com"
  description = "Temporary key for migration"
  expiration  = 86400 # 24 hours in seconds
}

# Rotate without downtime: bump `rotation` to replace the key. The new key is
# created and written to Vault first; the old one is deleted 5 minutes later.
variable "rotation" {
  type    = number
  default = 1
}

resource "terraform_data" "rotation" {
  input = var.rotation
}

resource "mailgun_domain_sending_key" "rotated" {
  domain         = "mail.example.com"
  description    = "Rotated application key"
  deletion_delay = 300

  lifecycle {
    create_before_destroy = true
    replace_triggered_by  = [terraform_data.rotation]
  }
}

output "sending_key_id" {
  value = mailgun_domain_sending_key.app.id
}

# Note: The secret is only available after creation and cannot be retrieved later.
# Make sure to store it securely (e.g., in HashiCorp Vault) immediately.
output "sending_key_secret" {
  value     = mailgun_domain_sending_key.app.secret
  sensitive = true
}

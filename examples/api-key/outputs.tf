output "api_key_id" {
  description = "The ID of the created API key"
  value       = mailgun_api_key.sending.id
}

output "api_key_secret" {
  description = "The API key secret (only available after creation)"
  value       = mailgun_api_key.sending.secret
  sensitive   = true
}

output "api_key_created_at" {
  description = "When the API key was created"
  value       = mailgun_api_key.sending.created_at
}

output "total_api_keys" {
  description = "Total number of API keys in the account"
  value       = data.mailgun_api_keys.all.total_count
}

output "domain_api_keys_count" {
  description = "Number of domain-specific API keys"
  value       = data.mailgun_api_keys.domain_keys.total_count
}

variable "mailgun_api_key" {
  description = "Mailgun API key"
  type        = string
  sensitive   = true
}

variable "mailgun_region" {
  description = "Mailgun region (US or EU)"
  type        = string
  default     = "US"
}

variable "domain_name" {
  description = "Domain name for sending email"
  type        = string
}

variable "smtp_login" {
  description = "SMTP login name (without @domain)"
  type        = string
  default     = "notifications"
}

variable "app_name" {
  description = "Application name (used for Vault path)"
  type        = string
  default     = "myapp"
}

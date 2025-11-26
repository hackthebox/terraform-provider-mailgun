variable "mailgun_api_key" {
  description = "Mailgun API key (must be an admin key to create other keys)"
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

variable "app_name" {
  description = "Application name (used for key description and Vault path)"
  type        = string
  default     = "myapp"
}

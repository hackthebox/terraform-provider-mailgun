# List all templates for a domain
data "mailgun_templates" "all" {
  domain = "mail.example.com"
}

# Output template count
output "total_templates" {
  value = data.mailgun_templates.all.total_count
}

# Output all template names
output "template_names" {
  value = [for t in data.mailgun_templates.all.templates : t.name]
}

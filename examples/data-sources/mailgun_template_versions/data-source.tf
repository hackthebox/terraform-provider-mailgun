# List all versions of a template
data "mailgun_template_versions" "welcome_versions" {
  domain        = "mail.example.com"
  template_name = "welcome-email"
}

# Output version count
output "total_versions" {
  value = data.mailgun_template_versions.welcome_versions.total_count
}

# Output all version tags
output "version_tags" {
  value = [for v in data.mailgun_template_versions.welcome_versions.versions : v.tag]
}

# Find the active version
output "active_version" {
  value = [for v in data.mailgun_template_versions.welcome_versions.versions : v.tag if v.active][0]
}

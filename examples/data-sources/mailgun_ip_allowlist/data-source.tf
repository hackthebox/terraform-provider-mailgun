# List all IP allowlist entries
data "mailgun_ip_allowlist" "all" {}

# Output all allowlisted IPs
output "allowlisted_ips" {
  value = [for entry in data.mailgun_ip_allowlist.all.entries : entry.address]
}

# Output entries with descriptions
output "allowlist_details" {
  value = {
    for entry in data.mailgun_ip_allowlist.all.entries :
    entry.address => entry.description
  }
}

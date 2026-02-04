# List all subaccounts
data "mailgun_subaccounts" "all" {}

# List only enabled subaccounts
data "mailgun_subaccounts" "enabled_only" {
  enabled = true
}

# Output the total number of subaccounts
output "total_subaccounts" {
  value = data.mailgun_subaccounts.all.total_count
}

# Output all enabled subaccount names
output "enabled_subaccount_names" {
  value = [for sa in data.mailgun_subaccounts.enabled_only.subaccounts : sa.name]
}

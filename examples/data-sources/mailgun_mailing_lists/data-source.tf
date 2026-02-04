# List all mailing lists
data "mailgun_mailing_lists" "all" {}

# Output mailing list count
output "total_mailing_lists" {
  value = data.mailgun_mailing_lists.all.total_count
}

# Output all mailing list addresses
output "mailing_list_addresses" {
  value = [for ml in data.mailgun_mailing_lists.all.lists : ml.address]
}

# List all members of a mailing list
data "mailgun_mailing_list_members" "developers" {
  list_address = "developers@lists.example.com"
}

# Output member count
output "total_members" {
  value = data.mailgun_mailing_list_members.developers.total_count
}

# Output all member email addresses
output "member_emails" {
  value = [for m in data.mailgun_mailing_list_members.developers.members : m.address]
}

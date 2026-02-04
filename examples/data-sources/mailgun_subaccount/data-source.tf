# Look up a subaccount by ID
data "mailgun_subaccount" "example" {
  id = "12345678-1234-5678-1234-123456789abc"
}

# Use the subaccount information
output "subaccount_name" {
  value = data.mailgun_subaccount.example.name
}

output "subaccount_status" {
  value = data.mailgun_subaccount.example.status
}

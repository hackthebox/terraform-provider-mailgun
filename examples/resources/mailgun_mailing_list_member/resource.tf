# Add a member to a mailing list
resource "mailgun_mailing_list_member" "member" {
  list_address = "developers@lists.example.com"
  address      = "john.doe@example.com"
  name         = "John Doe"
  subscribed   = true
}

# Add member with custom variables
resource "mailgun_mailing_list_member" "member_with_vars" {
  list_address = "newsletter@lists.example.com"
  address      = "jane.doe@example.com"
  name         = "Jane Doe"
  subscribed   = true
  vars = jsonencode({
    first_name = "Jane"
    last_name  = "Doe"
    company    = "Example Inc"
  })
}

# Reference existing mailing list
resource "mailgun_mailing_list" "team" {
  address = "team@lists.example.com"
  name    = "Team List"
}

resource "mailgun_mailing_list_member" "team_member" {
  list_address = mailgun_mailing_list.team.address
  address      = "member@example.com"
  name         = "Team Member"
  subscribed   = true
}

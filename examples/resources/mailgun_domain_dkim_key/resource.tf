# Create a DKIM key for a domain
resource "mailgun_domain_dkim_key" "example" {
  domain   = "mail.example.com"
  selector = "mailo1"
  key_size = 2048
}

# Reference an existing domain
resource "mailgun_domain" "example" {
  name = "mail.example.com"
}

resource "mailgun_domain_dkim_key" "with_domain" {
  domain   = mailgun_domain.example.name
  selector = "mailo1"
  key_size = 2048
}

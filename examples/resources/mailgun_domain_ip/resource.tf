# Assign a dedicated IP to a domain
resource "mailgun_domain_ip" "example" {
  domain = "mail.example.com"
  ip     = "192.168.1.100"
}

# With domain reference
resource "mailgun_domain" "example" {
  name = "mail.example.com"
}

resource "mailgun_domain_ip" "with_domain" {
  domain = mailgun_domain.example.name
  ip     = "192.168.1.100"
}

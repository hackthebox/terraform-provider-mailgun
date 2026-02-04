# Basic mailing list
resource "mailgun_mailing_list" "developers" {
  address      = "developers@lists.example.com"
  name         = "Developer Team"
  description  = "Internal developer mailing list"
  access_level = "members"
}

# Mailing list with custom settings
resource "mailgun_mailing_list" "newsletter" {
  address          = "newsletter@lists.example.com"
  name             = "Company Newsletter"
  description      = "Public newsletter subscription"
  access_level     = "readonly"
  reply_preference = "sender"
}

# Announcement list (only admins can post)
resource "mailgun_mailing_list" "announcements" {
  address      = "announcements@lists.example.com"
  name         = "Company Announcements"
  access_level = "readonly"
}

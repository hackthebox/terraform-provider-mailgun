# Import an IP allowlist entry using its IP address
terraform import mailgun_ip_allowlist.office 203.0.113.50

# Import a CIDR range entry
terraform import mailgun_ip_allowlist.vpn 198.51.100.0/24

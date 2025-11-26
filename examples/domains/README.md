# Mailgun Domains Example

This example demonstrates how to manage Mailgun domains using the Terraform provider, including:

- Creating domains with basic configuration
- Creating domains with advanced configuration
- Using data sources to list and filter domains
- Importing existing domains into Terraform state

## Prerequisites

- Terraform >= 1.0
- A Mailgun account with an API key
- Domain(s) you want to manage
- Provider installed locally (run `make install` from the repository root)

## Usage

### 1. Setup

Copy the example variables file and fill in your values:

```bash
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars` and set your Mailgun API key and domain names.

### 2. Plan and Apply

Review the changes:

```bash
terraform plan
```

Apply the configuration:

```bash
terraform apply
```

### 3. View Outputs

After applying, you can view the outputs:

```bash
terraform output
```

To view DNS records you need to configure:

```bash
terraform output basic_domain_receiving_dns_records
terraform output basic_domain_sending_dns_records
```

### 4. Import Existing Domains

If you have an existing domain in Mailgun that you want to manage with Terraform, you can import it:

```bash
# Import syntax: terraform import <resource_address> <domain_name>
terraform import mailgun_domain.imported_example existing-domain.com
```

After importing, add the domain to your Terraform configuration:

```hcl
resource "mailgun_domain" "imported_example" {
  name        = "existing-domain.com"
  spam_action = "disabled"
  wildcard    = false
}
```

Then run `terraform plan` to see if any changes are needed to match your configuration.

### 5. Destroy Resources

When you're done, you can destroy all created resources:

```bash
terraform destroy
```

**Note**: This will permanently delete the domains from Mailgun!

## Features Demonstrated

### Basic Domain Creation

The `mailgun_domain.example` resource demonstrates basic domain creation with minimal configuration:

- Domain name
- Spam action setting
- Wildcard subdomain support

### Advanced Domain Configuration

The `mailgun_domain.advanced` resource demonstrates advanced features:

- Custom spam action (tag)
- Wildcard subdomain enabled
- Automatic sender security (SPF/DKIM/DMARC)
- Force DKIM authority
- Custom DKIM key size (2048 bits)
- Custom web scheme for tracking (HTTPS)

### Data Sources

Three data source examples show how to:

1. **Lookup a single domain**: Retrieve information about a specific domain by name (`mailgun_domain`)
2. **List all domains**: Retrieve all domains in your account with a configurable limit (`mailgun_domains`)
3. **Filter domains**: Search for specific domains using the search parameter (`mailgun_domains` with filters)

**When to use which data source:**

- Use `data "mailgun_domain"` when:
  - You need to lookup ONE specific domain by its exact name
  - You want to reference an existing domain's attributes in your configuration
  - Example: `data.mailgun_domain.lookup.state`

- Use `data "mailgun_domains"` when:
  - You need to list MULTIPLE domains
  - You want to iterate over all domains in your account
  - You need to filter or search across domains
  - Example: `data.mailgun_domains.all.items[*].name`

## DNS Configuration

After creating a domain, you need to configure DNS records with your DNS provider. The provider outputs the required DNS records:

- **Receiving DNS records**: MX records for receiving email
- **Sending DNS records**: SPF, DKIM, and other records for sending email

You can view these with:

```bash
terraform output -json basic_domain_receiving_dns_records
terraform output -json basic_domain_sending_dns_records
```

## Importing Existing Domains

The import workflow allows you to bring existing Mailgun domains under Terraform management:

1. **Import the domain** into Terraform state using its domain name as the identifier
2. **Add the resource** to your Terraform configuration matching the imported domain
3. **Run `terraform plan`** to verify the state matches your configuration
4. **Adjust your configuration** if needed to match the actual state

Example import commands:

```bash
# Import a single domain
terraform import mailgun_domain.my_domain example.com

# Import multiple domains
terraform import mailgun_domain.domain1 domain1.com
terraform import mailgun_domain.domain2 domain2.com
terraform import mailgun_domain.domain3 domain3.com
```

## Common Operations

### Create a domain

```bash
terraform apply
```

### Read domain information

```bash
terraform show
terraform output
```

### Update a domain

Edit the domain configuration in `main.tf` (note: some attributes like `name` require domain recreation), then:

```bash
terraform apply
```

### Delete a domain

Remove the domain resource from your configuration or run:

```bash
terraform destroy -target=mailgun_domain.example
```

## Important Notes

1. **Domain names** must be unique and properly formatted
2. **DNS propagation** can take time - your domain won't work immediately for sending/receiving email until DNS is properly configured
3. **Wildcard domains** accept email for all subdomains
4. **Spam action** determines what happens to spam emails:
   - `disabled`: No spam filtering
   - `tag`: Tag spam but deliver it
   - `delete`: Delete spam emails
5. **Import limitations**: Some create-only fields (like `smtp_password`, `dkim_key_size`) won't be imported and should be set in your configuration with the `lifecycle { ignore_changes = [...] }` block if they differ from defaults

## Resources

- [Mailgun API Documentation](https://documentation.mailgun.com/en/latest/api_reference.html)
- [Terraform Mailgun Provider Registry](https://registry.terraform.io/providers/hackthebox/mailgun/latest)

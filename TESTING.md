# Testing the Mailgun Terraform Provider

This guide explains how to test the provider locally.

## Step 1: Build and Install the Provider Locally

From the repository root:

```bash
# Build and install the provider
make install
```

This installs the provider to your local Terraform plugin directory.

## Step 2: Quick Test

The fastest way to test is to use the complete example:

```bash
# Navigate to the example
cd examples/complete-setup

# Copy and edit variables
cp terraform.tfvars.example terraform.tfvars
vim terraform.tfvars  # Add your Mailgun API key and domain

# Initialize and apply
terraform init
terraform plan
terraform apply
```

## Step 3: Verify Resources

After applying, check the Mailgun dashboard at https://app.mailgun.com/ to verify:
- Domain was created
- Webhooks are configured
- SMTP credentials exist
- Route is active

## Step 4: View Outputs

```bash
# See all outputs
terraform output

# View sensitive outputs individually
terraform output -raw smtp_password
terraform output -raw api_key_secret
```

## Step 5: Cleanup

```bash
terraform destroy
```

## Detailed Testing Guide

See the complete example README at `examples/complete-setup/README.md` for:
- Prerequisites
- Detailed usage instructions
- Testing individual resources
- Import examples
- Troubleshooting tips

## Running Tests

```bash
# Unit tests
make test

# Build verification
make build

# Full acceptance tests (requires MAILGUN_API_KEY env var)
make testacc
```

## IP Allowlist Tests

The `mailgun_ip_allowlist` resource acceptance tests automatically manage the test runner's IP:

1. **Before each test**: The test runner's public IP is detected and added to the Mailgun IP allowlist
2. **During test**: Tests safely add/remove RFC 5737 documentation IPs (192.0.2.X)
3. **After test**: The test runner's IP is automatically removed via `t.Cleanup()`

This ensures tests work from any environment (local development, GitHub Actions, self-hosted runners) without manual setup.

**Note**: If the test runner's IP is already in the allowlist (manually added), it will be left untouched and not removed after the test.

#!/bin/bash
# Copyright Hack The Box 2025, 2026
# SPDX-License-Identifier: MPL-2.0

#
# Mailgun Workspace Audit Script
# READ-ONLY - Only GET operations, no modifications
#
# Usage:
#   export MAILGUN_API_KEY="your-api-key"
#   ./mailgun-audit.sh

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
AUDIT_DIR="${REPO_ROOT}/audit"

# API endpoints
US_API="https://api.mailgun.net"
EU_API="https://api.eu.mailgun.net"

# Check for API key
if [ -z "${MAILGUN_API_KEY:-}" ]; then
    echo "Error: MAILGUN_API_KEY environment variable is not set"
    exit 1
fi

# Check for jq
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed"
    exit 1
fi

# Create output directory
mkdir -p "$AUDIT_DIR"
echo "Output directory: $AUDIT_DIR"

# Make API request (GET only)
api_get() {
    local url="$1"
    curl -s -u "api:${MAILGUN_API_KEY}" "$url" 2>/dev/null
}

# Audit a region
audit_region() {
    local region="$1"
    local api_base="$2"
    local REGION_UPPER=$(echo "$region" | tr '[:lower:]' '[:upper:]')

    echo ""
    echo "========================================"
    echo "Auditing $REGION_UPPER region ($api_base)"
    echo "========================================"

    # Test connectivity
    echo "Testing API connectivity..."
    local test=$(api_get "${api_base}/v4/domains?limit=1")
    if ! echo "$test" | jq -e '.items' &>/dev/null; then
        echo "Failed to connect to $REGION_UPPER API"
        echo "Response: $test"
        return 1
    fi
    echo "API connection successful"

    # Fetch domains
    echo ""
    echo "Fetching domains..."
    local domains_response=$(api_get "${api_base}/v4/domains")
    local domains=$(echo "$domains_response" | jq '.items')
    local domain_count=$(echo "$domains" | jq 'length')
    echo "Found $domain_count domains"

    # Save domains list
    echo "$domains" | jq '.' > "${AUDIT_DIR}/${region}_domains.json"

    # Fetch domain details for each domain
    echo ""
    echo "Fetching domain details..."
    local domain_details="[]"
    for domain_name in $(echo "$domains" | jq -r '.[].name'); do
        echo "  - $domain_name"
        local detail=$(api_get "${api_base}/v4/domains/${domain_name}")
        domain_details=$(echo "$domain_details" | jq --argjson d "$detail" '. + [$d]')
    done
    echo "$domain_details" | jq '.' > "${AUDIT_DIR}/${region}_domains_detail.json"

    # Fetch SMTP credentials for each domain (uses v3 API with pagination)
    echo ""
    echo "Fetching SMTP credentials..."
    local all_creds="{}"
    for domain_name in $(echo "$domains" | jq -r '.[].name'); do
        echo "  - $domain_name"
        local domain_creds="[]"
        local skip=0
        local limit=100
        while true; do
            local creds_response=$(api_get "${api_base}/v3/domains/${domain_name}/credentials?limit=${limit}&skip=${skip}")
            local page_items=$(echo "$creds_response" | jq '.items // []')
            local page_count=$(echo "$page_items" | jq 'length')
            if [ "$page_count" -eq 0 ]; then
                break
            fi
            domain_creds=$(echo "$domain_creds" | jq --argjson items "$page_items" '. + $items')
            if [ "$page_count" -lt "$limit" ]; then
                break
            fi
            skip=$((skip + limit))
        done
        all_creds=$(echo "$all_creds" | jq --arg d "$domain_name" --argjson c "$domain_creds" '. + {($d): $c}')
    done
    echo "$all_creds" | jq '.' > "${AUDIT_DIR}/${region}_credentials.json"
    local total_creds=$(echo "$all_creds" | jq '[.[] | length] | add // 0')
    echo "Found $total_creds SMTP credentials"

    # Fetch domain sending keys (uses /v1/keys with kind=domain filter)
    echo ""
    echo "Fetching domain sending keys..."
    local all_keys="{}"
    for domain_name in $(echo "$domains" | jq -r '.[].name'); do
        echo "  - $domain_name"
        # Use the /v1/keys endpoint with domain_name and kind=domain filters
        local keys_response=$(api_get "${api_base}/v1/keys?domain_name=${domain_name}&kind=domain")
        # Filter for sending keys only (role=sending)
        local keys=$(echo "$keys_response" | jq '[.items // [] | .[] | select(.role == "sending")]')
        all_keys=$(echo "$all_keys" | jq --arg d "$domain_name" --argjson k "$keys" '. + {($d): $k}')
    done
    echo "$all_keys" | jq '.' > "${AUDIT_DIR}/${region}_sending_keys.json"
    local total_keys=$(echo "$all_keys" | jq '[.[] | length] | add // 0')
    echo "Found $total_keys sending keys"

    # Fetch webhooks for each domain
    echo ""
    echo "Fetching webhooks..."
    local all_webhooks="{}"
    for domain_name in $(echo "$domains" | jq -r '.[].name'); do
        local webhooks=$(api_get "${api_base}/v4/domains/${domain_name}/webhooks")
        all_webhooks=$(echo "$all_webhooks" | jq --arg d "$domain_name" --argjson w "$webhooks" '. + {($d): $w}')
    done
    echo "$all_webhooks" | jq '.' > "${AUDIT_DIR}/${region}_webhooks.json"

    # Fetch routes
    echo ""
    echo "Fetching routes..."
    local routes_response=$(api_get "${api_base}/v4/routes")
    local routes=$(echo "$routes_response" | jq '.items // []')
    echo "$routes" | jq '.' > "${AUDIT_DIR}/${region}_routes.json"
    local route_count=$(echo "$routes" | jq 'length')
    echo "Found $route_count routes"

    return 0
}

# Generate summary
generate_summary() {
    echo ""
    echo "========================================"
    echo "Generating Summary"
    echo "========================================"

    local summary_file="${AUDIT_DIR}/summary.txt"

    {
        echo "Mailgun Workspace Audit Summary"
        echo "Generated: $(date)"
        echo "============================================"
        echo ""

        for region in us eu; do
            local REGION_UPPER=$(echo "$region" | tr '[:lower:]' '[:upper:]')
            echo "== $REGION_UPPER REGION =="

            local domain_file="${AUDIT_DIR}/${region}_domains.json"
            if [ -f "$domain_file" ]; then
                local domain_count=$(jq 'length' "$domain_file")
                local active_count=$(jq '[.[] | select(.state == "active")] | length' "$domain_file")
                local disabled_count=$(jq '[.[] | select(.is_disabled == true)] | length' "$domain_file")
                echo "Domains: $domain_count total (active: $active_count, disabled: $disabled_count)"
            else
                echo "Domains: N/A"
            fi

            local creds_file="${AUDIT_DIR}/${region}_credentials.json"
            if [ -f "$creds_file" ]; then
                local creds_count=$(jq '[.[] | length] | add // 0' "$creds_file")
                echo "SMTP Credentials: $creds_count"
            fi

            local keys_file="${AUDIT_DIR}/${region}_sending_keys.json"
            if [ -f "$keys_file" ]; then
                local keys_count=$(jq '[.[] | length] | add // 0' "$keys_file")
                echo "Domain Sending Keys: $keys_count"
            fi

            local routes_file="${AUDIT_DIR}/${region}_routes.json"
            if [ -f "$routes_file" ]; then
                local routes_count=$(jq 'length' "$routes_file")
                echo "Routes: $routes_count"
            fi

            echo ""
        done

        # List all domains
        echo "== ALL DOMAINS =="
        echo ""
        echo "US Region:"
        if [ -f "${AUDIT_DIR}/us_domains.json" ]; then
            jq -r '.[] | "  - \(.name) [\(.state)] (created: \(.created_at))"' "${AUDIT_DIR}/us_domains.json"
        fi
        echo ""
        echo "EU Region:"
        if [ -f "${AUDIT_DIR}/eu_domains.json" ]; then
            jq -r '.[] | "  - \(.name) [\(.state)] (created: \(.created_at))"' "${AUDIT_DIR}/eu_domains.json"
        fi

    } | tee "$summary_file"

    echo ""
    echo "Summary saved to $summary_file"
}

# Generate analysis
generate_analysis() {
    echo ""
    echo "========================================"
    echo "Generating Analysis"
    echo "========================================"

    local analysis_file="${AUDIT_DIR}/analysis.md"

    {
        echo "# Mailgun Audit Analysis"
        echo ""
        echo "Generated: $(date)"
        echo ""

        # Wrong region check
        echo "## Domains Potentially in Wrong Region"
        echo ""
        echo "| Domain | Current Region | Expected | Reason |"
        echo "|--------|----------------|----------|--------|"

        # .eu domains in US region
        if [ -f "${AUDIT_DIR}/us_domains.json" ]; then
            jq -r '.[] | select(.name | test("\\.eu$")) | "| \(.name) | US | EU | .eu TLD |"' "${AUDIT_DIR}/us_domains.json" 2>/dev/null || true
        fi

        # Every EU-region domain is listed for review rather than judged, since
        # .com/.ai/.social placement there is often deliberate.
        if [ -f "${AUDIT_DIR}/eu_domains.json" ]; then
            jq -r '.[] | "| \(.name) | EU | Review | Non-.eu in EU region |"' "${AUDIT_DIR}/eu_domains.json" 2>/dev/null || true
        fi

        echo ""

        # Disabled domains
        echo "## Disabled Domains"
        echo ""
        local any_disabled=0
        for region in us eu; do
            local REGION_UPPER=$(echo "$region" | tr '[:lower:]' '[:upper:]')
            if [ -f "${AUDIT_DIR}/${region}_domains.json" ]; then
                local disabled=$(jq '[.[] | select(.is_disabled == true)]' "${AUDIT_DIR}/${region}_domains.json")
                local count=$(echo "$disabled" | jq 'length')
                if [ "$count" -gt 0 ]; then
                    any_disabled=1
                    echo "### $REGION_UPPER Region"
                    echo "$disabled" | jq -r '.[] | "- \(.name)"'
                    echo ""
                fi
            fi
        done
        if [ "$any_disabled" -eq 0 ]; then
            echo "None found."
            echo ""
        fi

        # DNS Issues (from detailed domain info)
        echo "## DNS Verification Status"
        echo ""
        for region in us eu; do
            local REGION_UPPER=$(echo "$region" | tr '[:lower:]' '[:upper:]')
            local detail_file="${AUDIT_DIR}/${region}_domains_detail.json"
            if [ -f "$detail_file" ]; then
                echo "### $REGION_UPPER Region"
                echo ""
                echo "| Domain | Sending DNS | Receiving DNS |"
                echo "|--------|-------------|---------------|"
                jq -r '.[] |
                    (.domain.name) as $name |
                    (if .sending_dns_records then ([.sending_dns_records[] | select(.valid == "valid" or .valid == true)] | length | tostring) + "/" + ([.sending_dns_records[]] | length | tostring) else "N/A" end) as $sending |
                    (if .receiving_dns_records then ([.receiving_dns_records[] | select(.valid == "valid" or .valid == true)] | length | tostring) + "/" + ([.receiving_dns_records[]] | length | tostring) else "N/A" end) as $receiving |
                    "| \($name) | \($sending) valid | \($receiving) valid |"
                ' "$detail_file" 2>/dev/null || echo "Error parsing DNS records"
                echo ""
            fi
        done

        # SMTP Credentials
        echo "## SMTP Credentials"
        echo ""
        for region in us eu; do
            local REGION_UPPER=$(echo "$region" | tr '[:lower:]' '[:upper:]')
            local creds_file="${AUDIT_DIR}/${region}_credentials.json"
            if [ -f "$creds_file" ]; then
                echo "### $REGION_UPPER Region"
                echo ""
                jq -r 'to_entries[] | select(.value | length > 0) | "**\(.key)**: \(.value | length) credentials\n\(.value | .[] | "  - \(.login) (created: \(.created_at))")"' "$creds_file" 2>/dev/null || echo "No credentials"
                echo ""
            fi
        done

        # Domain Sending Keys
        echo "## Domain Sending Keys"
        echo ""
        for region in us eu; do
            local REGION_UPPER=$(echo "$region" | tr '[:lower:]' '[:upper:]')
            local keys_file="${AUDIT_DIR}/${region}_sending_keys.json"
            if [ -f "$keys_file" ]; then
                echo "### $REGION_UPPER Region"
                echo ""
                jq -r 'to_entries[] | select(.value | length > 0) | "**\(.key)**: \(.value | length) keys\n\(.value | .[] | "  - \(.id): \(.description // "no description") [disabled: \(.disabled)]")"' "$keys_file" 2>/dev/null || echo "No keys"
                echo ""
            fi
        done

        # Routes
        echo "## Routes"
        echo ""
        for region in us eu; do
            local REGION_UPPER=$(echo "$region" | tr '[:lower:]' '[:upper:]')
            local routes_file="${AUDIT_DIR}/${region}_routes.json"
            if [ -f "$routes_file" ]; then
                local count=$(jq 'length' "$routes_file")
                echo "### $REGION_UPPER Region ($count routes)"
                echo ""
                if [ "$count" -gt 0 ]; then
                    echo "| Priority | Expression | Actions |"
                    echo "|----------|------------|---------|"
                    jq -r 'sort_by(.priority) | .[] | "| \(.priority) | \(.expression | .[0:60]) | \(.actions | join("; ") | .[0:40]) |"' "$routes_file" 2>/dev/null || true
                else
                    echo "No routes configured."
                fi
                echo ""
            fi
        done

        # Webhooks
        echo "## Webhooks"
        echo ""
        for region in us eu; do
            local REGION_UPPER=$(echo "$region" | tr '[:lower:]' '[:upper:]')
            local webhooks_file="${AUDIT_DIR}/${region}_webhooks.json"
            if [ -f "$webhooks_file" ]; then
                echo "### $REGION_UPPER Region"
                echo ""
                jq -r 'to_entries[] | "**\(.key)**:\n\(.value.webhooks // {} | to_entries | if length == 0 then "  No webhooks configured" else (.[] | "  - \(.key): \(.value.urls // [] | join(", "))") end)"' "$webhooks_file" 2>/dev/null || echo "Error parsing webhooks"
                echo ""
            fi
        done

        echo "## Recommendations"
        echo ""
        echo "1. **Review domains in wrong region** - Consider if .eu domains should be in EU region"
        echo "2. **Check DNS records** - Ensure all domains have valid DNS configuration"
        echo "3. **Audit SMTP credentials** - Remove unused credentials"
        echo "4. **Review domain keys** - Remove disabled or unused API keys"
        echo "5. **Consolidate routes** - Review and clean up email routing rules"

    } > "$analysis_file"

    echo "Analysis saved to $analysis_file"
}

# Main
echo "============================================"
echo "   Mailgun Workspace Audit Tool"
echo "   READ-ONLY - GET operations only"
echo "============================================"

# Audit both regions
audit_region "us" "$US_API" || true
audit_region "eu" "$EU_API" || true

# Generate reports
generate_summary
generate_analysis

echo ""
echo "============================================"
echo "Audit complete!"
echo "============================================"
echo ""
echo "Output files:"
ls -la "$AUDIT_DIR"

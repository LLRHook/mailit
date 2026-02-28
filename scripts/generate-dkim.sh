#!/usr/bin/env bash
set -euo pipefail

# Generate DKIM keys for MailIt
SELECTOR="${DKIM_SELECTOR:-mailit}"
DOMAIN="${MAILIT_DOMAIN:-example.com}"
KEY_DIR="./data/dkim"
KEY_BITS="${DKIM_KEY_BITS:-2048}"

# Validate inputs
if [ "$DOMAIN" = "example.com" ]; then
    echo "Warning: MAILIT_DOMAIN is still set to example.com"
    echo "Please set MAILIT_DOMAIN to your actual domain before running this script."
    exit 1
fi

if ! command -v openssl &> /dev/null; then
    echo "Error: openssl is required but not installed"
    exit 1
fi

mkdir -p "$KEY_DIR"

echo "Generating ${KEY_BITS}-bit RSA DKIM key pair for domain: $DOMAIN"
openssl genrsa -out "$KEY_DIR/dkim_private.pem" "$KEY_BITS" 2>/dev/null
openssl rsa -in "$KEY_DIR/dkim_private.pem" -pubout -outform DER 2>/dev/null | openssl base64 -A > "$KEY_DIR/dkim_public.txt"

# Verify files were created
if [ ! -f "$KEY_DIR/dkim_private.pem" ] || [ ! -f "$KEY_DIR/dkim_public.txt" ]; then
    echo "Error: Failed to generate DKIM keys"
    exit 1
fi

# Set restrictive permissions on private key
chmod 600 "$KEY_DIR/dkim_private.pem"

PUBLIC_KEY=$(cat "$KEY_DIR/dkim_public.txt")

echo ""
echo "════════════════════════════════════════════════════════════════"
echo "✓ DKIM Keys Generated Successfully"
echo "════════════════════════════════════════════════════════════════"
echo ""
echo "Private key saved to: $KEY_DIR/dkim_private.pem"
echo "Public key saved to:  $KEY_DIR/dkim_public.txt"
echo ""
echo "════════════════════════════════════════════════════════════════"
echo "ADD THESE DNS RECORDS TO YOUR DOMAIN REGISTRAR"
echo "════════════════════════════════════════════════════════════════"
echo ""
echo "1. SPF Record (recommended, existing SPF may vary):"
echo "   Type:  TXT"
echo "   Name:  @  (or your domain)"
echo "   Value: v=spf1 a mx ip4:YOUR_SERVER_IP ~all"
echo ""
echo "2. DKIM Record (IMPORTANT - required for email delivery):"
echo "   Type:  TXT"
echo "   Name:  ${SELECTOR}._domainkey"
echo "   Value: v=DKIM1; k=rsa; p=${PUBLIC_KEY}"
echo ""
echo "3. DMARC Record (recommended):"
echo "   Type:  TXT"
echo "   Name:  _dmarc"
echo "   Value: v=DMARC1; p=none; rua=mailto:dmarc@${DOMAIN}"
echo ""
echo "4. MX Record (if not already set):"
echo "   Type:     MX"
echo "   Name:     @  (or your domain)"
echo "   Value:    ${DOMAIN}"
echo "   Priority: 10"
echo ""
echo "5. A Record (if not already set):"
echo "   Type:  A"
echo "   Name:  mail  (creates mail.${DOMAIN})"
echo "   Value: YOUR_SERVER_IP"
echo ""
echo "Note: DNS changes may take up to 24 hours to propagate."
echo "════════════════════════════════════════════════════════════════"
echo ""

#!/usr/bin/env bash
set -euo pipefail

# Generate DKIM keys for MailIt
SELECTOR="${DKIM_SELECTOR:-mailit}"
DOMAIN="${MAILIT_DOMAIN:-example.com}"
KEY_DIR="./data/dkim"

mkdir -p "$KEY_DIR"

echo "Generating 2048-bit RSA DKIM key pair..."
openssl genrsa -out "$KEY_DIR/dkim_private.pem" 2048 2>/dev/null
openssl rsa -in "$KEY_DIR/dkim_private.pem" -pubout -outform DER 2>/dev/null | openssl base64 -A > "$KEY_DIR/dkim_public.txt"

PUBLIC_KEY=$(cat "$KEY_DIR/dkim_public.txt")

echo ""
echo "=== DKIM Key Generated ==="
echo ""
echo "Private key saved to: $KEY_DIR/dkim_private.pem"
echo ""
echo "=== Add these DNS records ==="
echo ""
echo "1. SPF Record:"
echo "   Type: TXT"
echo "   Name: $DOMAIN"
echo "   Value: v=spf1 a mx ip4:YOUR_SERVER_IP ~all"
echo ""
echo "2. DKIM Record:"
echo "   Type: TXT"
echo "   Name: ${SELECTOR}._domainkey.${DOMAIN}"
echo "   Value: v=DKIM1; k=rsa; p=${PUBLIC_KEY}"
echo ""
echo "3. DMARC Record:"
echo "   Type: TXT"
echo "   Name: _dmarc.${DOMAIN}"
echo "   Value: v=DMARC1; p=none; rua=mailto:dmarc@${DOMAIN}"
echo ""
echo "4. MX Record:"
echo "   Type: MX"
echo "   Name: ${DOMAIN}"
echo "   Value: ${DOMAIN}"
echo "   Priority: 10"
echo ""
echo "5. Return-Path:"
echo "   Type: CNAME"
echo "   Name: bounce.${DOMAIN}"
echo "   Value: ${DOMAIN}"
echo ""

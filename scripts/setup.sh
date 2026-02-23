#!/usr/bin/env bash
set -euo pipefail

echo "=== MailIt First-Run Setup ==="
echo ""

# Check dependencies
command -v docker >/dev/null 2>&1 || { echo "Error: docker is required"; exit 1; }
command -v docker compose >/dev/null 2>&1 || command -v docker-compose >/dev/null 2>&1 || { echo "Error: docker compose is required"; exit 1; }

# Check for .env
if [ ! -f .env ]; then
    echo "Creating .env from .env.example..."
    cp .env.example .env
    echo "Please edit .env with your settings, then run this script again."
    exit 0
fi

# Source .env
set -a
source .env
set +a

# Validate required vars
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD must be set in .env}"
: "${JWT_SECRET:?JWT_SECRET must be set in .env}"
: "${MAILIT_DOMAIN:?MAILIT_DOMAIN must be set in .env}"

# Generate DKIM keys
echo "Generating DKIM keys..."
./scripts/generate-dkim.sh

# Start services
echo ""
echo "Starting services..."
docker compose up -d

echo ""
echo "Waiting for services to be healthy..."
sleep 10

# Run migrations
echo "Running database migrations..."
docker compose exec mailit-api mailit migrate --up --config /etc/mailit/mailit.yaml

# Create admin account
echo ""
echo "Creating admin account..."
docker compose exec -T mailit-api mailit setup --config /etc/mailit/mailit.yaml

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Dashboard: http://localhost:${WEB_PORT:-3000}"
echo "API:       http://localhost:${HTTP_PORT:-8080}"
echo ""
echo "Next steps:"
echo "1. Add the DNS records printed above"
echo "2. Open the dashboard and verify your domain"
echo "3. Create an API key and start sending!"

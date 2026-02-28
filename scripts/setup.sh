#!/usr/bin/env bash
set -euo pipefail

echo "=== MailIt First-Run Setup ==="
echo ""

# Check dependencies
if ! command -v docker &> /dev/null; then
    echo "Error: docker is required but not installed"
    exit 1
fi

# Check for docker compose (supports both 'docker compose' and 'docker-compose')
if docker compose version &> /dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
elif command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    echo "Error: 'docker compose' is required but not installed"
    echo "Install Docker Desktop 4.0+ or docker-compose plugin"
    exit 1
fi

# Check for .env
if [ ! -f .env ]; then
    echo "Error: .env file not found"
    echo "Creating .env from .env.example..."
    cp .env.example .env
    echo ""
    echo "IMPORTANT: Please edit .env with your settings:"
    echo "  - Set POSTGRES_PASSWORD to a strong password"
    echo "  - Set JWT_SECRET to a random 32+ char string"
    echo "  - Set NEXTAUTH_SECRET to a random 32+ char string"
    echo "  - Set DKIM_MASTER_KEY to 32 bytes of hex (openssl rand -hex 16)"
    echo "  - Change MAILIT_DOMAIN to your domain"
    echo "  - Update NEXT_PUBLIC_API_URL and NEXTAUTH_URL"
    echo ""
    echo "Then run this script again."
    exit 0
fi

# Source .env
set -a
source .env
set +a

# Validate required vars
echo "Validating environment variables..."
MISSING_VARS=()
[ -z "${POSTGRES_PASSWORD}" ] && MISSING_VARS+=("POSTGRES_PASSWORD")
[ -z "${JWT_SECRET}" ] && MISSING_VARS+=("JWT_SECRET")
[ -z "${NEXTAUTH_SECRET}" ] && MISSING_VARS+=("NEXTAUTH_SECRET")
[ -z "${MAILIT_DOMAIN}" ] && MISSING_VARS+=("MAILIT_DOMAIN")
[ -z "${DKIM_MASTER_KEY}" ] && MISSING_VARS+=("DKIM_MASTER_KEY")

if [ ${#MISSING_VARS[@]} -gt 0 ]; then
    echo "Error: The following required variables are not set in .env:"
    printf '  - %s\n' "${MISSING_VARS[@]}"
    exit 1
fi

# Validate secret lengths
if [ ${#JWT_SECRET} -lt 32 ]; then
    echo "Warning: JWT_SECRET should be at least 32 characters (currently ${#JWT_SECRET})"
fi

if [ ${#NEXTAUTH_SECRET} -lt 32 ]; then
    echo "Warning: NEXTAUTH_SECRET should be at least 32 characters (currently ${#NEXTAUTH_SECRET})"
fi

# Generate DKIM keys
echo ""
echo "Generating DKIM keys..."
./scripts/generate-dkim.sh

# Start services
echo ""
echo "Starting services with: $DOCKER_COMPOSE"
$DOCKER_COMPOSE up -d

echo ""
echo "Waiting for services to be healthy..."
for i in {1..30}; do
    if $DOCKER_COMPOSE exec -T postgres pg_isready -U "${POSTGRES_USER}" &> /dev/null; then
        echo "✓ Database is healthy"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "Error: Database did not become healthy in time"
        $DOCKER_COMPOSE logs postgres
        exit 1
    fi
    echo "  Waiting for database... ($i/30)"
    sleep 1
done

sleep 2

# Run migrations
echo ""
echo "Running database migrations..."
$DOCKER_COMPOSE exec -T mailit-api mailit migrate --up --config /etc/mailit/mailit.yaml

# Create admin account
echo ""
echo "Creating admin account..."
$DOCKER_COMPOSE exec -T mailit-api mailit setup --config /etc/mailit/mailit.yaml

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "✓ Setup Complete!"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "Dashboard: ${NEXTAUTH_URL:-http://localhost:3000}"
echo "API:       ${NEXT_PUBLIC_API_URL:-http://localhost:8080}"
echo ""
echo "Next steps:"
echo "1. Review the DNS records printed above"
echo "2. Add the DNS records to your domain registrar"
echo "3. Open the dashboard and verify your domain setup"
echo "4. Create an API key to start sending emails"
echo ""
echo "View logs with: $DOCKER_COMPOSE logs -f"

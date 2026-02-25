# MailIt

Self-hosted email platform for developers. Send transactional emails, manage contacts, and broadcast campaigns — all from your own infrastructure.

MailIt ships as a single Go binary that runs an HTTP API, background workers, and an inbound SMTP server. A Next.js dashboard provides a web UI for managing domains, templates, audiences, and more.

## Features

- **Transactional email** — Send via REST API with DKIM signing and automatic retries
- **Direct MX delivery** — Connects directly to recipient mail servers (no relay required)
- **Inbound SMTP** — Receive bounces and replies on your own domain
- **Contact management** — Audiences, contacts, segments, and custom properties
- **Broadcasts** — Send campaigns to audience segments with template support
- **Templates** — HTML email templates with versioning and publish flow
- **Webhooks** — Get notified of delivery events (sent, bounced, complained) with signed payloads
- **DKIM signing** — Automatic DKIM key generation and DNS record guidance
- **Domain verification** — SPF, DKIM, and MX record verification
- **API keys** — Scoped API key authentication (Resend-compatible `re_` prefix)
- **Rate limiting** — Per-endpoint rate limiting backed by Redis
- **Dashboard** — Next.js 15 web UI with dark theme

## Architecture

```
┌─────────────────────────────────────────────────┐
│                  mailit binary                   │
│                                                  │
│  ┌──────────┐  ┌──────────────┐  ┌───────────┐  │
│  │ HTTP API │  │ Asynq Workers│  │ SMTP Server│  │
│  │ (chi)    │  │ (background) │  │ (inbound)  │  │
│  └────┬─────┘  └──────┬───────┘  └─────┬─────┘  │
│       │               │                │         │
│       └───────┬───────┴────────┬───────┘         │
│               │                │                 │
│         ┌─────┴─────┐   ┌─────┴─────┐           │
│         │ PostgreSQL │   │   Redis   │           │
│         │  (pgx/v5)  │   │ (asynq)  │           │
│         └────────────┘   └───────────┘           │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│              Next.js Dashboard                   │
│          (separate process / container)          │
└─────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 16+
- Redis 7+
- Node.js 18+ (for the dashboard)

### Local Development

```bash
# Start PostgreSQL and Redis
make dev
# This runs docker compose for dependencies, then starts the Go server

# Or start dependencies separately
docker compose -f docker-compose.dev.yml up -d postgres redis
go run ./cmd/mailit serve
```

### First-Run Setup

```bash
# Run migrations and create admin user + DKIM keys
go run ./cmd/mailit migrate --up
go run ./cmd/mailit setup
```

The setup wizard will prompt for admin credentials and generate DKIM keys with DNS records to configure.

### Docker Compose (Production)

```bash
cp config/mailit.example.yaml config/mailit.yaml
# Edit config/mailit.yaml with your settings

docker compose up -d
```

## Configuration

MailIt uses a YAML config file with environment variable overrides. Copy the example config to get started:

```bash
cp config/mailit.example.yaml config/mailit.yaml
```

Every setting can be overridden with environment variables using the `MAILIT_` prefix:

```
server.http_addr     → MAILIT_SERVER_HTTP_ADDR
database.host        → MAILIT_DATABASE_HOST
auth.jwt_secret      → MAILIT_AUTH_JWT_SECRET
smtp_outbound.hostname → MAILIT_SMTP_OUTBOUND_HOSTNAME
```

See [`config/mailit.example.yaml`](config/mailit.example.yaml) for all available settings.

## API

MailIt exposes a REST API on port 8080 (default). Authenticate with a JWT token (from `/auth/login`) or an API key in the `Authorization: Bearer` header.

### Sending Email

```bash
curl -X POST http://localhost:8080/emails \
  -H "Authorization: Bearer re_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "from": "hello@yourdomain.com",
    "to": ["user@example.com"],
    "subject": "Hello from MailIt",
    "html": "<p>Your email body here</p>"
  }'
```

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/auth/register` | Register a new account |
| `POST` | `/auth/login` | Log in and receive a JWT |
| `POST` | `/emails` | Send an email |
| `POST` | `/emails/batch` | Send a batch of emails |
| `GET` | `/emails` | List emails |
| `GET` | `/emails/{emailId}` | Get email details |
| `POST` | `/domains` | Add a sending domain |
| `GET` | `/domains` | List domains |
| `POST` | `/domains/{domainId}/verify` | Verify domain DNS records |
| `POST` | `/api-keys` | Create an API key |
| `GET` | `/api-keys` | List API keys |
| `POST` | `/audiences` | Create an audience |
| `POST` | `/audiences/{audienceId}/contacts` | Add a contact |
| `GET` | `/audiences/{audienceId}/contacts` | List contacts |
| `POST` | `/templates` | Create an email template |
| `POST` | `/templates/{templateId}/publish` | Publish a template version |
| `POST` | `/broadcasts` | Create a broadcast |
| `POST` | `/broadcasts/{broadcastId}/send` | Send a broadcast |
| `POST` | `/webhooks` | Register a webhook endpoint |
| `GET` | `/inbound/emails` | List received inbound emails |
| `GET` | `/logs` | View system logs |
| `GET` | `/healthz` | Health check |

## CLI

```
mailit serve   [--config path]             Start API server, workers, and SMTP
mailit migrate [--config path] --up/--down Run database migrations
mailit setup   [--config path]             First-run setup (admin + DKIM)
mailit version                             Print version
```

## Dashboard

The Next.js dashboard runs as a separate process on port 3000 (default).

```bash
cd web
npm install
npm run dev
```

Pages: Overview, Emails, Domains, API Keys, Audiences, Templates, Broadcasts, Webhooks, Logs, Metrics, Settings.

## Deployment

### Docker

```bash
# Build images
make docker-build

# Run with Docker Compose
docker compose up -d
```

The API image exposes ports 8080 (HTTP), 25 (SMTP inbound), and 587 (submission).

### Kubernetes

A Helm chart is included:

```bash
helm install mailit deploy/helm/mailit \
  --set database.password=your-password \
  --set auth.jwtSecret=your-secret \
  --set dkim.masterEncryptionKey=your-32-byte-hex-key
```

## DNS Setup

For sending emails from your domain, configure these DNS records:

| Type | Name | Value |
|------|------|-------|
| MX | `yourdomain.com` | `mail.yourdomain.com` |
| TXT | `yourdomain.com` | `v=spf1 include:mail.yourdomain.com ~all` |
| TXT | `mailit._domainkey.yourdomain.com` | `v=DKIM1; k=rsa; p=<your-public-key>` |
| A | `mail.yourdomain.com` | `<your-server-ip>` |
| PTR | `<your-server-ip>` | `mail.yourdomain.com` |

Run `mailit setup` to generate DKIM keys and get the exact DNS record values.

## Development

```bash
make build            # Build binary → bin/mailit
make test             # Run Go tests with race detector
make test-web         # Run frontend tests (vitest)
make test-all         # Run all tests
make lint             # Run golangci-lint
make test-coverage    # Generate coverage report
```

## Project Structure

```
cmd/mailit/              Entry point (serve/migrate/setup/version)
internal/
  config/                Koanf config loader
  dto/                   Request/response DTOs
  engine/                SMTP sender, DKIM, DNS, bounce classification
  handler/               HTTP handlers
  model/                 Domain models
  pkg/                   Crypto, response helpers, validation
  repository/postgres/   PostgreSQL repositories
  repository/redis/      Cache layer
  server/                HTTP server setup (chi) + middleware
  service/               Business logic
  smtp/                  Inbound SMTP server (go-smtp)
  webhook/               Webhook dispatcher
  worker/                Asynq task handlers
db/migrations/           SQL migration files
web/                     Next.js 15 dashboard
deploy/                  Dockerfile, Helm chart, CI/CD
config/                  Example configuration
```

## License

MIT

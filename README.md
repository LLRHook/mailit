<p align="center">
  <h1 align="center">MailIt</h1>
  <p align="center">
    Open-source, self-hosted email platform for developers.
    <br />
    Send transactional emails, manage contacts, and broadcast campaigns — all from your own infrastructure.
  </p>
</p>

<p align="center">
  <a href="https://github.com/LLRHook/mailit/actions/workflows/ci.yml"><img src="https://github.com/LLRHook/mailit/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/LLRHook/mailit/releases"><img src="https://img.shields.io/github/v/release/LLRHook/mailit?label=release" alt="Release"></a>
  <a href="https://github.com/LLRHook/mailit/blob/main/LICENSE"><img src="https://img.shields.io/github/license/LLRHook/mailit" alt="License"></a>
  <a href="https://goreportcard.com/report/github.com/mailit-dev/mailit"><img src="https://goreportcard.com/badge/github.com/mailit-dev/mailit" alt="Go Report Card"></a>
</p>

---

## Why MailIt?

Services like Resend, Postmark, and SendGrid are great — until you need full control over your email infrastructure, want to avoid per-email pricing, or need to keep data on your own servers.

MailIt gives you a production-grade email platform that you own entirely:

- **No per-email fees** — send as much as your server can handle
- **Your data stays yours** — self-hosted, no third-party data sharing
- **Resend-compatible API** — familiar `re_` prefixed API keys, similar endpoint structure
- **Single binary** — one Go process runs the API, background workers, and SMTP server
- **Dashboard included** — Next.js web UI for managing everything visually

## Features

- **Transactional email** — Send via REST API with DKIM signing and automatic retries
- **Direct MX delivery** — Connects directly to recipient mail servers (no relay needed)
- **Inbound SMTP** — Receive and process incoming emails on your own domain
- **Contact management** — Audiences, contacts, segments, and custom properties
- **Broadcasts** — Send campaigns to audience segments with template personalization
- **Templates** — HTML email templates with versioning and a publish workflow
- **Webhooks** — Signed payloads for delivery events (`email.sent`, `email.bounced`, `email.inbound`, etc.)
- **DKIM signing** — Automatic key generation and DNS record guidance
- **Domain verification** — SPF, DKIM, and MX record verification
- **Open & click tracking** — Per-recipient tracking with automatic pixel/link injection
- **Suppression lists** — Auto-suppress hard bounces and spam complaints
- **Idempotent sends** — Replay-safe API with 24-hour idempotency keys
- **Rate limiting** — Per-endpoint rate limiting backed by Redis
- **Dashboard** — Next.js 15 web UI with dark theme

## How It Works

MailIt runs as a single Go binary containing three servers — an HTTP API, an asynq background worker, and an inbound SMTP server — backed by PostgreSQL and Redis.

### Architecture

```
                        ┌──────────────────────────────────────────────────────┐
                        │                    mailit binary                     │
                        │                                                      │
  REST API clients ───▶ │  ┌──────────┐  ┌───────────────┐  ┌──────────────┐  │ ◀── Incoming mail
                        │  │ HTTP API │  │ Asynq Workers │  │ SMTP Server  │  │
                        │  │  (chi)   │  │  (background) │  │  (inbound)   │  │
                        │  └────┬─────┘  └───────┬───────┘  └──────┬───────┘  │
                        │       │                │                 │           │
                        │       └────────┬───────┴─────────┬──────┘           │
                        │                │                 │                   │
                        │          ┌─────┴──────┐   ┌─────┴──────┐            │
                        │          │ PostgreSQL │   │   Redis    │            │
                        │          │   (pgx)    │   │  (asynq)   │            │
                        │          └────────────┘   └────────────┘            │
                        └──────────────────────────────────────────────────────┘

                        ┌──────────────────────────────────────────────────────┐
                        │              Next.js Dashboard (port 3000)           │
                        │          Manages domains, templates, audiences       │
                        └──────────────────────────────────────────────────────┘
```

### Sending an Email (Transactional Flow)

When you call `POST /emails`, here's what happens end to end:

```
API Request                     Background Worker                   Recipient
    │                                  │                                │
    ▼                                  │                                │
 1. Validate request                   │                                │
 2. Check idempotency (Redis)          │                                │
 3. Check suppression list             │                                │
 4. Create email record (PG)           │                                │
 5. Enqueue "email:send" task ────────▶│                                │
    │                                  ▼                                │
    │                           6. Fetch email from DB                  │
    │                           7. Look up DKIM key for domain          │
    │                           8. Inject tracking (open pixel, links)  │
    │                           9. Build MIME message                   │
    │                          10. Sign with DKIM (RSA-SHA256)          │
    │                          11. Resolve MX records (DNS)             │
    │                          12. Connect to MX → STARTTLS → deliver ─▶│
    │                                  │                                │
    │                          13. On success: dispatch "email.sent" webhook
    │                              On 5xx: classify bounce → suppress → webhook
    │                              On 4xx: retry with exponential backoff
```

**Key details:**
- Recipients are checked against the suppression list before sending — bounced/complained addresses are automatically blocked
- DKIM private keys are encrypted with AES-256-GCM and stored in PostgreSQL
- MX records are resolved per-domain, tried in priority order, with A/AAAA fallback per RFC 5321
- STARTTLS is attempted opportunistically by default (configurable to mandatory or off)
- Each recipient gets individual delivery tracking — partial failures trigger retries only for failed recipients

### Receiving Email (Inbound Flow)

MailIt runs a go-smtp server that accepts incoming mail for your verified domains:

```
Sender's MX ──▶ Inbound SMTP (port 25)
                     │
                     ▼
                1. RCPT TO: verify domain is registered + verified
                2. DATA: parse MIME (headers, text, HTML, attachments)
                3. Store attachments to disk/S3
                4. Create InboundEmail record (PG)
                5. Enqueue "inbound:process" task
                     │
                     ▼
                Worker marks email processed
                Dispatches "email.inbound" webhook
```

### Broadcasting to Audiences

Broadcasts let you send campaigns to an audience (optionally filtered by segment):

```
POST /broadcasts/{id}/send
         │
         ▼
    1. Validate broadcast (has audience, content, from address)
    2. Set status → "sending"
    3. Paginate contacts (500 at a time)
    4. For each contact:
       a. Substitute variables: {{contact.email}}, {{contact.first_name}}, etc.
       b. Create individual Email record
       c. Enqueue "email:send" task
    5. Each email follows the standard transactional flow above
```

Templates with versioning can be attached to broadcasts — the published version's subject and body are used, with contact-specific variable substitution.

### Webhook Delivery

Every significant event dispatches a signed webhook:

| Event | Trigger |
|-------|---------|
| `email.sent` | Recipient MX accepted the message |
| `email.bounced` | Hard bounce (5xx) from recipient MX |
| `email.failed` | Temporary failure (4xx) after retries exhausted |
| `email.inbound` | Inbound email received and processed |

Webhooks are signed with HMAC-SHA256 and include `X-Webhook-Signature` and `X-Webhook-Timestamp` headers for verification. Failed deliveries retry with exponential backoff (30s → 2m → 10m → 30m → 2h).

### Bounce & Suppression

MailIt automatically classifies bounces and maintains a suppression list:

- **Hard bounce** (5xx) → address added to suppression list, no future sends
- **Soft bounce** (4xx) → retried with backoff, not suppressed
- **Spam complaint** → address suppressed immediately
- **Special case**: 552 "mailbox full" is treated as a soft bounce (temporary)

Every send checks the suppression list first — suppressed addresses are rejected before any SMTP connection is made.

## Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 16+
- Redis 7+
- Node.js 20+ (for the dashboard)

### Local Development

```bash
# Clone the repository
git clone https://github.com/LLRHook/mailit.git
cd mailit

# Start PostgreSQL and Redis, then run the Go server
make dev

# Or start dependencies separately
docker compose -f docker-compose.dev.yml up -d postgres redis
go run ./cmd/mailit serve
```

### First-Run Setup

```bash
# Run database migrations
go run ./cmd/mailit migrate --up

# Create admin user + generate DKIM keys
go run ./cmd/mailit setup
```

The setup wizard prompts for admin credentials and generates DKIM keys with the exact DNS records you need to configure.

### Docker Compose (Production)

```bash
# Configure environment
cp .env.example .env
# Edit .env with your passwords, secrets, and domain

# Run the setup script (validates config, generates DKIM, starts everything)
chmod +x scripts/setup.sh scripts/generate-dkim.sh
./scripts/setup.sh
```

Or manually:

```bash
cp .env.example .env
# Edit .env

./scripts/generate-dkim.sh
docker compose up -d
sleep 10
docker compose exec -T mailit-api mailit migrate --up
docker compose exec -T mailit-api mailit setup
```

Open the dashboard at `http://localhost:3000`.

## Configuration

MailIt uses a YAML config file with environment variable overrides. Every setting can be overridden with the `MAILIT_` prefix:

```
server.http_addr        → MAILIT_SERVER_HTTP_ADDR
database.host           → MAILIT_DATABASE_HOST
auth.jwt_secret         → MAILIT_AUTH_JWT_SECRET
smtp_outbound.hostname  → MAILIT_SMTP_OUTBOUND_HOSTNAME
```

See [`config/mailit.example.yaml`](config/mailit.example.yaml) for all settings with descriptions.

### Key Environment Variables

| Variable | Description |
|----------|-------------|
| `POSTGRES_PASSWORD` | Database password |
| `JWT_SECRET` | API authentication secret (min 32 chars) |
| `NEXTAUTH_SECRET` | Dashboard auth secret (min 32 chars) |
| `MAILIT_DOMAIN` | Your mail server FQDN (e.g., `mail.yourdomain.com`) |
| `DKIM_MASTER_KEY` | 32-byte hex key for DKIM key encryption |
| `NEXT_PUBLIC_API_URL` | API URL as seen by browser |

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

The Next.js dashboard runs as a separate process on port 3000.

```bash
cd web
npm install
npm run dev
```

**Pages:** Overview, Emails, Domains, API Keys, Audiences, Templates, Broadcasts, Webhooks, Logs, Metrics, Settings.

## DNS Setup

For sending emails from your domain, configure these DNS records:

| Type | Name | Value |
|------|------|-------|
| A | `mail.yourdomain.com` | `<your-server-ip>` |
| MX | `yourdomain.com` | `mail.yourdomain.com` (priority 10) |
| TXT | `yourdomain.com` | `v=spf1 include:mail.yourdomain.com ~all` |
| TXT | `mailit._domainkey.yourdomain.com` | `v=DKIM1; k=rsa; p=<your-public-key>` |
| PTR | `<your-server-ip>` | `mail.yourdomain.com` |

Run `mailit setup` to generate DKIM keys and get the exact DNS record values.

## Deployment

### Docker Compose

See the [Quick Start](#quick-start) section above. For production, ensure you:

1. **Use a reverse proxy** (Nginx or Caddy) for TLS termination
2. **Set strong secrets** in `.env` — never use defaults
3. **Configure DNS** with SPF, DKIM, and PTR records
4. **Open ports** 25 (SMTP), 587 (submission), 8080 (API), 3000 (dashboard)

### Kubernetes (Helm)

```bash
helm install mailit deploy/helm/mailit \
  --set database.password=$(openssl rand -base64 32) \
  --set auth.jwtSecret=$(openssl rand -base64 32) \
  --set auth.nextauthSecret=$(openssl rand -base64 32) \
  --set dkim.masterEncryptionKey=$(openssl rand -hex 16) \
  --set domain=mail.yourdomain.com \
  --set ingress.enabled=true \
  --set ingress.host=mail.yourdomain.com
```

See [`deploy/helm/mailit/values.yaml`](deploy/helm/mailit/values.yaml) for all Helm options.

### Docker Images

Pre-built multi-arch images (amd64 + arm64) are published to GitHub Container Registry on each release:

```
ghcr.io/llrhook/mailit/api:<version>
ghcr.io/llrhook/mailit/web:<version>
```

## Development

```bash
make build            # Build binary → bin/mailit
make test             # Run Go tests with race detector
make test-web         # Run frontend tests (vitest)
make test-all         # Run all tests
make lint             # Run golangci-lint
make test-coverage    # Generate coverage report
```

### Project Structure

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
db/migrations/           SQL migration files (18 pairs)
web/                     Next.js 15 dashboard
deploy/                  Dockerfile, Helm chart, CI/CD
config/                  Example configuration
```

## Contributing

Contributions are welcome! Here's how to get started:

1. **Fork** the repository
2. **Create a branch** for your feature or fix
3. **Write tests** — the project has 369 tests (314 Go + 55 frontend) and we'd like to keep coverage high
4. **Run checks** before submitting:
   ```bash
   make test       # Go tests pass
   make lint       # No lint errors
   make test-web   # Frontend tests pass
   ```
5. **Open a pull request** against `main`

### Development Setup

- **Go 1.25** — backend
- **Node.js 20** — frontend (Next.js dashboard)
- **Docker** — for PostgreSQL and Redis in development
- **golangci-lint** — Go linting

### Areas Where Help Is Needed

- Integration tests (testcontainers for PostgreSQL/Redis)
- SMTP backend tests
- Dashboard API wiring (currently uses placeholder data)
- Documentation improvements
- Bug reports and feature requests

## License

MIT

## Acknowledgments

Built with [chi](https://github.com/go-chi/chi), [pgx](https://github.com/jackc/pgx), [asynq](https://github.com/hibiken/asynq), [go-smtp](https://github.com/emersion/go-smtp), [go-msgauth](https://github.com/emersion/go-msgauth), [Next.js](https://nextjs.org), and [shadcn/ui](https://ui.shadcn.com).

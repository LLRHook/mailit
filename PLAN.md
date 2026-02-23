# MailIt — Implementation Plan

## Context

**Problem:** Resend.com charges $20-90/mo with artificial limits (1-day data retention, domain caps, team member limits, account suspensions). No self-hosting option.

**Solution:** MailIt — a fully open-source, self-hosted Resend competitor. Resend-compatible API, built-in SMTP server, unlimited everything. Docker Compose one-command deploy. Optional $10 donation.

**Decisions made during brainstorming:**
- Backend: Go (single binary: API + SMTP + workers)
- Frontend: Next.js 15 + Tailwind + shadcn/ui (dark theme)
- SMTP: Built-in direct delivery to recipient MX servers
- Database: PostgreSQL + Redis (job queue)
- Deployment: Docker Compose + Kubernetes Helm chart
- Monetization: Fully open source + optional donation

---

## Phase 1: Foundation (Go project skeleton + database)

### 1.1 Initialize Go module and directory structure

```
mailit/
├── cmd/mailit/main.go           # Entry point with serve/migrate subcommands
├── internal/
│   ├── config/config.go         # YAML + env config (koanf)
│   ├── server/
│   │   ├── server.go            # Chi router + middleware wiring
│   │   └── middleware/
│   │       ├── auth.go          # API key + JWT dual auth
│   │       ├── ratelimit.go     # Per-team rate limiting
│   │       └── requestid.go     # Request ID injection
│   ├── handler/                 # HTTP handlers (one per resource)
│   ├── service/                 # Business logic layer
│   ├── repository/postgres/     # DB access (pgx)
│   ├── repository/redis/        # Cache + idempotency
│   ├── model/                   # Domain models (Go structs)
│   ├── dto/                     # Request/Response DTOs
│   ├── engine/                  # SMTP sending: sender.go, dkim.go, dns.go, bounce.go
│   ├── smtp/                    # Inbound SMTP: server.go, backend.go
│   ├── worker/                  # Asynq task handlers
│   ├── webhook/                 # Webhook dispatch
│   └── pkg/                     # Shared utilities
├── db/migrations/               # SQL migration files (golang-migrate)
├── config/mailit.example.yaml
├── web/                         # Next.js dashboard
├── deploy/
│   ├── docker/Dockerfile
│   └── helm/mailit/
├── scripts/
├── Makefile
├── docker-compose.yml
└── .github/workflows/
```

**Key Go dependencies:**

| Purpose | Library |
|---------|---------|
| HTTP Router | `github.com/go-chi/chi/v5` |
| PostgreSQL | `github.com/jackc/pgx/v5` |
| Redis | `github.com/redis/go-redis/v9` |
| Job Queue | `github.com/hibiken/asynq` |
| SMTP Server | `github.com/emersion/go-smtp` |
| DKIM | `github.com/emersion/go-msgauth` |
| DNS | `github.com/miekg/dns` |
| JWT | `github.com/golang-jwt/jwt/v5` |
| Config | `github.com/knadh/koanf/v2` |
| Migrations | `github.com/golang-migrate/migrate/v4` |
| Validation | `github.com/go-playground/validator/v10` |
| Logging | `log/slog` (stdlib) |
| UUID | `github.com/google/uuid` |

### 1.2 Database schema (18 migrations)

Tables in order:
1. `users` — id, email, password (bcrypt), name
2. `teams` + `team_members` — team_id, user_id, role (owner/admin/member)
3. `api_keys` — key_hash (SHA-256), key_prefix, permission (full/sending), domain scope
4. `domains` — name, status, region, dkim_private_key (encrypted), dkim_selector, open/click tracking, tls_policy
5. `domain_dns_records` — record_type (SPF/DKIM/MX/DMARC), dns_type, name, value, status
6. `emails` — from, to[], cc[], bcc[], subject, html/text body, status (queued→sending→sent→delivered→bounced→failed), scheduled_at, tags, headers, attachments, idempotency_key
7. `email_events` — type (sent/delivered/bounced/opened/clicked/complained), payload
8. `audiences` — name per team
9. `contacts` — email, first_name, last_name, unsubscribed, per audience
10. `contact_properties` + `contact_property_values` — custom properties
11. `topics` + `contact_topics` — subscription preferences
12. `segments` + `segment_contacts` — audience segmentation
13. `templates` + `template_versions` — versioned templates with variables
14. `broadcasts` — campaign status (draft→queued→sending→sent), audience targeting
15. `webhooks` + `webhook_events` — URL, events[], signing secret, delivery tracking
16. `suppression_list` — email, reason (bounce/complaint/unsubscribe/manual)
17. `logs` — API request logs (level, message, metadata)
18. `inbound_emails` — received emails with raw message storage

### 1.3 Config file structure (`config/mailit.example.yaml`)

Sections: server, database, redis, auth (jwt_secret, api_key_prefix "re_"), smtp_outbound (hostname, tls_policy), smtp_inbound (listen_addr, max_message_bytes), dkim (selector, key_bits, master_encryption_key), workers (concurrency, queues, retry delays), rate_limit, webhooks, dns, logging, storage (local/s3), suppression.

### 1.4 Application entry point (`cmd/mailit/main.go`)

Subcommands:
- `mailit serve` — loads config, connects DB + Redis, runs migrations, starts HTTP server + Asynq workers + SMTP servers concurrently via errgroup
- `mailit migrate --up/--down` — standalone migration runner
- `mailit setup` — first-run admin account creation + DKIM key generation

**Verification:** `go build ./cmd/mailit` compiles. `mailit migrate --up` creates all tables. `mailit serve` starts and responds to `GET /healthz`.

---

## Phase 2: API Core (REST endpoints + auth)

### 2.1 Auth middleware (`internal/server/middleware/auth.go`)

Dual auth: API key (prefix "re_", SHA-256 hash lookup) and JWT (for dashboard). Injects team_id + permission into request context.

### 2.2 REST API endpoints (Resend-compatible)

```
POST   /emails                    POST   /emails/batch
GET    /emails                    GET    /emails/{id}
PATCH  /emails/{id}               POST   /emails/{id}/cancel

POST   /domains                   GET    /domains
GET    /domains/{id}              PATCH  /domains/{id}
DELETE /domains/{id}              POST   /domains/{id}/verify

POST   /api-keys                  GET    /api-keys
DELETE /api-keys/{id}

POST   /audiences                 GET    /audiences
GET    /audiences/{id}            DELETE /audiences/{id}

POST   /audiences/{id}/contacts   GET    /audiences/{id}/contacts
GET    /audiences/{aid}/contacts/{cid}
PATCH  /audiences/{aid}/contacts/{cid}
DELETE /audiences/{aid}/contacts/{cid}

POST   /contact-properties        GET    /contact-properties
PATCH  /contact-properties/{id}   DELETE /contact-properties/{id}

POST   /topics       GET /topics
PATCH  /topics/{id}  DELETE /topics/{id}

POST   /audiences/{id}/segments   GET    /audiences/{id}/segments
PATCH  /audiences/{aid}/segments/{sid}
DELETE /audiences/{aid}/segments/{sid}

POST   /templates                 GET    /templates
GET    /templates/{id}            PATCH  /templates/{id}
DELETE /templates/{id}            POST   /templates/{id}/publish

POST   /broadcasts               GET    /broadcasts
GET    /broadcasts/{id}           PATCH  /broadcasts/{id}
DELETE /broadcasts/{id}           POST   /broadcasts/{id}/send

POST   /webhooks                  GET    /webhooks
GET    /webhooks/{id}             PATCH  /webhooks/{id}
DELETE /webhooks/{id}

GET    /inbound/emails            GET    /inbound/emails/{id}
```

Rate limiting: 10 req/s default, 2 req/s for send, 1 req/s for batch. Redis-backed counters. Returns `X-RateLimit-*` headers.

**Verification:** curl `POST /emails` with API key returns email ID. `GET /emails` returns paginated list.

---

## Phase 3: Email Engine (SMTP sending + DKIM)

### 3.1 DKIM signing (`internal/engine/dkim.go`)

Auto-generate 2048-bit RSA key pair per domain. Sign with `go-msgauth`. Store private key AES-256-GCM encrypted. Generate DNS TXT record for public key.

### 3.2 SMTP sender (`internal/engine/sender.go`)

1. Build RFC 5322 MIME message (headers, multipart body, attachments)
2. DKIM sign the message
3. Resolve recipient MX records via `miekg/dns`
4. Connect to MX host on port 25, EHLO, STARTTLS (opportunistic), MAIL FROM, RCPT TO, DATA
5. Try MX hosts in priority order; classify errors as permanent (5xx) vs temporary (4xx)

### 3.3 Job queue workers (`internal/worker/`)

Asynq task types: `email:send`, `email:send_batch`, `broadcast:send`, `domain:verify`, `webhook:deliver`, `bounce:process`, `inbound:process`, `cleanup:expired`

Queue priorities: critical (6), default (3), low (1). Concurrency: 20 goroutines.

Retry: exponential backoff [10s, 30s, 1m, 5m, 15m, 30m, 1h, 2h], max 8 retries. Permanent failures (5xx) are not retried.

### 3.4 Bounce handling (`internal/engine/bounce.go`)

Classify bounces: hard (5xx → suppress + don't retry), soft (4xx → retry), spam (complaint → suppress). Auto-add hard bounces to suppression list.

### 3.5 Inbound SMTP server (`internal/smtp/`)

`go-smtp` server on port 25. Validates recipient domain is registered. Reads message up to 25MB. Enqueues for async processing (parse MIME, store, dispatch webhooks).

### 3.6 Webhook dispatch (`internal/webhook/dispatcher.go`)

HTTP POST to registered endpoints. HMAC-SHA256 signing. Retry with backoff [30s, 2m, 10m, 30m, 2h], max 5 attempts.

**Verification:** Send email via API → appears in recipient inbox with valid DKIM signature. `dig` confirms SPF/DKIM/DMARC pass.

---

## Phase 4: Dashboard (Next.js)

### 4.1 Initialize project

```bash
cd web/
npx create-next-app@latest . --typescript --tailwind --eslint --app --src-dir
npx shadcn@latest init  # dark default, neutral base
npx shadcn@latest add button badge card dialog dropdown-menu input label \
  select separator sheet sidebar skeleton table tabs textarea tooltip avatar \
  collapsible command popover calendar chart form sonner pagination \
  scroll-area toggle-group switch checkbox radio-group alert
npm install @tanstack/react-table @tanstack/react-query next-themes \
  lucide-react recharts date-fns axios zustand
```

### 4.2 Design system

- Background: `#0a0a0a`, card: `#0e0e0e`, border: `#242424`
- Accent color: teal/cyan (hue 173) — differentiates from Resend's blue
- Fonts: Inter (UI), JetBrains Mono (code/keys)
- Status badges: green (delivered/verified/2xx), red (failed/bounced/4xx-5xx), yellow (pending/queued)

### 4.3 Route structure

```
(auth)/login, /register, /forgot-password  — centered card layout, no sidebar
(dashboard)/                               — sidebar layout
  emails/           emails/[emailId]
  broadcasts/       broadcasts/new        broadcasts/[id]
  templates/        templates/new         templates/[id]
  audience/         (tabs: Contacts, Properties, Segments, Topics)
  metrics/
  domains/          domains/[domainId]
  logs/
  api-keys/
  webhooks/         webhooks/new
  settings/         (tabs: Usage, Billing, Team, SMTP, Integrations, Unsubscribe, Documents)
```

### 4.4 Key shared components

- **DataTable** — generic TanStack Table with sort/filter/paginate (used by 8+ pages)
- **StatusBadge** — maps status → colored badge
- **ApiDrawer** — `</>` button opening Sheet with curl examples per page
- **PageHeader** — title + description + action button + API drawer
- **EmptyState** — "No data yet" with icon + CTA
- **CopyButton**, **CodeBlock**, **StatCard**, **DateRangePicker**

### 4.5 API client + auth

- Axios instance with JWT cookie interceptors
- TanStack Query for server state (caching, mutations)
- Zustand for UI state (sidebar collapse)
- Next.js middleware for route protection (redirect unauthenticated → /login)

### 4.6 Sidebar

Navigation: Emails, Broadcasts, Templates, Audience, Metrics, Domains, Logs, API Keys, Webhooks, Settings. Team switcher at top, user avatar at bottom.

**Verification:** Login → see emails list → click email → see detail with preview/HTML/events. All pages render with data from Go API.

---

## Phase 5: Deployment Infrastructure

### 5.1 Dockerfiles

- **Go** (`deploy/docker/Dockerfile`): multi-stage (golang:1.23-alpine builder → alpine:3.20 runtime). CGO_ENABLED=0 static binary. Non-root user. Migrations bundled at `/migrations`. Healthcheck on `/healthz`.
- **Next.js** (`web/Dockerfile`): 3-stage (deps → builder → runner with standalone output). Non-root user. Healthcheck.

### 5.2 Docker Compose (`docker-compose.yml`)

4 services:
- `mailit-api` (Go) — ports 8080, 25, 587. Depends on postgres + redis (healthy).
- `mailit-web` (Next.js) — port 3000. Depends on mailit-api (healthy).
- `postgres:16-alpine` — persistent volume, health check via pg_isready.
- `redis:7-alpine` — appendonly, persistent volume, health check via redis-cli ping.

Required env vars: `POSTGRES_PASSWORD`, `MAILIT_DOMAIN`, `NEXTAUTH_SECRET`.

### 5.3 Kubernetes Helm chart (`deploy/helm/mailit/`)

- API Deployment (2 replicas, init container for migrations, liveness/readiness probes)
- Web Deployment (2 replicas)
- PostgreSQL StatefulSet with PVC (10Gi default)
- Redis StatefulSet with PVC (2Gi default)
- Ingress (nginx, cert-manager TLS)
- LoadBalancer Services for SMTP inbound (25) + outbound (587)
- HPA for API + Web
- Secrets (auto-generated postgres password + nextauth secret)
- Supports external postgres/redis via values.yaml

### 5.4 Development environment

- `docker-compose.dev.yml` with Go hot reload (air) + Next.js dev server
- Source code volume mounts, exposed DB/Redis ports for local tooling
- Makefile with targets: dev, build, test, lint, migrate, setup-dkim, docker-build

### 5.5 CI/CD (GitHub Actions)

- **CI** (`ci.yml`): Go test + lint, Next.js test + lint + build, Helm lint. Runs on push/PR to main.
- **Release** (`release.yml`): On tag push, builds multi-arch images (amd64 + arm64), pushes to GHCR, packages Helm chart, uploads to GitHub Release.

### 5.6 First-run setup flow

1. `git clone` + `cp .env.example .env` + edit
2. `make setup-dkim` — generates DKIM keys, prints DNS records
3. `docker compose up -d` — starts everything, runs migrations, creates admin account (password printed to logs)
4. Open dashboard → DNS wizard shows required records → verify button checks DNS

**Verification:** `docker compose up -d` starts all 4 containers healthy. `curl localhost:8080/healthz` returns 200. Dashboard loads at localhost:3000.

---

## Build Order (Phase Dependencies)

```
Phase 1 (Foundation) ──→ Phase 2 (API) ──→ Phase 3 (Engine) ──→ Phase 5 (Deploy)
                                                    ↕
                                             Phase 4 (Dashboard)
```

- Phases 3 and 4 can be worked on in parallel after Phase 2.
- Phase 5 can start as soon as Phase 1 is done (Dockerfiles), but Helm chart needs Phase 2-3.

## Files to Create (in order)

1. `go.mod`, `Makefile`, `.env.example`, `config/mailit.example.yaml`
2. `cmd/mailit/main.go` — entry point
3. `internal/config/config.go` — config loader
4. `db/migrations/000001_*.sql` through `000018_*.sql`
5. `internal/model/*.go` — all domain models
6. `internal/repository/postgres/*.go` — DB layer
7. `internal/server/middleware/auth.go` — auth middleware
8. `internal/server/server.go` — router setup
9. `internal/handler/*.go` — all HTTP handlers
10. `internal/service/*.go` — business logic
11. `internal/engine/*.go` — SMTP sender, DKIM, DNS, bounce
12. `internal/smtp/*.go` — inbound SMTP server
13. `internal/worker/*.go` — job handlers
14. `internal/webhook/dispatcher.go` — webhook delivery
15. `web/` — full Next.js dashboard
16. `deploy/docker/Dockerfile`, `web/Dockerfile`
17. `docker-compose.yml`, `docker-compose.dev.yml`
18. `deploy/helm/mailit/` — full Helm chart
19. `.github/workflows/ci.yml`, `.github/workflows/release.yml`
20. `scripts/generate-dkim.sh`, `scripts/setup.sh`

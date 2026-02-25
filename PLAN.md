# MailIt v1.0 — Road to Production

All 5 original build phases are complete. This plan covers everything needed to reach
a fully functional, self-hostable v1.0 release.

Based on a full end-to-end test of every page and API endpoint (Feb 2026).

---

## Phase 1 — Critical Bug Fixes

These block the core email sending flow. Nothing else matters until these work.

### 1.1 Fix worker queue name mismatch
- **Bug**: `internal/service/email.go` enqueues to `asynq.Queue("email")` but the worker
  mux in `internal/worker/server.go` only monitors queues `critical`, `default`, `low`
- **Impact**: Emails are enqueued but **never processed**
- **Fix**: Replace all hardcoded queue strings in services with constants from
  `internal/worker/tasks.go` (`QueueCritical`, `QueueDefault`, `QueueLow`).
  Route email sends to `QueueCritical`.
- **Files**: `internal/service/email.go`, `internal/service/broadcast.go`,
  `internal/service/domain.go`

### 1.2 Fix domain DNS records varchar overflow
- **Bug**: `record_type VARCHAR(10)` in migration 000005 but `RETURN_PATH` is 11 chars
- **Impact**: Domain creation crashes with 500
- **Fix**: New migration — `ALTER COLUMN record_type TYPE VARCHAR(20)`
- **Files**: New `db/migrations/000019_fix_dns_record_type_length.{up,down}.sql`

### 1.3 Fix suppression list "not found" handling
- **Bug**: `internal/service/email.go` treats ALL suppression lookup errors as failures.
  `ErrNotFound` should mean "not suppressed, proceed".
- **Impact**: Email sending returns 500 for every recipient
- **Fix**: Check for `ErrNotFound` and treat as "not suppressed". Only error on real DB failures.
- **Files**: `internal/service/email.go`

### 1.4 Add batch email send handler
- **Bug**: `TaskEmailBatchSend` is defined in `internal/worker/tasks.go` but **no handler
  is registered** in `NewMux()`
- **Impact**: `POST /emails/batch` enqueues a task that is silently dropped
- **Fix**: Implement `BatchEmailSendHandler` that expands the batch into individual
  `TaskEmailSend` tasks
- **Files**: New `internal/worker/batch_email_handler.go`, update `internal/worker/server.go`

### 1.5 Fix CORS defaults
- **Bug**: Default `cors_origins` is `["http://localhost:3000"]`. Any other frontend port
  (or production domain) fails silently with no error.
- **Fix**: Add `cors_origins` to `config/mailit.example.yaml` with clear documentation.
  Default should include both `:3000` and `:3001` for dev.
- **Files**: `internal/config/config.go`, `config/mailit.example.yaml`

### 1.6 Fix API key permission values in frontend
- **Bug**: Frontend sends `"full_access"` / `"sending_access"`, backend validates
  `oneof=full sending`
- **Fix**: Change SelectItem values to `"full"` and `"sending"`
- **Files**: `web/src/app/(dashboard)/api-keys/page.tsx`

---

## Phase 2 — Frontend Foundations

Fix the dashboard so every page works end-to-end with the backend.

### 2.1 Root page redirect
- Replace `web/src/app/page.tsx` (still the default Next.js starter) with a redirect
  to `/login`
- **Files**: `web/src/app/page.tsx`

### 2.2 Add logout
- Add a logout button to the sidebar (bottom, near the version string)
- On click: clear `localStorage` token, clear `mailit_token` cookie, redirect to `/login`
- **Files**: `web/src/components/layout/app-sidebar.tsx`

### 2.3 Fix token storage consistency
- **Problem**: Login sets both `localStorage` and cookie. API client reads `localStorage`.
  Middleware reads cookie. 401 interceptor clears `localStorage` but not cookie.
- **Fix**: Keep dual storage but ensure both are always set AND cleared together.
  Fix the 401 interceptor to clear both. Add `HttpOnly` in a future iteration.
- **Files**: `web/src/lib/api.ts`, `web/src/app/(auth)/login/page.tsx`,
  `web/src/app/(auth)/register/page.tsx`

### 2.4 Add error toasts to all mutations
- **Problem**: No mutation anywhere has an `onError` callback. The Toaster (sonner)
  is mounted in `layout.tsx` but never used.
- **Fix**: Add `onError` with `toast.error()` and `onSuccess` with `toast.success()`
  to every `useMutation` across all pages.
- **Pages**: `domains/page.tsx`, `api-keys/page.tsx`, `audience/page.tsx`,
  `broadcasts/new/page.tsx`, `templates/new/page.tsx`, `templates/[id]/page.tsx`,
  `webhooks/new/page.tsx`

### 2.5 Fix template edit form state
- **Bug**: Form state initializes as `null`, setter functions crash with `...f!`
  when `f` is null
- **Fix**: Initialize form from fetched template data via `useEffect` after query loads
- **Files**: `web/src/app/(dashboard)/templates/[id]/page.tsx`

### 2.6 Add missing delete operations to UI
- Add delete buttons with confirmation dialogs to: Domains, Broadcasts, Webhooks
- Backend already supports `DELETE` for all of these — just needs frontend UI
- **Files**: `domains/page.tsx`, `broadcasts/page.tsx`, `webhooks/page.tsx`

---

## Phase 3 — Metrics & Analytics

Build the metrics backend and wire the frontend.

### 3.1 Create metrics aggregation table
```sql
CREATE TABLE email_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id),
    period_start TIMESTAMPTZ NOT NULL,
    period_type VARCHAR(10) NOT NULL CHECK (period_type IN ('hourly', 'daily')),
    sent INT DEFAULT 0,
    delivered INT DEFAULT 0,
    bounced INT DEFAULT 0,
    failed INT DEFAULT 0,
    opened INT DEFAULT 0,
    clicked INT DEFAULT 0,
    complained INT DEFAULT 0,
    UNIQUE (team_id, period_start, period_type)
);
CREATE INDEX idx_email_metrics_team_period ON email_metrics(team_id, period_type, period_start);
```
- **Files**: New migration pair (next available number after Phase 1 migrations)

### 3.2 Metrics worker
- New asynq periodic task (`TaskMetricsAggregate`) that runs hourly
- Queries `email_events` table grouped by hour/day and upserts into `email_metrics`
- Also increment counters in real-time from `EmailSendHandler` on delivery/bounce events
- **Files**: New `internal/worker/metrics_handler.go`, update `internal/worker/tasks.go`,
  `internal/worker/server.go`

### 3.3 Metrics API endpoint
- `GET /metrics?period=7d|30d|24h` — returns time-series + aggregate totals for the team
- Response: stat card values (total sent, delivery rate, open rate, bounce rate) +
  chart data (per-day or per-hour breakdown)
- **Files**: New `internal/handler/metrics.go`, new `internal/service/metrics.go`,
  new `internal/repository/postgres/metrics.go`, update `internal/server/server.go`

### 3.4 Wire frontend metrics page
- Replace empty data arrays with `useQuery` fetching from `GET /metrics`
- Populate stat cards with real totals, charts with time-series data
- Show empty state when no data, charts when data exists
- **Files**: `web/src/app/(dashboard)/metrics/page.tsx`

---

## Phase 4 — Settings & Team Management

### 4.1 Settings API endpoints
- `GET /settings/usage` — counts (emails today/month, domains, API keys, webhooks, contacts)
- `GET /settings/team` — team info + members list with roles
- `PATCH /settings/team` — update team name
- `GET /settings/smtp` — SMTP config for the team (hostname, ports, auth display)
- **Files**: New `internal/handler/settings.go`, new `internal/service/settings.go`,
  update `internal/server/server.go`

### 4.2 Team invitations & roles
- New `team_invitations` table: id, team_id, email, role, token, expires_at, accepted_at
- `POST /settings/team/invite` — send invitation email
- `POST /auth/accept-invite?token=...` — accept invitation, create account, join team
- Roles: `owner` (full control), `admin` (manage keys/domains), `member` (read + send)
- Role-based checks in auth middleware for destructive operations
- **Files**: New migration, new `internal/handler/team.go`, new `internal/service/team.go`,
  new `internal/repository/postgres/team_invitation.go`

### 4.3 Wire frontend settings
- **Usage tab**: Fetch from `GET /settings/usage` (currently hardcoded zeros are fine
  since it's computing from real data)
- **Team tab**: Fetch members from `GET /settings/team`. Wire "Invite Member" dialog
  to `POST /settings/team/invite`. Show role badges. Allow owner to remove members.
- **SMTP tab**: Fetch from `GET /settings/smtp` instead of hardcoded values
- **Integrations tab**: Remove for v1.0 (or keep "Coming soon" placeholder)
- **Files**: `web/src/app/(dashboard)/settings/page.tsx`

---

## Phase 5 — Email Delivery Features

### 5.1 Unsubscribe support
- Generate unique unsubscribe tokens per contact per audience
- Add `List-Unsubscribe` and `List-Unsubscribe-Post` headers to all outbound emails
  (RFC 8058 one-click unsubscribe)
- New public endpoint: `POST /unsubscribe?token=...` (no auth required)
- On unsubscribe: mark contact as unsubscribed, create `unsubscribed` event,
  dispatch webhook
- **Files**: New `internal/handler/unsubscribe.go`, update `internal/engine/sender.go`,
  new migration for unsubscribe tokens table

### 5.2 Open tracking
- When `domain.open_tracking` is enabled, inject a 1x1 transparent tracking pixel
  at the bottom of HTML email bodies
- Pixel URL: `GET /track/open/{tracking_id}` (public, no auth)
- On load: create `opened` event, increment metrics counter, dispatch webhook
- Return a transparent 1x1 GIF
- **Files**: New `internal/handler/tracking.go`, update `internal/engine/sender.go`
  (pixel injection), update `internal/server/server.go` (public route)

### 5.3 Click tracking
- When `domain.click_tracking` is enabled, rewrite all `<a href="...">` links
  in HTML email bodies
- Rewritten URL: `GET /track/click/{tracking_id}?url={encoded_original_url}`
- On click: create `clicked` event, increment metrics, dispatch webhook,
  302 redirect to original URL
- **Files**: Update `internal/handler/tracking.go`, update `internal/engine/sender.go`
  (link rewriting logic)

### 5.4 Fix template-based broadcasts
- **Bug**: `internal/worker/broadcast_handler.go` builds emails from inline broadcast
  fields (FromAddress, Subject, HTMLBody, TextBody), completely ignoring `TemplateID`
- **Fix**: When broadcast has a `TemplateID`, fetch the published template version
  and use its subject/HTML/text body. Support variable substitution
  (e.g., `{{contact.first_name}}`, `{{contact.email}}`)
- **Files**: `internal/worker/broadcast_handler.go`

---

## Phase 6 — Contact Management

### 6.1 Bulk contact import (CSV)
- New endpoint: `POST /audiences/{audienceId}/contacts/import` (multipart form upload)
- Accept CSV with columns: email (required), first_name, last_name, + custom properties
- Process in background via asynq task (large files can have 100k+ rows)
- Return import job ID, allow polling status via `GET /audiences/{id}/contacts/import/{jobId}`
- Handle duplicates: skip existing or update on match
- **Files**: New `internal/handler/contact_import.go`,
  new `internal/worker/import_handler.go`, new `internal/service/contact_import.go`

### 6.2 Contact export
- `GET /audiences/{audienceId}/contacts/export?format=csv`
- Stream CSV download with all contacts + property values
- **Files**: Update `internal/handler/contact.go`

### 6.3 Segment-based broadcast targeting
- **Gap**: Broadcasts reference an audience but ignore segments entirely
- **Fix**: Add optional `SegmentID` to broadcast model + DTO. When set, only
  expand contacts matching the segment filter during broadcast send.
- **Files**: Update `internal/worker/broadcast_handler.go`,
  update `internal/dto/broadcast.go`, update `internal/model/broadcast.go`

### 6.4 Frontend: import/export UI
- Add "Import CSV" button on Audience page → file upload dialog with column mapping
- Add "Export" button → triggers CSV download
- Progress indicator for large imports
- **Files**: Update `web/src/app/(dashboard)/audience/page.tsx`

---

## Phase 7 — Inbound SMTP

### 7.1 Attachment parsing
- **Gap**: `internal/smtp/backend.go` always returns empty attachments array
- **Fix**: Parse MIME multipart body, extract attachments (filename, content-type, size)
- Store attachment content based on `storage.type` config (local filesystem or S3)
- Link attachments to inbound email record
- **Files**: `internal/smtp/backend.go`, new `internal/service/attachment.go`

### 7.2 Inbound email webhook forwarding
- When an inbound email is received and stored, dispatch a webhook event (`email.inbound`)
- Payload: from, to, subject, text body, HTML body, attachment metadata (names, sizes, URLs)
- **Files**: Update `internal/worker/inbound_handler.go`,
  update `internal/webhook/dispatcher.go` (add `email.inbound` event type)

---

## Phase 8 — Polish & Hardening

### 8.1 Error handling consistency
- Standardize repository error wrapping across all postgres repos
- Ensure all services return typed errors that handlers map to correct HTTP status codes
- **Files**: All `internal/repository/postgres/*.go`, all `internal/service/*.go`

### 8.2 Missing database indexes
- Add compound index on `domain_dns_records(domain_id, record_type)`
- Add index on `email_events(recipient)`
- Review slow-query patterns and add indexes as needed
- **Files**: New migration

### 8.3 API documentation
- OpenAPI/Swagger spec for all endpoints
- Serve interactive docs at `GET /docs`
- **Files**: New `api/openapi.yaml` or code-generated spec

### 8.4 Rate limiting on sensitive endpoints
- Stricter limits on: `POST /auth/register` (5/min), `POST /auth/login` (10/min),
  `POST /api-keys` (10/min)
- Prevent brute-force and enumeration
- **Files**: Update `internal/server/server.go`

### 8.5 Frontend accessibility & UX
- Add `autocomplete` attributes to auth forms
- Add loading skeletons during data fetch (replace blank states during loading)
- Dynamic page titles (browser tab shows current section, e.g., "Domains — MailIt")
- **Files**: Various frontend components

### 8.6 Integration tests
- Testcontainers-based tests: register → create domain → create API key → send email
  → verify delivery event → check metrics
- SMTP backend tests for inbound email parsing + attachment extraction
- **Files**: New `internal/integration_test/` directory

---

## Phase Dependency Map

```
Phase 1 (Bug Fixes) ──→ Phase 2 (Frontend) ──→ Phase 4 (Teams)
       │                                              │
       ├──→ Phase 3 (Metrics) ────────────────────────┤
       │                                              │
       ├──→ Phase 5 (Delivery Features) ──→ Phase 6 (Contacts)
       │                                              │
       ├──→ Phase 7 (Inbound SMTP) ───────────────────┤
       │                                              │
       └──────────────────────────────────────→ Phase 8 (Polish)
```

Phases 3, 5, 7 can run in parallel after Phase 1. Phase 4 depends on Phase 2.
Phase 6 depends on Phases 1 + 5. Phase 8 is ongoing but formalized last.

---

## Bugs Found During Testing (Reference)

| # | Severity | Description | Phase |
|---|----------|-------------|-------|
| 1 | CRITICAL | Worker queue name mismatch — emails never processed | 1.1 |
| 2 | CRITICAL | `varchar(10)` overflow on `RETURN_PATH` in DNS records | 1.2 |
| 3 | CRITICAL | Suppression "not found" treated as error — all sends fail | 1.3 |
| 4 | CRITICAL | `TaskEmailBatchSend` has no handler — batch silently dropped | 1.4 |
| 5 | CRITICAL | CORS not configured by default — frontend can't reach API | 1.5 |
| 6 | MEDIUM | Frontend sends `full_access`, backend expects `full` | 1.6 |
| 7 | MEDIUM | Template edit form crashes (null state) | 2.5 |
| 8 | MEDIUM | No error toasts on any mutation — failures are silent | 2.4 |
| 9 | MEDIUM | No logout button anywhere in the dashboard | 2.2 |
| 10 | MEDIUM | Token storage inconsistency (localStorage vs cookie) | 2.3 |
| 11 | MEDIUM | Settings page: dead invite button, hardcoded SMTP | 4.3 |
| 12 | MEDIUM | Broadcast handler ignores TemplateID | 5.4 |
| 13 | LOW | Root page shows Next.js default starter | 2.1 |
| 14 | LOW | No delete UI for domains, broadcasts, webhooks | 2.6 |
| 15 | LOW | Metrics page was hardcoded dummy data (now zeroed) | 3.4 |
| 16 | LOW | Missing DB indexes on dns_records and email_events | 8.2 |

## Missing Features (Reference)

| # | Feature | Phase |
|---|---------|-------|
| 1 | Metrics aggregation backend | 3 |
| 2 | Multi-user team management + invites | 4 |
| 3 | List-Unsubscribe headers (RFC 8058) | 5.1 |
| 4 | Open tracking (pixel injection) | 5.2 |
| 5 | Click tracking (link rewriting) | 5.3 |
| 6 | Template variable substitution in broadcasts | 5.4 |
| 7 | CSV contact import/export | 6.1, 6.2 |
| 8 | Segment-based broadcast targeting | 6.3 |
| 9 | Inbound SMTP attachment parsing | 7.1 |
| 10 | Inbound email webhook events | 7.2 |
| 11 | OpenAPI documentation | 8.3 |
| 12 | Integration tests | 8.6 |

---

## Definition of Done (v1.0)

A user can:

1. Register an account and log in (and log out)
2. Add a sending domain and verify DNS records (SPF, DKIM, MX, DMARC)
3. Create API keys with scoped permissions
4. Send transactional emails via REST API with DKIM signing
5. Send batch emails via REST API
6. Create and manage versioned email templates
7. Import contacts via CSV and organize into audiences
8. Create broadcasts targeting audiences or segments with template support
9. View real delivery metrics (sent, delivered, bounced, opened, clicked)
10. Receive webhook notifications for all email events
11. Manage unsubscribes with RFC 8058 one-click support
12. Receive inbound emails with attachment parsing
13. Invite team members with role-based access
14. Self-host the entire stack via Docker Compose or Helm
15. Read API documentation at `/docs`

All of the above work through both the REST API and the dashboard UI.

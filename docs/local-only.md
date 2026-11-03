# Local-only Mailit

Mailit now has a complete local profile with no Railway or other hosting account.
It runs the API, dashboard, Postgres, Redis, and a Mailpit test inbox in Docker.
All published ports bind to `127.0.0.1`; the database and Redis have no host ports.
Outgoing mail goes to Mailpit inside the stack, so local tests do not send mail
to real recipients.

```sh
scripts/local.sh init   # create private .env.local secrets once
scripts/local.sh up     # build, start, and wait for health checks
scripts/local.sh setup  # create an admin for a new database; skip after a restore
```

- Dashboard: <http://localhost:13000>
- API and health: <http://localhost:18080/healthz>
- Captured mail: <http://localhost:18025>
- Inbound test SMTP: `localhost:11025`

Override `LOCAL_WEB_PORT`, `LOCAL_API_PORT`, `LOCAL_INBOX_PORT`, or
`LOCAL_SMTP_PORT` in `.env.local` if a port is occupied. Existing migrated users
retain their passwords. The named Docker volumes survive `scripts/local.sh down`.

## Preserve and restore data

`scripts/local.sh backup` saves a Postgres dump under ignored `data/backups/`.
Keep `.env.local` privately with the backups: its DKIM master encryption key is
needed to decrypt previously stored signing keys. A complete migration also
preserves the old Redis queue, any attachment files, and any extra databases.
Never commit any of these files.

For a fresh destination, start Postgres first, then restore a dump before starting
the API. Stop the API during a restore and use `psql -v ON_ERROR_STOP=1` so partial
restore errors cannot be mistaken for success. Redis and attachments have their
own volumes; keep their snapshots alongside the database backup.

`scripts/local.sh down` stops the stack without removing volumes.
`scripts/local.sh logs` shows recent logs. Running locally requires Docker and
the computer to remain on while you use Mailit.

The local inbox uses [Mailpit](https://mailpit.axllent.org/docs/install/docker/).
Receiving Internet mail, DNS/MX records, and delivery to real recipient servers
require a publicly reachable mail host and are outside this local-only profile.

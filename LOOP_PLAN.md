# LOOP_PLAN.md — Feature Development Loop

## Constraints
- **Go NOT installed** — only frontend (Next.js/TypeScript) work is possible
- **Do NOT touch**: `.env`, secrets, `db/migrations/`, `LOOP_QUALITY_NOTES.md`, `LOOP_QUALITY_PLAN.md`
- **Branch**: `worktree-loop-features` (isolated from `worktree-loop-quality`)
- **Test suite**: `cd web && npx vitest run` + `cd web && npx eslint .` + `cd web && npm run build`
- **Full validation command**: `cd web && npm run lint && npx vitest run && npm run build`

## Current State
- **Last completed task**: task-10 (iteration 10 — API drawer with code examples)
- **Next task to pick**: task-11
- **Known conflict**: api-drawer.test.tsx exists on both features and quality branches — will need resolution on merge
- **Branch conflict check**: Quality branch touches emails/page.tsx, webhooks/new/page.tsx, layout.tsx, app-sidebar.tsx — new tasks avoid these files

## Batch 1 (tasks 01–10) — COMPLETE

<details>
<summary>All 10 tasks done — 97 tests added</summary>

```json
[
  { "id": "task-01", "title": "Add search/filter input to DataTable component", "status": "done" },
  { "id": "task-02", "title": "Add confirmation dialog for destructive delete actions", "status": "done" },
  { "id": "task-03", "title": "Build a real dashboard home page with summary stats", "status": "done" },
  { "id": "task-04", "title": "Add error handling to logs page", "status": "done" },
  { "id": "task-05", "title": "Add relative time display to timestamps", "status": "done" },
  { "id": "task-06", "title": "Add keyboard shortcut for global search navigation", "status": "done" },
  { "id": "task-07", "title": "Add dark mode toggle to sidebar", "status": "done" },
  { "id": "task-08", "title": "Add auto-refresh toggle to emails page", "status": "done" },
  { "id": "task-09", "title": "Add webhook event type descriptions", "status": "done" },
  { "id": "task-10", "title": "Add API documentation drawer with code examples", "status": "done" }
]
```

</details>

## Batch 2 (tasks 11–20) — Task List

```json
[
  {
    "id": "task-11",
    "title": "Wire searchKey into broadcasts, templates, audience DataTables",
    "priority": 0,
    "status": "pending",
    "deps": [],
    "description": "The DataTable component supports searchKey/searchPlaceholder props (added in task-01) but no page uses them yet. Add searchKey to broadcasts (name), templates (name), and audience (email) list pages.",
    "files": [
      "web/src/app/(dashboard)/broadcasts/page.tsx",
      "web/src/app/(dashboard)/templates/page.tsx",
      "web/src/app/(dashboard)/audience/page.tsx"
    ],
    "test_files": ["web/src/app/(dashboard)/broadcasts/__tests__/page.test.tsx"]
  },
  {
    "id": "task-12",
    "title": "Wire searchKey into webhooks, domains, api-keys, logs DataTables",
    "priority": 1,
    "status": "pending",
    "deps": [],
    "description": "Continue wiring DataTable search into remaining list pages. Add searchKey to webhooks (url), domains (name), api-keys (name), and logs (message).",
    "files": [
      "web/src/app/(dashboard)/webhooks/page.tsx",
      "web/src/app/(dashboard)/domains/page.tsx",
      "web/src/app/(dashboard)/api-keys/page.tsx",
      "web/src/app/(dashboard)/logs/page.tsx"
    ],
    "test_files": ["web/src/app/(dashboard)/api-keys/__tests__/page.test.tsx"]
  },
  {
    "id": "task-13",
    "title": "Add error handling to audience, broadcasts, api-keys pages",
    "priority": 2,
    "status": "pending",
    "deps": [],
    "description": "These 3 list pages have no error handling. Add the isError/error/refetch pattern with an error card and Retry button, matching the emails/logs page pattern.",
    "files": [
      "web/src/app/(dashboard)/audience/page.tsx",
      "web/src/app/(dashboard)/broadcasts/page.tsx",
      "web/src/app/(dashboard)/api-keys/page.tsx"
    ],
    "test_files": ["web/src/app/(dashboard)/audience/__tests__/page.test.tsx"]
  },
  {
    "id": "task-14",
    "title": "Add error handling to templates, webhooks, domains pages",
    "priority": 3,
    "status": "pending",
    "deps": [],
    "description": "Continue adding error handling to remaining list pages. Add the isError/error/refetch pattern with error card and Retry button.",
    "files": [
      "web/src/app/(dashboard)/templates/page.tsx",
      "web/src/app/(dashboard)/webhooks/page.tsx",
      "web/src/app/(dashboard)/domains/page.tsx"
    ],
    "test_files": ["web/src/app/(dashboard)/templates/__tests__/page.test.tsx"]
  },
  {
    "id": "task-15",
    "title": "Add error handling to detail pages",
    "priority": 4,
    "status": "pending",
    "deps": [],
    "description": "Detail pages (broadcasts/[id], templates/[id], domains/[domainId]) show skeletons on load but have no error state. Add error handling with retry for each.",
    "files": [
      "web/src/app/(dashboard)/broadcasts/[id]/page.tsx",
      "web/src/app/(dashboard)/templates/[id]/page.tsx",
      "web/src/app/(dashboard)/domains/[domainId]/page.tsx"
    ],
    "test_files": ["web/src/app/(dashboard)/broadcasts/__tests__/detail-page.test.tsx"]
  },
  {
    "id": "task-16",
    "title": "Add loading skeletons to list pages",
    "priority": 5,
    "status": "pending",
    "deps": [],
    "description": "List pages (audience, broadcasts, api-keys, templates, webhooks, domains) show nothing while loading. Add skeleton cards/tables matching the detail page skeleton pattern.",
    "files": [
      "web/src/app/(dashboard)/audience/page.tsx",
      "web/src/app/(dashboard)/broadcasts/page.tsx",
      "web/src/app/(dashboard)/api-keys/page.tsx",
      "web/src/app/(dashboard)/templates/page.tsx",
      "web/src/app/(dashboard)/webhooks/page.tsx",
      "web/src/app/(dashboard)/domains/page.tsx"
    ],
    "test_files": []
  },
  {
    "id": "task-17",
    "title": "Add zod validation to broadcasts/new form",
    "priority": 6,
    "status": "pending",
    "deps": [],
    "description": "The broadcasts/new page uses manual validation (isValid = name && from && subject && audienceId && (html || text)). Replace with zod schema validation with inline error messages for each field.",
    "files": ["web/src/app/(dashboard)/broadcasts/new/page.tsx"],
    "test_files": ["web/src/app/(dashboard)/broadcasts/__tests__/new-page.test.tsx"]
  },
  {
    "id": "task-18",
    "title": "Add zod validation to templates/new form",
    "priority": 7,
    "status": "pending",
    "deps": [],
    "description": "The templates/new page uses manual validation. Replace with zod schema validation with inline error messages. Similar pattern to task-17.",
    "files": ["web/src/app/(dashboard)/templates/new/page.tsx"],
    "test_files": ["web/src/app/(dashboard)/templates/__tests__/new-page.test.tsx"]
  },
  {
    "id": "task-19",
    "title": "Add empty state illustrations to list pages",
    "priority": 8,
    "status": "pending",
    "deps": [],
    "description": "List pages show an empty DataTable when there's no data. Add a friendly empty state with an icon, message, and CTA button (e.g. 'No broadcasts yet — create your first one') to broadcasts, templates, and audience pages.",
    "files": [
      "web/src/app/(dashboard)/broadcasts/page.tsx",
      "web/src/app/(dashboard)/templates/page.tsx",
      "web/src/app/(dashboard)/audience/page.tsx"
    ],
    "test_files": []
  },
  {
    "id": "task-20",
    "title": "Add metrics page error handling and loading skeleton",
    "priority": 9,
    "status": "pending",
    "deps": [],
    "description": "The metrics page uses multiple useQuery calls but lacks error states and loading skeletons. Add consistent error handling and skeleton UI.",
    "files": ["web/src/app/(dashboard)/metrics/page.tsx"],
    "test_files": ["web/src/app/(dashboard)/metrics/__tests__/page.test.tsx"]
  }
]
```

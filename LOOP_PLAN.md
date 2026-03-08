# LOOP_PLAN.md — Feature Development Loop

## Constraints
- **Go NOT installed** — only frontend (Next.js/TypeScript) work is possible
- **Do NOT touch**: `.env`, secrets, `db/migrations/`, `LOOP_QUALITY_NOTES.md`, `LOOP_QUALITY_PLAN.md`
- **Branch**: `worktree-loop-features` (isolated from `worktree-loop-quality`)
- **Test suite**: `cd web && npx vitest run` + `cd web && npx eslint .` + `cd web && npm run build`
- **Full validation command**: `cd web && npm run lint && npx vitest run && npm run build`

## Current State
- **Last completed task**: task-06 (iteration 6 — command palette with Cmd+K)
- **Next task to pick**: task-07
- **Branch conflict check**: Zero file overlap with worktree-loop-quality (quality touches setup.ts, error-boundary/stat-card/use-mobile/utils tests; features touches dashboard, data-table, confirm-dialog, page components)

## Task List

```json
[
  {
    "id": "task-01",
    "title": "Add search/filter input to DataTable component",
    "priority": 0,
    "status": "done",
    "deps": [],
    "description": "The DataTable component (web/src/components/shared/data-table.tsx) lacks text search. Add an optional searchKey prop and a search input that filters rows using getFilteredRowModel. This immediately improves UX across emails, broadcasts, contacts, API keys, webhooks, logs pages.",
    "files": ["web/src/components/shared/data-table.tsx"],
    "test_files": ["web/src/components/shared/__tests__/data-table.test.tsx"]
  },
  {
    "id": "task-02",
    "title": "Add confirmation dialog for destructive delete actions",
    "priority": 1,
    "status": "done",
    "deps": [],
    "description": "Broadcasts, API keys, and webhooks pages have delete buttons with no confirmation. Create a reusable ConfirmDialog component and wire it into delete mutations on broadcasts, api-keys, and webhooks pages.",
    "files": [
      "web/src/components/shared/confirm-dialog.tsx",
      "web/src/app/(dashboard)/broadcasts/page.tsx",
      "web/src/app/(dashboard)/api-keys/page.tsx",
      "web/src/app/(dashboard)/webhooks/page.tsx"
    ],
    "test_files": ["web/src/components/shared/__tests__/confirm-dialog.test.tsx"]
  },
  {
    "id": "task-03",
    "title": "Build a real dashboard home page with summary stats",
    "priority": 2,
    "status": "done",
    "deps": [],
    "description": "The dashboard root (web/src/app/(dashboard)/page.tsx) just redirects to /emails. Replace with a proper dashboard showing key metrics (emails sent, delivery rate, recent activity) using the existing /metrics and /settings/usage API endpoints.",
    "files": ["web/src/app/(dashboard)/page.tsx"],
    "test_files": ["web/src/app/(dashboard)/__tests__/page.test.tsx"]
  },
  {
    "id": "task-04",
    "title": "Add error handling to logs page",
    "priority": 3,
    "status": "done",
    "deps": [],
    "description": "The logs page (web/src/app/(dashboard)/logs/page.tsx) doesn't handle API errors — no error state, no retry button. The emails page already has this pattern. Add consistent error handling following the same pattern.",
    "files": ["web/src/app/(dashboard)/logs/page.tsx"],
    "test_files": ["web/src/app/(dashboard)/logs/__tests__/page.test.tsx"]
  },
  {
    "id": "task-05",
    "title": "Add relative time display to timestamps",
    "priority": 4,
    "status": "done",
    "deps": [],
    "description": "All timestamps across the app show absolute dates. Add a RelativeTime component using date-fns formatDistanceToNow that shows relative time (e.g. '5 min ago') with absolute time on hover via a tooltip. Use it in emails list and logs table.",
    "files": [
      "web/src/components/shared/relative-time.tsx",
      "web/src/app/(dashboard)/emails/page.tsx",
      "web/src/app/(dashboard)/logs/page.tsx"
    ],
    "test_files": ["web/src/components/shared/__tests__/relative-time.test.tsx"]
  },
  {
    "id": "task-06",
    "title": "Add keyboard shortcut for global search navigation",
    "priority": 5,
    "status": "done",
    "deps": ["task-01"],
    "description": "Add Cmd+K / Ctrl+K keyboard shortcut that opens a command palette (using the existing cmdk dependency) for quick navigation between pages. The cmdk package is already in package.json but unused.",
    "files": [
      "web/src/components/shared/command-palette.tsx",
      "web/src/app/(dashboard)/layout.tsx"
    ],
    "test_files": ["web/src/components/shared/__tests__/command-palette.test.tsx"]
  },
  {
    "id": "task-07",
    "title": "Add dark mode toggle to sidebar",
    "priority": 6,
    "status": "todo",
    "deps": [],
    "description": "The app uses next-themes but there's no visible toggle for switching between light/dark mode. Add a theme toggle button to the sidebar footer using the existing next-themes dependency.",
    "files": ["web/src/components/layout/app-sidebar.tsx"],
    "test_files": ["web/src/components/layout/__tests__/app-sidebar.test.tsx"]
  },
  {
    "id": "task-08",
    "title": "Add auto-refresh toggle to emails page",
    "priority": 7,
    "status": "todo",
    "deps": [],
    "description": "The emails page has refetchInterval: 15_000 hardcoded. Add a visible toggle button so users can enable/disable auto-refresh and see when the last refresh happened.",
    "files": ["web/src/app/(dashboard)/emails/page.tsx"],
    "test_files": ["web/src/app/(dashboard)/emails/__tests__/page.test.tsx"]
  },
  {
    "id": "task-09",
    "title": "Add webhook event type descriptions",
    "priority": 8,
    "status": "todo",
    "deps": [],
    "description": "The webhook creation page (web/src/app/(dashboard)/webhooks/new/page.tsx) lists event types but doesn't describe what each event means. Add descriptive text for each event type to help users configure webhooks.",
    "files": ["web/src/app/(dashboard)/webhooks/new/page.tsx"],
    "test_files": ["web/src/app/(dashboard)/webhooks/__tests__/new-page.test.tsx"]
  },
  {
    "id": "task-10",
    "title": "Add API documentation drawer with code examples",
    "priority": 9,
    "status": "todo",
    "deps": [],
    "description": "The api-drawer component exists but could be enhanced with language-specific code examples (curl, Node.js, Python) for sending emails via the API. This helps developer onboarding.",
    "files": ["web/src/components/shared/api-drawer.tsx"],
    "test_files": ["web/src/components/shared/__tests__/api-drawer.test.tsx"]
  }
]
```

# Loop Quality Plan

> Auto-generated quality improvement task list for the mailit codebase.
> **Constraint**: Go is not installed on this machine — only frontend (web/) tasks are actionable.
> **Do not touch**: .env, secrets, migration files (`db/migrations/`), or files modified on `loop-features` branch.

## Tasks

| # | Priority | Type | File | Description | Status | Depends |
|---|----------|------|------|-------------|--------|---------|
| 1 | P0 | lint-fix | `web/src/test/setup.ts` | Fix `@typescript-eslint/no-unused-vars` warnings for `_fill` and `_priority` — destructured but unused vars in next/image mock | done | — |
| 2 | P1 | missing-test | `web/src/lib/utils.ts` | Add unit tests for utility functions (cn helper, etc.) | done | — |
| 3 | P1 | missing-test | `web/src/hooks/use-mobile.ts` | Add unit test for `useMobile` hook | done | — |
| 4 | P1 | missing-test | `web/src/components/shared/error-boundary.tsx` | Add unit test for ErrorBoundary component | done | — |
| 5 | P1 | missing-test | `web/src/components/shared/stat-card.tsx` | Add unit test for StatCard component | done | — |
| 6 | P1 | missing-test | `web/src/components/shared/date-range-picker.tsx` | Add unit test for DateRangePicker component | done | — |
| 7 | P2 | missing-test | `web/src/components/shared/api-drawer.tsx` | Add unit test for ApiDrawer component | done | — |
| 8 | P2 | missing-test | `web/src/lib/api.ts` | Add unit tests for API client functions | skip (jsdom localStorage broken on this Node) | — |
| 9 | P2 | missing-test | `web/src/lib/store.ts` | Add unit tests for Zustand store | done | — |
| 10 | P2 | missing-test | `web/src/app/(dashboard)/audience/page.tsx` | Add page test for audience dashboard | done | — |
| 11 | P2 | missing-test | `web/src/app/(dashboard)/metrics/page.tsx` | Add page test for metrics dashboard | done | — |
| 12 | P2 | missing-test | `web/src/app/(dashboard)/broadcasts/page.tsx` | Add page test for broadcasts list page | skip (modified on loop-features) | — |
| 13 | P3 | missing-test | `web/src/app/(dashboard)/broadcasts/new/page.tsx` | Add page test for new broadcast page | done | — |
| 14 | P3 | missing-test | `web/src/app/(dashboard)/broadcasts/[id]/page.tsx` | Add page test for broadcast detail page | done | — |
| 15 | P3 | missing-test | `web/src/app/(dashboard)/webhooks/new/page.tsx` | Add page test for new webhook page | todo | — |
| 16 | P3 | missing-test | `web/src/app/(dashboard)/templates/new/page.tsx` | Add page test for new template page | todo | — |
| 17 | P3 | missing-test | `web/src/app/(dashboard)/templates/[id]/page.tsx` | Add page test for template detail page | todo | — |
| 18 | P3 | missing-test | `web/src/app/(dashboard)/domains/[domainId]/page.tsx` | Add page test for domain detail page | todo | — |
| 19 | P3 | missing-test | `web/src/app/(dashboard)/emails/[emailId]/page.tsx` | Add page test for email detail page | todo | — |
| 20 | P3 | missing-test | `web/src/middleware.ts` | Add unit test for Next.js middleware (auth redirect logic) | todo | — |

## State

- **Current iteration**: 13
- **Last completed task**: #14 (add page test for broadcast detail page)
- **Next task**: #15

# apps/ui/AGENTS.md

Local conventions for the **frontend app only**. Read this before touching anything under `apps/ui/`. The root `AGENTS.md` still applies for general rules.

---

## Stack

- **Vite** + **React 18** + **TypeScript** (strict)
- **Tailwind CSS v4** (no `tailwind.config.js` — token-driven via `@theme` in CSS)
- **TanStack Query** for server state
- **Zustand** for client state
- **react-router-dom** for routing
- Font: **Stack Sans Notch** (TTFs in `public/fonts/`, declared in `src/styles/fonts.css`)

---

## File layout (target)

When the user is building page-by-page, structure pages and supporting code like this:

```
src/
├── app/                 # App shell: providers, router, route guards
├── routes/              # One folder per top-level route, with its own pages
│   ├── customer/        # /c/* — end-user voice/chat widget
│   ├── business/        # /b/* — org admin dashboard
│   └── platform/        # /p/* — super admin
├── components/
│   ├── ui/              # Reusable primitives (Button, Input, Modal, ...)
│   └── layout/          # App shell, sidebars, top bars
├── features/            # Feature-specific components (used in one route group)
├── lib/                 # Generic helpers, API client, query keys
├── stores/              # Zustand stores
├── mocks/               # Mock data, page-by-page (see Mocks below)
│   ├── customer/
│   ├── business/
│   └── platform/
├── labels/              # String constants for future i18n (see Labels below)
├── styles/              # CSS entry + token definitions
├── types/               # Shared TS types
└── main.tsx
```

**Only create folders when something needs to live in them.** Empty folders are noise. Add folders lazily, as the page that needs them is being built.

---

## Design tokens (CSS variables only)

- Every color, radius, shadow, spacing override lives as a CSS variable in `src/styles/index.css` under `@theme { ... }`.
- Components **must consume tokens via Tailwind utilities** that resolve to those variables (`bg-brand-600`, `rounded-md`, `shadow-lg`).
- **Never** write raw hex / `rgb()` / `px` values inside component files. If a needed token doesn't exist, add it to `index.css` first, then use it.
- Dark-mode overrides go in the `@media (prefers-color-scheme: dark)` block, redefining the same variable names.

---

## Labels (i18n-ready)

- **Never** hard-code visible strings in components.
- All UI strings live in `src/labels/{common,customer,business,platform}.ts` as exported constants.
- Group by feature/page: `start.title`, `session.mute`, `actions.save`, etc.
- Components import the relevant label bundle and reference keys.

---

## Mocks (page-by-page)

- Mock data lives under `src/mocks/{interface}/{page}.ts`.
- One file per page, exported as typed objects/arrays/factories.
- Components import the mock for the page they render — never hard-code sample data in component files.
- When real API integration lands, swap the import; the page component shouldn't change shape.

---

## Components

- **Decompose.** Anything likely to be reused goes into `components/ui/` (presentational) or `components/layout/` (structural).
- Feature-specific composition (e.g. a `LiveConversationCard` only used in the business dashboard) goes in `features/`.
- Keep components small and single-purpose. If a component grows past ~150 lines, split it.
- Co-locate the page with its own small subcomponents in the same folder, then promote them to `components/` only when reused.

---

## Auth-guarded routes

- All three interfaces live in the same app.
- Each interface's routes are wrapped in a `<ProtectedRoute roles={[...]}>` that checks the session.
- Routing structure:
  - `/c/*` → customer
  - `/b/*` → business (sidebar shell)
  - `/p/*` → platform (sidebar shell)
- Customer route typically has no shell — it's a focused widget.

---

## Work style for this area

- **One page at a time.** User picks the next page; build that page, its mock, its labels, its components, and stop.
- Don't pre-build "in case we need it" stuff.
- Run `npm run dev` to verify the page renders before marking the task done.

---

## Reference

- Root conventions: `../../AGENTS.md`
- PRD: `../../docs/prd.md`
- TRD: `../../docs/trd.md`

# xkit

Daily PKM (Personal Knowledge Management) monorepo. Daily 是一个专注于个人碎片化知识管理的本地化工具。

Frontend: Vue 3 + Go backend. Also contains the `@xkit/hotkeys` library, `@xkit/pi-kit` pi-coding-agent extensions, and skill packages.

## Toolchain

- **Lint / format**: `oxlint` + `oxfmt` (not ESLint/Prettier). Config in `.oxlintrc.json`, `.oxfmtrc.jsonc`.
- **TS / Vue**: TypeScript 6, Vue 3.5+, Vite 8, Vitest 4, vue-tsc 3.
- **Go**: Go 1.24+, Gin, SQLite (FTS5), sqlc, Bubble Tea TUI, Cobra CLI.
- **Package manager**: pnpm 11 (workspace defined in `pnpm-workspace.yaml`).
- **`pi-*` packages** (pi-kit, pi-theme-everforest) are excluded from the workspace — they use a `pi` field in package.json for the pi-coding-agent framework.

## Commands

```sh
pnpm lint            # oxlint .
pnpm fmt             # oxfmt . --check
pnpm fmt:fix         # oxfmt . --write
pnpm typecheck       # vue-tsc --noEmit --project packages/daily/web/tsconfig.app.json
pnpm dev:web         # pnpm --filter web dev
pnpm build:web       # pnpm --filter web build
pnpm build:hotkeys   # pnpm --filter @xkit/hotkeys build
pnpm test:web        # pnpm --filter web test (Vitest)
pnpm test:hotkeys    # pnpm --filter @xkit/hotkeys test (Vitest)
pnpm test:cli        # Go CLI tests (daily-cli)
pnpm xkit <task>     # Go server management (see below)
pnpm ci              # Full local CI (typecheck → test → vet → lint → fmt)
pnpm web-check       # API endpoint sync check between frontend docs & backend routes
```

### Go server tasks (`pnpm xkit <task>`)

| task | description |
|---|---|
| `dev` | Run development server |
| `test` | Run Go tests |
| `test-cli` | Run CLI integration tests |
| `build` | Build Go binary |
| `build-web` | Build Vue frontend + Go server |
| `vet` | Run `go vet` |
| `sqlc-gen` | Regenerate sqlc Go code after query changes |
| `doc-check` | Validate API docs consistency |
| `clean` | Clean build artifacts |

No Makefile or justfile — all server operations go through `scripts/server.mts`.

## CI order (from `scripts/ci.mts`)

```
pnpm typecheck → pnpm test:hotkeys → pnpm xkit vet → pnpm xkit test → pnpm test:cli → pnpm lint → pnpm fmt
```

Run `pnpm ci` before push.

## Scripts (`scripts/`)

| file | description |
|---|---|
| `scripts/ci.mts` | Full local CI pipeline runner |
| `scripts/server.mts` | Go server build/test/dev task runner |
| `scripts/web-check.mts` | API endpoint sync check between frontend & backend docs |

## GitHub Actions

- `.github/workflows/daily-web-check.yml` — PR/push check for daily/web changes (pnpm install + `check:daily:web:fast`)
- Other checks are manual/local (no CI/CD automation for the rest).

## Package structure

| path | description |
|---|---|
| `packages/daily/web/` | Vue 3 SPA (Pinia, Vue Router 5, TDesign Vue Next, Tailwind 4) |
| `packages/daily/server/` | Go backend (Gin, SQLite, Bubble Tea TUI, Cobra CLI) |
| `packages/daily/docs/` | Business docs, API specifications |
| `packages/hotkeys/` | `@xkit/hotkeys` — Vue 3 hotkey binding & command palette library |
| `packages/strata/` | `@xkit/strata` — Go config library (env/file layered config with source tracking) |
| `packages/skills/` | Agent skills collection (20+ skills, see below) |
| `packages/pi-kit/` | `@xkit/pi-kit` — pi-coding-agent extensions (safe-run, daily, parallel) |
| `packages/pi-theme-everforest/` | Everforest theme for pi-coding-agent |

### Skills (`packages/skills/`)

20+ agent skills covering API design, architecture deepening, code cleanup, code review, critique, debugging, deprecation/migration, git workflow, idea refinement, module interfaces, performance tuning, pre-push checks, requirements docs, security review, source-driven development, spec-to-plan, TDD, and more.

Loaded dynamically by the pi-coding-agent via `packages/skills/skills-router/`.

### Agent runtime (`packages/skills/skills-router/`)

- `router.json` — maps skill names to their SKILL.md paths
- `default-skills.json` — default set of loaded skills

## Architecture

**Both `daily/web` and `daily/server` use strict Clean Architecture layers:**

| layer | web (Vue) | server (Go) |
|---|---|---|
| outermost | `presentation/` (components, views) | `interfaces/` (handlers, CLI, TUI) |
| middle | `application/` (use cases) | `application/` (use cases) |
| middle | `domain/` (entities) | `domain/` (entities) |
| innermost | `infra/` (gateways, stores) | `infrastructure/` (persistence) |

- Frontend uses a `Result<T>` monad (`Success` / `Failure`) instead of throwing for API calls.
- `@xkit/hotkeys` uses static registration at app startup (commands, bindings) + dynamic context nodes per component (`useCtx()`, `useCmd()`). Hotkey matching favors deeper context nodes along the active path.
- `@xkit/strata` provides layered config loading with priority: struct defaults → `.env` → `~/.xkit/config.json[ns]` → config file → explicit env vars, with source tracking.

## Testing

- **Vitest config** (`vitest.config.ts`) defines **two projects**:
  - `node` environment: `packages/*/tests/**/*.test.ts`
  - `happy-dom` environment: `packages/hotkeys/src/**/*.test.ts`
- **`daily/web` tests** live in `tests/unit/` organized by layer.
- **`daily/server` tests** live in `tests/integration/` (standard Go `testing` package).
- GitHub Actions runs `check:daily:web:fast` on web changes.

## Database & codegen

- **SQLite** with FTS5, managed via sqlc code generation.
- Schema: `migrations/sqlite/`. Queries: `internal/infrastructure/persistence/sqlite/queries/`.
- Generated code goes to `internal/infrastructure/persistence/sqlite/db/` (package `sldb`).
- Run `pnpm xkit sqlc-gen` after changing queries.

## Project configuration

- `.translate` — translation config file
- `.oxlintrc.json` — oxlint configuration
- `.oxfmtrc.jsonc` — oxfmt configuration
- `.npmrc` — npm/pnpm configuration
- `tsconfig.base.json` — shared TypeScript base config
- `vitest.config.ts` — Vitest test runner config

## Local agent config

- `.agents/` — pi-coding-agent hooks and skills plugin cache
- `.pi/` — pi-coding-agent local settings
- `.xkit/` — xkit tool local configuration
- `.opencode/` — opencode tool configuration

## Other notes

- Codebase is **Chinese-first** (comments, naming, READMEs).
- The web build (`pnpm build:web`) includes `vue-tsc -b` (typecheck step) then vite build.
- `web-check` script enforces that API endpoints documented in `daily/web/docs/business/CAPABILITIES.md` match `daily/server/docs/api/` docs.

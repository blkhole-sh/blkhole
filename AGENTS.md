# AGENTS.md

Guidelines for AI coding agents working on this project. Merge these with any runtime-specific defaults.

---

## 1. Think Before Coding

Before implementing anything:
- State assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them — don't pick silently.
- If a simpler approach exists, say so.
- If something is unclear, stop and name what's confusing.

## 2. Simplicity First

- No features beyond what was asked.
- No abstractions for single-use code.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

## 3. Surgical Changes

- Don't improve adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- Remove imports/variables/functions that *your* changes made unused — not pre-existing dead code.

## 4. Goal-Driven Execution

Transform tasks into verifiable goals before starting:
- "Fix the bug" → write a test that reproduces it, then make it pass.
- "Add feature X" → define what done looks like, then verify it.

---

## Branch & PR Rules

| Branch | Purpose |
|--------|---------|
| `main` | Production — only receives merges from `development` |
| `development` | Integration — all PRs target this |
| `feat/*` | New features |
| `fix/*` | Bug fixes |

- Always branch from `development`
- Always submit work via PR — never push directly to any branch
- PRs always target `development`

## Commit Messages

**Never** add AI attribution to commits or PRs. This is strictly forbidden:
- No `Co-Authored-By: Claude` or any AI tool
- No `Generated with Claude Code` or similar footers
- No `AI-assisted`, `written by AI`, or any equivalent phrasing

Commits must read as if written by the human developer. No exceptions.

## Stack

**Backend** — Go (`internal/`):
- SQLite via `modernc.org/sqlite`
- Migrations via goose (`internal/db/migrations/`)
- DNS resolver, content blocker, stats cache

**Frontend** — SolidJS + TypeScript (`web/src/`):
- UnoCSS (Tailwind-compatible utility classes)
- Biome for linting and formatting
- Bun as package manager

## Key Commands

```sh
# Backend
go test ./...                  # run all tests
go build ./...                 # verify build

# Frontend (run from web/)
bun run dev                    # dev server
bun run build                  # production build
bunx biome check src/          # lint + format check
bunx biome check --write src/  # auto-fix
```

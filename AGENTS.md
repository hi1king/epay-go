# Repository Guidelines

## Project Structure & Module Organization
- `cmd/server/`: Go entrypoint that wires config, router, database, and payment adapters.
- `internal/`: core backend code. Key packages are `handler/` (HTTP layer), `service/` (business logic), `repository/` (data access), `model/` (GORM models), `payment/` (provider adapters), `router/`, and `middleware/`.
- `pkg/`: shared helpers such as `jwt`, `response`, `sign`, and utility functions.
- `web/src/`: Vue 3 frontend. API clients live in `web/src/api`, pages in `web/src/views`, layouts in `web/src/layouts`, and payment option mapping in `web/src/utils`.
- `deploy/`, `docker-compose*.yml`, and `.github/workflows/`: deployment and CI/CD assets.
- Go tests live next to the code they cover as `*_test.go`.

## Build, Test, and Development Commands
- `go run cmd/server/main.go`: run the backend locally.
- `go build ./cmd/server`: compile the backend binary.
- `go test ./...`: run all Go tests.
- `cd web && npm ci && npm run build`: install frontend dependencies and produce a production build.
- `docker compose pull && docker compose up -d`: start the default stack from GHCR images.
- `docker compose -f docker-compose.prod.caddy.yml up -d --build`: build and run the Caddy-based production example locally.

## Coding Style & Naming Conventions
- Go code must be `gofmt`-formatted; keep package names lowercase and exported identifiers in `CamelCase`.
- Vue/TypeScript files use the existing 2-space indentation and `script setup` style.
- Keep handlers thin, move business rules into `internal/service`, and register new payment providers in `internal/payment` via `init()`.
- Name tests after behavior, for example `TestResolvePayRoutingStripe`.

## Testing Guidelines
- Add or update `*_test.go` files when changing payment routing, adapters, or request parsing.
- Prefer small unit tests around normalization, signature validation, and amount conversion before adding integration-heavy coverage.
- Run `go test ./...` before opening a PR; if frontend logic changed, also run `cd web && npm run build`.

## Commit & Pull Request Guidelines
- Recent history uses short Chinese commit subjects, sometimes with a scope prefix, for example `功能（支付）：添加 Stripe Checkout 支付支持`.
- Keep commits focused and imperative; avoid mixing backend, frontend, and deployment refactors without a clear reason.
- PRs should include the problem, the main changes, config impacts, and screenshots for UI changes. Link related issues and mention any new env vars, webhook paths, or external provider setup.

## Security & Configuration Tips
- Never commit live payment keys, webhook secrets, or populated `.env` files.
- Start from `.env.example` and `config.example.yaml`.
- For Stripe, Alipay, and WeChat changes, document callback URLs and required channel config fields in the PR.

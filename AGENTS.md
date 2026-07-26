# Repository Guidelines

## Project Structure & Module Organization

`main.go` starts the Go application. HTTP handlers live in `service/api/`, shared services in `service/`, database-facing types in `models/`, and reusable infrastructure in `pkg/` (cache, database, middleware, routing, TLS). General helpers are under `utils/`, while configuration parsing and defaults live in `config/`.

The Vue 3/Vite client is in `frontend/`; source files are under `frontend/src/` and public assets under `frontend/public/`. Editor integrations are independent TypeScript packages in `plugin/vscode/` and `plugin/obsidian/`. Deployment definitions live in `docker/`, `deploy/`, and `.github/workflows/`. Go tests sit beside their packages as `*_test.go`.

## Build, Test, and Development Commands

- `make dev-backend`: run the API locally with debug mode enabled.
- `make dev-frontend`: start the Vite development server.
- `go test ./...`: run all Go tests.
- `make build`: build the Go binary, install/build frontend dependencies, and copy `frontend/dist/` into `static/`.
- `make backend-build`: write the backend binary to `build/bin/mkBlog`.
- `make docker-build`: build the local `mkblog:latest` image.
- `cd plugin/vscode && npm run build`: compile the VS Code extension.
- `cd plugin/obsidian && pnpm build`: type-check and bundle the Obsidian plugin.

Use Node `^20.19` or `>=22.12` for the frontend, as declared in its `package.json`.

## Coding Style & Naming Conventions

Format Go changes with `gofmt`; use idiomatic package names, exported `PascalCase` identifiers, and unexported `camelCase` identifiers. Keep handlers grouped by resource, following files such as `service/api/article.go`. TypeScript and Vue files use two-space indentation, `camelCase` functions/variables, `PascalCase` Vue components, and existing semicolon conventions. No repository-wide linter is configured; treat `go vet ./...` and package type-check scripts as pre-review checks.

## Testing Guidelines

Add table-driven Go tests beside changed code and name them `TestBehavior` in `*_test.go`. Run `go test ./...` before submitting. Frontend changes currently have no automated test suite; run `cd frontend && npm run build` to perform Vue/TypeScript checks and a production build.

## Commit & Pull Request Guidelines

Recent history follows Conventional Commit-style subjects such as `feat(deploy): ...`, `chore(webui): ...`, and `fix: ...`. Keep subjects imperative, concise, and scoped when useful. Pull requests should explain behavior and configuration changes, link relevant issues, list validation commands, and include screenshots for visible UI updates. Never commit runtime data or secrets from `data/`, `config.yaml`, certificates, or auth/database credentials; update `config.yaml.tmpl` with safe placeholders instead.

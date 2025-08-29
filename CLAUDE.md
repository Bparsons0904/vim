# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## CRITICAL: Request Clarification Protocol

**If you cannot complete a request for ANY reason, STOP immediately and ask for clarification.**

- Don't make assumptions about unclear requirements
- Don't proceed with partial implementations
- Don't guess what the user wants
- Simply state what you don't understand and ask for specific clarification

This prevents wasted time and ensures accurate implementation.

## Testing Philosophy

**CRITICAL RULE: Never add business logic to make tests pass - use mocks instead**

When writing or fixing tests, follow these principles:

- **Use mocks for external dependencies**: Database connections, cache clients, HTTP clients, file system operations
- **Never modify business logic to accommodate test requirements**: If a test needs specific behavior, mock the dependency rather than changing production code
- **Prefer unit tests with mocked dependencies over integration tests**: Integration tests should be minimal and focused on critical paths
- **Test behavior, not implementation**: Focus on what the code should do, not how it does it
- **Mock at service boundaries**: Mock database interfaces, external APIs, and other services rather than internal function calls

**Examples of what NOT to do:**

- Adding conditional logic to skip database operations in test mode
- Exposing internal functions just to make them testable
- Adding test-specific configuration or flags to production code
- Creating test-specific database schemas or data structures

**Examples of what TO do:**

- Mock database interfaces using testify/mock or similar tools
- Use dependency injection to replace real implementations with mocks
- Create test doubles for external services
- Use in-memory implementations for caches and databases in tests

This keeps production code clean and ensures tests accurately reflect real-world behavior.

## Common Development Commands

### Development Environment

- **Start development**: `tilt up` (starts all services with hot reloading)
- **Stop development**: `tilt down`
- **View logs**: `tilt up --stream`
- **Tilt dashboard**: http://localhost:10350

### Testing & Linting

- **Server tests**: `tilt trigger server-tests` or `go test -C ./server ./...`
- **Server test coverage**: `go test -C ./server -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html`
- **Server linting**: `tilt trigger server-lint` or `golangci-lint run -C ./server`
- **Client linting**: `tilt trigger client-lint` or `npm run lint:check -C ./client`
- **Client tests**: `tilt trigger client-tests` or `npm run test -C ./client`
- **Client test UI**: `npm run test:ui -C ./client` (interactive Vitest UI)
- **Client test watch**: `npm run test:run -C ./client` (run tests without watch mode)
- **TypeScript check**: `npm run typecheck -C ./client`
- **Client preview build**: `npm run serve -C ./client` (preview production build)

### Database Operations

- **Run migrations**: `tilt trigger migrate-up`
- **Rollback migration**: `tilt trigger migrate-down`
- **Seed database**: `tilt trigger migrate-seed`
- **Valkey info**: `tilt trigger valkey-info`

### Manual Development (without Tilt)

- **Server**: `go run -C ./server cmd/api/main.go`
- **Client**: `npm run dev -C ./client`
- **Full stack**: `docker compose -f docker-compose.dev.yml up --build`

### Development Workspace Script

- **Start development environment**: `tilt up`
  - Starts all services with hot reloading
  - Access via Tilt dashboard at http://localhost:10350

### MCP Tools Usage

**CRITICAL: Always prioritize MCP (Model Context Protocol) tools over bash commands when available.**

Available MCP tools and their preferred usage:

- **Git Operations**: Use `mcp__git__*` tools instead of bash git commands
  - `mcp__git__git_status` instead of `git status`
  - `mcp__git__git_commit` instead of `git commit`
  - `mcp__git__git_add` instead of `git add`
  - `mcp__git__git_create_branch` instead of `git checkout -b`
  - `mcp__git__git_checkout` instead of `git checkout`
- **GitHub Operations**: Use `mcp__github__*` tools for GitHub interactions
  - `mcp__github__create_pull_request` for PRs
  - `mcp__github__list_pull_requests` for PR management
  - `mcp__github__get_pull_request` for PR details
- **File Operations**: Use `mcp__filesystem__*` tools when available
  - `mcp__filesystem__read_file` instead of `cat`
  - `mcp__filesystem__write_file` for file creation
  - `mcp__filesystem__list_directory` instead of `ls`

### Important Note: cd Command Aliasing

The `cd` command is aliased to zoxide and cannot be used directly in bash commands. When using bash commands, use one of these alternatives:

- **Use the -C flag**: `go test -C ./server ./...` (preferred)
- **Use builtin cd**: `\cd server && go test ./...` (escapes the alias)
- **Use absolute paths**: `go test /home/bobparsons/Development/billywu/claude/server/...`
- **NEVER use**: `cd server && go test ./...` (this will fail due to zoxide aliasing)

## Architecture Overview

### High-Level Structure

Full-stack application for vim action workflows with Go backend, SolidJS frontend, and Valkey cache:

- **Backend**: Fiber framework with SQLite + GORM, JWT auth, WebSockets
- **Frontend**: SolidJS with TypeScript, Vite, CSS Modules, Solid Query
- **Cache**: Valkey (Redis-compatible) for sessions and caching
- **Orchestration**: Docker Compose + Tilt for development

### Key Ports

- Server API: http://localhost:8288 (WebSocket: ws://localhost:8288/ws)
- Client App: http://localhost:3020
- Valkey DB: localhost:6399 (note: non-standard port to avoid conflicts)

### Backend Architecture (Go)

- **Dependency Injection**: App struct (`internal/app/app.go`) contains all services
- **Controllers**: Interface-based design (`internal/interfaces/`)
- **Database**: Dual database setup - SQLite (primary) + Valkey (cache)
- **Auth**: JWT tokens with bcrypt, middleware-based protection
- **WebSockets**: Manager pattern with hub for real-time communication
- **Routing**: Fiber router with middleware chain

### Frontend Architecture (SolidJS)

- **State Management**: AuthContext + Solid Query for server state
- **API Layer**: Axios with interceptors for token management (`services/api/`)
- **WebSocket**: Auto-connecting WebSocket context with auth token header
- **Routing**: @solidjs/router with protected routes
- **Styling**: SCSS with CSS Modules pattern

### Database Layer

- **Primary**: SQLite with GORM (initialization/seeding in `cmd/migration/`)
- **Cache**: Valkey client for sessions and temporary data
- **Models**: GORM models with methods (`internal/models/`)

### Authentication Flow

1. Login via `/users/login` returns JWT
2. Token stored in HTTP-only cookie and sent via `X-Auth-Token` header
3. AuthContext manages client state and API interceptors
4. WebSocket auth uses same token in connection headers
5. Middleware validates JWT on protected routes

### WebSocket Architecture

- Hub pattern managing client connections
- Auth token required in connection headers
- Real-time communication between authenticated clients
- Automatic reconnection and auth token refresh on client

## Development Notes

### File Structure Guidelines

- **NEVER create index.js/ts files unless absolutely necessary** - Use direct imports instead
- Index files create confusion and make navigation harder as projects grow
- Prefer explicit imports like `import { Modal } from "./components/Modal/Modal"`
- Only create index files for very large component libraries where re-exports are essential

### Key Files to Understand

- `server/internal/app/app.go` - Main dependency injection container
- `client/src/context/AuthContext.tsx` - Auth state management
- `server/internal/handlers/router.go` - API route definitions
- `client/src/services/api/api.service.ts` - API client with interceptors
- `Tiltfile` - Development environment configuration
- `docs/API_IMPLEMENTATION_GUIDE.md` - Comprehensive API development guide
- `client/StyleGuide.md` - CSS architecture and design token conventions

### Environment Configuration

All environment variables in `.env` at project root, shared between services.

**Local Environment Overrides:**

- Copy `.env.local.example` to `.env.local` for local development overrides
- `.env.local` is git-ignored and will override values from `.env`
- Useful for local database paths, different ports, or testing configurations
- Both backend (Go) and frontend (Vite) support this pattern

**Running Multiple Development Instances:**

- Use `.env.local` to configure different ports for parallel development
- Each environment gets isolated Docker resources via `DOCKER_ENV` variable
- Multiple docker-compose files available: `docker-compose.dev.yml`, `docker-compose.claude.yml`, `docker-compose.gemini.yml`
- Example configuration for running alongside main instance:
  - Server: `localhost:8289` (instead of 8288)
  - Client: `localhost:3021` (instead of 3020)
  - Valkey: `localhost:6400` (instead of 6399)
  - Separate database: `data/local_dev.db`
- See `.env.local.example` for complete configuration examples

### Testing Strategy

- **Go tests**: Standard `go test` with table-driven tests using testify for assertions
- **Frontend tests**: Vitest with custom jest-dom-like matchers (defined in `client/src/test/setup.ts`)
- **Path aliases**: Comprehensive import alias setup configured in vitest.config.ts
- **Linting**: golangci-lint for Go (config in `.golangci.yml`), ESLint with TypeScript for frontend
- Manual testing via Tilt dashboard utilities

### Database Operations

- Database initialization and seeding handled in `server/cmd/migration/`
- Use `tilt trigger migrate-up` for database initialization
- Seed data available via `tilt trigger migrate-seed`
- No traditional SQL migrations - database schema managed through GORM models
- Do not write test for this project unless asked
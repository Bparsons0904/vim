# Vim Actions Project

A modern full-stack web application designed for building vim action workflows. Features a Go backend with Fiber framework, SolidJS frontend, and Valkey cache database. The development environment is orchestrated using Docker and Tilt for an optimal developer experience.

## 🏗️ Architecture

```
vim-actions/
├── server/          # Go backend (Repository pattern + Fiber + GORM + SQLite)
│   ├── internal/
│   │   ├── repositories/    # Data access layer with interfaces
│   │   ├── controllers/     # Business logic with DI
│   │   ├── app/            # Dependency injection container
│   │   └── ...
├── client/          # SolidJS frontend (TypeScript + Vite)
├── database/valkey/ # Valkey cache database
└── docker-compose.dev.yml
```

## 🚀 Quick Start

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Tilt](https://tilt.dev/) - Modern development environment orchestrator
- [Node.js v22](https://nodejs.org/) (for local development)
- [Go 1.24+](https://golang.org/) (for local development)

- 🔧 **Server API**: http://localhost:8288 (with WebSocket at /ws)
- 🎨 **Client App**: http://localhost:3020 (with automatic WebSocket auth)
- 💾 **Valkey DB**: localhost:6399
- 📊 **Tilt Dashboard**: http://localhost:10350

### Development Environment

The easiest way to get started is with Tilt, which provides hot reloading, service orchestration, and a web dashboard:

```bash
# Start the entire development environment
tilt up

# Access the Tilt dashboard
open http://localhost:10350
```

This will start:

- 🔧 **Server API**: http://localhost:8288
- 🎨 **Client App**: http://localhost:3020
- 💾 **Valkey DB**: localhost:6399
- 📊 **Tilt Dashboard**: http://localhost:10350

### Alternative: Docker Compose Only

If you prefer to use Docker Compose directly:

```bash
# Start all services
docker compose -f docker-compose.dev.yml up --build

# Stop all services
docker compose -f docker-compose.dev.yml down
```

## 📁 Project Structure

### Server (`/server`)

Go backend using Fiber framework with repository pattern architecture.

- **API Framework**: Fiber v2
- **Database**: SQLite with GORM + Valkey cache
- **Architecture**: Repository pattern with dependency injection
- **Authentication**: JWT with bcrypt
- **WebSockets**: Real-time communication support
- **Data Access**: Interface-based repositories with dual database/cache strategy

#### Repository Layer

The server implements a clean repository pattern for data access:

- **User Repository**: Handles user data with cache-first strategy and database fallback
- **Session Repository**: Manages JWT sessions exclusively in Valkey cache
- **Interface-based Design**: All repositories implement contracts for easy testing and swapping

#### Dependency Injection

The App struct serves as a centralized dependency injection container:

- **Constructor Injection**: Repositories and services injected via constructors
- **Interface Contracts**: Loose coupling through interface-based design
- **Circular Dependency Handling**: WebSocket manager uses setter injection
- **Centralized Configuration**: Single App struct manages all service dependencies

[📖 Server Documentation](./server/README.md)

### Client (`/client`)

Modern SolidJS frontend application with TypeScript.

- **Framework**: SolidJS with TypeScript
- **Build Tool**: Vite
- **Styling**: SCSS with CSS Modules
- **Routing**: @solidjs/router
- **State Management**: Solid Query + Context API

[📖 Client Documentation](./client/README.md)

### Database (`/database/valkey`)

Valkey cache database for session management and caching.

- **Database**: Valkey (Redis-compatible)
- **Configuration**: Optimized for development
- **Persistence**: AOF + RDB snapshots

[📖 Database Documentation](./database/valkey/README.md)

## 🛠️ Development Tools

### Tilt Dashboard Features

The Tilt dashboard at http://localhost:10350 provides:

- **Live Service Status**: Real-time health monitoring
- **Log Streaming**: Aggregated logs from all services
- **Manual Triggers**: Run tests, linting, and utilities
- **Resource Management**: Easy service restart and debugging

### Available Commands

```bash
# Development shortcuts via Tilt
tilt trigger server-tests    # Run Go tests
tilt trigger server-lint     # Run Go linting
tilt trigger client-tests    # Run frontend tests
tilt trigger client-lint     # Run frontend linting
tilt trigger valkey-info     # Show Valkey database info

# Stop all services
tilt down

# Start with streaming logs
tilt up --stream
```

### Development Utilities

A comprehensive development script is available:

```bash
# Setup development environment
./scripts/dev-tools.sh setup

# Run all tests
./scripts/dev-tools.sh test

# Run all linters
./scripts/dev-tools.sh lint

# Build production assets
./scripts/dev-tools.sh build

# Clean all artifacts
./scripts/dev-tools.sh clean

# Database operations
./scripts/dev-tools.sh db reset
./scripts/dev-tools.sh db migrate
```

## 🔧 Configuration

### Centralized Environment Configuration

All environment variables are managed in a single `.env` file at the project root:

```bash
# .env (project root)

# General
GENERAL_VERSION=0.0.1

# Server Configuration
SERVER_PORT=8288
DB_PATH=data/vim.db
DB_CACHE_ADDRESS=valkey
DB_CACHE_PORT=6379

# CORS - must expose X-Auth-Token header for WebSocket auth
CORS_ALLOW_ORIGINS=http://localhost:3020

# Security & Authentication
SECURITY_SALT=12
SECURITY_PEPPER=your-secure-pepper-string
SECURITY_JWT_SECRET=your-secure-jwt-secret

# Client Configuration
VITE_API_URL=http://localhost:8288
VITE_WS_URL=ws://localhost:8288/ws
VITE_ENV=local
```

## 🧪 Testing & Linting

Each component has its own testing and linting setup:

- **Server**: Go tests with `go test`, linting with `golangci-lint`
  - Repository interface testing with mock implementations
  - Controller unit tests with dependency injection
  - Interface compliance testing
- **Client**: TypeScript tests (ready for setup), ESLint for linting
- **Integration**: Manual testing utilities via Tilt dashboard

## 🚢 Production Deployment

While the current setup is optimized for development, production deployment considerations:

- Use multi-stage Docker builds for optimized images
- Configure proper environment variables for production
- Set up proper database backups for Valkey
- Configure reverse proxy for the frontend
- Enable HTTPS and security headers

## 🤝 Contributing

1. **Development Setup**: Use `tilt up` for the best development experience
2. **Code Style**: Follow the established patterns in each component
3. **Testing**: Run tests before submitting changes
4. **Documentation**: Update README files when adding new features

## 📚 Additional Resources

- [Tilt Documentation](https://docs.tilt.dev/)
- [Fiber Documentation](https://docs.gofiber.io/)
- [SolidJS Documentation](https://www.solidjs.com/docs)
- [Valkey Documentation](https://valkey.io/documentation/)

## 🔍 Troubleshooting

### Common Issues

1. **Port Conflicts**: Ensure ports 8288, 3020, and 6399 are available
2. **Docker Issues**: Try `docker system prune` to clean up resources
3. **Tilt Issues**: Check the Tilt dashboard logs for detailed error information
4. **Database Issues**: Use `./scripts/dev-tools.sh db reset` to reset the database

### Getting Help

- Check the Tilt dashboard for real-time service status
- Review individual component README files for specific issues
- Check Docker container logs: `docker compose -f docker-compose.dev.yml logs [service]`

---

**Happy coding! 🎉**

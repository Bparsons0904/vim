# Vim Actions Server

A high-performance Go backend API built with Fiber framework, featuring authentication, WebSocket support, and comprehensive database management. Uses an App Container architecture pattern to eliminate circular dependencies and ensure clean initialization.

## 🏗️ Technology Stack

- **Framework**: [Fiber v2](https://docs.gofiber.io/) - Express-inspired web framework
- **Database**: SQLite with [GORM](https://gorm.io/) ORM
- **Cache**: [Valkey](https://valkey.io/) (Redis-compatible) client
- **Authentication**: JWT with bcrypt password hashing
- **WebSockets**: Real-time communication with token-based auth
- **Migration**: SQL migrations with versioning
- **Configuration**: Viper with `.env` file support
- **Logging**: Structured logging with slog
- **Architecture**: App Container pattern for dependency injection

## 🏛️ Architecture Overview

The server uses an **App Container pattern** to manage dependencies and eliminate circular dependencies:

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ App Container│    │ Interfaces  │    │  Database   │
│             │◄──►│             │◄──►│             │
│  - Database │    │  - UserCtrl │    │  - SQL      │
│  - UserCtrl │    │  - Middlewr │    │  - Cache    │
│  - Middlewr │    └─────────────┘    └─────────────┘
└─────────────┘            │                   │
       │                   ▼                   ▼
       ▼            ┌─────────────┐    ┌─────────────┐
┌─────────────┐     │   Routes    │    │   Models    │
│   Server    │◄────│             │    │             │
│             │     │ - UserRoute │    │ - User      │
│ - Fiber     │     │ - Health    │    │ - Session   │
│ - Routes    │     └─────────────┘    └─────────────┘
└─────────────┘
```

**Key Benefits:**

- ✅ No circular dependencies
- ✅ Single database connection across application
- ✅ Clean layer separation: Routes → Controllers → Models
- ✅ Interface-based design for easy testing
- ✅ Centralized dependency management

## 📁 Project Structure

```
server/
├── cmd/
│   ├── api/
│   │   └── main.go              # Application entry point & app container setup
│   └── migration/
│       ├── main.go              # Migration runner
│       ├── seed/                # Database seeding
│       └── migrations/          # SQL migration files
├── config/
│   └── config.go                # Configuration with .env support
├── internal/
│   ├── app/
│   │   └── app.go               # App container - dependency management
│   ├── interfaces/
│   │   └── interfaces.go        # Service contracts & interfaces
│   ├── controllers/
│   │   └── users/               # Domain-specific controllers
│   │       └── userController.go
│   ├── routes/
│   │   ├── router.go            # Main router setup
│   │   ├── user.routes.go       # User route handlers
│   │   ├── health.routes.go     # Health check endpoints
│   │   └── middleware/          # Authentication middleware
│   ├── models/                  # Data models & database access
│   │   ├── base.model.go        # Base model with common fields
│   │   ├── user.model.go        # User model & methods
│   │   └── session.models.go    # Session management
│   ├── database/                # Database layer
│   │   ├── database.go          # Database connection & setup
│   │   └── cache.database.go    # Valkey cache operations
│   ├── websockets/              # WebSocket management
│   │   ├── websocket.go         # Connection handling & auth
│   │   └── hub.websocket.go     # Client management & caching
│   ├── logger/                  # Structured logging
│   │   └── logger.go            # Logger interface & implementation
│   ├── utils/                   # Utility functions
│   │   ├── auth.utils.go        # Password hashing
│   │   ├── cookie.utils.go      # Cookie management
│   │   └── token.utils.go       # JWT token operations
│   └── server/
│       └── server.go            # Fiber app configuration
├── tmp/                         # Temporary files (development)
├── .env                         # Environment configuration
├── .air.toml                    # Air hot-reload configuration
├── .dockerignore                # Docker ignore patterns
├── Dockerfile.dev               # Development Docker image
├── go.mod                       # Go module dependencies
└── go.sum                       # Go module checksums
```

## 🚀 Getting Started

### Prerequisites

- Go 1.24+
- Air (for hot reloading): `go install github.com/air-verse/air@latest`
- golangci-lint (for linting): `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

### Local Development

1. **Setup Configuration**:

   ```bash
   # Copy example .env from project root or create your own
   cp ../.env .env
   # Edit .env with your settings
   ```

2. **Install Dependencies**:

   ```bash
   go mod download
   ```

3. **Run with Hot Reload**:

   ```bash
   air
   ```

4. **Run Normally**:
   ```bash
   go run cmd/api/main.go
   ```

### Docker Development

The recommended way is through the main project's Tilt setup, but you can also run the server container directly:

```bash
# Build and run development container
docker build -f Dockerfile.dev -t vim-actions-server-dev .
docker run -p 8288:8288 vim-actions-server-dev
```

## 🔧 Configuration

Configuration is managed through a `.env` file using Viper:

### .env Configuration

```bash
# General
GENERAL_VERSION=0.0.1
ENVIRONMENT=development

# Server
SERVER_PORT=8288

# Database
DB_PATH=data/vim.db
DB_CACHE_ADDRESS=valkey  # or localhost for local development
DB_CACHE_PORT=6379

# CORS - must expose X-Auth-Token header for WebSocket auth
CORS_ALLOW_ORIGINS=http://localhost:3020

# Security & Authentication
SECURITY_SALT=12
SECURITY_PEPPER=your-secure-pepper-string
SECURITY_JWT_SECRET=your-secure-jwt-secret
```

**Environment Variables Override**: All config values can be overridden with environment variables using the same names.

## 📡 API Endpoints

### Authentication Flow

| Method | Endpoint            | Description           | Response Headers     |
| ------ | ------------------- | --------------------- | -------------------- |
| POST   | `/api/users/login`  | User login            | `X-Auth-Token` (JWT) |
| POST   | `/api/users/logout` | User logout           | -                    |
| GET    | `/api/users`        | Get current user info | `X-Auth-Token` (JWT) |

### Health Check

| Method | Endpoint      | Description           |
| ------ | ------------- | --------------------- |
| GET    | `/api/health` | Service health status |

### WebSocket

| Endpoint                 | Description                                      | Authentication     |
| ------------------------ | ------------------------------------------------ | ------------------ |
| `ws://localhost:8288/ws` | WebSocket connection for real-time communication | JWT Token Required |

**WebSocket Authentication:**

1. Connect to WebSocket endpoint
2. Server sends `auth_request` message immediately
3. Client responds with `auth_response` containing JWT token from login
4. Server validates and sends `auth_success` or `auth_failure`
5. Authenticated connections are cached in Valkey

```javascript
// Client-side WebSocket auth flow
const ws = new WebSocket("ws://localhost:8288/ws");

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);

  if (message.type === "auth_request") {
    // Send JWT token obtained from login response header
    ws.send(
      JSON.stringify({
        type: "auth_response",
        data: { token: "your-jwt-token" },
      }),
    );
  }
};
```

## 🗄️ Database

### Models

#### User Model

```go
type User struct {
    BaseModel
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Login     string `json:"login"`
    Password  string `json:"-"`        // Hidden from JSON
    IsAdmin   bool   `json:"is_admin"`
}
```

#### Session Model

```go
type Session struct {
    ID        string    `json:"id"`
    UserID    string    `json:"userId"`
    Token     string    `json:"token"`
    ExpiresAt time.Time `json:"expiresAt"`
    RefreshAt time.Time `json:"refreshAt"`
}
```

### Migrations

Run migrations using the migration command:

```bash
# Run all migrations up
go run cmd/migration/main.go up

# Run migrations down (1 step)
go run cmd/migration/main.go down

# Run migrations down (multiple steps)
go run cmd/migration/main.go down 3

# Seed database with test data
go run cmd/migration/main.go seed
```

**Adding a New Migration**:

1. Create a new SQL file: `cmd/migration/migrations/0002_add_feature.sql`
2. Use the migration format:

   ```sql
   -- +migrate Up
   CREATE TABLE new_table (
     id TEXT PRIMARY KEY,
     created_at DATETIME DEFAULT CURRENT_TIMESTAMP
   );

   -- +migrate Down
   DROP TABLE IF EXISTS new_table;
   ```

### Database Seeding

The database includes default test users (created via `seed` command):

- `deadstyle` / `password` (admin)
- `bobb` / `password` (admin)
- `ada` / `password` (user)

## 🔐 Authentication & Security

### JWT Authentication

- Sessions are managed with JWT tokens stored in cache
- Tokens are provided via `X-Auth-Token` response header for client storage
- WebSocket connections require token-based authentication
- Automatic session refresh for active users
- Configurable expiration times (7 days default, 5 days refresh)

### Password Security

- bcrypt hashing with configurable salt cost
- Additional pepper for enhanced security
- Secure password comparison with timing attack protection

### CORS Configuration

- Configurable allowed origins
- Credentials support for cookie-based auth
- **Important**: `X-Auth-Token` header exposed for WebSocket authentication

## 🌐 WebSocket Support

The server provides WebSocket support with comprehensive connection management:

**Features**:

- Token-based authentication on connection
- Connection state management (unauthenticated → authenticated)
- Client tracking and caching in Valkey
- Automatic cleanup on disconnection
- Ping/pong heartbeat for connection health
- Message routing based on authentication status

**Connection Lifecycle**:

1. Client connects → Server registers as unauthenticated
2. Server sends auth request → Client provides JWT token
3. Server validates token → Promotes to authenticated
4. Server caches connection data in Valkey
5. Client can send/receive messages
6. On disconnect → Server cleans up cache entries

## 🧪 Testing & Development

### Running Tests

```bash
go test ./...
```

### Linting

```bash
golangci-lint run
```

### Hot Reload Development

```bash
# Air will watch for changes and automatically rebuild
air
```

### Development Tools

```bash
# Format code
go fmt ./...

# Tidy dependencies
go mod tidy

# Check for vulnerabilities
go mod download && go mod verify
```

## 🏗️ App Container Pattern

The app container (`internal/app/app.go`) manages all dependencies:

```go
type App struct {
    Database   database.DB
    Middleware middleware.Middleware
    Websocket  *websockets.Manager
    Config     config.Config

    // Controllers
    UserController interfaces.UserController
}
```

**Initialization Flow:**

1. Config loaded from `.env`
2. Database connection established
3. Controllers instantiated with dependencies
4. Middleware configured with shared database
5. WebSocket manager initialized
6. App validation ensures all dependencies are present
7. Server started with complete app container

## 🔧 Troubleshooting

### Common Issues

1. **Database Connection Errors**:

   ```bash
   # Check database directory exists
   mkdir -p tmp
   # Verify permissions
   chmod 755 tmp
   ```

2. **Cache Connection Errors**:

   ```bash
   # Verify Valkey is running
   docker ps | grep valkey
   # Check connection settings in .env
   ```

3. **Port Already in Use**:

   ```bash
   # Find process using port 8288
   lsof -i :8288
   # Kill process if needed
   kill -9 <PID>
   ```

4. **WebSocket Authentication Issues**:
   - Ensure login provides `X-Auth-Token` header
   - Verify JWT secret is consistent between login and WebSocket auth
   - Check WebSocket client sends proper `auth_response` format

### Development Tips

- Use `air` for the best development experience with hot reloading
- Check logs for detailed error information with structured logging
- Use the health endpoint to verify service status
- Monitor WebSocket connections through server logs
- App container validation will catch missing dependencies at startup

## 🤝 Contributing

1. Follow Go conventions and use `gofmt`
2. Add tests for new functionality
3. Update interfaces when adding new controller methods
4. Use structured logging for debugging
5. Ensure migrations are reversible
6. Add new controllers to app container and interfaces

## 📚 Additional Resources

- [Fiber Documentation](https://docs.gofiber.io/)
- [GORM Documentation](https://gorm.io/docs/)
- [Valkey Documentation](https://valkey.io/documentation/)
- [Air Documentation](https://github.com/air-verse/air)
- [JWT Go Documentation](https://github.com/golang-jwt/jwt)

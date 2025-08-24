# Tiltfile for Vim Actions development environment

# Load environment variables from client/.env
load('ext://dotenv', 'dotenv')
load('ext://restart_process', 'docker_build_with_restart')

dotenv('./.env')
# Load .env.local for local overrides (if it exists)
if os.path.exists('./.env.local'):
    dotenv('./.env.local')

# Configuration - use environment variables with defaults
SERVER_PORT = os.getenv('SERVER_PORT', '8288')
CLIENT_PORT = os.getenv('CLIENT_PORT', '3020')
DB_CACHE_PORT= os.getenv('DB_CACHE_PORT', '6379')
DOCKER_ENV = os.getenv('DOCKER_ENV', 'dev')
TILT_PORT = os.getenv('TILT_PORT', '10350')

# Development mode toggle
DEV_MODE = True

# Go Server with Air hot reloading - Volume mount approach
docker_build(
    'vim-actions-server-dev-' + DOCKER_ENV,
    context='./server',
    dockerfile='./server/Dockerfile.dev',
    target='development',
    # No live_update - use volume mounts instead
    ignore=[
        'tmp/', 
        '*.log', 
        'main',
        '.git/',
        'Dockerfile*',
        '.dockerignore',
        'data/',
        '*.db',
        '*.db-journal',
    ]
)

# SolidJS Client with Vite hot reloading  
docker_build(
    'vim-actions-client-dev-' + DOCKER_ENV,
    context='./client',
    dockerfile='./client/Dockerfile.dev',
    live_update=[
        # ALL SYNC STEPS MUST COME FIRST
        # Sync package files for dependency management
        sync('./client/package.json', '/app/package.json'),
        sync('./client/package-lock.json', '/app/package-lock.json'),
        # Sync source directories for hot reloading
        sync('./client/src', '/app/src'),
        sync('./client/public', '/app/public'),
        # Sync config files
        sync('./client/vite.config.ts', '/app/vite.config.ts'),
        sync('./client/tsconfig.json', '/app/tsconfig.json'),
        sync('./client/index.html', '/app/index.html'),
        # ALL RUN STEPS MUST COME AFTER SYNC STEPS
        # Run npm install when package files change
        run('npm install', trigger=['./client/package.json', './client/package-lock.json']),
    ],
    ignore=[
        'node_modules/', 
        'dist/', 
        'build/', 
        '.vite/',
        '.*.swp',
        '.*.swo',
        '*~',
        '.DS_Store',
        '.git/',
        '.gitignore',
        'Dockerfile*',
        '.dockerignore',
    ]
)

# Valkey database service
docker_build(
    'vim-actions-valkey-dev-' + DOCKER_ENV,
    context='./database/valkey',
    dockerfile='./database/valkey/Dockerfile.dev',
    live_update=[
        # Sync configuration changes
        sync('./database/valkey/valkey.conf', '/usr/local/etc/valkey/valkey.conf'),
        # Restart container when config changes (Valkey needs restart for config changes)
        restart_container(),
    ],
    ignore=[
        '.*.swp',
        '.*.swo',
        '*~',
        '.DS_Store',
        '.git/',
        '.gitignore',
    ]
)

# Use docker-compose for orchestration - environment-specific file
docker_compose('./docker-compose.' + DOCKER_ENV + '.yml')

# ==========================================
# CORE SERVICES
# ==========================================

dc_resource('server',
    labels=['1-services'],
    resource_deps=['valkey'],
)

dc_resource('client',
    labels=['1-services'],
    resource_deps=['server']
)

dc_resource('valkey',
    labels=['1-services'],
    resource_deps=[],
)

# Development utilities
if DEV_MODE:
    # ==========================================
    # SERVER/BACKEND QUALITY CHECKS
    # ==========================================
    
    # Server full check - runs tests and linting
    local_resource(
        'server-1-check-all',
        cmd='cd server && go test ./... && golangci-lint run',
        deps=['./server'],
        ignore=['./server/tmp', './server/*.log', './server/main'],
        labels=['2-server'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL
    )

    # Server tests
    local_resource(
        'server-2-tests',
        cmd='cd server && go test ./...',
        deps=['./server'],
        ignore=['./server/tmp', './server/*.log', './server/main'],
        labels=['2-server'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL
    )

    # Server linting
    local_resource(
        'server-3-lint',
        cmd='cd server && golangci-lint run',
        deps=['./server'],
        ignore=['./server/tmp', './server/*.log', './server/main'],
        labels=['2-server'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL
    )

    # Server test coverage (additional utility)
    local_resource(
        'server-4-coverage',
        cmd='cd server && go test -cover ./...',
        deps=['./server'],
        ignore=['./server/tmp', './server/*.log', './server/main'],
        labels=['2-server'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL
    )

    # HTML coverage report (additional utility)
    local_resource(
        'server-5-coverage-html',
        cmd='cd server && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html',
        deps=['./server'],
        ignore=['./server/tmp', './server/*.log', './server/main'],
        labels=['2-server'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL
    )

    # ==========================================
    # CLIENT/FRONTEND QUALITY CHECKS
    # ==========================================
    
    # Client full check - runs all three checks
    local_resource(
        'client-1-check-all',
        cmd='cd client && [ ! -d node_modules ] && echo "❌ ERROR: node_modules not found. Run npm install first." && exit 1 || (npm run test:run && npm run lint:check && npx tsc --noEmit --skipLibCheck)',
        deps=['./client/src', './client/package.json'],
        ignore=['./client/node_modules', './client/dist'],
        labels=['3-client'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL,
        resource_deps=['client']  # Wait for client service to be ready
    )

    # Client tests
    local_resource(
        'client-2-tests',
        cmd='cd client && [ ! -d node_modules ] && echo "❌ ERROR: node_modules not found. Run npm install first." && exit 1 || npm run test:run',
        deps=['./client/src'],
        ignore=['./client/node_modules', './client/dist'],
        labels=['3-client'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL,
        resource_deps=['client']
    )

    # Client linting
    local_resource(
        'client-3-lint',
        cmd='cd client && [ ! -d node_modules ] && echo "❌ ERROR: node_modules not found. Run npm install first." && exit 1 || npm run lint:check',
        deps=['./client/src'],
        ignore=['./client/node_modules', './client/dist'],
        labels=['3-client'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL,
        resource_deps=['client']
    )

    # TypeScript compilation check
    local_resource(
        'client-4-typecheck',
        cmd='cd client && [ ! -d node_modules ] && echo "❌ ERROR: node_modules not found. Run npm install first." && exit 1 || npx tsc --noEmit --skipLibCheck',
        deps=['./client/src'],
        ignore=['./client/node_modules', './client/dist'],
        labels=['3-client'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL,
        resource_deps=['client']
    )

    # ==========================================
    # VALKEY/DATABASE UTILITIES
    # ==========================================
    
    # Valkey utilities
    local_resource(
        'valkey-info',
        cmd='docker compose -f docker-compose.' + DOCKER_ENV + '.yml exec valkey valkey-cli info',
        labels=['4-valkey'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL
    )

    # Database migration commands
    local_resource(
        'migrate-up',
        cmd='docker compose -f docker-compose.' + DOCKER_ENV + '.yml exec server go run cmd/migration/main.go up',
        deps=['./server/cmd/migration', './server/internal', './server/config'],
        ignore=['./server/tmp', './server/*.log', './server/main'],
        labels=['4-valkey'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL,
        resource_deps=['server'] 
    )

    local_resource(
        'migrate-down',
        cmd='docker compose -f docker-compose.' + DOCKER_ENV + '.yml exec server go run cmd/migration/main.go down 1',
        deps=['./server/cmd/migration', './server/internal', './server/config'],
        ignore=['./server/tmp', './server/*.log', './server/main'],
        labels=['4-valkey'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL,
        resource_deps=['server']
    )

    local_resource(
        'migrate-seed',
        cmd='docker compose -f docker-compose.' + DOCKER_ENV + '.yml exec server go run cmd/migration/main.go seed',
        deps=['./server/cmd/migration', './server/internal', './server/config'],
        ignore=['./server/tmp', './server/*.log', './server/main'],
        labels=['4-valkey'],
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL,
        resource_deps=['server']
    )


print("🚀 Vim Actions Development Environment (Environment: %s)" % DOCKER_ENV)
print("📊 Tilt Dashboard: http://localhost:%s" % TILT_PORT)
print("🔧 Server API: http://localhost:%s" % SERVER_PORT)
print("🎨 Client App: http://localhost:%s" % CLIENT_PORT)
print("💾 Valkey DB: localhost:%s" % DB_CACHE_PORT)
print("💡 Hot reloading enabled for all services!")
print("🧪 Manual test/lint resources available in Tilt UI")

# Development shortcuts
print("\n📋 Quick Commands:")
print("\n🔧 SERVER (Backend):")
print("• tilt trigger server-1-check-all     - Run ALL server checks (tests + lint)")
print("• tilt trigger server-2-tests         - Run Go tests")
print("• tilt trigger server-3-lint          - Run Go linting")
print("• tilt trigger server-4-coverage      - Run tests with coverage")
print("• tilt trigger server-5-coverage-html - Generate HTML coverage report")
print("\n🎨 CLIENT (Frontend):")
print("• tilt trigger client-1-check-all     - Run ALL client checks (tests + lint + types)")
print("• tilt trigger client-2-tests         - Run frontend tests")
print("• tilt trigger client-3-lint          - Run frontend linting") 
print("• tilt trigger client-4-typecheck     - Run TypeScript checking")
print("\n💾 VALKEY (Database):")
print("• tilt trigger migrate-up             - Run database migrations")
print("• tilt trigger migrate-down           - Rollback 1 migration")
print("• tilt trigger migrate-seed           - Reset and seed database")
print("• tilt trigger valkey-info            - Show Valkey info")
print("\n⚡ GENERAL:")
print("• tilt down                           - Stop all services")
print("• tilt up --stream                    - Start with streaming logs")

# Billy Wu Valkey Database

A high-performance Valkey cache database configured for session management, caching, and real-time data storage. Valkey is a Redis-compatible database that provides enhanced performance and modern features.

## 🏗️ Technology Stack

- **Database**: [Valkey 7.2](https://valkey.io/) - High-performance Redis-compatible database
- **Configuration**: Custom optimized configuration for development
- **Persistence**: AOF (Append Only File) + RDB snapshots
- **Features**: Enhanced multi-threading, improved memory management
- **Monitoring**: Built-in latency monitoring and slow query logging

## 📁 Project Structure

```
database/valkey/
├── valkey.conf              # Valkey configuration file
├── Dockerfile.dev           # Development Docker image
├── .dockerignore           # Docker ignore patterns
└── README.md               # This documentation
```

## 🚀 Getting Started

### Prerequisites

- Docker (for containerized deployment)
- Valkey CLI (for direct database interaction)

### Development Setup

The Valkey database is automatically started through the main project's Tilt or Docker Compose setup. However, you can also run it independently:

```bash
# Build and run the Valkey container
docker build -f Dockerfile.dev -t billy-wu-valkey-dev .
docker run -p 6379:6379 billy-wu-valkey-dev
```

### Connection Information

- **Host**: `localhost` (local development) or `valkey` (Docker network)
- **Port**: `6379`
- **Database**: `0` (default)
- **Authentication**: None (development mode)

## ⚙️ Configuration

### Development Optimizations

The Valkey configuration (`valkey.conf`) is optimized for development with the following key settings:

#### Network & Basic Settings

```conf
bind 0.0.0.0                 # Accept connections from any interface
port 6379                    # Standard Valkey/Redis port
timeout 0                    # No client timeout
tcp-keepalive 300           # Keep connections alive
```

#### Threading (Valkey Enhancement)

```conf
io-threads 4                 # Enable multi-threading for I/O
io-threads-do-reads yes      # Use threads for read operations
```

#### Memory Management

```conf
maxmemory-policy allkeys-lru # Evict least recently used keys when memory is full
```

#### Persistence Strategy

```conf
# RDB Snapshots
save 900 1                   # Save if at least 1 key changed in 900 seconds
save 300 10                  # Save if at least 10 keys changed in 300 seconds
save 60 10000               # Save if at least 10000 keys changed in 60 seconds

# AOF (Append Only File)
appendonly yes               # Enable AOF persistence
appendfilename "appendonly.aof"
appendfsync everysec        # Sync AOF every second
```

#### Development Features

```conf
# Slow Query Logging
slowlog-log-slower-than 10000  # Log queries slower than 10ms
slowlog-max-len 128            # Keep last 128 slow queries

# Latency Monitoring
latency-monitor-threshold 100   # Monitor operations taking >100ms

# Debugging
loglevel notice                 # Appropriate logging for development
```

## 💾 Data Storage & Usage

### Primary Use Cases

1. **Session Management**

   - User authentication sessions
   - JWT token storage and validation
   - Session expiration handling

2. **Application Caching**

   - API response caching
   - Database query result caching
   - Computed data caching

3. **Real-time Data**
   - WebSocket connection tracking
   - Real-time event storage
   - Temporary data storage

### Data Patterns

#### Session Storage

```bash
# Session key pattern
SET session:uuid:abc123 '{"userId":"user123","expires":"2024-01-01T00:00:00Z"}'
EXPIRE session:uuid:abc123 604800  # 7 days

# User session lookup
SET user:sessions:user123 '["abc123","def456"]'
```

#### Cache Storage

```bash
# API response caching
SET cache:api:users:list '{"users":[...],"timestamp":"2024-01-01T00:00:00Z"}'
EXPIRE cache:api:users:list 300  # 5 minutes

# Database query caching
SET cache:db:user:user123 '{"id":"user123","name":"John Doe"}'
EXPIRE cache:db:user:user123 3600  # 1 hour
```

#### Real-time Data

```bash
# WebSocket connections
SADD websocket:connections "conn123"
SET websocket:conn:conn123 '{"userId":"user123","connected":"2024-01-01T00:00:00Z"}'

# Event storage
LPUSH events:user:user123 '{"type":"login","timestamp":"2024-01-01T00:00:00Z"}'
LTRIM events:user:user123 0 99  # Keep last 100 events
```

## 🔧 Operations & Management

### Development Commands

#### Database Information

```bash
# Connect to Valkey CLI
docker exec -it <container_name> valkey-cli

# Basic information
INFO
INFO server
INFO memory
INFO persistence

# Database statistics
DBSIZE
INFO keyspace
```

#### Key Management

```bash
# List all keys (development only - avoid in production)
KEYS *

# Search for specific patterns
KEYS session:*
KEYS cache:*

# Key information
TYPE key_name
TTL key_name
OBJECT encoding key_name
```

#### Monitoring

```bash
# Monitor all commands in real-time
MONITOR

# View slow queries
SLOWLOG GET 10
SLOWLOG RESET

# Check latency
LATENCY LATEST
LATENCY HISTORY command-name
```

#### Memory Management

```bash
# Memory usage information
MEMORY USAGE key_name
MEMORY STATS

# Memory cleanup
FLUSHDB        # Clear current database
FLUSHALL       # Clear all databases (use with caution)
```

### Health Checks

The Docker container includes a health check:

```bash
# Manual health check
valkey-cli ping
# Expected response: PONG

# Detailed health verification
valkey-cli INFO server | grep uptime
valkey-cli INFO persistence | grep aof_enabled
```

### Backup & Recovery

#### Manual Backup

```bash
# Create RDB snapshot
BGSAVE

# Check backup status
LASTSAVE

# AOF rewrite (optimize AOF file)
BGREWRITEAOF
```

#### Data Export/Import

```bash
# Export data (development)
valkey-cli --rdb backup.rdb

# Import data
valkey-cli --pipe < backup.txt
```

## 🔍 Monitoring & Debugging

### Performance Monitoring

#### Slow Query Analysis

```bash
# View slow queries
SLOWLOG GET 10

# Example slow query analysis
1) 1) (integer) 0          # Query ID
   2) (integer) 1609459200 # Timestamp
   3) (integer) 12000      # Execution time (microseconds)
   4) 1) "GET"             # Command
      2) "large_key"       # Arguments
```

#### Memory Analysis

```bash
# Memory usage by data type
INFO memory

# Key space information
INFO keyspace

# Sample memory usage
MEMORY USAGE session:uuid:abc123
```

#### Latency Monitoring

```bash
# Check latency events
LATENCY LATEST

# Monitor specific commands
LATENCY MONITOR
```

### Development Debugging

#### Connection Testing

```bash
# Test connection
valkey-cli ping

# Test from application container
docker exec -it server_container valkey-cli -h valkey ping
```

#### Data Verification

```bash
# Check if session exists
EXISTS session:uuid:abc123

# Verify session data
GET session:uuid:abc123
TTL session:uuid:abc123
```

#### Performance Testing

```bash
# Simple benchmark
valkey-cli --latency
valkey-cli --latency-history -i 1

# Throughput testing
valkey-cli eval "for i=1,100000 do redis.call('set','key'..i,'value'..i) end" 0
```

## 🛠️ Configuration Tuning

### Memory Optimization

```conf
# Adjust based on available memory
maxmemory 256mb
maxmemory-policy allkeys-lru

# Hash table optimizations
hash-max-ziplist-entries 512
hash-max-ziplist-value 64
```

### Performance Tuning

```conf
# I/O threading (Valkey advantage)
io-threads 4
io-threads-do-reads yes

# TCP settings
tcp-backlog 511
tcp-keepalive 300
```

### Persistence Tuning

```conf
# For session-heavy workloads
save 300 10
save 60 10000

# AOF tuning
appendfsync everysec
no-appendfsync-on-rewrite no
```

## 🔧 Troubleshooting

### Common Issues

1. **Connection Refused**

   ```bash
   # Check if container is running
   docker ps | grep valkey

   # Check logs
   docker logs <valkey_container>

   # Verify port binding
   docker port <valkey_container>
   ```

2. **Memory Issues**

   ```bash
   # Check memory usage
   valkey-cli INFO memory

   # Check if keys are expiring
   valkey-cli INFO keyspace

   # Manual cleanup if needed
   valkey-cli FLUSHDB
   ```

3. **Performance Issues**

   ```bash
   # Check slow queries
   valkey-cli SLOWLOG GET 10

   # Monitor latency
   valkey-cli --latency

   # Check if AOF rewrite is needed
   valkey-cli INFO persistence
   ```

4. **Persistence Issues**

   ```bash
   # Check AOF status
   valkey-cli INFO persistence | grep aof

   # Force AOF rewrite
   valkey-cli BGREWRITEAOF

   # Check RDB save status
   valkey-cli LASTSAVE
   ```

### Development Tips

- Use `MONITOR` command to debug application interactions
- Check `SLOWLOG` regularly to identify performance bottlenecks
- Use `INFO` command to monitor memory usage and connection counts
- Enable latency monitoring for performance analysis
- Use `EXPLAIN` (when available) to understand query performance

## 🔒 Security Considerations

### Development Security

- No authentication required (development only)
- Bind to all interfaces for container accessibility
- Protected mode disabled for easy development

### Production Recommendations

```conf
# Enable authentication
requirepass your_secure_password

# Restrict network access
bind 127.0.0.1

# Enable protected mode
protected-mode yes

# Disable dangerous commands
rename-command FLUSHDB ""
rename-command FLUSHALL ""
rename-command CONFIG "CONFIG_9a8b7c6d"
```

## 🚀 Production Migration

### Production Configuration Changes

1. **Security**: Enable authentication and restrict network access
2. **Memory**: Set appropriate `maxmemory` limits
3. **Persistence**: Adjust save intervals based on data criticality
4. **Monitoring**: Enable comprehensive logging and monitoring
5. **Backup**: Implement automated backup strategies

### High Availability

- Consider Redis Sentinel for automatic failover
- Implement read replicas for scaling
- Use Redis Cluster for horizontal scaling

## 🤝 Contributing

1. **Configuration Changes**: Test thoroughly in development environment
2. **Performance Tuning**: Benchmark before and after changes
3. **Documentation**: Update README for any configuration changes
4. **Monitoring**: Ensure changes don't negatively impact observability

## 📚 Additional Resources

- [Valkey Documentation](https://valkey.io/documentation/)
- [Valkey vs Redis Comparison](https://valkey.io/blog/introducing-valkey/)
- [Valkey Configuration Reference](https://valkey.io/topics/config/)
- [Performance Best Practices](https://valkey.io/topics/memory-optimization/)
- [Persistence Configuration](https://valkey.io/topics/persistence/)

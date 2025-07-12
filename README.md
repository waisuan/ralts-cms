# Ralts-CMS

A web service built in Go for managing machines and their maintenance records. It provides a RESTful API for querying and mutating machine and maintenance data stored in PostgreSQL.

## Features

- Machine management (CRUD operations)
- Maintenance record management (CRUD operations)
- JWT Token authentication
- PostgreSQL database support
- Configurable HTTP timeouts
- RESTful API design

## Prerequisites

- Go 1.24 or later
- Docker and Docker Compose (for local development)
- PostgreSQL client tools (optional, for database management)

## Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd ralts-cms
   ```

2. **Setup and run the application**
   ```bash
   make dev
   ```

This command will:
- Start a local PostgreSQL instance
- Create the required database schema
- Start the web server on port 8080

## Database Setup

The project includes PostgreSQL support via Docker Compose for development and testing.

### Starting PostgreSQL

```bash
# Start PostgreSQL only
docker compose up postgres

# Start all services
docker compose up
```

### Running Database Migrations

The project uses [golang-migrate/migrate](https://github.com/golang-migrate/migrate) for database migrations.

#### Install Migration Tool

```bash
# Using Go
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Using Docker
docker pull migrate/migrate
```

#### Basic Migration Commands

```bash
# Apply all pending migrations
migrate -path ./db/migrations -database "postgresql://ralts_user:ralts_password@localhost:5432/ralts_cms?sslmode=disable" up

# Apply specific number of migrations
migrate -path ./db/migrations -database "postgresql://ralts_user:ralts_password@localhost:5432/ralts_cms?sslmode=disable" up 1

# Rollback last migration
migrate -path ./db/migrations -database "postgresql://ralts_user:ralts_password@localhost:5432/ralts_cms?sslmode=disable" down 1

# Rollback all migrations
migrate -path ./db/migrations -database "postgresql://ralts_user:ralts_password@localhost:5432/ralts_cms?sslmode=disable" down

# Check migration status
migrate -path ./db/migrations -database "postgresql://ralts_user:ralts_password@localhost:5432/ralts_cms?sslmode=disable" version

# Force migration version (use with caution)
migrate -path ./db/migrations -database "postgresql://ralts_user:ralts_password@localhost:5432/ralts_cms?sslmode=disable" force VERSION
```

#### Using Docker for Migrations

```bash
# Apply migrations using Docker
docker run -v $(pwd)/db/migrations:/migrations --network host migrate/migrate \
  -path=/migrations/ \
  -database "postgresql://ralts_user:ralts_password@localhost:5432/ralts_cms?sslmode=disable" \
  up

# Check version using Docker
docker run -v $(pwd)/db/migrations:/migrations --network host migrate/migrate \
  -path=/migrations/ \
  -database "postgresql://ralts_user:ralts_password@localhost:5432/ralts_cms?sslmode=disable" \
  version
```

#### Environment Variables

You can use environment variables for the database connection:

```bash
export DATABASE_URL="postgresql://ralts_user:ralts_password@localhost:5432/ralts_cms?sslmode=disable"

# Then use the environment variable
migrate -path ./db/migrations -database $DATABASE_URL up
```

#### Creating New Migrations

```bash
# Create a new migration
migrate create -ext sql -dir ./db/migrations -seq add_user_table

# This creates:
# - 000002_add_user_table.up.sql
# - 000002_add_user_table.down.sql
```

#### Migration Files

Migrations are numbered sequentially and have both up and down versions:

- `000001_init_schema.up.sql` - Creates the initial database schema
- `000001_init_schema.down.sql` - Reverses the initial schema creation

**Migration Naming Convention:**
- Format: `{version}_{description}.{up|down}.sql`
- Version: 6-digit zero-padded number (e.g., 000001, 000002)
- Description: Descriptive name in snake_case
- Direction: `up` for applying, `down` for reversing

#### Best Practices

1. **Always test both up and down migrations**
2. **Use transactions** (migrate handles this automatically)
3. **Keep migrations idempotent** when possible
4. **Use descriptive names** for migration files
5. **Include comments** explaining complex migrations
6. **Test migrations on a copy of production data**
7. **Never modify existing migrations** that have been applied

#### Troubleshooting

**Common Issues:**
1. **Migration already applied**: Use `force` command to set correct version
2. **Connection issues**: Check database URL and network connectivity
3. **Permission errors**: Ensure database user has proper permissions

**Reset Database:**
```bash
# Drop and recreate database (WARNING: This will delete all data)
migrate -path ./db/migrations -database "postgresql://ralts_user:ralts_password@localhost:5432/ralts_cms?sslmode=disable" down
migrate -path ./db/migrations -database "postgresql://ralts_user:ralts_password@localhost:5432/ralts_cms?sslmode=disable" up
```

### Connecting to PostgreSQL with psql

#### Method 1: Using Docker exec (Recommended)
```bash
# Connect directly to the PostgreSQL container
docker exec -it postgres psql -U ralts_user -d ralts_cms

# Or with explicit host and port
docker exec -it postgres psql -h localhost -p 5432 -U ralts_user -d ralts_cms
```

#### Method 2: Using local psql client
If you have PostgreSQL client tools installed locally:
```bash
# Connect from your host machine
psql -h localhost -p 5432 -U ralts_user -d ralts_cms

# When prompted for password, enter: ralts_password
```

#### Method 3: Using connection string
```bash
# Using connection string format
psql "postgresql://ralts_user:ralts_password@localhost:5432/ralts_cms"
```

### PostgreSQL Connection Details

- **Host**: localhost
- **Port**: 5432
- **Database**: ralts_cms
- **Username**: ralts_user
- **Password**: ralts_password

### PostgreSQL Connection Pool Configuration

The application uses a configurable connection pool to manage database connections efficiently. You can customize these settings based on your application's needs:

#### Connection Pool Settings

| Setting | Environment Variable | Default | Description |
|---------|---------------------|---------|-------------|
| Max Open Connections | `POSTGRES_MAX_OPEN_CONNS` | 25 | Maximum number of open database connections |
| Max Idle Connections | `POSTGRES_MAX_IDLE_CONNS` | 5 | Maximum number of idle database connections |
| Connection Max Lifetime | `POSTGRES_CONN_MAX_LIFETIME` | 5m | Maximum lifetime of database connections |

#### Configuration Examples

**Development (default):**
```env
POSTGRES_MAX_OPEN_CONNS=25
POSTGRES_MAX_IDLE_CONNS=5
POSTGRES_CONN_MAX_LIFETIME=5m
```

**Production (high load):**
```env
POSTGRES_MAX_OPEN_CONNS=50
POSTGRES_MAX_IDLE_CONNS=10
POSTGRES_CONN_MAX_LIFETIME=10m
```

**Production (very high load):**
```env
POSTGRES_MAX_OPEN_CONNS=100
POSTGRES_MAX_IDLE_CONNS=20
POSTGRES_CONN_MAX_LIFETIME=15m
```

#### Best Practices

1. **Max Open Connections**: Should not exceed your PostgreSQL server's `max_connections` setting
2. **Max Idle Connections**: Keep some connections warm for better performance
3. **Connection Lifetime**: Set to a reasonable value to prevent connection staleness
4. **Monitor**: Use PostgreSQL monitoring tools to track connection usage

#### Troubleshooting Connection Pool Issues

**Common Issues:**
- **"too many connections"**: Increase `max_connections` in PostgreSQL or reduce `POSTGRES_MAX_OPEN_CONNS`
- **Connection timeouts**: Check network connectivity and increase timeouts if needed
- **Performance issues**: Monitor connection pool metrics and adjust settings accordingly

### Useful psql Commands

Once connected to PostgreSQL, here are some useful commands:

```sql
-- List all tables
\dt

-- Describe table structure
\d machines
\d maintenance

-- View sample data
SELECT * FROM machines LIMIT 5;
SELECT * FROM maintenance LIMIT 5;

-- Check database size
SELECT pg_size_pretty(pg_database_size('ralts_cms'));

-- List all databases
\l

-- Switch databases
\c another_database

-- Exit psql
\q
```

### Data Persistence

PostgreSQL data is persisted using Docker volumes. Your data will survive:
- Container restarts
- Docker Compose down/up cycles
- System reboots

The data is stored in the `postgres_data` volume and can be found at:
```bash
# View volume information
docker volume ls | grep postgres_data

# Inspect volume details
docker volume inspect ralts-cms_postgres_data
```

## Manual Setup

If you prefer to set up manually:

1. **Start PostgreSQL**
   ```bash
   docker compose up postgres
   ```

2. **Run the application**
   ```bash
   make run
   ```

## Configuration

The application uses environment variables for configuration. Create a `.env.development` file:

### Development Configuration
```env
APP_NAME=ralts-cms
APP_ENV=development
DATABASE_URL=postgresql://ralts_user:ralts_password@localhost:5432/ralts_cms?sslmode=disable
SERVER_PORT=8080
HTTP_READ_TIMEOUT=15s
HTTP_WRITE_TIMEOUT=15s
HTTP_IDLE_TIMEOUT=60s
JWT_SECRET=your-jwt-secret-key
DEFAULT_MACHINE_LIMIT=50
```

### Production Configuration
```env
APP_NAME=ralts-cms
APP_ENV=production
DATABASE_URL=postgresql://username:password@your-postgres-host:5432/ralts_cms?sslmode=require
SERVER_PORT=8080
HTTP_READ_TIMEOUT=30s
HTTP_WRITE_TIMEOUT=30s
HTTP_IDLE_TIMEOUT=120s
JWT_SECRET=your-production-jwt-secret-key
DEFAULT_MACHINE_LIMIT=100
```

### Configuration Options

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `APP_NAME` | Application name | - | Yes |
| `APP_ENV` | Environment (development/production) | development | No |
| `DATABASE_URL` | PostgreSQL connection string | - | Yes |
| `SERVER_PORT` | HTTP server port | 8080 | No |
| `HTTP_READ_TIMEOUT` | HTTP read timeout | 15s | No |
| `HTTP_WRITE_TIMEOUT` | HTTP write timeout | 15s | No |
| `HTTP_IDLE_TIMEOUT` | HTTP idle timeout | 60s | No |
| `JWT_SECRET` | JWT secret key | - | Yes |
| `DEFAULT_MACHINE_LIMIT` | Default number of machines returned per page | 50 | No |
| `POSTGRES_MAX_OPEN_CONNS` | Maximum number of open database connections | 25 | No |
| `POSTGRES_MAX_IDLE_CONNS` | Maximum number of idle database connections | 5 | No |
| `POSTGRES_CONN_MAX_LIFETIME` | Maximum lifetime of database connections | 5m | No |

## Authentication

Ralts-CMS uses JWT (JSON Web Token) authentication for all protected API endpoints.

### Configuration

Set the `JWT_SECRET` environment variable with your secret key:

```bash
# Development
JWT_SECRET=your-jwt-secret-key

# Production
JWT_SECRET=your-production-jwt-secret-key
```

### Usage

Include the JWT token in the `Authorization` header for all API requests:

```bash
curl -H "Authorization: Bearer <your-jwt-token>" \
     http://localhost:8080/machines
```

### JWT Token Structure

JWT tokens should contain the following claims:
- `sub` (subject): User identifier
- `exp` (expiration): Token expiration time
- `iat` (issued at): Token creation time

### Protected Endpoints

All endpoints except `/health` require authentication:

- `GET /machines` - List machines
- `GET /machines/{serial_number}` - Get specific machine
- `POST /machines` - Create machine
- `PUT /machines` - Update machine
- `DELETE /machines/{serial_number}` - Delete machine
- `GET /machines/{serial_number}/maintenance` - List maintenance records
- `GET /machines/{serial_number}/maintenance/{work_order_number}` - Get specific maintenance
- `POST /machines/{serial_number}/maintenance` - Create maintenance record
- `PUT /machines/{serial_number}/maintenance` - Update maintenance record
- `DELETE /machines/{serial_number}/maintenance/{work_order_number}` - Delete maintenance record

### Public Endpoints

- `GET /health` - Health check (no authentication required)

### Error Responses

- `401 Unauthorized` - Missing or invalid Authorization header
- `401 Unauthorized` - Invalid JWT token format
- `401 Unauthorized` - Invalid JWT token signature
- `401 Unauthorized` - JWT token expired

### Security Best Practices

1. Use a strong, randomly generated secret key (at least 32 characters)
2. Keep the secret key secure and never commit it to version control
3. Rotate secret keys regularly
4. Use HTTPS in production
5. Set appropriate token expiration times
6. Validate token claims in your application logic

### Example with curl

```bash
# List all machines
curl -H "Authorization: Bearer <your-jwt-token>" \
     http://localhost:8080/machines

# Get specific machine
curl -H "Authorization: Bearer <your-jwt-token>" \
     http://localhost:8080/machines/MACHINE123

# Create a new machine
curl -X POST \
     -H "Authorization: Bearer <your-jwt-token>" \
     -H "Content-Type: application/json" \
     -d '{"serial_number":"MACHINE123","customer":"ACME Corp"}' \
     http://localhost:8080/machines
```

### Generating JWT Tokens

You can generate JWT tokens using various tools or libraries. Here's an example using the `jwt-cli` tool:

```bash
# Install jwt-cli
go install github.com/golang-jwt/jwt/v5/cmd/jwt@latest

# Generate a token
jwt encode --secret "your-jwt-secret-key" --claim "sub=test-user" --claim "exp=$(date -d '+1 hour' +%s)"
```

## Production Deployment

For production deployment, the application uses PostgreSQL for data storage:

1. **Set environment to production**
   ```env
   APP_ENV=production
   ```

2. **Configure PostgreSQL connection**
   ```env
   DATABASE_URL=postgresql://username:password@your-postgres-host:5432/ralts_cms?sslmode=require
   ```

3. **Configure HTTP timeouts for production load**
   ```env
   HTTP_READ_TIMEOUT=30s
   HTTP_WRITE_TIMEOUT=30s
   HTTP_IDLE_TIMEOUT=120s
   ```

4. **Configure PostgreSQL connection pool for production load**
   ```env
   POSTGRES_MAX_OPEN_CONNS=50
   POSTGRES_MAX_IDLE_CONNS=10
   POSTGRES_CONN_MAX_LIFETIME=10m
   ```

5. **Set a secure JWT secret**
   ```env
   JWT_SECRET=your-production-jwt-secret-key
   ```

## API Endpoints

All endpoints require Bearer Token authentication. The token is specified in the `JWT_SECRET` environment variable.

### Machine Endpoints

- `GET /machines` - List all machines
- `GET /machines/{serial_number}` - Get a machine by serial number
- `POST /machines` - Create a new machine
- `PUT /machines` - Update an existing machine
- `DELETE /machines/{serial_number}` - Delete a machine

### Maintenance Endpoints

- `GET /machines/{serial_number}/maintenance` - List maintenance records for a machine
- `GET /machines/{serial_number}/maintenance/{work_order_number}` - Get a specific maintenance record
- `POST /machines/{serial_number}/maintenance` - Create a new maintenance record
- `PUT /machines/{serial_number}/maintenance` - Update an existing maintenance record
- `DELETE /machines/{serial_number}/maintenance/{work_order_number}` - Delete a maintenance record

### Health Check

- `GET /health` - Health check endpoint (no authentication required)

## Data Models

### Machine

```json
{
  "serial_number": "string",
  "customer": "string",
  "state": "string",
  "account_type": "string",
  "model": "string",
  "status": "string",
  "brand": "string",
  "district": "string",
  "person_in_charge": "string",
  "reported_by": "string",
  "additional_notes": "string",
  "attachment": "string",
  "ppm_status": "string",
  "tnc_date": "string (ISO 8601)",
  "ppm_date": "string (ISO 8601)",
  "created_at": "string (ISO 8601)",
  "updated_at": "string (ISO 8601)"
}
```

### Maintenance

```json
{
  "machine_serial_number": "string",
  "work_order_number": "string",
  "work_order_date": "string (ISO 8601)",
  "action_taken": "string",
  "reported_by": "string",
  "worker_order_type": "string",
  "attachment": "string",
  "created_at": "string (ISO 8601)",
  "updated_at": "string (ISO 8601)"
}
```

## Development

### Building

```bash
make build
```

### Running Tests

```bash
make test
```

### Formatting Code

```bash
make fmt
```

## Scripts

The `scripts/` directory contains utility scripts for development and testing:

### JWT Token Generator

- `scripts/jwt/` - JWT token generator for local testing

Generate a valid JWT token for testing:

```bash
go run scripts/jwt/main.go
```

This generates a JWT token with:
- **Secret**: Uses the default JWT secret from your config
- **Subject**: `test-user`
- **Expiration**: 24 hours from generation
- **Algorithm**: HS256

The script provides the complete Authorization header and example curl commands for testing API endpoints.

## Project Structure

```
ralts-cms/
├── cmd/
│   └── web/
│       └── main.go          # Application entry point
├── internal/
│   ├── deps/
│   │   ├── config.go        # Configuration management
│   │   ├── deps.go          # Dependency injection
│   │   └── database.go      # Database setup and configuration
│   ├── handler/
│   │   ├── machine.go       # Machine HTTP handlers
│   │   ├── maintenance.go   # Maintenance HTTP handlers
│   │   ├── health.go        # Health check handler
│   │   └── middleware.go    # HTTP middleware
│   ├── machine/
│   │   ├── machine.go       # Machine domain model
│   │   └── repository.go    # Machine data access
│   ├── maintenance/
│   │   ├── maintenance.go   # Maintenance domain model
│   │   └── repository.go    # Maintenance data access
│   └── router/
│       └── router.go        # HTTP routing
├── db/
│   └── migrations/          # Database migrations
│       ├── 000001_init_schema.up.sql
│       └── 000001_init_schema.down.sql
├── scripts/
│   └── jwt/                 # JWT token generator
├── docker-compose.yml       # Local PostgreSQL setup
├── Makefile                 # Build and development commands
└── README.md               # This file
```

## License

[Add your license information here] 
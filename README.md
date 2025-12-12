# Ralts-CMS

A web service built in Go for managing machines and their maintenance records. It provides a RESTful API for querying and mutating machine and maintenance data stored in PostgreSQL.

## Features

- Machine management (CRUD operations)
- Maintenance record management (CRUD operations)
- File attachment support (S3-compatible storage)
- JWT Token authentication
- PostgreSQL database support
- Configurable HTTP timeouts
- RESTful API design
- **Audit Logging** - Comprehensive event tracking for all user actions

## Prerequisites

- Go 1.24 or later
- Docker and Docker Compose (for local development)
- PostgreSQL client tools (optional, for database management)

## Quick Start

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd ralts-cms
   ```

2. **Setup development environment**
   ```bash
   make dev
   ```

This command will:
- Create environment files (`.env.development` and `.env.test`)
- Start both development and test PostgreSQL databases

3. **Run database migrations**
   ```bash
   make migrate-up
   ```

This will set up the database schema in both development and test databases.

4. **Start the development server**
   ```bash
   make run
   ```

Or use the combined command:
```bash
make dev-with-db
```

This will start the development server with the database automatically.

### Additional Development Setup

**Seed Test Data** (Optional):
```bash
make seed-dev
```
This populates your development database with sample data for testing:
- 100 sample machines with realistic random data (customers, states, models, etc.)
- 10 sample users with test credentials and various roles
- Associated maintenance records for each machine

**LocalStack for S3 Testing** (Optional):
```bash
make localstack-up
```
This starts a local AWS S3-compatible service at `http://localhost:4566` for testing file attachments without needing real AWS credentials. The service provides:
- S3 bucket simulation for file uploads/downloads
- Compatible with AWS SDK calls
- Automatically included when running `make dev`

**Complete Development Workflow**:
```bash
# One-time setup
make dev           # Sets up environment, databases, and LocalStack
make migrate-up    # Apply database schema
make seed-dev      # Add sample data (optional)

# Daily development
make run           # Start the web server
```

## Database Setup

The project includes PostgreSQL support via Docker Compose for development and testing.

### Starting PostgreSQL

```bash
# Start development database
make db-dev-up

# Start test database
make db-test-up

# Start both databases
make db-up
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

## Database Setup Guide

This project uses PostgreSQL with Docker for both development and testing environments. The databases are completely isolated to prevent data conflicts.

### Quick Start

#### Development Database
```bash
# Start development database
make db-dev-up

# Start development server with database
make dev-with-db

# Stop development database
make db-dev-down
```

#### Test Database
```bash
# Start test database and run tests
make test-with-db

# Start test database only
make db-test-up

# Stop test database
make db-test-down
```

#### Combined Commands
```bash
# Start both databases
make db-up

# Stop both databases
make db-down

# Check database status
make db-status

# Clean up all databases and volumes
make db-clean
```

### Database Configuration

#### Development Database
- **Port**: 5432
- **Database**: `ralts_cms_dev`
- **User**: `ralts_user`
- **Password**: `ralts_password`
- **Container**: `postgres-dev`
- **Volume**: `postgres_dev_data`

#### Test Database
- **Port**: 5433
- **Database**: `ralts_cms_test`
- **User**: `ralts_user`
- **Password**: `ralts_password`
- **Container**: `postgres-test`
- **Volume**: `postgres_test_data`

### Environment Variables

For your application to connect to the databases, set these environment variables:

#### Development
```bash
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=ralts_cms_dev
POSTGRES_USER=ralts_user
POSTGRES_PASSWORD=ralts_password
```

#### Testing
```bash
POSTGRES_HOST=localhost
POSTGRES_PORT=5433
POSTGRES_DB=ralts_cms_test
POSTGRES_USER=ralts_user
POSTGRES_PASSWORD=ralts_password
```

### Available Make Commands

#### Database Management
- `make db-dev-up` - Start development database
- `make db-dev-down` - Stop development database
- `make db-dev-logs` - View development database logs
- `make db-dev-connect` - Connect to development database via psql

- `make db-test-up` - Start test database
- `make db-test-down` - Stop test database
- `make db-test-logs` - View test database logs
- `make db-test-connect` - Connect to test database via psql

#### Combined Commands
- `make db-up` - Start both databases
- `make db-down` - Stop both databases
- `make db-status` - Check status of all databases
- `make db-clean` - Remove all containers and volumes

#### Development Workflow
- `make dev` - Setup development environment (env files + databases)
- `make dev-with-db` - Start development server with database
- `make test-with-db` - Run tests with test database

#### Migration Commands
- `make migrate-dev` - Run migrations on development database
- `make migrate-test` - Run migrations on test database
- `make migrate-up` - Run migrations on both databases
- `make migrate-down` - Rollback migrations on both databases

### Data Isolation

The databases are completely isolated:

1. **Separate Containers**: Each environment has its own container
2. **Separate Volumes**: Data is stored in different Docker volumes
3. **Separate Ports**: Development uses port 5432, testing uses port 5433
4. **Separate Database Names**: `ralts_cms_dev` vs `ralts_cms_test`

### Troubleshooting

#### Database Won't Start
```bash
# Check if port is already in use
lsof -i :5432
lsof -i :5433

# Clean up and restart
make db-clean
make db-up
```

#### Connection Issues
```bash
# Check database status
make db-status

# View logs
make db-dev-logs
make db-test-logs

# Test connection
make db-dev-connect
make db-test-connect
```

#### Reset Database
```bash
# Remove all data and start fresh
make db-clean
make db-up
```

### Docker Compose Files

- `docker-compose.dev.yml` - Development database configuration
- `docker-compose.test.yml` - Test database configuration

### Migration Scripts

Database initialization scripts should be placed in the `init-scripts/` directory. These will be automatically executed when the containers start for the first time.

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
PORT=8080
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
PORT=8080
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
| `PORT` | HTTP server port | 8080 | No |
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

- `GET /machines` - List machines with pagination and sorting
- `GET /machines/{serial_number}` - Get specific machine
- `POST /machines` - Create machine
- `PUT /machines` - Update machine
- `DELETE /machines/{serial_number}` - Delete machine
- `GET /machines/due-ppm` - Get machines due for PPM
- `GET /machines/{serial_number}/maintenance` - List maintenance records with pagination and sorting
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
# List all machines with pagination
curl -H "Authorization: Bearer <your-jwt-token>" \
     "http://localhost:8080/machines?limit=10&offset=0&sort=created_at_desc"

# Get specific machine
curl -H "Authorization: Bearer <your-jwt-token>" \
     http://localhost:8080/machines/MACHINE123

# Get machines due for PPM
curl -H "Authorization: Bearer <your-jwt-token>" \
     http://localhost:8080/machines/due-ppm

# Create a new machine
curl -X POST \
     -H "Authorization: Bearer <your-jwt-token>" \
     -H "Content-Type: application/json" \
     -d '{"serial_number":"MACHINE123","customer":"ACME Corp"}' \
     http://localhost:8080/machines

# List maintenance records for a machine
curl -H "Authorization: Bearer <your-jwt-token>" \
     "http://localhost:8080/machines/MACHINE123/maintenance?limit=5&sort=work_order_date_desc"
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

6. **Configure S3 and AWS settings for file attachments**
   ```env
   S3_BUCKET_NAME=your-production-s3-bucket-name
   AWS_ACCESS_KEY_ID=your-aws-access-key-id
   AWS_SECRET_ACCESS_KEY=your-aws-secret-access-key
   AWS_DEFAULT_REGION=your-aws-region
   AWS_ENDPOINT_URL=https://s3.amazonaws.com
   AWS_S3_FORCE_PATH_STYLE=false
   ```

## API Endpoints

All endpoints require Bearer Token authentication except for the health check endpoint. The token is specified in the `JWT_SECRET` environment variable.

### Machine Endpoints

- `GET /machines` - List all machines with pagination and sorting
- `GET /machines/{serial_number}` - Get a machine by serial number
- `POST /machines` - Create a new machine
- `PUT /machines` - Update an existing machine
- `DELETE /machines/{serial_number}` - Delete a machine
- `GET /machines/due-ppm` - Get machines that are due for PPM (Preventive Planned Maintenance)

### Maintenance Endpoints

- `GET /machines/{serial_number}/maintenance` - List maintenance records for a machine with pagination and sorting
- `GET /machines/{serial_number}/maintenance/{work_order_number}` - Get a specific maintenance record
- `POST /machines/{serial_number}/maintenance` - Create a new maintenance record
- `PUT /machines/{serial_number}/maintenance` - Update an existing maintenance record
- `DELETE /machines/{serial_number}/maintenance/{work_order_number}` - Delete a maintenance record

### Health Check

- `GET /health` - Health check endpoint (no authentication required)

### Query Parameters

#### Machine List Endpoint (`GET /machines`)

- `limit` (optional): Number of machines to return (1-100, default: 50)
- `offset` (optional): Number of machines to skip (default: 0)
- `sort` (optional): Sort order
  - `created_at_desc` (default): Newest first
  - `created_at_asc`: Oldest first

#### Maintenance List Endpoint (`GET /machines/{serial_number}/maintenance`)

- `limit` (optional): Number of maintenance records to return (1-100, default: 50)
- `offset` (optional): Number of maintenance records to skip (default: 0)
- `sort` (optional): Sort order
  - `work_order_date_desc` (default): Most recent work order first
  - `work_order_date_asc`: Oldest work order first
  - `created_at_desc`: Newest created first
  - `created_at_asc`: Oldest created first

### Example Requests

```bash
# List machines with pagination
curl -H "Authorization: Bearer <your-jwt-token>" \
     "http://localhost:8080/machines?limit=10&offset=0&sort=created_at_desc"

# Get machines due for PPM
curl -H "Authorization: Bearer <your-jwt-token>" \
     http://localhost:8080/machines/due-ppm

# List maintenance records for a machine
curl -H "Authorization: Bearer <your-jwt-token>" \
     "http://localhost:8080/machines/MACHINE123/maintenance?limit=5&sort=work_order_date_desc"
```

## Data Models

### Machine

```json
{
  "id": 1,
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
  "tnc_date": "2024-01-01T00:00:00Z",
  "ppm_date": "2024-01-01T00:00:00Z",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Maintenance

```json
{
  "id": 1,
  "machine_serial_number": "string",
  "work_order_number": "string",
  "work_order_date": "2024-01-01T00:00:00Z",
  "action_taken": "string",
  "reported_by": "string",
  "work_order_type": "string",
  "attachment": "string",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### List Response Format

Both machine and maintenance list endpoints return responses in the following format:

```json
{
  "machines": [...], // or "maintenance": [...] for maintenance endpoints
  "count": 10,
  "limit": 50,
  "offset": 0,
  "sort": "created_at_desc"
}
```

## Audit Logging

The application includes comprehensive audit logging that tracks all user interactions with the system. Audit events are stored in the database and can be queried for compliance, debugging, and monitoring purposes.

### What Gets Audited

Every API endpoint is audited with the following information:
- **User ID**: The authenticated user who performed the action
- **Action**: The type of action (created, updated, deleted, viewed, listed, login, logout, password_changed)
- **Resource Type**: The type of resource (machine, maintenance, user, session, attachment)
- **Resource ID**: The identifier of the affected resource
- **Details**: Additional context-specific information (JSON)
- **Timestamp**: When the event occurred

### Audit Actions

| Action | Description |
|--------|-------------|
| `created` | A new resource was created |
| `updated` | An existing resource was modified |
| `deleted` | A resource was removed |
| `viewed` | A single resource was retrieved |
| `listed` | Multiple resources were queried/searched |
| `login` | User successfully authenticated |
| `logout` | User logged out |
| `password_changed` | User changed their password |

### Audit Resource Types

| Resource | Description |
|----------|-------------|
| `machine` | Machine records |
| `maintenance` | Maintenance records |
| `user` | User accounts |
| `session` | Authentication sessions |
| `attachment` | File attachments |

### Architecture

Audit logging is implemented asynchronously to avoid impacting API performance:

1. **Non-blocking**: API handlers queue events to a buffered channel and return immediately
2. **Background Worker**: A goroutine processes events and persists them to PostgreSQL
3. **Graceful Shutdown**: The service drains remaining events before the application exits
4. **Buffer Overflow**: If the buffer is full (1000 events), events are dropped with a warning log

### Database Schema

Audit events are stored in the `audit_logs` table:

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

Indexes are created for efficient querying:
- `idx_audit_logs_created_at` - For time-based queries
- `idx_audit_logs_user_id` - For user-specific queries
- `idx_audit_logs_resource` - For resource-based queries

## CLI Tools

### Admin Account Creation

The application includes a CLI tool for creating admin accounts. This is useful for initial setup when you need to create the first admin user.

```bash
# Build the CLI tool
go build -o admin-cli ./cmd/cli

# Create an admin account with username and password
./admin-cli -type admin -username myadmin -password mypassword123
```

Requirements:
- **Username**: Must be at least 3 characters long
- **Password**: Must be at least 8 characters long
- **Email**: Automatically generated as `{username}@admin.local`

The admin account will be created with:
- Role: `ADMIN`
- Status: `APPROVED` (automatically approved)
- Full access to the admin panel

### Query Audit Events

The CLI tool can query audit events from the database for debugging and monitoring:

```bash
# Show events from the last 30 minutes (default)
./admin-cli -type events

# Show events from the last 60 minutes
./admin-cli -type events -minutes 60

# Filter by action type
./admin-cli -type events -action created
./admin-cli -type events -action login

# Filter by resource type
./admin-cli -type events -resource machine
./admin-cli -type events -resource user

# Combine filters
./admin-cli -type events -minutes 120 -action deleted -resource machine

# Limit results
./admin-cli -type events -limit 100
```

#### Event Query Options

| Flag | Default | Description |
|------|---------|-------------|
| `-minutes` | 30 | Show events from the last N minutes |
| `-action` | (all) | Filter by action type |
| `-resource` | (all) | Filter by resource type |
| `-limit` | 50 | Maximum number of events to return |

#### Example Output

```
📋 Audit Events (last 30 minutes)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Found 5 events:

[1] 2024-01-15 14:32:15
    Action:   created
    Resource: machine (SN-001234)
    User ID:  42
    Details:  {
                "customer": "Acme Corp",
                "model": "X100"
              }

[2] 2024-01-15 14:30:00
    Action:   login
    Resource: session (42)
    User ID:  42
    Details:  {
                "username": "admin"
              }

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Total: 5 events
By action: created=3, login=2
By resource: machine=3, session=2
```

### Test Data Generation

You can also use the CLI tool to generate test data:

```bash
# Generate test machines (will delete existing data)
./admin-cli -type machine -count 50

# Generate test users (will delete existing data)
./admin-cli -type user -count 20
```

**Note**: The test data generation commands will delete all existing data of that type before creating new records.

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

### Linting Code

```bash
make lint
```

This command will:
- Download the [Revive](https://github.com/mgechev/revive) linter binary to the `bin/` folder if it doesn't exist
- Run the linter with friendly formatting on all Go files in the project
- Use the existing `revive.toml` configuration file for linting rules

## Frontend

The frontend is a modern Next.js application built with React and TypeScript, providing a user-friendly interface for machine and maintenance management.

### Frontend Architecture

- **Framework**: Next.js 15 with React 19
- **Language**: TypeScript for type safety
- **Styling**: Tailwind CSS for modern, responsive design
- **Testing**: Jest with Testing Library for comprehensive testing

### Key Features

- Machine management interface (view, create, edit, delete)
- Maintenance record tracking and management
- Responsive design for desktop and mobile devices
- JWT authentication integration with the backend API
- Real-time data synchronization with the Go backend

### Getting Started

Navigate to the `ui/` directory and follow the setup instructions in [`ui/README.md`](ui/README.md) for detailed frontend development information.

### Backend Integration

The frontend communicates with the Go backend through RESTful API calls, handling:
- JWT authentication for secure access
- Machine data CRUD operations
- Maintenance record management
- Real-time updates and error handling

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
├── ui/                      # Next.js frontend application
│   ├── src/
│   │   ├── app/             # Next.js app router pages
│   │   ├── components/      # Reusable UI components
│   │   ├── services/        # API client services
│   │   ├── hooks/           # Custom React hooks
│   │   ├── types/           # TypeScript type definitions
│   │   └── utils/           # Utility functions
│   ├── package.json         # Frontend dependencies
│   └── README.md           # Frontend setup guide
├── db/
│   └── migrations/          # Database migrations
│       ├── 000001_add_machine_table.up.sql
│       ├── 000001_add_machine_table.down.sql
│       ├── 000002_add_maintenance_table.up.sql
│       └── 000002_add_maintenance_table.down.sql
├── scripts/
│   └── jwt/                 # JWT token generator
├── docker-compose.dev.yml   # Development PostgreSQL setup
├── docker-compose.test.yml  # Test PostgreSQL setup
├── Makefile                 # Build and development commands
└── README.md               # This file
```
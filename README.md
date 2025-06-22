# Ralts-CMS

A web service built in Go for managing machines and their maintenance records. It provides a RESTful API for querying and mutating machine and maintenance data stored in AWS DynamoDB.

## Features

- Machine management (CRUD operations)
- Maintenance record management (CRUD operations)
- Basic authentication
- Local DynamoDB development setup
- Production-ready DynamoDB configuration
- Configurable HTTP timeouts
- RESTful API design

## Prerequisites

- Go 1.24 or later
- Docker and Docker Compose (for local development)
- AWS CLI (for local DynamoDB operations)

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
- Start a local DynamoDB instance
- Create the required DynamoDB table
- Start the web server on port 8080

## Manual Setup

If you prefer to set up manually:

1. **Start DynamoDB**
   ```bash
   make start-dynamodb
   ```

2. **Create the table**
   ```bash
   make create-table
   ```

3. **Run the application**
   ```bash
   make run
   ```

## Configuration

The application uses environment variables for configuration. Create a `.env.development` file:

### Development Configuration
```env
APP_NAME=ralts-cms
APP_ENV=development
DYNAMODB_ENDPOINT=http://localhost:8000
DYNAMODB_REGION=us-east-1
DYNAMODB_TABLE=ralts
AWS_ACCESS_KEY_ID=local
AWS_SECRET_ACCESS_KEY=local
SERVER_PORT=8080
HTTP_READ_TIMEOUT=15s
HTTP_WRITE_TIMEOUT=15s
HTTP_IDLE_TIMEOUT=60s
CREDENTIALS=admin:password
```

### Production Configuration
```env
APP_NAME=ralts-cms
APP_ENV=production
DYNAMODB_REGION=us-east-1
DYNAMODB_TABLE=ralts-prod
SERVER_PORT=8080
HTTP_READ_TIMEOUT=30s
HTTP_WRITE_TIMEOUT=30s
HTTP_IDLE_TIMEOUT=120s
CREDENTIALS=your-production-credentials
```

### Configuration Options

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `APP_NAME` | Application name | - | Yes |
| `APP_ENV` | Environment (development/production) | development | No |
| `DYNAMODB_ENDPOINT` | DynamoDB endpoint (local only) | http://localhost:8000 | No |
| `DYNAMODB_REGION` | AWS region for DynamoDB | us-east-1 | No |
| `DYNAMODB_TABLE` | DynamoDB table name | ralts | No |
| `AWS_ACCESS_KEY_ID` | AWS access key (local only) | local | No |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key (local only) | local | No |
| `SERVER_PORT` | HTTP server port | 8080 | No |
| `HTTP_READ_TIMEOUT` | HTTP read timeout | 15s | No |
| `HTTP_WRITE_TIMEOUT` | HTTP write timeout | 15s | No |
| `HTTP_IDLE_TIMEOUT` | HTTP idle timeout | 60s | No |
| `CREDENTIALS` | Basic auth credentials | admin:password | No |

## Production Deployment

For production deployment, the application automatically detects the environment and configures DynamoDB accordingly:

1. **Set environment to production**
   ```env
   APP_ENV=production
   ```

2. **Configure AWS credentials** (one of the following):
   - Use IAM roles (recommended for EC2/EKS)
   - Set AWS environment variables
   - Use AWS credentials file

3. **Set production DynamoDB settings**
   ```env
   DYNAMODB_REGION=your-aws-region
   DYNAMODB_TABLE=your-production-table
   ```

4. **Configure HTTP timeouts for production load**
   ```env
   HTTP_READ_TIMEOUT=30s
   HTTP_WRITE_TIMEOUT=30s
   HTTP_IDLE_TIMEOUT=120s
   ```

## API Endpoints

All endpoints require Basic Authentication. The credentials are specified in the `CREDENTIALS` environment variable.

### Authentication

Include the Authorization header with base64-encoded credentials:
```
Authorization: Basic <base64-encoded-username:password>
```

### Machine Endpoints

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
│   │   └── dynamodb.go      # DynamoDB setup and configuration
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
├── scripts/
│   ├── start-dynamodb.sh    # Start local DynamoDB
│   └── create-table.sh      # Create DynamoDB table
├── docker-compose.yml       # Local DynamoDB setup
├── Makefile                 # Build and development commands
└── README.md               # This file
```

## License

[Add your license information here] 
# Authentication Middleware

This document explains how the JWT-based authentication middleware works in the Ralts-CMS application.

## Overview

The authentication middleware validates JWT tokens from the `Authorization` header and adds user context to requests for protected endpoints. It ensures that all API requests (except public endpoints) are properly authenticated.

## How It Works

### 1. Token Generation (Login)

When a user logs in via `/users/login`, the system:

1. Validates the user's credentials
2. Generates a JWT token with the user's email as the `entity_id`
3. Returns the token in the response

```go
// Example login response
{
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "role": "user",
    "status": "active"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 2. Token Validation (Protected Endpoints)

For all protected endpoints, the middleware:

1. Extracts the `Authorization` header
2. Validates the JWT token format (`Bearer <token>`)
3. Verifies the token signature using the JWT secret
4. Extracts user information from the token claims
5. Adds the user context to the request

### 3. User Context Access

Handlers can access the authenticated user information:

```go
func (h *MachinesHandler) CreateMachine(w http.ResponseWriter, r *http.Request) {
    // Get authenticated user from context
    user, err := middlewares.GetUserFromContext(r.Context())
    if err != nil {
        http.Error(w, "Authentication required", http.StatusUnauthorized)
        return
    }

    // Use user information
    fmt.Printf("Machine created by user: %s (email: %s)\n", user.EntityID, user.Email)
    
    // ... rest of handler logic
}
```

## Protected vs Public Endpoints

### Public Endpoints (No Authentication Required)
- `GET /health` - Health check
- `POST /users` - User registration
- `POST /users/login` - User login

### Protected Endpoints (Authentication Required)
- `GET /machines` - List machines
- `GET /machines/{serial_number}` - Get specific machine
- `POST /machines` - Create machine
- `PUT /machines` - Update machine
- `DELETE /machines/{serial_number}` - Delete machine
- `GET /machines/due-ppm` - Get machines due for PPM
- All maintenance endpoints under `/machines/{serial_number}/maintenance`

## JWT Token Structure

The JWT tokens contain the following claims:

```json
{
  "entity_id": "user@example.com",
  "exp": 1640995200,
  "iat": 1640908800
}
```

- `entity_id`: The user's email address
- `exp`: Token expiration time (24 hours from creation)
- `iat`: Token creation time

## Configuration

The authentication middleware requires the `JWT_SECRET` environment variable:

```env
JWT_SECRET=your-jwt-secret-key
```

## Usage Examples

### Making Authenticated Requests

```bash
# Login to get a token
curl -X POST http://localhost:8080/users/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'

# Use the token for authenticated requests
curl -H "Authorization: Bearer <your-jwt-token>" \
  http://localhost:8080/machines

# Create a machine with authentication
curl -X POST http://localhost:8080/machines \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{"serial_number": "MACHINE123", "customer": "ACME Corp"}'
```

### Error Responses

The middleware returns appropriate HTTP status codes:

- `401 Unauthorized` - Missing or invalid Authorization header
- `401 Unauthorized` - Invalid JWT token format
- `401 Unauthorized` - Invalid JWT token signature
- `401 Unauthorized` - JWT token expired
- `401 Unauthorized` - Missing entity_id in token claims

## Security Best Practices

1. **Strong JWT Secret**: Use a strong, randomly generated secret key (at least 32 characters)
2. **HTTPS in Production**: Always use HTTPS in production environments
3. **Token Expiration**: Tokens expire after 24 hours for security
4. **Secret Management**: Never commit JWT secrets to version control
5. **Token Storage**: Store tokens securely on the client side
6. **Regular Rotation**: Consider rotating JWT secrets periodically

## Testing

### Running Authentication Tests

```bash
# Test the auth package
go test ./pkg/auth -v

# Test the middleware
go test ./internal/middlewares -v
```

### Manual Testing

1. Start the application:
   ```bash
   make dev
   ```

2. Create a user:
   ```bash
   curl -X POST http://localhost:8080/users \
     -H "Content-Type: application/json" \
     -d '{"name": "Test User", "email": "test@example.com", "password": "password123"}'
   ```

3. Login to get a token:
   ```bash
   curl -X POST http://localhost:8080/users/login \
     -H "Content-Type: application/json" \
     -d '{"email": "test@example.com", "password": "password123"}'
   ```

4. Use the token for authenticated requests:
   ```bash
   curl -H "Authorization: Bearer <token-from-login>" \
     http://localhost:8080/machines
   ```

## Implementation Details

### Middleware Chain

The authentication middleware is applied in the router:

```go
// Apply middleware to protected endpoints only
api.Use(middlewares.LoggingMiddleware)
api.Use(middlewares.CORSMiddleware)
api.Use(middlewares.AuthenticationMiddleware(deps.Config.JWTSecret))
```

### User Context

The middleware adds user information to the request context:

```go
type UserContext struct {
    EntityID string `json:"entity_id"`
    Email    string `json:"email"`
}
```

### Helper Functions

- `GetUserFromContext(ctx context.Context) (*UserContext, error)` - Extract user from context
- `RequireAuth(w http.ResponseWriter, r *http.Request) (*UserContext, error)` - Ensure authentication

## Troubleshooting

### Common Issues

1. **"Authorization header is required"**
   - Ensure you're including the `Authorization: Bearer <token>` header

2. **"Invalid token"**
   - Check that the token is valid and not expired
   - Verify the JWT secret is correct

3. **"Invalid authorization header format"**
   - Ensure the header starts with "Bearer " (note the space)

4. **Token expiration**
   - Re-login to get a new token

### Debugging

Enable debug logging to see authentication details:

```go
// In your handler
user, err := middlewares.GetUserFromContext(r.Context())
if err != nil {
    log.Printf("Authentication error: %v", err)
    http.Error(w, "Authentication required", http.StatusUnauthorized)
    return
}
log.Printf("Authenticated user: %+v", user)
``` 
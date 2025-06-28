# JWT Token Generator

This directory contains a simple JWT token generator for local testing of the Ralts-CMS API.

## Usage

Generate a new JWT token:

```bash
go run scripts/jwt/main.go
```

## What it generates

The script generates a JWT token with the following characteristics:

- **Secret**: Uses the default JWT secret from the application config (`your-jwt-secret-key`)
- **Subject**: `test-user`
- **Expiration**: 24 hours from generation
- **Algorithm**: HS256

## Output

The script provides:

1. The JWT secret being used
2. The complete Authorization header
3. The full JWT token
4. Example curl commands for testing
5. Token details (subject, issued at, expires at)

## Example Output

```
=== Valid JWT Token for Local Testing ===

JWT Secret: your-jwt-secret-key

Authorization Header: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

Full Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

=== Usage Examples ===

curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
     http://localhost:8080/api/v1/machines

=== Token Details ===
Subject (sub): test-user
Issued At (iat): 2024-01-01T12:00:00Z
Expires At (exp): 2024-01-02T12:00:00Z
Valid Until: 2024-01-02T12:00:00Z
```

## Testing with curl

Copy the generated Authorization header and use it in your API requests:

```bash
# List machines
curl -H "Authorization: Bearer YOUR_TOKEN_HERE" \
     http://localhost:8080/api/v1/machines

# Create a machine
curl -H "Authorization: Bearer YOUR_TOKEN_HERE" \
     -H "Content-Type: application/json" \
     -d '{"serial_number":"TEST123","customer":"Test Customer"}' \
     http://localhost:8080/api/v1/machines
```

## Notes

- The token is valid for 24 hours
- Uses the same JWT secret as your local development environment
- Compatible with all protected API endpoints
- For production, use a different JWT secret and shorter expiration times 
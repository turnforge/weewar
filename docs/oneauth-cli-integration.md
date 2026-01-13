# OneAuth CLI Integration

This document describes the CLI authentication integration with oneauth.

## Overview

The CLI uses oneauth's `APIAuth` handler for token-based authentication. This provides:

1. OAuth2-compatible password grant for email/password login
2. JWT access tokens with configurable expiration
3. Refresh tokens for session management
4. Bearer token validation in API requests

## Server Configuration

In `web/server/auth.go`, the APIAuth handler is configured:

```go
import (
    oa "github.com/panyam/oneauth"
    oafs "github.com/panyam/oneauth/stores/fs"
)

// In setupAuthService():
jwtSecret := os.Getenv("JWT_CLI_SECRET")
if jwtSecret == "" {
    jwtSecret = "lilbattle-dev-secret-change-in-production"
}

refreshTokenStore := oafs.NewFSRefreshTokenStore(storagePath)
apiAuth := &oa.APIAuth{
    RefreshTokenStore:   refreshTokenStore,
    JWTSecretKey:        jwtSecret,
    JWTIssuer:           "lilbattle",
    JWTAudience:         "cli",
    AccessTokenExpiry:   30 * 24 * time.Hour, // 30 days
    RefreshTokenExpiry:  90 * 24 * time.Hour, // 90 days
    ValidateCredentials: authService.ValidateLocalCredentials,
}
oneauth.AddAuth("/cli/token", apiAuth)
```

## Token Endpoint

**Endpoint**: `POST /auth/cli/token`

**Request** (OAuth2 password grant):
```json
{
  "grant_type": "password",
  "username": "user@example.com",
  "password": "secret",
  "scope": "read write profile offline",
  "client_id": "cli"
}
```

**Success Response** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 2592000,
  "refresh_token": "abc123...",
  "scope": "read write profile offline"
}
```

**Error Response**:
```json
{
  "error": "invalid_grant",
  "error_description": "Invalid credentials"
}
```

## CLI Usage

```bash
# Interactive login (prompts for email/password)
ww login http://localhost:8080

# Token-based login (for pre-generated tokens)
ww login http://localhost:8080 --token eyJhbGc...

# Check authentication status
ww whoami

# Logout from a server
ww logout http://localhost:8080

# Migrate worlds between servers
ww migrate http://localhost:6060/api/v1/worlds/Desert \
           http://localhost:8080/api/v1/worlds/Desert
```

## Credential Storage

Credentials are stored in `~/.config/lilbattle/credentials.json` with 0600 permissions:

```json
{
  "servers": {
    "http://localhost:8080": {
      "token": "eyJhbGc...",
      "user_id": "",
      "user_email": "user@example.com",
      "expires_at": "2025-03-01T00:00:00Z",
      "created_at": "2025-01-15T00:00:00Z"
    }
  }
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `JWT_CLI_SECRET` | Secret key for signing JWT tokens | Dev fallback (change in production!) |

## Security Considerations

1. **Production Secret**: Always set `JWT_CLI_SECRET` in production
2. **Token Storage**: Credentials file has restricted permissions (0600)
3. **Token Expiration**: Access tokens expire after 30 days by default
4. **Refresh Tokens**: Available for extending sessions without re-authentication

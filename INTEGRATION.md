# Auth Microservice Integration Guide

Quick guide for integrating with the authentication microservice from other services.

## Authentication Flow

### 1. User Login (Get Tokens)

Your frontend/client first obtains tokens by logging in:

```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 86400
}
```

### 2. Verify Token (From Your Microservice)

When your microservice receives a request with a token, verify it:

```bash
POST /api/v1/auth/verify
Content-Type: application/json

{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response (200 OK):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "role": "customer",
  "active": true,
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

**Response (401 Unauthorized):**
```json
{
  "error": "Invalid or expired token"
}
```

## Quick Integration Example (Go)

```go
func verifyUserToken(token string) (*UserInfo, error) {
    reqBody, _ := json.Marshal(map[string]string{
        "token": token,
    })
    
    resp, err := http.Post(
        "http://auth-service:8080/api/v1/auth/verify",
        "application/json",
        bytes.NewBuffer(reqBody),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, errors.New("invalid token")
    }
    
    var user UserInfo
    json.NewDecoder(resp.Body).Decode(&user)
    return &user, nil
}
```

## User Roles

- `customer` - Regular users
- `employee` - Staff members
- `admin` - Administrators

Use the `role` field from the verify response to implement role-based access control in your service.

## Best Practices

1. **Cache verification results** - Cache user info for a few minutes to reduce load on auth service
2. **Handle token expiration** - Return 401 when token verification fails so clients can refresh
3. **Validate role** - Check the user's role for authorization in your service
4. **Check active status** - Ensure `active: true` before allowing operations
5. **Extract from headers** - Tokens typically come in `Authorization: Bearer <token>` header

## Token Refresh

When access tokens expire, clients should refresh them:

```bash
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

Returns new access and refresh tokens with the same structure as login.

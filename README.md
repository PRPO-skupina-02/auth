# Auth Microservice

Authentication and authorization service for the PRPO project.

## Features

- **User Registration** - Customers can self-register
- **User Authentication** - JWT-based login with access and refresh tokens
- **Role-Based Access Control (RBAC)** - Three roles: customer, employee, admin
- **Token Verification** - Verify JWT tokens for inter-service communication
- **User Management** - Admin-only endpoints for managing all users
- **Profile Management** - Users can view and update their own profile

## User Roles

- **Customer** - Can self-register and manage their own profile only
- **Employee** - Cannot self-register (created by admin), has limited access
- **Admin** - Full access to all user management operations

## API Endpoints

### Public Endpoints

- `POST /api/v1/auth/register` - Register a new customer account
- `POST /api/v1/auth/login` - Login and receive JWT tokens
- `POST /api/v1/auth/refresh` - Refresh access token using refresh token
- `POST /api/v1/auth/verify` - Verify a JWT token and get user info

### Protected Endpoints (Require Authentication)

- `GET /api/v1/auth/me` - Get current user information
- `PUT /api/v1/auth/me` - Update current user profile
- `PUT /api/v1/auth/me/password` - Change password

### Admin Endpoints (Admin Only)

- `GET /api/v1/auth/users` - List all users
- `GET /api/v1/auth/users/:userID` - Get user by ID
- `POST /api/v1/auth/users` - Create user with any role (including employees)
- `PUT /api/v1/auth/users/:userID` - Update user
- `DELETE /api/v1/auth/users/:userID` - Delete user

## Environment Variables

- `JWT_SECRET` - Secret key for signing JWT tokens (default: "dev-secret-key-change-in-production")
- `DATABASE_URL` - PostgreSQL connection string

## Development

### Prerequisites

- Go 1.25+
- PostgreSQL
- Make

### Setup

1. Install CLI tools:
```bash
make install-cli-tools
```

2. Run migrations:
```bash
make migrate
```

3. Load fixtures (optional):
```bash
make fixtures
```

4. Generate Swagger docs:
```bash
make docs
```

### Running the Service

```bash
go run main.go
```

The service will start on `http://localhost:8080`

### Swagger Documentation

Access the API documentation at `http://localhost:8080/swagger/index.html`

## Testing

Run tests:
```bash
make test
```

## Authentication Flow

### For Customer Registration
1. Customer calls `POST /register` with email and password
2. System creates user with role "customer"
3. Customer can login with credentials

### For Employee Creation
1. Admin logs in and gets JWT token
2. Admin calls `POST /users` with employee details and role "employee"
3. Employee can login with provided credentials

### Token Usage
1. Login returns both access_token (24h) and refresh_token (7 days)
2. Include access token in Authorization header: `Bearer <token>`
3. When access token expires, use refresh token to get new tokens

### Inter-Service Authentication
Other microservices can verify user tokens by calling:
```bash
POST /api/v1/auth/verify
{
  "token": "<jwt-token>"
}
```

This returns user information including role for authorization decisions.

## Example Requests

### Register Customer
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "customer@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "customer@example.com",
    "password": "password123"
  }'
```

### Create Employee (Admin Only)
```bash
curl -X POST http://localhost:8080/api/v1/auth/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin-token>" \
  -d '{
    "email": "employee@example.com",
    "password": "password123",
    "first_name": "Jane",
    "last_name": "Smith",
    "role": "employee",
    "active": true
  }'
```

### Verify Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/verify \
  -H "Content-Type: application/json" \
  -d '{
    "token": "<jwt-token>"
  }'
```

## Default Test Users (from fixtures)

- **Admin**: admin@example.com / admin123
- **Employee**: employee@example.com / employee123
- **Customer**: customer@example.com / customer123

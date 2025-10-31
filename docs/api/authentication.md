# Authentication API Documentation

This document describes the authentication endpoints available in the Portfolios API.

## Base URL

All endpoints are prefixed with `/api/auth`

For local development: `http://localhost:8080/api/auth`

## Authentication

Most endpoints require authentication via JWT Bearer token in the Authorization header:

```
Authorization: Bearer <access_token>
```

## Rate Limiting

Authentication endpoints are rate-limited to **5 requests per minute per IP address** to prevent brute force attacks.

---

## Endpoints

### 1. Register User

Create a new user account.

**Endpoint:** `POST /api/auth/register`

**Authentication:** Not required

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123"
}
```

**Request Fields:**
- `email` (string, required): Valid email address, must be unique
- `password` (string, required): Minimum 8 characters, must contain at least one uppercase letter, one lowercase letter, and one number

**Success Response (201 Created):**
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "created_at": "2025-10-31T12:00:00Z",
    "last_login_at": null
  },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 1800
}
```

**Response Fields:**
- `user.id` (UUID): Unique user identifier
- `user.email` (string): User's email address
- `user.created_at` (timestamp): Account creation timestamp
- `user.last_login_at` (timestamp, nullable): Last login timestamp
- `access_token` (string): JWT access token (30-60 minute lifespan)
- `refresh_token` (string): JWT refresh token (7-30 day lifespan)
- `expires_in` (integer): Access token expiration time in seconds

**Error Responses:**

**400 Bad Request** - Invalid input data
```json
{
  "error": "Invalid request data: password does not meet requirements",
  "code": "INVALID_PASSWORD"
}
```

**409 Conflict** - Email already registered
```json
{
  "error": "A user with this email already exists",
  "code": "EMAIL_ALREADY_EXISTS"
}
```

**Example cURL:**
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "SecurePass123"
  }'
```

---

### 2. Login

Authenticate a user and receive access and refresh tokens.

**Endpoint:** `POST /api/auth/login`

**Authentication:** Not required

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123",
  "remember_me": false
}
```

**Request Fields:**
- `email` (string, required): User's email address
- `password` (string, required): User's password
- `remember_me` (boolean, optional, default: false): If true, extends access token to 24 hours and refresh token to 30 days

**Success Response (200 OK):**
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "created_at": "2025-10-31T12:00:00Z",
    "last_login_at": "2025-10-31T14:30:00Z"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 1800
}
```

**Error Responses:**

**401 Unauthorized** - Invalid credentials
```json
{
  "error": "Invalid email or password",
  "code": "INVALID_CREDENTIALS"
}
```

**Example cURL:**
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123",
    "remember_me": true
  }'
```

---

### 3. Refresh Access Token

Exchange a refresh token for a new access token without requiring the user to log in again.

**Endpoint:** `POST /api/auth/refresh`

**Authentication:** Not required (uses refresh token)

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Request Fields:**
- `refresh_token` (string, required): Valid refresh token received from login or registration

**Success Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 1800
}
```

**Response Fields:**
- `access_token` (string): New JWT access token
- `expires_in` (integer): Access token expiration time in seconds

**Error Responses:**

**401 Unauthorized** - Invalid or expired refresh token
```json
{
  "error": "Invalid or expired refresh token",
  "code": "INVALID_REFRESH_TOKEN"
}
```

**Example cURL:**
```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

---

### 4. Logout

Revoke the user's refresh token and invalidate their session.

**Endpoint:** `POST /api/auth/logout`

**Authentication:** Required (Bearer token)

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Request Fields:**
- `refresh_token` (string, required): Refresh token to revoke

**Success Response (200 OK):**
```json
{
  "message": "Logged out successfully"
}
```

**Error Responses:**

**401 Unauthorized** - Not authenticated or invalid access token
```json
{
  "error": "User not authenticated",
  "code": "NOT_AUTHENTICATED"
}
```

**Example cURL:**
```bash
curl -X POST http://localhost:8080/api/auth/logout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

---

### 5. Get Current User

Retrieve the authenticated user's profile information.

**Endpoint:** `GET /api/auth/me`

**Authentication:** Required (Bearer token)

**Headers:**
```
Authorization: Bearer <access_token>
```

**Success Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "created_at": "2025-10-31T12:00:00Z",
  "last_login_at": "2025-10-31T14:30:00Z"
}
```

**Response Fields:**
- `id` (UUID): User's unique identifier
- `email` (string): User's email address
- `created_at` (timestamp): Account creation timestamp
- `last_login_at` (timestamp, nullable): Last successful login timestamp

**Error Responses:**

**401 Unauthorized** - Not authenticated or invalid token
```json
{
  "error": "User not authenticated",
  "code": "NOT_AUTHENTICATED"
}
```

**404 Not Found** - User not found
```json
{
  "error": "User not found",
  "code": "USER_NOT_FOUND"
}
```

**Example cURL:**
```bash
curl -X GET http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

### 6. Forgot Password

Initiate the password reset process by sending a reset link to the user's email.

**Endpoint:** `POST /api/auth/forgot-password`

**Authentication:** Not required

**Request Body:**
```json
{
  "email": "user@example.com"
}
```

**Request Fields:**
- `email` (string, required): User's email address

**Success Response (200 OK):**
```json
{
  "message": "If an account exists with this email, a password reset link has been sent"
}
```

**Note:** The endpoint always returns success to prevent email enumeration attacks. An email is only sent if the account exists.

**Error Responses:**

**400 Bad Request** - Invalid email format
```json
{
  "error": "Invalid request data: email is invalid",
  "code": "INVALID_REQUEST"
}
```

**Example cURL:**
```bash
curl -X POST http://localhost:8080/api/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com"
  }'
```

**Email Content:**

Users will receive an email with a password reset link:
```
https://app.example.com/reset-password?token=<reset_token>
```

The reset token is valid for **1 hour** and can only be used **once**.

---

### 7. Reset Password

Complete the password reset process using the token received via email.

**Endpoint:** `POST /api/auth/reset-password`

**Authentication:** Not required (uses reset token)

**Request Body:**
```json
{
  "token": "abc123def456...",
  "new_password": "NewSecurePass123"
}
```

**Request Fields:**
- `token` (string, required): Password reset token from email
- `new_password` (string, required): New password (minimum 8 characters, must contain at least one uppercase letter, one lowercase letter, and one number)

**Success Response (200 OK):**
```json
{
  "message": "Password reset successfully"
}
```

**Error Responses:**

**400 Bad Request** - Invalid or expired token
```json
{
  "error": "Invalid or expired reset token",
  "code": "INVALID_RESET_TOKEN"
}
```

**400 Bad Request** - Invalid password format
```json
{
  "error": "invalid password: must be at least 8 characters",
  "code": "INVALID_PASSWORD"
}
```

**Example cURL:**
```bash
curl -X POST http://localhost:8080/api/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "token": "abc123def456...",
    "new_password": "NewSecurePass123"
  }'
```

---

## Status Codes

| Status Code | Description |
|-------------|-------------|
| 200 OK | Request successful |
| 201 Created | Resource created successfully (registration) |
| 400 Bad Request | Invalid request data or validation error |
| 401 Unauthorized | Authentication required or invalid credentials |
| 403 Forbidden | User does not have permission to access resource |
| 404 Not Found | Resource not found |
| 409 Conflict | Resource already exists (duplicate email) |
| 429 Too Many Requests | Rate limit exceeded |
| 500 Internal Server Error | Server error |

---

## Error Response Format

All error responses follow this format:

```json
{
  "error": "Human-readable error message",
  "code": "MACHINE_READABLE_ERROR_CODE"
}
```

**Common Error Codes:**
- `INVALID_REQUEST`: Malformed request or validation failure
- `INVALID_PASSWORD`: Password does not meet requirements
- `EMAIL_ALREADY_EXISTS`: Email is already registered
- `INVALID_CREDENTIALS`: Invalid email or password
- `INVALID_REFRESH_TOKEN`: Refresh token is invalid or expired
- `INVALID_RESET_TOKEN`: Password reset token is invalid or expired
- `NOT_AUTHENTICATED`: User is not authenticated
- `USER_NOT_FOUND`: User does not exist
- `LOGOUT_FAILED`: Failed to logout
- `RESET_PASSWORD_FAILED`: Failed to reset password
- `REGISTRATION_FAILED`: Failed to register user

---

## Security Considerations

1. **Password Requirements**
   - Minimum 8 characters
   - At least one uppercase letter
   - At least one lowercase letter
   - At least one number
   - Passwords are hashed with bcrypt (cost factor 12) before storage

2. **JWT Tokens**
   - Access tokens: 30-60 minutes (or 24 hours with remember_me)
   - Refresh tokens: 7 days (or 30 days with remember_me)
   - Tokens are signed with HS256 algorithm
   - Access tokens contain minimal claims: user_id, exp, iat

3. **Rate Limiting**
   - 5 requests per minute per IP on all auth endpoints
   - Helps prevent brute force attacks

4. **Email Enumeration Prevention**
   - Forgot password always returns success
   - Does not reveal whether email exists

5. **HTTPS**
   - Always use HTTPS in production
   - Never send tokens over unencrypted connections

6. **CORS**
   - Configured to allow only trusted frontend origins
   - Set via CORS_ALLOWED_ORIGINS environment variable

---

## Token Management

### Access Token

- Short-lived (30-60 minutes by default, 24 hours with remember_me)
- Used for API authentication
- Sent in Authorization header: `Bearer <access_token>`
- Contains user_id claim for identifying the user

### Refresh Token

- Long-lived (7 days by default, 30 days with remember_me)
- Used to obtain new access tokens without re-authentication
- Stored hashed in database
- Can be revoked on logout
- Should be stored securely on client (httpOnly cookie recommended)

### Token Refresh Flow

1. Client detects access token is about to expire (check `expires_in`)
2. Client calls `/api/auth/refresh` with refresh token
3. Server validates refresh token and issues new access token
4. Client continues making authenticated requests with new access token

---

## Example Usage Flow

### Registration and Login Flow

```javascript
// 1. Register new user
const registerResponse = await fetch('http://localhost:8080/api/auth/register', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'SecurePass123'
  })
});

const { user, access_token, refresh_token, expires_in } = await registerResponse.json();

// Store tokens securely (localStorage or httpOnly cookie)
localStorage.setItem('access_token', access_token);
localStorage.setItem('refresh_token', refresh_token);

// 2. Make authenticated request
const profileResponse = await fetch('http://localhost:8080/api/auth/me', {
  headers: {
    'Authorization': `Bearer ${access_token}`
  }
});

const profile = await profileResponse.json();

// 3. Refresh token when needed
const refreshResponse = await fetch('http://localhost:8080/api/auth/refresh', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    refresh_token: refresh_token
  })
});

const { access_token: newAccessToken } = await refreshResponse.json();
localStorage.setItem('access_token', newAccessToken);

// 4. Logout
await fetch('http://localhost:8080/api/auth/logout', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${access_token}`
  },
  body: JSON.stringify({
    refresh_token: refresh_token
  })
});

// Clear tokens
localStorage.removeItem('access_token');
localStorage.removeItem('refresh_token');
```

### Password Reset Flow

```javascript
// 1. User requests password reset
await fetch('http://localhost:8080/api/auth/forgot-password', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'user@example.com'
  })
});

// 2. User receives email with reset link
// Link format: https://app.example.com/reset-password?token=abc123...

// 3. User submits new password with token
await fetch('http://localhost:8080/api/auth/reset-password', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    token: 'abc123...',
    new_password: 'NewSecurePass123'
  })
});

// 4. User can now login with new password
```

---

## Swagger/OpenAPI Documentation

Interactive API documentation is available via Swagger UI when running the server:

**URL:** http://localhost:8080/api/docs

The Swagger UI provides:
- Interactive API testing
- Detailed request/response schemas
- Authentication flows
- Example requests and responses

---

## Support

For issues or questions about the authentication API, please:
1. Check the error response code and message
2. Review this documentation
3. Check server logs for detailed error information
4. Contact the development team

---

**Last Updated:** 2025-10-31
**API Version:** 1.0.0

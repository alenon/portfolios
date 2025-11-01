# Specification: User Authentication & Authorization

## Goal
Implement secure user registration, login, JWT-based authentication with access and refresh tokens, and password reset functionality to provide the foundational authentication system for the portfolios application, ensuring each user has isolated access to only their own portfolio data.

## User Stories
- As a new user, I want to register with my email and password so that I can create an account and start managing my portfolios
- As a registered user, I want to log in with my credentials and optionally select "Remember Me" so that I can access my portfolios with appropriate session duration
- As a user who forgot my password, I want to receive a password reset email with a secure link so that I can regain access to my account

## Specific Requirements

**User Registration Flow**
- Users register with email (as username) and password only
- Email must be unique across all users, validated for proper format
- Password must meet validation: minimum 8 characters, at least one uppercase, one lowercase, one number
- Passwords are hashed using bcrypt (cost factor 10-12) before database storage
- No email verification required - users can log in immediately after registration
- Successful registration automatically logs user in, returning access token, refresh token, and user data
- Duplicate email returns 409 Conflict error with clear message
- User record includes: id (UUID), email, password_hash, created_at, updated_at, last_login_at

**User Login & Session Management**
- Users log in with email and password
- "Remember Me" checkbox controls session duration: 24 hours (unchecked) or 7 days (checked)
- Dual-token JWT strategy: access token (15-60 minutes lifespan) and refresh token (7-30 days lifespan)
- Access token contains minimal claims: user_id, exp (expiration), iat (issued at)
- Refresh token is hashed before storage in database, can be revoked on logout
- JWT signed with HS256 algorithm using secure secret from environment variable
- Failed login returns 401 Unauthorized with generic error message (don't reveal if email exists)
- Last login timestamp updated on successful authentication
- Token refresh endpoint allows obtaining new access token without re-authentication

**Password Reset Flow**
- "Forgot Password" link on login page leads to email entry form
- System generates cryptographically secure reset token, valid for 1 hour
- Reset token is hashed before database storage
- Password reset email sent via SMTP with configurable settings (host, port, username, password from environment)
- Email contains link: `https://app.example.com/reset-password?token=xxx`
- Always return success message to prevent email enumeration attacks
- Reset password page validates token, allows entry of new password meeting requirements
- Tokens are single-use (marked as used_at after successful reset)
- Expired or invalid tokens return 400 Bad Request with appropriate error
- After successful reset, user must log in with new password

**API Authentication & Authorization Middleware**
- Authentication middleware extracts JWT from Authorization header (Bearer token format)
- Validates token signature, checks expiration, extracts user_id from claims
- Attaches user_id to request context for downstream handlers
- Returns 401 Unauthorized if token missing, invalid, or expired
- Authorization middleware checks resource ownership before allowing access
- User can only access resources where user_id matches authenticated user
- Returns 403 Forbidden if user attempts to access another user's resources
- Protected endpoints require authentication middleware in routing configuration

**Security Implementation**
- All passwords hashed with bcrypt, never stored in plain text
- JWT secret stored in environment variable, never committed to version control
- CORS configured to allow only frontend origin in production
- Rate limiting on authentication endpoints: 5 requests per minute per IP for login/register
- Input validation on all request bodies using go-playground/validator
- SQL injection prevention via parameterized queries (sqlc or GORM)
- Refresh tokens and password reset tokens hashed before database storage
- HTTPS enforced in production (configured at ingress/load balancer level)

**Database Schema Design**
- Users table with UUID primary key, unique email index, timestamp columns
- Refresh tokens table with foreign key to users, indexes on user_id and token_hash
- Password reset tokens table with foreign key to users, indexes on token_hash and user_id
- ON DELETE CASCADE for refresh tokens and reset tokens when user deleted
- All timestamp fields use TIMESTAMP type with DEFAULT CURRENT_TIMESTAMP where appropriate
- Database migrations managed via golang-migrate/migrate for version control

**API Endpoint Specification**
- POST /api/auth/register: Create new user account, returns user + tokens
- POST /api/auth/login: Authenticate user, returns user + tokens based on remember_me flag
- POST /api/auth/refresh: Exchange refresh token for new access token
- POST /api/auth/logout: Revoke refresh token, requires authenticated user
- GET /api/auth/me: Return current user profile, requires authentication
- POST /api/auth/forgot-password: Initiate password reset, accepts email
- POST /api/auth/reset-password: Complete password reset with token and new password
- All endpoints accept/return JSON, follow RESTful conventions

**Frontend Page Implementations**
- Registration page (/register): Email, password inputs with validation, password visibility toggle, requirements display, register button, link to login
- Login page (/login): Email, password inputs, remember me checkbox, forgot password link, login button, link to register
- Forgot password page (/forgot-password): Email input, submit button, back to login link, success message display
- Reset password page (/reset-password?token=xxx): New password, confirm password inputs with visibility toggles, requirements display, validation for password match
- Authenticated layout: Header with app name/logo, user email display, logout button
- Protected route wrapper component that checks authentication status and redirects to login if not authenticated

**Error Handling & User Feedback**
- Form validation errors displayed inline with specific field feedback
- API errors displayed as user-friendly messages (not raw error codes)
- Loading states shown during all async operations (registration, login, password reset)
- Success messages for operations like password reset email sent, password changed
- Token expiration handled gracefully with automatic refresh or redirect to login
- 401 responses trigger automatic redirect to login page with return URL
- Clear validation feedback for password requirements in real-time

## Visual Design
No visual assets provided - implementation will follow industry best practices for authentication UI/UX patterns with clean, accessible forms.

## Existing Code to Leverage
This is a greenfield project with no existing codebase. This authentication system will serve as the foundation for all subsequent features.

**Future features will leverage:**
- Authentication middleware for protecting portfolio and transaction endpoints
- User ID from JWT claims for data isolation and ownership verification
- Authorization patterns established here for resource-level access control
- JWT token management strategy for session handling across the application
- Database schema patterns (UUID primary keys, timestamps, indexes) for consistency

## Out of Scope
- Email verification during registration (users can login immediately)
- Social login integration (Google, GitHub, Facebook, Apple)
- Two-factor authentication (2FA) or multi-factor authentication (MFA)
- Account deletion or deactivation functionality
- User profile editing (name, avatar, bio, preferences)
- Admin roles or role-based access control (RBAC) beyond basic user isolation
- Password strength meter UI component
- Login attempt rate limiting per user account (only IP-based rate limiting)
- Account lockout after multiple failed login attempts
- Password history tracking to prevent password reuse
- OAuth 2.0 provider functionality (allowing other apps to authenticate via this system)
- API key authentication for programmatic access
- Session management dashboard showing active sessions across devices
- "Sign in with email link" passwordless authentication
- Biometric authentication support

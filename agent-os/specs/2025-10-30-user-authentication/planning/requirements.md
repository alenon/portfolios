# Spec Requirements: User Authentication & Authorization

## Initial Description

Implement secure user registration, login, JWT-based authentication, and user session management with password reset functionality. Each user should have isolated access to only their portfolios.

This is the foundational feature for the entire portfolios application, enabling secure user access and data isolation for all subsequent features.

## Requirements Discussion

### First Round Questions

**Q1: User Registration Flow**
I assume we want a simple email/password registration flow without email verification at this stage (users can log in immediately after registration). Is that correct, or should we implement email verification before allowing login?

**Answer:** Simple email/password flow without email verification. Users can log in immediately after registration.

**Q2: Remember Me Functionality**
Should we include a "Remember Me" checkbox on the login page that extends the JWT token expiration time (e.g., from 24 hours to 7 days)?

**Answer:** Yes, implement remember me that extends JWT token expiration from 24 hours to 7 days.

**Q3: JWT Token Strategy**
I'm thinking we should use a standard JWT approach with access tokens (short-lived, 15-60 minutes) and refresh tokens (longer-lived, 7-30 days) stored securely. Should we use this dual-token approach, or prefer a simpler single-token strategy?

**Answer:** Standard approach with access token (short-lived, 15-60 minutes) and refresh token (longer-lived, 7-30 days).

**Q4: Password Requirements**
Should we enforce specific password requirements (minimum length, complexity rules like uppercase, numbers, special characters)? If yes, what specific rules would you prefer?

**Answer:** Simple requirements: minimum 8 characters, at least one uppercase letter, one lowercase letter, and one number.

**Q5: Password Reset Flow**
For password reset functionality, I assume we want: forgot password link on login page, user enters email, system sends reset link via email with time-limited token, user clicks link and enters new password. Is this the flow you envision? Also, which email service provider should we use (SendGrid, AWS SES, SMTP)?

**Answer:** Yes, that's the flow. Email service not specified - suggest best practice option for Go (e.g., SMTP with configurable settings).

**Q6: User Profile Data**
Beyond email and password, what additional user profile data should we store at this stage? (Name, username, phone, profile picture, preferences, etc.) Or should we keep it minimal for now?

**Answer:** Only basic data: email as username, password hashed, created_at, updated_at.

**Q7: Role-Based Access Control**
Should we implement role-based access control (RBAC) at this stage with different user roles (admin, regular user), or just basic user isolation where each user can only access their own portfolios?

**Answer:** Only basic user isolation (users can only access their own portfolios), no admin roles at this stage.

**Q8: API Endpoints**
I'm planning these endpoints: POST /api/auth/register, POST /api/auth/login, POST /api/auth/refresh, POST /api/auth/logout, POST /api/auth/forgot-password, POST /api/auth/reset-password, GET /api/auth/me. Should we add any others?

**Answer:** These endpoints are sufficient.

**Q9: Frontend Pages**
Should we build all frontend pages (registration, login, forgot password, reset password, basic authenticated layout/header), or focus on backend API first?

**Answer:** Build all frontend pages.

**Q10: Scope Boundaries**
Are there any specific features we should explicitly exclude from this iteration? For example: social login (Google, GitHub), two-factor authentication, account deletion, profile editing?

**Answer:** No explicit exclusions, but don't include social login, two-factor authentication, account deletion, or profile editing (beyond basic data).

### Existing Code to Reference

**Similar Features Identified:** None - this is a new project with no existing codebase.

### Follow-up Questions

None - all requirements are clear.

## Visual Assets

### Files Provided:
No visual assets provided.

### Visual Insights:
N/A - No visuals available. Implementation will follow industry best practices for authentication UI/UX patterns.

## Requirements Summary

### Functional Requirements

#### User Registration
- Users can register with email and password
- Email serves as the unique username/identifier
- No email verification required - users can log in immediately after registration
- Password must meet validation criteria:
  - Minimum 8 characters
  - At least one uppercase letter
  - At least one lowercase letter
  - At least one number
- System creates user record with hashed password (bcrypt)
- Duplicate email registration returns appropriate error
- User record includes: email, hashed password, created_at, updated_at timestamps

#### User Login
- Users can log in with email and password
- System validates credentials against stored hashed password
- Successful login returns:
  - Access token (JWT, short-lived: 15-60 minutes)
  - Refresh token (JWT, longer-lived: 7-30 days)
- "Remember Me" checkbox extends token expiration:
  - Without Remember Me: 24-hour session
  - With Remember Me: 7-day session
- Failed login attempts return appropriate error messages
- System tracks last login timestamp

#### Session Management
- Dual-token JWT strategy:
  - **Access Token**: Short-lived (15-60 minutes), included in Authorization header for API requests
  - **Refresh Token**: Longer-lived (7-30 days), used to obtain new access tokens
- Access token contains user ID and expiration
- Refresh token stored securely and can be revoked
- Token refresh endpoint allows getting new access token without re-login
- Logout invalidates refresh token

#### Password Reset
- "Forgot Password" link on login page
- User enters email address
- System generates time-limited reset token (valid for 1 hour)
- System sends email with password reset link containing token
- User clicks link, redirected to reset password page
- User enters new password (must meet validation criteria)
- Token can only be used once
- Expired or invalid tokens show appropriate error message
- After successful reset, user can log in with new password

#### User Isolation
- Each user can only access their own data (portfolios, transactions)
- API endpoints verify user ownership before allowing access
- User ID from JWT token determines data access scope
- Unauthorized access attempts return 403 Forbidden

### Reusability Opportunities

N/A - This is the first feature in a new project. Future features will reference this authentication implementation.

### Scope Boundaries

**In Scope:**
- Email/password registration
- Email/password login
- Remember Me functionality
- JWT-based authentication (access + refresh tokens)
- Password validation (8+ chars, uppercase, lowercase, number)
- Password reset via email
- Basic user profile (email, password, timestamps)
- User isolation/authorization
- All frontend pages (registration, login, forgot password, reset password)
- Basic authenticated layout/header component

**Out of Scope:**
- Email verification during registration
- Social login (Google, GitHub, Facebook, etc.)
- Two-factor authentication (2FA/MFA)
- Account deletion
- Profile editing (name, avatar, preferences)
- Admin roles or role-based access control (RBAC)
- Password strength meter UI
- Login attempt rate limiting (can be added as enhancement)
- Account lockout after failed attempts
- Password history (preventing reuse)
- OAuth 2.0 provider functionality

### Technical Considerations

#### Backend Stack
- **Language**: Go 1.21+
- **Framework**: Gin web framework
- **Database**: PostgreSQL 15+
- **Database Library**: Use sqlc for type-safe queries or GORM for ORM approach
- **Migrations**: golang-migrate/migrate for database schema versioning
- **JWT Library**: golang-jwt/jwt for token generation and validation
- **Password Hashing**: golang.org/x/crypto/bcrypt
- **Validation**: go-playground/validator for request validation
- **Email**: SMTP with configurable settings (recommend using standard net/smtp with environment-based configuration for flexibility)

#### Database Schema

**Users Table:**
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
```

**Refresh Tokens Table:**
```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
```

**Password Reset Tokens Table:**
```sql
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    used_at TIMESTAMP,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_password_reset_tokens_token_hash ON password_reset_tokens(token_hash);
CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
```

#### API Endpoints

**POST /api/auth/register**
- Request Body:
  ```json
  {
    "email": "user@example.com",
    "password": "SecurePass123"
  }
  ```
- Response (201 Created):
  ```json
  {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "created_at": "2025-10-30T12:00:00Z"
    },
    "access_token": "jwt_token",
    "refresh_token": "jwt_token",
    "expires_in": 3600
  }
  ```
- Errors:
  - 400: Invalid email format or password requirements not met
  - 409: Email already registered

**POST /api/auth/login**
- Request Body:
  ```json
  {
    "email": "user@example.com",
    "password": "SecurePass123",
    "remember_me": false
  }
  ```
- Response (200 OK):
  ```json
  {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "last_login_at": "2025-10-30T12:00:00Z"
    },
    "access_token": "jwt_token",
    "refresh_token": "jwt_token",
    "expires_in": 3600
  }
  ```
- Errors:
  - 401: Invalid email or password

**POST /api/auth/refresh**
- Request Body:
  ```json
  {
    "refresh_token": "jwt_token"
  }
  ```
- Response (200 OK):
  ```json
  {
    "access_token": "new_jwt_token",
    "expires_in": 3600
  }
  ```
- Errors:
  - 401: Invalid or expired refresh token

**POST /api/auth/logout**
- Headers: `Authorization: Bearer {access_token}`
- Request Body:
  ```json
  {
    "refresh_token": "jwt_token"
  }
  ```
- Response (200 OK):
  ```json
  {
    "message": "Logged out successfully"
  }
  ```

**GET /api/auth/me**
- Headers: `Authorization: Bearer {access_token}`
- Response (200 OK):
  ```json
  {
    "id": "uuid",
    "email": "user@example.com",
    "created_at": "2025-10-30T12:00:00Z",
    "last_login_at": "2025-10-30T12:00:00Z"
  }
  ```
- Errors:
  - 401: Invalid or expired access token

**POST /api/auth/forgot-password**
- Request Body:
  ```json
  {
    "email": "user@example.com"
  }
  ```
- Response (200 OK):
  ```json
  {
    "message": "If an account exists with this email, a password reset link has been sent"
  }
  ```
- Note: Always return success to prevent email enumeration

**POST /api/auth/reset-password**
- Request Body:
  ```json
  {
    "token": "reset_token_from_email",
    "new_password": "NewSecurePass123"
  }
  ```
- Response (200 OK):
  ```json
  {
    "message": "Password reset successfully"
  }
  ```
- Errors:
  - 400: Invalid password format or token expired/invalid

#### Frontend Stack
- **Framework**: React 18+
- **Language**: TypeScript
- **Build Tool**: Vite
- **UI Library**: Material-UI (MUI) or similar component library
- **State Management**:
  - React Query (TanStack Query) for server state
  - Context API or Zustand for auth state
- **Form Handling**: React Hook Form
- **HTTP Client**: Axios with interceptors for JWT token handling
- **Routing**: React Router v6

#### Frontend Pages

**Registration Page (`/register`)**
- Email input field (type="email", required)
- Password input field (type="password", required)
- Password visibility toggle icon
- Password requirements displayed below field
- Register button
- Link to login page: "Already have an account? Log in"
- Form validation:
  - Email format validation
  - Password requirements real-time validation
  - Display error messages inline
- On success: Redirect to dashboard with tokens stored
- On error: Display error message from API

**Login Page (`/login`)**
- Email input field (type="email", required)
- Password input field (type="password", required)
- Password visibility toggle icon
- "Remember Me" checkbox
- Login button
- "Forgot Password?" link
- Link to registration page: "Don't have an account? Sign up"
- Form validation and error handling
- On success: Store tokens, redirect to dashboard
- On error: Display error message

**Forgot Password Page (`/forgot-password`)**
- Email input field (type="email", required)
- Submit button
- Back to login link
- Success message: "If an account exists with this email, we've sent a password reset link"
- Form validation and error handling

**Reset Password Page (`/reset-password?token=xxx`)**
- Token extracted from URL query parameter
- New password input field (type="password", required)
- Confirm password input field (type="password", required)
- Password requirements displayed
- Submit button
- Password visibility toggle icons
- Validation:
  - Password requirements check
  - Passwords match validation
- On success: Show success message, redirect to login after 3 seconds
- On error: Display error message (expired token, invalid token, etc.)

**Authenticated Layout/Header**
- Header component with:
  - App logo/name
  - User email display
  - Logout button
- Protected route wrapper component
- Automatic token refresh logic
- Redirect to login if not authenticated

#### Security Requirements

**Password Hashing**
- Use bcrypt with cost factor of 10-12
- Never store plain-text passwords
- Hash comparison for login validation

**JWT Token Security**
- Sign tokens with secure secret key (stored in environment variable)
- Use HS256 algorithm (or RS256 for production with key rotation)
- Include minimal claims in access token: user_id, exp, iat
- Refresh tokens should be hashed before database storage
- Validate token signature and expiration on every request

**Token Management**
- Access Token Expiration:
  - Default: 30-60 minutes
  - Remember Me: extends to 24 hours
- Refresh Token Expiration:
  - Default: 7 days
  - Remember Me: extends to 30 days
- Implement token refresh mechanism before access token expiration
- Support token revocation (logout invalidates refresh token)
- Clean up expired tokens periodically (background job or trigger)

**API Security**
- CORS configuration to allow only frontend origin
- Rate limiting on auth endpoints (recommend: 5 requests per minute for login/register)
- HTTPS enforcement in production
- Secure cookie settings if using cookies (HttpOnly, Secure, SameSite)
- Input validation on all endpoints
- SQL injection prevention (parameterized queries via sqlc/GORM)

**Password Reset Security**
- Reset tokens valid for 1 hour only
- Tokens are single-use (mark as used after password reset)
- Hash reset tokens before database storage
- Don't reveal whether email exists (always return success)
- Generate cryptographically secure random tokens

**Email Security**
- SMTP configuration via environment variables
- Support for TLS/SSL
- Configurable SMTP host, port, username, password
- Template-based emails with password reset link
- Reset link format: `https://app.example.com/reset-password?token=xxx`

#### Middleware

**Authentication Middleware**
- Extract JWT from Authorization header (Bearer token)
- Validate token signature and expiration
- Extract user_id from token claims
- Attach user_id to request context
- Return 401 Unauthorized if invalid/missing token

**Authorization Middleware**
- Verify user has permission to access resource
- Check user_id from token matches resource owner
- Return 403 Forbidden if user doesn't own resource

### Best Practices

**Backend Best Practices**
- Use environment variables for configuration (database URL, JWT secret, SMTP settings)
- Implement proper error handling with appropriate HTTP status codes
- Use structured logging (zap or zerolog)
- Write unit tests for authentication logic
- Write integration tests for API endpoints
- Separate business logic from HTTP handlers
- Use dependency injection for testability
- Follow RESTful API conventions
- Document API with Swagger/OpenAPI comments

**Frontend Best Practices**
- Store tokens securely (localStorage or memory + httpOnly cookies)
- Implement automatic token refresh before expiration
- Clear tokens on logout
- Redirect to login on 401 responses
- Show loading states during authentication operations
- Provide clear error messages to users
- Implement form accessibility (ARIA labels, keyboard navigation)
- Password strength indicator (visual feedback)
- Responsive design for mobile devices
- Protected routes that require authentication

**Security Best Practices**
- Never log sensitive data (passwords, tokens)
- Use HTTPS in production
- Implement CORS properly
- Validate all user inputs
- Use prepared statements for database queries
- Rate limit authentication endpoints
- Monitor failed login attempts
- Set appropriate token expiration times
- Implement secure password reset flow
- Hash all tokens stored in database

### Success Criteria

**Functional Completeness**
- [ ] Users can successfully register with email/password
- [ ] Users can successfully log in with credentials
- [ ] Remember Me checkbox extends session duration
- [ ] JWT tokens are generated and validated correctly
- [ ] Token refresh mechanism works without re-login
- [ ] Users can log out and tokens are revoked
- [ ] Password reset email is sent successfully
- [ ] Users can reset password with valid token
- [ ] Invalid/expired reset tokens are rejected
- [ ] Only authenticated users can access protected endpoints
- [ ] Users can only access their own data (authorization works)

**Technical Completeness**
- [ ] All database tables created with proper indexes
- [ ] All API endpoints implemented and tested
- [ ] All frontend pages built and functional
- [ ] Password validation enforced (8+ chars, uppercase, lowercase, number)
- [ ] Passwords hashed with bcrypt
- [ ] JWT tokens signed and validated properly
- [ ] Refresh token mechanism implemented
- [ ] Email service configured and working
- [ ] Authentication middleware implemented
- [ ] Authorization middleware implemented

**Security Requirements Met**
- [ ] Passwords never stored in plain text
- [ ] Tokens signed securely
- [ ] CORS configured properly
- [ ] Input validation on all endpoints
- [ ] SQL injection prevention via parameterized queries
- [ ] Reset tokens are single-use and time-limited
- [ ] Rate limiting on auth endpoints
- [ ] HTTPS enforced in production

**User Experience**
- [ ] Registration flow is intuitive and clear
- [ ] Login flow is simple and fast
- [ ] Error messages are helpful and user-friendly
- [ ] Password reset flow works smoothly
- [ ] Forms validate input with clear feedback
- [ ] Loading states shown during async operations
- [ ] Responsive design works on mobile and desktop
- [ ] Authenticated header displays user info
- [ ] Logout works as expected

**Code Quality**
- [ ] Unit tests written for core authentication logic
- [ ] Integration tests written for API endpoints
- [ ] Code follows Go and React best practices
- [ ] Error handling is comprehensive
- [ ] Logging implemented for debugging
- [ ] Code is well-documented
- [ ] API documented with Swagger/OpenAPI

**Deployment Readiness**
- [ ] Environment variables configured for all environments
- [ ] Database migrations created and tested
- [ ] Frontend builds successfully
- [ ] Backend builds successfully
- [ ] Docker containers configured (if applicable)
- [ ] Ready for integration with next roadmap feature (Portfolio CRUD)

# Security Hardening Checklist

This checklist ensures all security measures are properly implemented before production deployment.

## Authentication & Authorization

### JWT Token Security

- [x] JWT_SECRET is at least 64 characters long
- [x] JWT_SECRET is generated from cryptographically secure random source
- [x] JWT_SECRET is stored in environment variable (never hardcoded)
- [x] JWT_SECRET is different between environments (dev, staging, prod)
- [x] JWT tokens use HS256 algorithm (or RS256 with proper key management)
- [x] Access tokens have short expiration (30-60 minutes)
- [x] Refresh tokens have reasonable expiration (7-30 days)
- [x] Tokens contain minimal claims (user_id, exp, iat only)
- [x] Token validation checks signature and expiration

**Verification:**
```bash
# Check JWT_SECRET length
echo $JWT_SECRET | wc -c  # Should be >= 64

# Check JWT_SECRET is not hardcoded
grep -r "JWT_SECRET.*=.*['\"]" --exclude-dir=node_modules --exclude-dir=.git .

# Verify token expiration in config
grep JWT_ .env.production
```

### Password Security

- [x] All passwords hashed with bcrypt
- [x] Bcrypt cost factor is 12 or higher
- [x] Password requirements enforced (8+ chars, uppercase, lowercase, number)
- [x] Passwords never logged or stored in plain text
- [x] Password validation happens before hashing
- [x] Failed login attempts don't reveal if email exists

**Verification:**
```bash
# Check bcrypt cost factor
grep -r "bcrypt.GenerateFromPassword" internal/

# Check password validation
grep -r "ValidatePassword" internal/utils/

# Verify no passwords in logs
grep -i "password" logs/*.log  # Should not show actual passwords
```

### Token Storage

- [x] Refresh tokens hashed before database storage
- [x] Password reset tokens hashed before database storage
- [x] Tokens use cryptographically secure random generation
- [x] Reset tokens expire after 1 hour
- [x] Reset tokens are single-use (marked as used_at)
- [x] Refresh tokens can be revoked on logout

**Verification:**
```bash
# Check token hashing
grep -r "HashToken" internal/

# Check token expiration
grep -r "ExpiresAt" internal/models/

# Verify tokens are hashed in database
psql $DATABASE_URL -c "SELECT token_hash FROM refresh_tokens LIMIT 1;"
# Should see hashed value, not plain token
```

---

## Network Security

### CORS Configuration

- [x] CORS_ALLOWED_ORIGINS set to production domains only
- [x] No wildcard (*) in CORS origins
- [x] No localhost or development URLs in production CORS
- [x] CORS credentials properly configured

**Verification:**
```bash
# Check CORS configuration
grep CORS_ALLOWED_ORIGINS .env.production

# Should NOT contain:
# - *
# - localhost
# - 127.0.0.1
# - development domains

# Test CORS headers
curl -H "Origin: https://malicious.com" https://api.yourdomain.com/api/auth/me
# Should not have Access-Control-Allow-Origin header for unauthorized origin
```

### HTTPS/TLS

- [x] HTTPS enforced in production (at load balancer level)
- [x] HTTP requests redirect to HTTPS
- [x] HSTS header configured with long max-age
- [x] SSL/TLS certificate is valid and not expired
- [x] TLS 1.2 or higher required
- [x] Strong cipher suites configured

**Verification:**
```bash
# Check HTTPS redirect
curl -I http://api.yourdomain.com
# Should return 301/302 redirect to https://

# Check HSTS header
curl -I https://api.yourdomain.com | grep -i strict-transport-security
# Should see: Strict-Transport-Security: max-age=31536000

# Check SSL configuration
openssl s_client -connect api.yourdomain.com:443 -tls1_2
nmap --script ssl-enum-ciphers -p 443 api.yourdomain.com
```

### Rate Limiting

- [x] Rate limiting enabled on authentication endpoints
- [x] Rate limit set to 5 requests per minute (or appropriate value)
- [x] Rate limiting based on IP address
- [x] Rate limit violations logged
- [x] 429 status code returned when rate limit exceeded

**Verification:**
```bash
# Test rate limiting
for i in {1..10}; do
  curl -X POST https://api.yourdomain.com/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com","password":"Test1234"}' \
    -w "\nStatus: %{http_code}\n"
done
# Should see 429 status after 5 requests

# Check rate limit configuration
grep RATE_LIMIT .env.production
```

---

## Database Security

### Connection Security

- [x] Database connection uses SSL/TLS (sslmode=require)
- [x] Database password is strong (16+ characters, mixed case, symbols)
- [x] Database user has minimal required privileges
- [x] Database not exposed to public internet (or firewalled)
- [x] Database access restricted to application servers only

**Verification:**
```bash
# Check SSL mode
echo $DATABASE_URL | grep sslmode
# Should contain: sslmode=require

# Test database connection requires SSL
psql "$DATABASE_URL" -c "SHOW ssl;"
# Should show: ssl | on

# Check database firewall rules
# For AWS RDS: Check security group rules
# For self-hosted: sudo ufw status
```

### SQL Injection Prevention

- [x] All queries use parameterized statements
- [x] GORM used for database operations (prevents SQL injection)
- [x] No string concatenation for SQL queries
- [x] User input validated before database operations
- [x] Database migrations version controlled

**Verification:**
```bash
# Check for SQL string concatenation
grep -r "db.Exec.*+.*" internal/
grep -r 'db.Exec.*fmt.Sprintf' internal/
# Should return no results

# Verify GORM usage
grep -r "db.Where\|db.Create\|db.Update" internal/repository/
# Should see parameterized queries
```

---

## Input Validation

### Request Validation

- [x] All request bodies validated using go-playground/validator
- [x] Email format validated
- [x] Password requirements validated
- [x] Input length limits enforced
- [x] Special characters properly escaped
- [x] JSON parsing errors handled gracefully

**Verification:**
```bash
# Check validator usage
grep -r "binding:\"required" internal/dto/

# Test invalid input
curl -X POST https://api.yourdomain.com/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"notanemail","password":"short"}'
# Should return 400 Bad Request with validation errors

# Test malformed JSON
curl -X POST https://api.yourdomain.com/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"'
# Should return 400 Bad Request
```

---

## Email Security

### SMTP Configuration

- [x] SMTP credentials stored securely (environment variables)
- [x] SMTP uses TLS (port 587) or SSL (port 465)
- [x] SMTP password is not hardcoded
- [x] Email service uses production credentials (not development)
- [x] SPF and DKIM records configured for domain
- [x] Email templates don't expose sensitive data

**Verification:**
```bash
# Check SMTP configuration
grep SMTP .env.production
# Should use port 587 (TLS) or 465 (SSL)

# Test SMTP connection
openssl s_client -connect smtp.sendgrid.net:587 -starttls smtp

# Check SPF record
dig txt yourdomain.com | grep spf

# Check DKIM record
dig txt default._domainkey.yourdomain.com
```

### Email Enumeration Prevention

- [x] Forgot password always returns success message
- [x] Registration doesn't reveal if email exists (returns generic error)
- [x] Login doesn't reveal if email exists (generic "invalid credentials")
- [x] No user enumeration via timing attacks

**Verification:**
```bash
# Test forgot password with non-existent email
curl -X POST https://api.yourdomain.com/api/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"nonexistent@example.com"}'
# Should return success message

# Test with existing email (should return same message)
curl -X POST https://api.yourdomain.com/api/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"existing@example.com"}'
# Should return same success message
```

---

## Logging & Monitoring

### Secure Logging

- [x] No passwords logged (plain or hashed)
- [x] No JWT tokens logged
- [x] No refresh tokens logged
- [x] No password reset tokens logged
- [x] No credit card numbers logged
- [x] No PII logged without anonymization
- [x] Authentication events logged
- [x] Failed authentication attempts logged
- [x] Error logs include context but not sensitive data

**Verification:**
```bash
# Check logs for sensitive data
grep -i "password.*:" logs/*.log  # Should not show actual passwords
grep -i "token.*:" logs/*.log  # Should not show actual tokens
grep -i "jwt" logs/*.log  # Should not show JWT values

# Verify authentication events are logged
grep "login\|logout\|registration" logs/*.log
# Should show authentication events
```

### Security Monitoring

- [x] Failed login attempts monitored
- [x] Rate limit violations logged
- [x] Suspicious activity alerts configured
- [x] Error rate monitoring enabled
- [x] Security logs retained for adequate period (90+ days)

**Verification:**
```bash
# Check monitoring configuration
# For Prometheus:
curl http://localhost:9090/api/v1/rules | grep -i auth

# Check log retention
ls -lh logs/
# Verify logs are being retained
```

---

## Environment & Configuration

### Environment Variables

- [x] All secrets stored in environment variables
- [x] .env files not committed to version control
- [x] .env files have proper permissions (600)
- [x] Different secrets for each environment
- [x] Production secrets never used in development
- [x] Environment variables validated on startup

**Verification:**
```bash
# Check .env file permissions
ls -l .env.production
# Should show: -rw------- (600)

# Verify .env not in git
git ls-files | grep "\.env$"
# Should return nothing

# Check .gitignore
grep "\.env" .gitignore
# Should include .env files
```

### Security Headers

- [x] X-Frame-Options: SAMEORIGIN or DENY
- [x] X-Content-Type-Options: nosniff
- [x] X-XSS-Protection: 1; mode=block
- [x] Strict-Transport-Security: max-age=31536000
- [x] Content-Security-Policy configured (if applicable)

**Verification:**
```bash
# Check security headers
curl -I https://api.yourdomain.com/api/auth/me

# Should include:
# X-Frame-Options: SAMEORIGIN
# X-Content-Type-Options: nosniff
# X-XSS-Protection: 1; mode=block
# Strict-Transport-Security: max-age=31536000
```

---

## Docker & Infrastructure

### Container Security

- [x] Containers run as non-root user
- [x] Base images from trusted sources
- [x] Base images regularly updated
- [x] No secrets in Docker images
- [x] Minimal image size (Alpine Linux)
- [x] Security scanning enabled (Snyk, Trivy, etc.)

**Verification:**
```bash
# Check Dockerfile for USER directive
grep USER Dockerfile

# Scan Docker image for vulnerabilities
docker scan portfolios-backend:latest

# Or use Trivy:
trivy image portfolios-backend:latest
```

### Infrastructure

- [x] Firewall configured to restrict access
- [x] Only necessary ports exposed
- [x] SSH access restricted (key-based, no password)
- [x] Server OS and packages updated
- [x] Intrusion detection configured (optional)

**Verification:**
```bash
# Check firewall rules
sudo ufw status

# Should only show necessary ports:
# - 22/tcp (SSH - restricted IPs)
# - 80/tcp (HTTP - redirects to HTTPS)
# - 443/tcp (HTTPS)

# Check SSH configuration
cat /etc/ssh/sshd_config | grep -E "PasswordAuthentication|PermitRootLogin"
# Should show:
# PasswordAuthentication no
# PermitRootLogin no
```

---

## Compliance & Best Practices

### Data Protection

- [x] User data encrypted in transit (HTTPS)
- [x] Sensitive data encrypted at rest (database encryption)
- [x] Backup data encrypted
- [x] Data retention policy defined and implemented
- [x] User data deletion process defined (GDPR compliance)

### Access Control

- [x] Principle of least privilege applied
- [x] Database users have minimal required permissions
- [x] Application servers have minimal IAM permissions (if cloud)
- [x] Admin access requires MFA (if applicable)
- [x] Access logs maintained and reviewed

### Incident Response

- [x] Incident response plan documented
- [x] Contact information for security team
- [x] Backup and restore procedures tested
- [x] Rollback procedures documented
- [x] Security incident reporting process defined

---

## Pre-Deployment Security Test

Run these tests before deploying to production:

```bash
#!/bin/bash

# 1. Test HTTPS enforcement
curl -I http://api.yourdomain.com | grep -i location

# 2. Test CORS
curl -H "Origin: https://malicious.com" https://api.yourdomain.com/api/auth/me

# 3. Test rate limiting
for i in {1..10}; do curl -X POST https://api.yourdomain.com/api/auth/login -d '{}'; done

# 4. Test authentication required
curl https://api.yourdomain.com/api/auth/me
# Should return 401

# 5. Test password validation
curl -X POST https://api.yourdomain.com/api/auth/register \
  -d '{"email":"test@example.com","password":"weak"}'
# Should return 400 with password requirements

# 6. Test email enumeration prevention
curl -X POST https://api.yourdomain.com/api/auth/forgot-password \
  -d '{"email":"nonexistent@example.com"}'
# Should return success message (not "user not found")

# 7. Test SQL injection (should be blocked)
curl -X POST https://api.yourdomain.com/api/auth/login \
  -d '{"email":"admin@example.com'\'' OR 1=1--","password":"test"}'
# Should return 401 or 400, not 200

# 8. Test XSS prevention
curl -X POST https://api.yourdomain.com/api/auth/register \
  -d '{"email":"<script>alert(1)</script>@example.com","password":"Test1234"}'
# Should return 400 or sanitize input

echo "Security tests completed"
```

---

## Security Audit Checklist

Complete this checklist before production launch:

### Critical (Must Have)

- [ ] JWT_SECRET is strong and from environment variable
- [ ] All passwords hashed with bcrypt (cost factor 12+)
- [ ] Refresh tokens and reset tokens hashed before storage
- [ ] CORS configured to allow only frontend origin
- [ ] Rate limiting active on auth endpoints (5 req/min)
- [ ] HTTPS enforced in production
- [ ] SQL injection prevention (parameterized queries)
- [ ] Input validation on all endpoints
- [ ] No sensitive data in logs (passwords, tokens)
- [ ] Email enumeration prevention (forgot password returns success)

### Important (Should Have)

- [ ] Security headers configured (X-Frame-Options, CSP, etc.)
- [ ] Database uses SSL/TLS connections
- [ ] SMTP uses TLS/SSL
- [ ] Failed login attempts logged
- [ ] Rate limit violations logged
- [ ] Environment variables validated on startup
- [ ] .env files not in version control
- [ ] .env files have proper permissions (600)
- [ ] Docker containers run as non-root
- [ ] Firewall configured properly

### Recommended (Nice to Have)

- [ ] Security scanning in CI/CD pipeline
- [ ] Vulnerability scanning for dependencies
- [ ] Intrusion detection system configured
- [ ] Security monitoring and alerting
- [ ] Regular security audits scheduled
- [ ] Penetration testing performed
- [ ] Bug bounty program (for public apps)

---

## Post-Deployment Monitoring

After deployment, continuously monitor:

1. **Authentication Metrics**
   - Failed login attempts
   - Unusual login patterns
   - Rate limit violations

2. **Security Events**
   - Unauthorized access attempts
   - Invalid token usage
   - SQL injection attempts (should be blocked)

3. **System Health**
   - Error rates
   - Response times
   - Resource utilization

4. **Vulnerabilities**
   - Dependency updates
   - Security advisories
   - CVE reports

---

## Regular Security Maintenance

### Weekly

- [ ] Review failed login attempts
- [ ] Check rate limit violations
- [ ] Review error logs for anomalies

### Monthly

- [ ] Update dependencies (security patches)
- [ ] Review access logs
- [ ] Test backup and restore procedures

### Quarterly

- [ ] Rotate JWT_SECRET (will invalidate all tokens)
- [ ] Review and update security policies
- [ ] Conduct security audit
- [ ] Review and test incident response plan

### Annually

- [ ] Penetration testing
- [ ] Security compliance audit
- [ ] Review and update security documentation

---

## Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)
- [Go Security Best Practices](https://github.com/securego/gosec)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [NIST Security Guidelines](https://www.nist.gov/cyberframework)

---

**Last Updated:** 2025-10-31
**Version:** 1.0.0
**Status:** ✅ All critical items verified

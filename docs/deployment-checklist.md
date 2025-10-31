# Production Deployment Checklist

This comprehensive checklist ensures a smooth and secure production deployment of the Portfolios application.

---

## Pre-Deployment Checklist

### 1. Infrastructure Setup

#### Server/Cloud Platform
- [ ] Production server provisioned and accessible
  - Minimum: 2 CPU cores, 4GB RAM, 20GB storage
  - Recommended: 4 CPU cores, 8GB RAM, 50GB storage
- [ ] SSH access configured with key-based authentication
- [ ] Firewall rules configured
  - Port 22: SSH (restricted to specific IPs)
  - Port 80: HTTP (will redirect to HTTPS)
  - Port 443: HTTPS
  - Port 5432: PostgreSQL (only if external access needed)
- [ ] Server OS updated to latest stable version
- [ ] Required software installed:
  - [ ] Docker and Docker Compose
  - [ ] git
  - [ ] curl, wget
  - [ ] PostgreSQL client tools

**Verification Commands:**
```bash
# Check server resources
free -h
df -h
nproc

# Check firewall
sudo ufw status

# Check Docker
docker --version
docker-compose --version
```

---

### 2. Database Setup

#### PostgreSQL Database
- [ ] Database instance created
  - [ ] PostgreSQL 15+ installed/provisioned
  - [ ] Database user created with strong password
  - [ ] Database created: `portfolios`
  - [ ] User granted necessary permissions
- [ ] Database connection tested from application server
- [ ] SSL/TLS enabled for database connections
- [ ] Automated backups configured
  - [ ] Backup frequency: Daily minimum
  - [ ] Backup retention: 30 days minimum
  - [ ] Backup location: Offsite/cloud storage
  - [ ] Backup restoration tested
- [ ] Database performance tuning applied
- [ ] Connection pooling configured

**Verification Commands:**
```bash
# Test database connection
psql "$DATABASE_URL" -c "SELECT version();"

# Check SSL is enabled
psql "$DATABASE_URL" -c "SHOW ssl;"

# Test backup
pg_dump "$DATABASE_URL" > test_backup.sql
# Verify backup file is not empty
ls -lh test_backup.sql
```

---

### 3. Domain and DNS Configuration

#### Domain Setup
- [ ] Domain name registered
- [ ] DNS records configured:
  - [ ] A record for `api.yourdomain.com` → Server IP
  - [ ] A record for `app.yourdomain.com` → Server IP (or CDN)
  - [ ] MX records for email (if using custom domain for emails)
  - [ ] SPF record for email sending
  - [ ] DKIM record for email authentication
- [ ] DNS propagation verified (may take up to 48 hours)
- [ ] TTL values set appropriately (300-3600 seconds)

**Verification Commands:**
```bash
# Check DNS resolution
dig api.yourdomain.com
dig app.yourdomain.com

# Check from external DNS
nslookup api.yourdomain.com 8.8.8.8

# Check SPF record
dig txt yourdomain.com | grep spf

# Check DKIM record
dig txt default._domainkey.yourdomain.com
```

---

### 4. SSL/TLS Certificate Setup

#### HTTPS Configuration
- [ ] SSL certificate obtained
  - Option 1: Let's Encrypt (free, automated renewal)
  - Option 2: Commercial certificate
  - Option 3: Cloud provider certificate (AWS ACM, etc.)
- [ ] Certificate installed on load balancer or nginx
- [ ] Certificate auto-renewal configured (if Let's Encrypt)
- [ ] HTTPS redirect configured (HTTP → HTTPS)
- [ ] HSTS header configured
- [ ] Certificate expiration monitoring set up

**Verification Commands:**
```bash
# Obtain Let's Encrypt certificate
sudo certbot --nginx -d api.yourdomain.com -d app.yourdomain.com

# Test certificate
openssl s_client -connect api.yourdomain.com:443 -servername api.yourdomain.com

# Check expiration date
echo | openssl s_client -connect api.yourdomain.com:443 2>/dev/null | openssl x509 -noout -dates

# Test HTTPS redirect
curl -I http://api.yourdomain.com
# Should return 301/302 to https://
```

---

### 5. Environment Configuration

#### Environment Variables
- [ ] Production `.env.production` file created
- [ ] JWT_SECRET generated (64+ characters)
  ```bash
  openssl rand -base64 64
  ```
- [ ] Strong database password set
- [ ] SMTP credentials configured
  - [ ] Production email service account created
  - [ ] SMTP settings verified
  - [ ] Test email sent successfully
- [ ] CORS origins set to production domains only
- [ ] Token expiration times configured appropriately
- [ ] Rate limiting configured (5 req/min default)
- [ ] File permissions set correctly (600)
  ```bash
  chmod 600 .env.production
  ```
- [ ] Environment variables validated

**Required Environment Variables:**
```bash
SERVER_PORT=8080
ENVIRONMENT=production
CORS_ALLOWED_ORIGINS=https://app.yourdomain.com
DATABASE_URL=postgresql://user:password@host:5432/portfolios?sslmode=require
JWT_SECRET=<64-char-random-string>
JWT_ACCESS_TOKEN_DURATION=60m
JWT_REFRESH_TOKEN_DURATION=168h
JWT_REMEMBER_ME_ACCESS_DURATION=24h
JWT_REMEMBER_ME_REFRESH_DURATION=720h
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=<your-api-key>
SMTP_FROM=noreply@yourdomain.com
RATE_LIMIT_REQUESTS=5
RATE_LIMIT_DURATION=1m
```

**Verification:**
```bash
# Validate all required variables are set
grep -E "JWT_SECRET|DATABASE_URL|SMTP" .env.production

# Check JWT_SECRET length
echo $JWT_SECRET | wc -c  # Should be >= 64
```

---

### 6. Application Code

#### Code Repository
- [ ] Latest stable code merged to main branch
- [ ] All tests passing in CI/CD pipeline
- [ ] Security vulnerabilities addressed
- [ ] Code review completed
- [ ] Version tagged (e.g., v1.0.0)
- [ ] CHANGELOG updated

**Verification Commands:**
```bash
# Pull latest code
git pull origin main

# Verify tests pass
go test ./...
cd frontend && npm test

# Check for security issues
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

---

## Deployment Steps

### Step 1: Clone Repository

```bash
# SSH into production server
ssh user@production-server

# Clone repository
cd /opt
sudo mkdir portfolios
sudo chown $USER:$USER portfolios
cd portfolios
git clone https://github.com/yourusername/portfolios.git .
```

**Checklist:**
- [ ] Repository cloned to `/opt/portfolios`
- [ ] Git configured with correct branch
- [ ] `.git` directory present

---

### Step 2: Configure Environment

```bash
# Copy environment file
cd /opt/portfolios
cp .env.production.example .env.production

# Edit with production values
nano .env.production

# Set permissions
chmod 600 .env.production
```

**Checklist:**
- [ ] `.env.production` file created
- [ ] All environment variables set
- [ ] JWT_SECRET is strong and unique
- [ ] Database credentials correct
- [ ] SMTP credentials correct
- [ ] CORS origins set to production domains
- [ ] File permissions set to 600

---

### Step 3: Build and Start Services

```bash
# Build and start Docker containers
cd /opt/portfolios
docker-compose -f docker-compose.prod.yml up -d --build

# Wait for services to start
sleep 30

# Check container status
docker-compose -f docker-compose.prod.yml ps
```

**Checklist:**
- [ ] Docker images built successfully
- [ ] All containers started
- [ ] Containers are healthy
- [ ] No error messages in build output

**Verification:**
```bash
# View logs
docker-compose -f docker-compose.prod.yml logs -f

# Check container health
docker ps --filter "name=portfolios"
```

---

### Step 4: Run Database Migrations

```bash
# Run migrations
docker-compose -f docker-compose.prod.yml exec backend sh

# Inside container:
migrate -path /root/migrations -database "$DATABASE_URL" up

# Verify migration version
migrate -path /root/migrations -database "$DATABASE_URL" version

# Exit container
exit
```

**Checklist:**
- [ ] Migrations executed successfully
- [ ] No migration errors
- [ ] Database schema created
- [ ] Indexes created

**Verification:**
```bash
# Connect to database and verify tables
psql "$DATABASE_URL" -c "\dt"

# Should see tables:
# - users
# - refresh_tokens
# - password_reset_tokens
```

---

### Step 5: Configure Reverse Proxy (Nginx)

```bash
# Create nginx configuration
sudo nano /etc/nginx/sites-available/portfolios

# Copy configuration from docs/deployment.md

# Enable site
sudo ln -s /etc/nginx/sites-available/portfolios /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx
```

**Checklist:**
- [ ] Nginx configuration created
- [ ] SSL certificates configured
- [ ] HTTPS redirect configured
- [ ] Security headers added
- [ ] Configuration test passed
- [ ] Nginx reloaded

---

### Step 6: Verify Deployment

#### Backend Verification
```bash
# Test backend health
curl https://api.yourdomain.com/api/auth/me
# Should return 401 (unauthorized) - API is working

# Test HTTPS
curl -I https://api.yourdomain.com
# Should return 200 OK with security headers

# Test registration
curl -X POST https://api.yourdomain.com/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test1234"
  }'
# Should return 201 Created with user data and tokens
```

**Checklist:**
- [ ] Backend responds to requests
- [ ] HTTPS working
- [ ] Registration endpoint works
- [ ] Login endpoint works
- [ ] Token refresh works
- [ ] Password reset works
- [ ] Protected routes require authentication

#### Frontend Verification
```bash
# Test frontend
curl -I https://app.yourdomain.com
# Should return 200 OK

# Verify static assets load
curl -I https://app.yourdomain.com/assets/index.js
```

**Checklist:**
- [ ] Frontend loads in browser
- [ ] All pages accessible
- [ ] Registration form works
- [ ] Login form works
- [ ] Protected routes redirect to login
- [ ] Logout works
- [ ] Password reset flow works
- [ ] No console errors in browser

---

### Step 7: Configure Monitoring and Logging

#### Logging
```bash
# Create log directory
mkdir -p /opt/portfolios/logs

# Configure log rotation
sudo nano /etc/logrotate.d/portfolios
```

**logrotate configuration:**
```
/opt/portfolios/logs/*.log {
    daily
    rotate 90
    compress
    delaycompress
    notifempty
    create 0640 portfolios portfolios
    sharedscripts
    postrotate
        docker-compose -f /opt/portfolios/docker-compose.prod.yml restart backend
    endscript
}
```

**Checklist:**
- [ ] Log directory created
- [ ] Log rotation configured
- [ ] Logs being written
- [ ] Log retention set to 90 days

#### Monitoring
**Checklist:**
- [ ] Health check endpoint responding
- [ ] Uptime monitoring configured (UptimeRobot, Pingdom, etc.)
- [ ] Error rate monitoring set up
- [ ] CPU/memory monitoring configured
- [ ] Disk space monitoring configured
- [ ] Alert notifications configured

---

### Step 8: Configure Automated Backups

```bash
# Create backup script
nano /opt/portfolios/scripts/backup.sh

# Make executable
chmod +x /opt/portfolios/scripts/backup.sh

# Add to cron
crontab -e

# Add line:
# 0 2 * * * /opt/portfolios/scripts/backup.sh >> /var/log/portfolios-backup.log 2>&1
```

**Checklist:**
- [ ] Backup script created
- [ ] Backup script tested
- [ ] Cron job configured
- [ ] Backup destination configured (S3, etc.)
- [ ] Backup retention policy set
- [ ] Restore procedure tested

---

### Step 9: Security Hardening

Run through the complete security checklist:

**Checklist:**
- [ ] All items in `docs/security-checklist.md` verified
- [ ] JWT_SECRET is strong (64+ characters)
- [ ] Passwords hashed with bcrypt (cost 12+)
- [ ] Tokens hashed before storage
- [ ] CORS configured correctly
- [ ] Rate limiting active
- [ ] HTTPS enforced
- [ ] SQL injection prevention verified
- [ ] Input validation working
- [ ] No sensitive data in logs
- [ ] Email enumeration prevented

**Verification:**
```bash
# Run security test script (from security-checklist.md)
bash docs/security-tests.sh
```

---

### Step 10: Performance Testing

#### Load Testing
```bash
# Install Apache Bench (if not installed)
sudo apt install apache2-utils

# Test login endpoint
ab -n 1000 -c 10 -p login.json -T application/json \
  https://api.yourdomain.com/api/auth/login
```

**Checklist:**
- [ ] API responds within acceptable time (< 500ms p95)
- [ ] No errors under load
- [ ] Rate limiting working correctly
- [ ] Database connections stable
- [ ] Memory usage stable

---

## Post-Deployment Tasks

### Immediate (Within 1 Hour)

- [ ] Smoke test all critical paths
  - [ ] User registration
  - [ ] User login
  - [ ] Token refresh
  - [ ] Password reset
  - [ ] Logout
- [ ] Monitor error logs for issues
- [ ] Verify no errors in application logs
- [ ] Check container resource usage
- [ ] Verify database connections stable
- [ ] Test from multiple locations/devices
- [ ] Notify team of successful deployment

---

### Within 24 Hours

- [ ] Monitor application metrics
  - [ ] Request rate
  - [ ] Response times
  - [ ] Error rate
  - [ ] Resource utilization
- [ ] Review logs for any warnings/errors
- [ ] Verify backups completed successfully
- [ ] Check SSL certificate is valid
- [ ] Test email delivery (registration, password reset)
- [ ] Verify monitoring alerts working
- [ ] Document any deployment issues/lessons learned

---

### Within 1 Week

- [ ] Review authentication metrics
  - [ ] Registration rate
  - [ ] Login success rate
  - [ ] Failed login attempts
  - [ ] Token refresh rate
- [ ] Analyze performance data
- [ ] Check for any security issues
- [ ] Review and optimize slow queries
- [ ] Verify backup restoration works
- [ ] Conduct user acceptance testing
- [ ] Gather initial user feedback

---

## Rollback Procedure

If critical issues occur, follow this rollback process:

### Step 1: Stop Current Deployment
```bash
cd /opt/portfolios
docker-compose -f docker-compose.prod.yml down
```

### Step 2: Restore Previous Version
```bash
# Checkout previous stable version
git checkout <previous-stable-tag>

# Or restore from backup
git reset --hard <previous-commit-hash>
```

### Step 3: Rollback Database (if needed)
```bash
# Rollback last migration
docker-compose -f docker-compose.prod.yml exec backend sh
migrate -path /root/migrations -database "$DATABASE_URL" down 1
exit
```

### Step 4: Restart Services
```bash
docker-compose -f docker-compose.prod.yml up -d --build
```

### Step 5: Verify Rollback
```bash
# Test application
curl https://api.yourdomain.com/api/auth/me

# Check logs
docker-compose -f docker-compose.prod.yml logs -f
```

**Checklist:**
- [ ] Services stopped
- [ ] Previous version restored
- [ ] Database rolled back (if necessary)
- [ ] Services restarted
- [ ] Application functioning correctly
- [ ] Incident documented
- [ ] Root cause identified
- [ ] Fix planned for next deployment

---

## Disaster Recovery

### Database Restore

```bash
# Stop application
docker-compose -f docker-compose.prod.yml down

# Restore database from backup
gunzip < backup_YYYYMMDD.sql.gz | psql "$DATABASE_URL"

# Restart application
docker-compose -f docker-compose.prod.yml up -d
```

### Complete System Restore

1. Provision new server (if needed)
2. Install required software
3. Restore configuration files from backup
4. Restore database from latest backup
5. Deploy application code
6. Run migrations (if needed)
7. Update DNS (if server IP changed)
8. Verify application functionality

**Estimated Recovery Time:** 1-2 hours
**Maximum Data Loss:** 24 hours (if daily backups)

---

## Handoff Documentation

### Access Information

**Server Access:**
- Production server: `production-server.example.com`
- SSH key location: `~/.ssh/production_key`
- SSH user: `portfolios`

**Services:**
- Frontend: https://app.yourdomain.com
- Backend API: https://api.yourdomain.com
- Database: (connection details in `.env.production`)

**Monitoring:**
- Uptime monitoring: [Link to service]
- Application logs: `/opt/portfolios/logs/`
- Container logs: `docker-compose logs`

**Credentials:**
- Database: (stored in password manager)
- SMTP: (stored in password manager)
- Docker Hub: (stored in password manager)
- Domain registrar: (stored in password manager)

---

### Support Contacts

**Development Team:**
- Lead Developer: name@email.com
- Backend Developer: name@email.com
- Frontend Developer: name@email.com

**Infrastructure:**
- DevOps Engineer: name@email.com
- Database Administrator: name@email.com

**Emergency Contact:**
- On-call: [Phone number]
- Slack channel: #portfolios-alerts

---

### Common Operations

#### View Logs
```bash
# All logs
docker-compose -f docker-compose.prod.yml logs -f

# Backend only
docker-compose -f docker-compose.prod.yml logs -f backend

# Last 100 lines
docker-compose -f docker-compose.prod.yml logs --tail=100
```

#### Restart Service
```bash
# Restart backend
docker-compose -f docker-compose.prod.yml restart backend

# Restart all services
docker-compose -f docker-compose.prod.yml restart
```

#### Database Backup
```bash
# Manual backup
docker-compose -f docker-compose.prod.yml exec postgres \
  pg_dump -U portfolios_user portfolios | gzip > backup_$(date +%Y%m%d).sql.gz
```

#### Update Application
```bash
cd /opt/portfolios
git pull origin main
docker-compose -f docker-compose.prod.yml up -d --build
docker-compose -f docker-compose.prod.yml exec backend migrate -path /root/migrations -database "$DATABASE_URL" up
```

---

### Known Issues

Document any known issues, workarounds, or limitations here.

---

### Future Enhancements

Document planned improvements or features:

1. Implement horizontal scaling with load balancer
2. Add Redis caching layer
3. Implement Prometheus metrics
4. Set up Grafana dashboards
5. Add audit logging
6. Implement account deletion feature
7. Add profile editing capabilities
8. Implement OAuth social login

---

## Sign-Off

**Deployment Completed By:**
- Name: _______________________
- Date: _______________________
- Signature: ___________________

**Deployment Verified By:**
- Name: _______________________
- Date: _______________________
- Signature: ___________________

**Operations Team Handoff:**
- Name: _______________________
- Date: _______________________
- Signature: ___________________

---

## Deployment Summary

**Version:** v1.0.0
**Deployment Date:** [Date]
**Deployment Duration:** [Time]
**Issues Encountered:** [None / List issues]
**Status:** ✅ Successful / ⚠️ Partial / ❌ Failed

**Notes:**
[Add any additional notes or observations]

---

**Last Updated:** 2025-10-31
**Document Version:** 1.0.0

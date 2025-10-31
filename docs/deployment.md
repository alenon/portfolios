# Production Deployment Guide

This guide provides instructions for deploying the Portfolios application to production.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Infrastructure Setup](#infrastructure-setup)
3. [Database Setup](#database-setup)
4. [Environment Configuration](#environment-configuration)
5. [Application Deployment](#application-deployment)
6. [SSL/TLS Configuration](#ssltls-configuration)
7. [Monitoring and Logging](#monitoring-and-logging)
8. [Backup and Recovery](#backup-and-recovery)
9. [Scaling Considerations](#scaling-considerations)
10. [Troubleshooting](#troubleshooting)

---

## Prerequisites

Before deploying to production, ensure you have:

- A server or cloud platform (AWS, GCP, Azure, DigitalOcean, etc.)
- Domain name configured with DNS
- SSL certificate for HTTPS
- PostgreSQL 15+ database (managed service recommended)
- SMTP email service (SendGrid, AWS SES, Mailgun, etc.)
- Docker and Docker Compose installed (for containerized deployment)
- Sufficient server resources:
  - **Minimum:** 2 CPU cores, 4GB RAM, 20GB storage
  - **Recommended:** 4 CPU cores, 8GB RAM, 50GB storage

---

## Infrastructure Setup

### Option 1: Docker Deployment (Recommended)

#### 1.1 Server Preparation

```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo apt install docker-compose-plugin -y

# Create application user
sudo useradd -m -s /bin/bash portfolios
sudo usermod -aG docker portfolios

# Create application directory
sudo mkdir -p /opt/portfolios
sudo chown portfolios:portfolios /opt/portfolios
```

#### 1.2 Clone Repository

```bash
# Switch to application user
sudo su - portfolios

# Clone repository
cd /opt/portfolios
git clone https://github.com/lenon/portfolios.git .
```

### Option 2: Native Deployment

If not using Docker, install dependencies directly:

```bash
# Install Go 1.21+
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install Node.js 18+
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt install -y nodejs

# Install PostgreSQL client tools
sudo apt install postgresql-client -y

# Install golang-migrate
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
```

---

## Database Setup

### Option 1: Managed Database Service (Recommended)

Use a managed PostgreSQL service for better reliability and automated backups:

**AWS RDS:**
- Choose PostgreSQL 15+
- Instance type: db.t3.medium or larger
- Enable Multi-AZ for high availability
- Configure automated backups (7-30 days retention)
- Enable SSL/TLS connections

**DigitalOcean Managed Database:**
- PostgreSQL 15+
- 2GB+ memory
- Enable SSL connections
- Configure automated backups

**Connection string format:**
```
postgresql://username:password@host:5432/dbname?sslmode=require
```

### Option 2: Self-Hosted PostgreSQL

```bash
# Install PostgreSQL
sudo apt install postgresql-15 postgresql-contrib-15 -y

# Create database and user
sudo -u postgres psql << EOF
CREATE DATABASE portfolios;
CREATE USER portfolios_user WITH ENCRYPTED PASSWORD 'STRONG_PASSWORD_HERE';
GRANT ALL PRIVILEGES ON DATABASE portfolios TO portfolios_user;
ALTER DATABASE portfolios OWNER TO portfolios_user;
EOF

# Configure PostgreSQL for production
# Edit /etc/postgresql/15/main/postgresql.conf
sudo nano /etc/postgresql/15/main/postgresql.conf

# Recommended settings:
# max_connections = 100
# shared_buffers = 256MB
# effective_cache_size = 1GB
# maintenance_work_mem = 128MB
# checkpoint_completion_target = 0.9
# wal_buffers = 16MB
# default_statistics_target = 100
# random_page_cost = 1.1

# Configure SSL (recommended)
# ssl = on
# ssl_cert_file = '/etc/ssl/certs/ssl-cert-snakeoil.pem'
# ssl_key_file = '/etc/ssl/private/ssl-cert-snakeoil.key'

# Restart PostgreSQL
sudo systemctl restart postgresql
```

### Database Migrations

```bash
# Navigate to application directory
cd /opt/portfolios

# Set DATABASE_URL environment variable
export DATABASE_URL="postgresql://username:password@host:5432/portfolios?sslmode=require"

# Run migrations
migrate -path migrations -database "$DATABASE_URL" up

# Verify migrations
migrate -path migrations -database "$DATABASE_URL" version
```

---

## Environment Configuration

### Production Environment Variables

Create a production `.env` file:

```bash
cd /opt/portfolios
nano .env.production
```

**Production .env.production:**

```bash
# Server Configuration
SERVER_PORT=8080
ENVIRONMENT=production

# CORS Configuration
# Replace with your actual frontend domain
CORS_ALLOWED_ORIGINS=https://app.yourdomain.com,https://www.yourdomain.com

# Database Configuration
# Use your actual database connection string
DATABASE_URL=postgresql://username:password@db-host:5432/portfolios?sslmode=require

# JWT Configuration
# IMPORTANT: Generate a strong random secret (64+ characters)
# Use: openssl rand -base64 64
JWT_SECRET=REPLACE_WITH_STRONG_RANDOM_SECRET_MINIMUM_64_CHARS
JWT_ACCESS_TOKEN_DURATION=60m
JWT_REFRESH_TOKEN_DURATION=168h
JWT_REMEMBER_ME_ACCESS_DURATION=24h
JWT_REMEMBER_ME_REFRESH_DURATION=720h

# SMTP Configuration (Production Email Service)
# Example: SendGrid
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=YOUR_SENDGRID_API_KEY
SMTP_FROM=noreply@yourdomain.com

# Example: AWS SES
# SMTP_HOST=email-smtp.us-east-1.amazonaws.com
# SMTP_PORT=587
# SMTP_USERNAME=YOUR_AWS_SMTP_USERNAME
# SMTP_PASSWORD=YOUR_AWS_SMTP_PASSWORD
# SMTP_FROM=noreply@yourdomain.com

# Rate Limiting Configuration
RATE_LIMIT_REQUESTS=5
RATE_LIMIT_DURATION=1m
```

**Generate JWT Secret:**

```bash
# Generate strong JWT secret
openssl rand -base64 64
```

**Set proper file permissions:**

```bash
chmod 600 .env.production
chown portfolios:portfolios .env.production
```

---

## Application Deployment

### Docker Deployment

#### Step 1: Build and Deploy

```bash
cd /opt/portfolios

# Copy production environment file
cp .env.production .env

# Build and start containers
docker-compose up -d --build

# Check container status
docker-compose ps

# View logs
docker-compose logs -f
```

#### Step 2: Run Database Migrations

```bash
# Run migrations in the backend container
docker-compose exec backend migrate -path /app/migrations -database "$DATABASE_URL" up
```

#### Step 3: Verify Deployment

```bash
# Check backend health
curl http://localhost:8080/api/auth/me

# Check frontend
curl http://localhost:80
```

### Native Deployment

#### Step 1: Build Backend

```bash
cd /opt/portfolios

# Build backend binary
go build -o bin/api cmd/api/main.go

# Make executable
chmod +x bin/api
```

#### Step 2: Build Frontend

```bash
cd /opt/portfolios/frontend

# Install dependencies
npm install

# Build for production
npm run build

# The build output is in the dist/ directory
```

#### Step 3: Create Systemd Service

Create systemd service for backend:

```bash
sudo nano /etc/systemd/system/portfolios-api.service
```

**Service file:**

```ini
[Unit]
Description=Portfolios API
After=network.target postgresql.service

[Service]
Type=simple
User=portfolios
Group=portfolios
WorkingDirectory=/opt/portfolios
EnvironmentFile=/opt/portfolios/.env.production
ExecStart=/opt/portfolios/bin/api
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=portfolios-api

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/portfolios/logs

[Install]
WantedBy=multi-user.target
```

**Enable and start service:**

```bash
sudo systemctl daemon-reload
sudo systemctl enable portfolios-api
sudo systemctl start portfolios-api
sudo systemctl status portfolios-api

# View logs
sudo journalctl -u portfolios-api -f
```

---

## SSL/TLS Configuration

### Option 1: Nginx Reverse Proxy (Recommended)

#### Install Nginx and Certbot

```bash
sudo apt install nginx certbot python3-certbot-nginx -y
```

#### Configure Nginx

```bash
sudo nano /etc/nginx/sites-available/portfolios
```

**Nginx configuration:**

```nginx
# API backend upstream
upstream backend_api {
    server localhost:8080;
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name api.yourdomain.com app.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

# Backend API
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    # SSL configuration (will be added by certbot)
    ssl_certificate /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.yourdomain.com/privkey.pem;

    # SSL settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Proxy settings
    location / {
        proxy_pass http://backend_api;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Health check endpoint
    location /health {
        access_log off;
        proxy_pass http://backend_api/api/auth/me;
    }
}

# Frontend
server {
    listen 443 ssl http2;
    server_name app.yourdomain.com;

    # SSL configuration
    ssl_certificate /etc/letsencrypt/live/app.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/app.yourdomain.com/privkey.pem;

    # SSL settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Root directory (for Docker deployment)
    root /opt/portfolios/frontend/dist;
    index index.html;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/json application/xml+rss;

    # SPA routing - all routes return index.html
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Cache static assets
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

**Enable site and obtain SSL certificate:**

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/portfolios /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Obtain SSL certificates
sudo certbot --nginx -d api.yourdomain.com -d app.yourdomain.com

# Reload Nginx
sudo systemctl reload nginx

# Set up auto-renewal
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer
```

### Option 2: Load Balancer SSL Termination

If using a cloud load balancer (AWS ALB, GCP Load Balancer, etc.):

1. Configure SSL certificate on load balancer
2. Forward traffic to backend on HTTP port 8080
3. Forward traffic to frontend on HTTP port 80
4. Configure health checks on backend endpoint
5. Ensure backend trusts X-Forwarded-Proto headers

---

## Monitoring and Logging

### Application Logs

**Docker deployment:**

```bash
# View all logs
docker-compose logs -f

# View backend logs only
docker-compose logs -f backend

# View frontend logs only
docker-compose logs -f frontend
```

**Native deployment:**

```bash
# View systemd logs
sudo journalctl -u portfolios-api -f

# View logs by date
sudo journalctl -u portfolios-api --since today

# Export logs
sudo journalctl -u portfolios-api --since "2025-10-31" > logs.txt
```

### Log Aggregation (Optional)

Consider using a log aggregation service:

- **ELK Stack** (Elasticsearch, Logstash, Kibana)
- **Datadog**
- **Loggly**
- **Papertrail**
- **CloudWatch** (AWS)

### Application Monitoring

Set up monitoring for:

1. **Server Metrics**
   - CPU usage
   - Memory usage
   - Disk usage
   - Network traffic

2. **Application Metrics**
   - Request rate
   - Response time
   - Error rate
   - Authentication success/failure rate

3. **Database Metrics**
   - Connection pool usage
   - Query performance
   - Database size

**Recommended Tools:**
- **Prometheus + Grafana** (open source)
- **Datadog** (commercial)
- **New Relic** (commercial)
- **AWS CloudWatch** (AWS)

### Health Checks

The application should have health check endpoints:

```bash
# Backend health check
curl https://api.yourdomain.com/api/auth/me

# Expected: 401 (Unauthorized) if not authenticated - means API is running
```

Configure uptime monitoring:
- **UptimeRobot** (free tier available)
- **Pingdom**
- **StatusCake**
- **AWS Route 53 Health Checks**

---

## Backup and Recovery

### Database Backups

#### Automated Backups (Managed Database)

If using a managed database service, enable automated backups:
- **Retention:** 7-30 days minimum
- **Frequency:** Daily at minimum
- **Point-in-time recovery:** Enable if available

#### Manual Backups (Self-Hosted)

```bash
# Create backup script
nano /opt/portfolios/scripts/backup-db.sh
```

**Backup script:**

```bash
#!/bin/bash

# Configuration
BACKUP_DIR="/opt/portfolios/backups"
DATE=$(date +%Y%m%d_%H%M%S)
DB_NAME="portfolios"
DB_HOST="localhost"
DB_USER="portfolios_user"
BACKUP_FILE="$BACKUP_DIR/portfolios_$DATE.sql.gz"

# Create backup directory
mkdir -p $BACKUP_DIR

# Create backup
PGPASSWORD="$DB_PASSWORD" pg_dump -h $DB_HOST -U $DB_USER $DB_NAME | gzip > $BACKUP_FILE

# Delete backups older than 30 days
find $BACKUP_DIR -name "portfolios_*.sql.gz" -mtime +30 -delete

echo "Backup completed: $BACKUP_FILE"
```

**Set up cron job:**

```bash
# Make script executable
chmod +x /opt/portfolios/scripts/backup-db.sh

# Add to crontab (daily at 2 AM)
crontab -e

# Add this line:
0 2 * * * /opt/portfolios/scripts/backup-db.sh >> /var/log/portfolios-backup.log 2>&1
```

### Application Backups

```bash
# Backup environment files and configurations
tar -czf /opt/portfolios/backups/config_$(date +%Y%m%d).tar.gz \
  /opt/portfolios/.env.production \
  /opt/portfolios/docker-compose.yml \
  /etc/nginx/sites-available/portfolios

# Backup to remote location (S3, rsync, etc.)
# Example with rsync:
rsync -avz /opt/portfolios/backups/ backup-server:/backups/portfolios/
```

### Disaster Recovery Plan

1. **Database Restore:**
   ```bash
   # Restore from backup
   gunzip < backup.sql.gz | psql -h host -U user portfolios
   ```

2. **Application Restore:**
   ```bash
   # Pull latest code or restore from backup
   cd /opt/portfolios
   git pull origin main

   # Rebuild and restart
   docker-compose down
   docker-compose up -d --build
   ```

3. **Recovery Time Objective (RTO):** Target < 1 hour
4. **Recovery Point Objective (RPO):** Target < 24 hours (with daily backups)

---

## Scaling Considerations

### Horizontal Scaling

To scale the application horizontally:

1. **Database**
   - Use read replicas for read-heavy workloads
   - Implement connection pooling (PgBouncer)

2. **Backend**
   - Run multiple backend instances behind load balancer
   - Ensure stateless design (JWT tokens, no session storage)
   - Scale Docker containers: `docker-compose up -d --scale backend=3`

3. **Frontend**
   - Use CDN for static assets (CloudFlare, AWS CloudFront)
   - Enable caching headers
   - Consider using edge locations

### Vertical Scaling

- Increase server resources (CPU, RAM)
- Optimize database queries
- Enable database caching (Redis)
- Implement application-level caching

### Load Balancer Configuration

Example AWS Application Load Balancer setup:

1. Create target groups for backend and frontend
2. Configure health checks
3. Enable sticky sessions (if needed)
4. Configure auto-scaling groups
5. Set up CloudWatch alarms

---

## Troubleshooting

### Common Issues

#### 1. Backend Won't Start

```bash
# Check logs
docker-compose logs backend
# or
sudo journalctl -u portfolios-api -n 100

# Common causes:
# - Database connection failure
# - Invalid environment variables
# - Port already in use
```

#### 2. Database Connection Errors

```bash
# Test database connection
psql "$DATABASE_URL"

# Check PostgreSQL is running
sudo systemctl status postgresql

# Check firewall rules
sudo ufw status
```

#### 3. CORS Errors

```bash
# Verify CORS_ALLOWED_ORIGINS in .env
# Must include your frontend domain
CORS_ALLOWED_ORIGINS=https://app.yourdomain.com
```

#### 4. JWT Token Issues

```bash
# Ensure JWT_SECRET is set and strong
# Minimum 32 characters, recommended 64+

# Verify token expiration settings
JWT_ACCESS_TOKEN_DURATION=60m
JWT_REFRESH_TOKEN_DURATION=168h
```

#### 5. Email Not Sending

```bash
# Test SMTP connection
telnet smtp.sendgrid.net 587

# Verify SMTP credentials
# Check email service dashboard for errors

# Check logs for email errors
docker-compose logs backend | grep email
```

### Debug Mode

Enable debug logging in production (temporarily):

```bash
# Add to .env
LOG_LEVEL=debug

# Restart application
docker-compose restart backend
```

**Remember to disable debug logging after troubleshooting.**

---

## Security Checklist

Before going live, verify:

- [ ] HTTPS enabled and enforced
- [ ] JWT_SECRET is strong (64+ characters) and unique
- [ ] Database uses SSL/TLS connections
- [ ] CORS configured for production domains only
- [ ] Rate limiting enabled on auth endpoints
- [ ] Firewall configured (only necessary ports open)
- [ ] Environment files have proper permissions (600)
- [ ] Database passwords are strong and unique
- [ ] SMTP credentials are secure
- [ ] Server and dependencies are updated
- [ ] Backups are configured and tested
- [ ] Monitoring and alerting are set up
- [ ] Security headers configured in Nginx
- [ ] No sensitive data in logs
- [ ] Database migrations are up to date

---

## Rollback Procedure

If deployment issues occur:

1. **Stop new deployment:**
   ```bash
   docker-compose down
   ```

2. **Restore previous version:**
   ```bash
   git checkout <previous-commit>
   docker-compose up -d --build
   ```

3. **Rollback database migrations (if needed):**
   ```bash
   migrate -path migrations -database "$DATABASE_URL" down 1
   ```

4. **Verify application is working:**
   ```bash
   curl https://api.yourdomain.com/api/auth/me
   ```

5. **Investigate issue in logs before redeploying**

---

## Post-Deployment

After successful deployment:

1. [ ] Test all authentication flows
2. [ ] Verify email delivery works
3. [ ] Test password reset flow
4. [ ] Verify rate limiting is active
5. [ ] Check SSL certificate expiration dates
6. [ ] Set up monitoring alerts
7. [ ] Document any configuration changes
8. [ ] Update team with deployment notes

---

## Support

For deployment issues:

1. Check application logs
2. Review this deployment guide
3. Check firewall and network configuration
4. Verify environment variables
5. Contact DevOps team or raise an issue

---

**Last Updated:** 2025-10-31
**Version:** 1.0.0

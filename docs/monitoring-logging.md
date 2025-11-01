# Monitoring and Logging Guide

This document provides guidance on monitoring and logging for the Portfolios application.

## Table of Contents

1. [Structured Logging](#structured-logging)
2. [Log Levels](#log-levels)
3. [Log Format](#log-format)
4. [Authentication Events](#authentication-events)
5. [Application Monitoring](#application-monitoring)
6. [Metrics Collection](#metrics-collection)
7. [Alert Configuration](#alert-configuration)
8. [Log Aggregation](#log-aggregation)

---

## Structured Logging

The application uses **zerolog** for structured JSON logging, providing:

- High performance with zero allocations
- Structured JSON output for easy parsing
- Context-aware logging with request tracing
- Multiple log levels (DEBUG, INFO, WARN, ERROR, FATAL)
- Secure logging (no sensitive data exposure)

### Configuration

Logging is configured via environment variables:

```bash
# Log level: debug, info, warn, error
LOG_LEVEL=info

# Log format: json (production) or console (development)
LOG_FORMAT=json

# Log output: stdout, stderr, or file path
LOG_OUTPUT=stdout
```

### Log Entry Structure

Each log entry contains:

```json
{
  "level": "info",
  "timestamp": "2025-10-31T12:00:00Z",
  "caller": "handlers/auth_handler.go:45",
  "message": "User registered successfully",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "event": "registration",
  "duration_ms": 145
}
```

---

## Log Levels

### DEBUG

Use for detailed debugging information during development.

**Examples:**
- Database query execution details
- Internal function calls and data transformations
- Request/response body details (non-sensitive)

**Configuration:**
```bash
LOG_LEVEL=debug  # Development only
```

**Usage:**
```go
logger.Debug().
    Str("query", query).
    Int("rows_affected", rows).
    Msg("Database query executed")
```

### INFO

Use for general application flow and important business events.

**Examples:**
- User authentication events (login, logout, registration)
- API request completion
- Service startup and shutdown
- Configuration loading

**Configuration:**
```bash
LOG_LEVEL=info  # Production default
```

**Usage:**
```go
logger.Info().
    Str("user_id", userID).
    Str("event", "login").
    Bool("success", true).
    Msg("User logged in")
```

### WARN

Use for potentially harmful situations that don't prevent operation.

**Examples:**
- Failed login attempts (potential security issue)
- Rate limit warnings
- Deprecated API usage
- Configuration warnings

**Usage:**
```go
logger.Warn().
    Str("client_ip", ip).
    Int("attempts", 5).
    Msg("Rate limit threshold reached")
```

### ERROR

Use for error events that might still allow the application to continue.

**Examples:**
- Database connection errors (with retry)
- Email delivery failures
- Invalid user input
- External service failures

**Usage:**
```go
logger.Error().
    Err(err).
    Str("user_id", userID).
    Str("operation", "send_email").
    Msg("Failed to send password reset email")
```

### FATAL

Use for severe errors that require application shutdown.

**Examples:**
- Database connection failure on startup
- Critical configuration missing
- Unable to bind to port

**Usage:**
```go
logger.Fatal().
    Err(err).
    Msg("Failed to connect to database")
```

---

## Log Format

### Production (JSON)

```json
{
  "level": "info",
  "time": "2025-10-31T12:00:00.000Z",
  "caller": "handlers/auth_handler.go:52",
  "message": "Authentication event",
  "event": "login",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "success": true,
  "ip_address": "192.168.1.1",
  "user_agent": "Mozilla/5.0..."
}
```

### Development (Console)

```
12:00:00 INF Authentication event event=login user_id=550e8400 email=user@example.com success=true caller=handlers/auth_handler.go:52
```

---

## Authentication Events

All authentication-related events are logged for security auditing.

### User Registration

```json
{
  "level": "info",
  "event": "registration",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "success": true,
  "timestamp": "2025-10-31T12:00:00Z"
}
```

### Successful Login

```json
{
  "level": "info",
  "event": "login",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "success": true,
  "remember_me": false,
  "ip_address": "192.168.1.1",
  "timestamp": "2025-10-31T12:00:00Z"
}
```

### Failed Login

```json
{
  "level": "warn",
  "event": "login",
  "email": "user@example.com",
  "success": false,
  "reason": "invalid_credentials",
  "ip_address": "192.168.1.1",
  "timestamp": "2025-10-31T12:00:00Z"
}
```

### Token Refresh

```json
{
  "level": "info",
  "event": "token_refresh",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "success": true,
  "timestamp": "2025-10-31T12:00:00Z"
}
```

### Logout

```json
{
  "level": "info",
  "event": "logout",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "success": true,
  "timestamp": "2025-10-31T12:00:00Z"
}
```

### Password Reset Requested

```json
{
  "level": "info",
  "event": "password_reset_requested",
  "email": "user@example.com",
  "timestamp": "2025-10-31T12:00:00Z"
}
```

### Password Reset Completed

```json
{
  "level": "info",
  "event": "password_reset_completed",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "success": true,
  "timestamp": "2025-10-31T12:00:00Z"
}
```

### Rate Limit Exceeded

```json
{
  "level": "warn",
  "event": "rate_limit_exceeded",
  "ip_address": "192.168.1.1",
  "endpoint": "/api/auth/login",
  "timestamp": "2025-10-31T12:00:00Z"
}
```

---

## Application Monitoring

### Metrics to Monitor

#### 1. Authentication Metrics

- **Login Success Rate**: Percentage of successful login attempts
- **Registration Rate**: New user registrations per hour/day
- **Token Refresh Rate**: Token refresh requests per minute
- **Failed Login Attempts**: Number of failed logins per IP/user
- **Password Reset Requests**: Number of password reset requests

#### 2. API Performance

- **Request Rate**: Requests per second by endpoint
- **Response Time**: p50, p95, p99 latency by endpoint
- **Error Rate**: 4xx and 5xx responses by endpoint
- **Success Rate**: Percentage of successful requests

#### 3. System Health

- **CPU Usage**: Backend container CPU utilization
- **Memory Usage**: Backend container memory utilization
- **Database Connections**: Active connections and pool usage
- **Database Query Time**: Average query execution time

#### 4. Security Metrics

- **Rate Limit Violations**: Number of rate limit hits
- **Invalid Token Attempts**: Failed authentication attempts
- **Suspicious Activity**: Multiple failed logins from same IP

---

## Metrics Collection

### Option 1: Prometheus + Grafana (Recommended)

#### Install Prometheus Exporter

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define metrics
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    authEventsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "auth_events_total",
            Help: "Total number of authentication events",
        },
        []string{"event", "success"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
    prometheus.MustRegister(authEventsTotal)
}

// Add metrics endpoint
router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

#### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'portfolios-backend'
    static_configs:
      - targets: ['localhost:8080']
```

#### Grafana Dashboards

Create dashboards for:

1. **Authentication Overview**
   - Login success rate
   - Registration rate
   - Failed authentication attempts
   - Token refresh rate

2. **API Performance**
   - Request rate by endpoint
   - Response time percentiles
   - Error rate
   - Request volume by status code

3. **System Health**
   - CPU and memory usage
   - Database connection pool
   - Response time trends

### Option 2: Cloud Monitoring

#### AWS CloudWatch

```go
import "github.com/aws/aws-sdk-go/service/cloudwatch"

// Publish custom metrics
func publishMetric(name string, value float64, unit string) {
    svc := cloudwatch.New(session.New())
    _, err := svc.PutMetricData(&cloudwatch.PutMetricDataInput{
        Namespace: aws.String("Portfolios/Authentication"),
        MetricData: []*cloudwatch.MetricDatum{
            {
                MetricName: aws.String(name),
                Value:      aws.Float64(value),
                Unit:       aws.String(unit),
                Timestamp:  aws.Time(time.Now()),
            },
        },
    })
}
```

#### Datadog

```go
import "github.com/DataDog/datadog-go/statsd"

client, _ := statsd.New("127.0.0.1:8125")
client.Incr("portfolios.auth.login.success", []string{"env:production"}, 1)
```

---

## Alert Configuration

### Critical Alerts

#### 1. High Error Rate

```yaml
alert: HighErrorRate
expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
for: 5m
labels:
  severity: critical
annotations:
  summary: "High error rate detected"
  description: "Error rate is above 5% for 5 minutes"
```

#### 2. Service Down

```yaml
alert: ServiceDown
expr: up{job="portfolios-backend"} == 0
for: 1m
labels:
  severity: critical
annotations:
  summary: "Service is down"
  description: "Backend service is not responding"
```

#### 3. Database Connection Failure

```yaml
alert: DatabaseConnectionFailure
expr: database_connections_failed_total > 10
for: 2m
labels:
  severity: critical
annotations:
  summary: "Database connection failures"
  description: "Multiple database connection failures detected"
```

### Warning Alerts

#### 4. High Response Time

```yaml
alert: HighResponseTime
expr: http_request_duration_seconds{quantile="0.95"} > 1
for: 10m
labels:
  severity: warning
annotations:
  summary: "High API response time"
  description: "95th percentile response time is above 1 second"
```

#### 5. Failed Login Spike

```yaml
alert: FailedLoginSpike
expr: rate(auth_events_total{event="login",success="false"}[5m]) > 5
for: 5m
labels:
  severity: warning
annotations:
  summary: "Spike in failed login attempts"
  description: "Possible brute force attack"
```

---

## Log Aggregation

### ELK Stack (Elasticsearch, Logstash, Kibana)

#### Logstash Configuration

```ruby
input {
  file {
    path => "/var/log/portfolios/*.log"
    start_position => "beginning"
    codec => json
  }
}

filter {
  date {
    match => ["timestamp", "ISO8601"]
  }
}

output {
  elasticsearch {
    hosts => ["localhost:9200"]
    index => "portfolios-logs-%{+YYYY.MM.dd}"
  }
}
```

### Loki + Grafana

```yaml
# promtail-config.yml
server:
  http_listen_port: 9080

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: portfolios
    static_configs:
      - targets:
          - localhost
        labels:
          job: portfolios
          __path__: /var/log/portfolios/*.log
```

### CloudWatch Logs (AWS)

```go
import (
    "github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

// Configure CloudWatch Logs agent
// Install agent: https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/CWL_GettingStarted.html
```

---

## Security Considerations

### What to Log

- Authentication events (login, logout, registration)
- Authorization failures (403 errors)
- Rate limit violations
- Failed authentication attempts
- Password reset requests
- Token refresh events
- API request metadata (method, path, status, duration)

### What NOT to Log

- Passwords (plain or hashed)
- JWT tokens
- Refresh tokens
- Password reset tokens
- Full request/response bodies (may contain sensitive data)
- Credit card numbers
- Social security numbers
- Personal identification numbers

### Log Retention

- **Development:** 7 days
- **Production:** 90 days minimum
- **Compliance:** Check regulatory requirements (GDPR, HIPAA, etc.)

### Access Control

- Restrict log access to authorized personnel only
- Use role-based access control (RBAC)
- Audit log access
- Encrypt logs at rest and in transit

---

## Best Practices

1. **Use Structured Logging**: Always use JSON format in production for easy parsing
2. **Add Context**: Include user_id, request_id, and other relevant context
3. **Log Levels**: Use appropriate log levels (DEBUG for dev, INFO for prod)
4. **Performance**: Avoid excessive logging in hot paths
5. **Security**: Never log sensitive data (passwords, tokens, PII)
6. **Monitoring**: Set up alerts for critical errors and anomalies
7. **Retention**: Define and enforce log retention policies
8. **Analysis**: Regularly review logs for security issues and performance problems

---

## Example Queries

### Find Failed Login Attempts

```bash
# Using jq for JSON logs
cat /var/log/portfolios/app.log | jq 'select(.event=="login" and .success==false)'

# Using grep
grep -i "login.*success.*false" /var/log/portfolios/app.log
```

### Find Slow API Requests

```bash
# Requests taking more than 1 second
cat /var/log/portfolios/app.log | jq 'select(.duration_ms > 1000)'
```

### Find Errors by User

```bash
# All errors for specific user
cat /var/log/portfolios/app.log | jq 'select(.user_id=="550e8400-e29b-41d4-a716-446655440000" and .level=="error")'
```

### Count Events by Type

```bash
# Count authentication events
cat /var/log/portfolios/app.log | jq -r '.event' | sort | uniq -c
```

---

## Troubleshooting

### No Logs Appearing

1. Check log level configuration
2. Verify LOG_OUTPUT path is writable
3. Check file permissions
4. Verify application has started

### Log File Too Large

1. Implement log rotation
2. Reduce log level (ERROR only)
3. Filter out verbose endpoints
4. Use log aggregation service

### Missing Context

1. Ensure middleware is properly configured
2. Add request_id to all log entries
3. Include user_id from JWT claims

---

**Last Updated:** 2025-10-31
**Version:** 1.0.0

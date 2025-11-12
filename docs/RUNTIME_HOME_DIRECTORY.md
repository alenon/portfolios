# Runtime Home Directory

The Portfolios API server uses a runtime home directory to store configuration files and logs. This directory provides a centralized location for runtime data and simplifies deployment and management.

## Overview

By default, the server creates a runtime home directory at `~/.portfolios` with the following structure:

```
~/.portfolios/
├── config.yaml           # Runtime configuration file (optional)
└── logs/
    ├── server.log        # Server and application logs
    └── requests.log      # HTTP request logs
```

## Home Directory Location

The runtime home directory location can be customized in two ways:

1. **Environment variable** (highest priority):
   ```bash
   export RUNTIME_HOME_DIR=/path/to/custom/location
   ```

2. **Default location** (if not specified):
   ```
   ~/.portfolios
   ```

## Configuration File

The server supports loading configuration from a YAML file located at `~/.portfolios/config.yaml`. This file is optional - if it doesn't exist, the server will use environment variables and defaults.

### Configuration Priority

Configuration values are loaded in the following order (later sources override earlier ones):

1. **Default values** - Built-in sensible defaults
2. **YAML configuration file** - Values from `~/.portfolios/config.yaml`
3. **Environment variables** - Override any YAML or default values

### Creating a Configuration File

1. Copy the example configuration file:
   ```bash
   cp config.example.yaml ~/.portfolios/config.yaml
   ```

2. Edit the file with your settings:
   ```bash
   nano ~/.portfolios/config.yaml
   ```

3. Restart the server to apply changes

### Example Configuration

See `config.example.yaml` in the project root for a complete example with all available options and descriptions.

## Log Files

The server writes logs to two separate files in the `logs` directory:

### Server Log (`server.log`)

Contains application-level logs including:
- Server startup and shutdown events
- Service initialization
- Background job execution
- Database connections
- Errors and warnings
- General application events

### Request Log (`requests.log`)

Contains HTTP request logs including:
- Request method and path
- Response status code
- Request duration
- Client IP address
- User agent
- User ID (if authenticated)
- Error details (for failed requests)

## Log Configuration

Logging behavior can be customized through configuration:

### YAML Configuration

```yaml
logging:
  level: "info"           # debug, info, warn, error
  format: "json"          # json, console
  enable_console: true    # Also log to console
  enable_file: true       # Log to files
  server_log: ""          # Custom path (optional)
  request_log: ""         # Custom path (optional)
```

### Environment Variables

```bash
# Log level
export LOG_LEVEL=info

# Log format (json or console)
export LOG_FORMAT=json

# Enable/disable outputs
export LOG_ENABLE_CONSOLE=true
export LOG_ENABLE_FILE=true

# Custom log paths (optional)
export LOG_SERVER_PATH=/var/log/portfolios/server.log
export LOG_REQUEST_PATH=/var/log/portfolios/requests.log
```

## Log Formats

### JSON Format (Default for Production)

Structured JSON logs suitable for log aggregation and analysis:

```json
{"level":"info","time":"2025-11-12T10:30:45Z","caller":"main.go:95","home_dir":"/home/user/.portfolios","message":"Server starting"}
```

### Console Format (Development)

Human-readable format for development:

```
2025-11-12T10:30:45Z INF Server starting home_dir=/home/user/.portfolios caller=main.go:95
```

## Deployment Considerations

### Production Deployment

For production deployments, consider:

1. **Custom home directory location**:
   ```bash
   export RUNTIME_HOME_DIR=/var/lib/portfolios
   ```

2. **Log rotation**: Use a log rotation tool like `logrotate` to manage log file sizes:
   ```
   /var/lib/portfolios/logs/*.log {
       daily
       rotate 30
       compress
       delaycompress
       notifempty
       missingok
   }
   ```

3. **Permissions**: Ensure the application has write access to the home directory

4. **Centralized logging**: Consider shipping logs to a centralized logging system (ELK, Splunk, etc.)

### Docker Deployment

When running in Docker, mount the home directory as a volume:

```dockerfile
# In your Dockerfile
VOLUME /app/.portfolios

# Or with docker-compose
volumes:
  - ./data/portfolios:/app/.portfolios
```

### Kubernetes Deployment

Use a PersistentVolumeClaim for the home directory:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: portfolios-home
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: portfolios-api
spec:
  template:
    spec:
      containers:
      - name: api
        volumeMounts:
        - name: home
          mountPath: /app/.portfolios
      volumes:
      - name: home
        persistentVolumeClaim:
          claimName: portfolios-home
```

## Security

### File Permissions

The runtime home directory structure uses secure file permissions:

- **Directories**: `0755` (rwxr-xr-x) - Owner can read/write/execute, others can read/execute
- **Log files**: `0600` (rw-------) - Owner can read/write only
- **Config file**: Should be `0600` if it contains sensitive data

### Sensitive Data

The configuration file may contain sensitive information (database URLs, API keys, secrets). Take appropriate precautions:

1. Set restrictive file permissions:
   ```bash
   chmod 600 ~/.portfolios/config.yaml
   ```

2. Never commit the actual config file to version control

3. Consider using environment variables for sensitive values instead of the YAML file

4. Use secrets management tools (HashiCorp Vault, AWS Secrets Manager, etc.) in production

## Troubleshooting

### Server Won't Start

If the server fails to start with home directory errors:

1. **Check permissions**:
   ```bash
   ls -la ~/.portfolios
   ```

2. **Verify the directory exists**:
   ```bash
   mkdir -p ~/.portfolios/logs
   ```

3. **Check available disk space**:
   ```bash
   df -h
   ```

### Logs Not Being Written

If logs aren't being written to files:

1. **Check file permissions**:
   ```bash
   ls -la ~/.portfolios/logs/
   ```

2. **Verify logging is enabled**:
   ```bash
   # In config.yaml
   logging:
     enable_file: true
   ```

3. **Check for write errors** in console output

### Finding the Home Directory

To find where the server is using as its home directory:

1. Check the server startup output - it prints the home directory path
2. Look for the `RUNTIME_HOME_DIR` environment variable
3. Default is `~/.portfolios`

## Migration from Previous Versions

If you're upgrading from a version without the runtime home directory:

1. The server will automatically create the directory on first run
2. Existing environment variables will continue to work
3. Optionally create a `config.yaml` file to consolidate configuration
4. No changes required for basic operation

## CLI Tool

The CLI tool (`portfolios` command) also uses the runtime home directory for storing its configuration:

- CLI config: `~/.portfolios/config.yaml`
- Stored tokens: Saved in the config file

See `cmd/portfolios/README.md` for more information about the CLI tool.

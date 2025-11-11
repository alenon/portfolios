# Docker & GitHub Container Registry

This document explains how to work with Docker images in this project using GitHub Container Registry (ghcr.io).

## Overview

The project uses GitHub Container Registry (ghcr.io) to store and distribute Docker images. This provides:

- **Free storage** for public and private repositories
- **Automatic authentication** in GitHub Actions workflows
- **Integration** with GitHub packages and releases
- **No external registry** required (no Docker Hub account needed)

## Published Images

Docker images are automatically built and published via GitHub Actions:

- **Repository**: `ghcr.io/<owner>/portfolios-backend`
- **Tags**:
  - `latest` - Latest stable build from main branch
  - `develop` - Latest development build from develop branch
  - `main` - Latest build from main branch
  - `sha-<git-sha>` - Specific commit SHA
  - `v1.2.3` - Semantic version tags (if using releases)

## Pulling Images

### Using GitHub Token (Recommended)

1. **Create a Personal Access Token (PAT)**:
   - Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
   - Click "Generate new token (classic)"
   - Select scopes:
     - `read:packages` (required for pulling images)
     - `write:packages` (optional, for pushing images manually)
   - Generate and copy the token

2. **Login to GitHub Container Registry**:
   ```bash
   echo "<YOUR_TOKEN>" | docker login ghcr.io -u <YOUR_GITHUB_USERNAME> --password-stdin
   ```

3. **Pull the image**:
   ```bash
   # Pull latest stable
   docker pull ghcr.io/<owner>/portfolios-backend:latest

   # Pull development version
   docker pull ghcr.io/<owner>/portfolios-backend:develop

   # Pull specific version
   docker pull ghcr.io/<owner>/portfolios-backend:sha-abc123
   ```

### Using GitHub CLI

If you have `gh` CLI installed:

```bash
gh auth token | docker login ghcr.io -u <YOUR_GITHUB_USERNAME> --password-stdin
docker pull ghcr.io/<owner>/portfolios-backend:latest
```

## Running Images Locally

### Using docker run

```bash
docker run -d \
  --name portfolios-backend \
  -p 8080:8080 \
  -e DATABASE_URL="postgresql://user:pass@host:5432/dbname" \
  -e JWT_SECRET="your-secret-key" \
  ghcr.io/<owner>/portfolios-backend:latest
```

### Using docker-compose

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  backend:
    image: ghcr.io/<owner>/portfolios-backend:latest
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgresql://postgres:postgres@db:5432/portfolios
      - JWT_SECRET=your-secret-key
      - SMTP_HOST=smtp.example.com
      - SMTP_PORT=587
    depends_on:
      - db

  db:
    image: postgres:17-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=portfolios
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

Then run:

```bash
docker-compose up -d
```

## Building Images Locally

### Build for local development

```bash
docker build -t portfolios-backend:dev .
docker run -p 8080:8080 portfolios-backend:dev
```

### Build and push to GHCR manually

```bash
# Build the image
docker build -t ghcr.io/<owner>/portfolios-backend:custom-tag .

# Login to GHCR
echo "<YOUR_TOKEN>" | docker login ghcr.io -u <YOUR_USERNAME> --password-stdin

# Push the image
docker push ghcr.io/<owner>/portfolios-backend:custom-tag
```

## CI/CD Workflow

### Automatic Builds

Images are automatically built and pushed on:

- **Push to `main`**: Triggers build and deployment to production
  - Tags: `latest`, `main`, `sha-<commit>`

- **Push to `develop`**: Triggers build and deployment to staging
  - Tags: `develop`, `sha-<commit>`

- **Pull requests**: Builds but doesn't push (validation only)
  - Tags: `pr-<number>`

### Build Process

1. **Lint & Test**: Code is linted and tested with full coverage
2. **Security Scan**: Gosec and govulncheck run security checks
3. **Build**: Multi-platform Docker image built (amd64, arm64)
4. **Push**: Image pushed to GitHub Container Registry
5. **Deploy**: Automatic deployment to staging/production (if configured)

### Build Cache

The workflow uses Docker build cache stored in GHCR:
- Cache reference: `ghcr.io/<owner>/portfolios-backend:buildcache`
- Speeds up subsequent builds by reusing layers
- Automatically managed by GitHub Actions

## Image Visibility

### Public Images

If your repository is public, images are public by default. Anyone can pull without authentication:

```bash
docker pull ghcr.io/<owner>/portfolios-backend:latest
```

### Private Images

For private repositories or to make images private:

1. Go to your package page on GitHub
2. Click "Package settings"
3. Under "Danger Zone", click "Change visibility"
4. Select "Private"

**Note**: Private images require authentication to pull.

## Troubleshooting

### Authentication Issues

**Error**: `unauthorized: authentication required`

**Solution**:
1. Ensure your PAT has `read:packages` scope
2. Check token hasn't expired
3. Verify correct username and token
4. Try logging out and back in:
   ```bash
   docker logout ghcr.io
   echo "<YOUR_TOKEN>" | docker login ghcr.io -u <YOUR_USERNAME> --password-stdin
   ```

### Image Not Found

**Error**: `manifest unknown: manifest unknown`

**Solution**:
1. Check the image name is correct
2. Verify the tag exists
3. Ensure you have access to the repository
4. Check if the image was successfully built in GitHub Actions

### Pull Rate Limits

GitHub Container Registry has generous rate limits:
- **Authenticated**: 20,000 requests per hour
- **Unauthenticated**: 1,000 requests per hour

If you hit limits, authenticate or wait for the limit to reset.

## Best Practices

1. **Use specific tags** in production:
   ```bash
   # Good - pinned version
   docker pull ghcr.io/<owner>/portfolios-backend:sha-abc123

   # Avoid - may change unexpectedly
   docker pull ghcr.io/<owner>/portfolios-backend:latest
   ```

2. **Keep credentials secure**:
   - Store PATs in environment variables or secrets managers
   - Never commit tokens to git
   - Rotate tokens periodically

3. **Use multi-stage builds**:
   - Keep images small
   - Separate build and runtime dependencies
   - Already implemented in the Dockerfile

4. **Leverage build cache**:
   - Cache is automatically managed in CI
   - For local builds, use `--cache-from` flag

5. **Tag semantically**:
   - Use semver for releases (v1.2.3)
   - Use branch names for development
   - Use commit SHAs for specific versions

## Additional Resources

- [GitHub Container Registry Documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Docker Documentation](https://docs.docker.com/)
- [GitHub Actions Docker Login](https://github.com/docker/login-action)
- [Docker Build Push Action](https://github.com/docker/build-push-action)

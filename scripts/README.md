# Scripts

This directory contains utility scripts for development and deployment.

## seed.go

Database seeding script for creating test users with known credentials.

### Usage

```bash
# From the project root directory
go run scripts/seed.go
```

### Prerequisites

- Database must be running and migrations must be completed
- `DATABASE_URL` environment variable must be set (or `.env` file present)

### Test Users

The script creates the following test users:

| Email | Password |
|-------|----------|
| test1@example.com | Test1234 |
| test2@example.com | Test5678 |
| admin@example.com | Admin123 |
| demo@example.com | Demo1234 |
| user@example.com | User1234 |

### Notes

- If users already exist, they will be skipped
- Passwords are hashed with bcrypt before storage
- All passwords meet the minimum requirements (8+ chars, uppercase, lowercase, number)
- This script is for **development and testing only** - do not run in production

### Safety

The script will warn you if the database already contains users and ask for confirmation before proceeding.

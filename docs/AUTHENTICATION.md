# Web Authentication System

## Overview

The SCMD web interface now includes a secure authentication system that requires users to log in with their email address and API key before accessing the application.

## Features

- **Session-based authentication**: Secure 24-hour sessions with HTTP-only cookies
- **Database-backed credentials**: Email and API key validation against PostgreSQL
- **Protected routes**: All web pages require authentication except login/logout
- **Automatic session cleanup**: Expired sessions are automatically removed
- **Logout functionality**: Users can securely log out from any page

## Setup

### 1. Create User Accounts

Generate an API key for a user using the command-line tool:

```bash
./scmd --create-api "user@example.com"
```

This will:
- Generate a random 32-character API key
- Store the email and API key in the `access` table (or custom table defined by `ACCESS_TB` env variable)
- Display the credentials
- Optionally update your `.env` file with the API key

### 2. Database Requirements

The authentication system uses the `access` table in your PostgreSQL database. This table should have the following structure:

```sql
CREATE TABLE access (
    email VARCHAR(255) PRIMARY KEY,
    api_key VARCHAR(64) NOT NULL
);
```

The table name can be customized using the `ACCESS_TB` environment variable in your `.env` file.

### 3. Environment Configuration

Ensure your `.env` file contains the database connection details:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASS=your_db_password
DB_NAME=your_db_name
ACCESS_TB=access  # Optional: custom access table name
```

## Usage

### Starting the Web Server

Start the web server as usual:

```bash
# HTTP mode
./scmd --web

# Custom port
./scmd --web -port 8080

# HTTPS mode
./scmd --web-ssl cert.crt cert.key
```

### Logging In

1. Navigate to the web interface (e.g., `http://localhost:3333`)
2. You will be automatically redirected to the login page
3. Enter your email address and API key
4. Click "Login"

Upon successful authentication:
- A secure session cookie is created
- You are redirected to the home page
- The session remains active for 24 hours

### Logging Out

Click the "Logout" link in the navigation bar on any page. This will:
- Destroy your session
- Clear the session cookie
- Redirect you to the login page

## Security Features

### Session Management

- **Session Duration**: 24 hours from creation
- **Session Storage**: In-memory session store (sessions are lost on server restart)
- **Session ID**: Cryptographically secure random 32-byte identifier
- **Cookie Security**: HTTP-only cookies prevent XSS attacks

### Password Security

- API keys are stored in the database (consider hashing in production)
- Session cookies are HTTP-only and use SameSite protection
- Failed login attempts are logged for security monitoring

### HTTPS Support

For production deployments, enable HTTPS to encrypt credentials in transit:

```bash
./scmd --web-ssl /path/to/cert.crt /path/to/cert.key
```

When using HTTPS, update `auth.go` to set `Secure: true` in the cookie configuration.

## User Management

### Creating Multiple Users

Create API keys for multiple users:

```bash
./scmd --create-api "admin@example.com"
./scmd --create-api "user1@example.com"
./scmd --create-api "user2@example.com"
```

### Updating API Keys

To update a user's API key, simply run the create command again with the same email:

```bash
./scmd --create-api "user@example.com"
```

This will generate a new API key and update the database.

### Removing Users

To remove a user's access, delete their record from the database:

```sql
DELETE FROM access WHERE email = 'user@example.com';
```

## Troubleshooting

### "Authentication Failed" Error

- Verify the email and API key are correct
- Check that the user exists in the `access` table
- Ensure the database connection is working

### Redirected to Login After Logging In

- Check that cookies are enabled in your browser
- Verify the session store is working (check server logs)
- Ensure the server time is correct (sessions expire based on server time)

### Session Expires Too Quickly

- Sessions last 24 hours by default
- Modify the `ExpiresAt` calculation in `auth.go` to change duration:
  ```go
  ExpiresAt: time.Now().Add(48 * time.Hour), // 48 hour session
  ```

## Production Recommendations

1. **Use HTTPS**: Always use HTTPS in production to encrypt credentials
2. **Hash API Keys**: Consider hashing API keys in the database
3. **Persistent Sessions**: Implement database-backed session storage for server restarts
4. **Rate Limiting**: Add rate limiting to prevent brute-force attacks
5. **Audit Logging**: Log all authentication events for security monitoring
6. **Session Timeout**: Consider adding idle timeout in addition to absolute expiration

## API Reference

### Authentication Functions

#### `AuthenticateUser(email, apiKey string) (bool, error)`
Validates email and API key against the database.

#### `CreateSession(email string) (string, error)`
Creates a new session and returns the session ID.

#### `GetSession(sessionID string) (*Session, bool)`
Retrieves a session by ID, returns nil if expired or not found.

#### `DeleteSession(sessionID string)`
Removes a session from the store.

#### `RequireAuth(next http.HandlerFunc) http.HandlerFunc`
Middleware that protects routes, redirects to login if not authenticated.

## Files Modified

- `auth.go` - New file containing authentication logic
- `server.go` - Updated with login/logout handlers and protected routes
- `templates/login.html` - New login page template
- `templates/*.html` - Updated navigation bars with logout button

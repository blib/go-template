# Backend

A Go-based HTTPS server built with Gin web framework, providing secure API endpoints with proper logging and configuration management.

## Features

- HTTPS server with TLS/SSL support
- Health check endpoint (`/healtz`)
- Structured logging with Zap
- Configuration management with Viper
- Graceful shutdown handling
- Command-line interface with Cobra

## Quick Start

### Prerequisites

- Go 1.24.1 or higher
- OpenSSL (for certificate generation)

### Installation

1. Clone the repository
2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Generate SSL certificates for development:

   ```bash
   ./generate-certs.sh
   ```

4. Build the application:
   ```bash
   go build -o backend .
   ```

### Running the Server

Start the HTTPS server with default settings:

```bash
./backend serve
```

The server will start on `https://0.0.0.0:8443` by default.

### Configuration Options

Backend supports multiple configuration methods in order of precedence:

1. Command-line flags (highest priority)
2. Environment variables
3. Configuration file (YAML)
4. Default values (lowest priority)

#### Configuration File

Create a `config.yaml` file for persistent configuration:

```yaml
# Basic configuration
env: "dev"
log_level: "info"

# Server settings
server:
  host: "0.0.0.0"
  port: 8443
  cert: "cert.pem"
  key: "key.pem"
  mode: "release"
  read_timeout: "15s"
  write_timeout: "15s"
  trusted_proxies: []

# CORS settings
cors:
  allow_origins: ["*"]
  allow_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  allow_headers:
    [
      "Origin",
      "Content-Length",
      "Content-Type",
      "Authorization",
      "X-Requested-With",
    ]
  expose_headers: ["Content-Length"]
  allow_credentials: false
  max_age: "12h"

# Optional error reporting
bugsnag:
  api_key: ""
  release_stage: "dev"
  log_level: "info"
```

Use the configuration file:

```bash
./backend serve --config config.yaml
```

#### Environment Variables

All configuration options can be set via environment variables using uppercase names with underscores:

```bash
export SERVER_HOST="127.0.0.1"
export SERVER_PORT="9443"
export LOG_LEVEL="debug"
export CORS_ALLOW_ORIGINS="https://example.com,https://app.example.com"
export SERVER_TRUSTED_PROXIES="10.0.0.1,192.168.1.100"
./backend serve
```

#### Configuration Reference

**Core Settings:**

- `env`: Environment - dev, staging, production (default: "dev")
- `log_level`: Log level - debug, info, warn, error (default: "info")

**Server Settings:**

- `server.host`: Server host address (default: "0.0.0.0")
- `server.port`: Server port (default: 8443)
- `server.cert`: Path to SSL certificate file (default: "cert.pem")
- `server.key`: Path to SSL private key file (default: "key.pem")
- `server.mode`: Gin server mode - debug, release, test (default: "release")
- `server.read_timeout`: HTTP read timeout (default: "15s")
- `server.write_timeout`: HTTP write timeout (default: "15s")
- `server.trusted_proxies`: List of trusted proxy IP addresses (default: [])

**CORS Settings:**

- `cors.allow_origins`: Allowed origins (default: ["*"])
- `cors.allow_methods`: Allowed HTTP methods (default: ["GET", "POST", "PUT", "DELETE", "OPTIONS"])
- `cors.allow_headers`: Allowed headers (default: ["Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With"])
- `cors.expose_headers`: Exposed headers (default: ["Content-Length"])
- `cors.allow_credentials`: Allow credentials (default: false)
- `cors.max_age`: Max age for preflight requests (default: "12h")

**Error Reporting (Bugsnag):**

- `bugsnag.api_key`: Bugsnag API key (optional)
- `bugsnag.release_stage`: Release stage for error reporting (default: "dev")
- `bugsnag.log_level`: Minimum log level for Bugsnag reporting (default: "info")

#### Production Configuration Example

For production environments, create a secure configuration:

```yaml
env: "production"
log_level: "warn"

server:
  host: "0.0.0.0"
  port: 8443
  cert: "/etc/ssl/certs/backend.pem"
  key: "/etc/ssl/private/backend.key"
  mode: "release"
  trusted_proxies:
    - "10.0.0.1"
    - "192.168.1.100"

cors:
  allow_origins:
    - "https://yourdomain.com"
    - "https://app.yourdomain.com"
  allow_credentials: true

bugsnag:
  api_key: "your-bugsnag-api-key"
  release_stage: "production"
  log_level: "error"
```

### API Endpoints

#### Health Check

- **GET** `/healtz`
- Returns server health status and timestamp
- Example response:
  ```json
  {
    "status": "healthy",
    "timestamp": "2024-01-15T10:30:00Z",
    "service": "backend"
  }
  ```

### Testing

Test the health endpoint:

```bash
curl -k https://localhost:8443/healtz
```

### SSL Certificates

For development, use the provided script to generate self-signed certificates:

```bash
./generate-certs.sh
```

For production, replace `cert.pem` and `key.pem` with certificates from a trusted Certificate Authority.

## Architecture

- **Gin**: High-performance HTTP web framework
- **Zap**: Structured, leveled logging
- **Viper**: Configuration management
- **Cobra**: Command-line interface
- **Fx**: Dependency injection framework

## Graceful Shutdown

The server supports graceful shutdown on receiving SIGINT or SIGTERM signals, allowing active connections to complete before terminating.

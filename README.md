# TeamCity MCP Server

A comprehensive Model Context Protocol (MCP) server that exposes JetBrains TeamCity as structured AI-ready resources and tools for LLM agents and IDE plugins.

## ✅ Quick Start

### 1. Build the Server

```bash
make build
# This creates ./bin/teamcity-mcp and a symlink ./server
```

### 2. Set Environment Variables

```bash
# Required
export TC_URL="https://your-teamcity-server.com"

# Optional (enables HMAC authentication)
export SERVER_SECRET="your-hmac-secret-key"

# Authentication (choose one method)
export TC_TOKEN="your-teamcity-api-token"        # Preferred
# OR
export TC_USERNAME="your-username"               # Fallback
export TC_PASSWORD="your-password"               # Fallback
```

### 3. Run the Server

```bash
./server
# Server starts on :8123 by default
```

### 4. Test the Server

```bash
# Health check
curl http://localhost:8123/healthz

# MCP protocol test
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-hmac-secret-key" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{}}}'
```

**Expected result**: Health endpoint should return `{"status":"ok"}` and MCP endpoint should return initialization response.

## Features

- **MCP Protocol Compliance**: Full JSON-RPC 2.0 over HTTP/WebSocket support
- **TeamCity Integration**: Complete REST API integration with authentication
- **Resource Access**: Projects, build types, builds, agents, and artifacts
- **Build Operations**: Trigger, cancel, pin builds, set tags, download artifacts, search builds
- **Advanced Search**: Comprehensive build search with multiple filters (status, branch, user, dates, tags)
- **Production Ready**: Docker, Kubernetes, monitoring, caching, and comprehensive logging
- **Environment-Based Configuration**: No config files needed, everything via environment variables

## Environment Variables Reference

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `TC_URL` | TeamCity server URL | `https://teamcity.company.com` |
| `SERVER_SECRET` | HMAC secret for client authentication (optional) | `my-secure-secret-123` |

### Authentication Variables (Choose One Method)

**Method 1: API Token (Recommended)**
| Variable | Description | Example |
|----------|-------------|---------|
| `TC_TOKEN` | TeamCity API token | `eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...` |

**Method 2: Basic Authentication (Fallback)**
| Variable | Description | Example |
|----------|-------------|---------|
| `TC_USERNAME` | TeamCity username | `admin` |
| `TC_PASSWORD` | TeamCity password | `password123` |

### Optional Variables

| Variable | Default | Description | Example |
|----------|---------|-------------|---------|
| `LISTEN_ADDR` | `:8123` | Server listen address | `:8080` or `0.0.0.0:8123` |
| `TC_TIMEOUT` | `30s` | TeamCity API timeout | `60s` or `2m` |
| `TLS_CERT` | | Path to TLS certificate | `/path/to/cert.pem` |
| `TLS_KEY` | | Path to TLS private key | `/path/to/key.pem` |
| `LOG_LEVEL` | `info` | Log level | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `json` | Log format | `json` or `console` |
| `CACHE_TTL` | `10s` | Cache TTL for API responses | `30s` or `1m` |

## Configuration Examples

### Development Environment

```bash
export TC_URL=http://localhost:8111
export TC_TOKEN=dev-token-123
export SERVER_SECRET=dev-secret
export LOG_LEVEL=debug
export LOG_FORMAT=console
./server
```

### Production Environment

```bash
export TC_URL=https://teamcity.company.com
export TC_TOKEN=$TEAMCITY_API_TOKEN
export SERVER_SECRET=$MCP_SERVER_SECRET
export TLS_CERT=/etc/ssl/certs/teamcity-mcp.pem
export TLS_KEY=/etc/ssl/private/teamcity-mcp.key
export LOG_LEVEL=warn
export CACHE_TTL=30s
./server
```

## Docker Deployment

### Build and Run

```bash
# Build Docker image
make docker-build

# Run with environment variables
docker run -p 8123:8123 \
  -e TC_URL=https://teamcity.company.com \
  -e TC_TOKEN=your-token \
  -e SERVER_SECRET=your-secret \
  teamcity-mcp:latest
```

### Docker Compose

```bash
# Start with docker-compose
docker-compose up -d

# Check logs
docker-compose logs -f teamcity-mcp
```

## Kubernetes Deployment

### Using Helm

```bash
# Deploy with Helm
helm install teamcity-mcp ./helm/teamcity-mcp \
  --set teamcity.url=https://teamcity.company.com \
  --set secrets.teamcityToken=your-token \
  --set secrets.serverSecret=your-secret
```

### Manual Kubernetes Deployment

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: teamcity-mcp-secrets
type: Opaque
stringData:
  teamcity-token: "your-teamcity-token"
  server-secret: "your-server-secret"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: teamcity-mcp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: teamcity-mcp
  template:
    metadata:
      labels:
        app: teamcity-mcp
    spec:
      containers:
      - name: teamcity-mcp
        image: teamcity-mcp:latest
        ports:
        - containerPort: 8123
        env:
        - name: TC_URL
          value: "https://teamcity.company.com"
        - name: TC_TOKEN
          valueFrom:
            secretKeyRef:
              name: teamcity-mcp-secrets
              key: teamcity-token
        - name: SERVER_SECRET
          valueFrom:
            secretKeyRef:
              name: teamcity-mcp-secrets
              key: server-secret
```

## Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `--help` | Show environment variable help | |
| `--version` | Show version information | |
| `--transport` | Transport mode: http or stdio | `http` |

### Help and Documentation

```bash
# Show environment variable help
./server --help

# Show version
./server --version

# Show command line usage
./server -h
```

## Testing and Verification

### Automated Verification

Use the included verification script to test all functionality:

```bash
# Run all tests
./scripts/verify.sh

# Available options:
./scripts/verify.sh help     # Show help
./scripts/verify.sh start    # Start server only
./scripts/verify.sh stop     # Stop server only
./scripts/verify.sh clean    # Clean up processes
```

### Manual Testing

```bash
# 1. Set environment variables
export TC_URL=http://localhost:8111
export TC_TOKEN=test-token
export SERVER_SECRET=test-secret

# 2. Start server
./server &

# 3. Test health
curl http://localhost:8123/healthz

# 4. Test MCP protocol
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-secret" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{}}}'

# 5. Stop server
pkill -f teamcity-mcp
```

### Development Testing

```bash
# Install dependencies
go mod download

# Run unit tests
make test

# Run linter
make lint

# Build for multiple platforms
make build-all
```

## MCP Protocol Testing

### Initialize MCP Session

```bash
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-03-26",
      "capabilities": {},
      "clientInfo": {
        "name": "test-client",
        "version": "1.0.0"
      }
    }
  }'
```

### List Resources

```bash
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "resources/list",
    "params": {}
  }'
```

### List Tools

```bash
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/list",
    "params": {}
  }'
```

## Available Tools

The TeamCity MCP server provides 6 powerful tools for managing builds:

### 1. trigger_build
Trigger a new build in TeamCity.

**Parameters:**
- `buildTypeId` (required): Build configuration ID
- `branchName` (optional): Branch name to build
- `properties` (optional): Build properties object

**Example:**
```bash
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "trigger_build",
      "arguments": {
        "buildTypeId": "YourProject_BuildConfiguration",
        "branchName": "main",
        "properties": {
          "env.DEPLOY_ENV": "staging"
        }
      }
    }
  }'
```

### 2. cancel_build
Cancel a running build.

**Parameters:**
- `buildId` (required): Build ID to cancel
- `comment` (optional): Cancellation comment

**Example:**
```bash
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret" \
  -d '{
    "jsonrpc": "2.0",
    "id": 5,
    "method": "tools/call",
    "params": {
      "name": "cancel_build",
      "arguments": {
        "buildId": "12345",
        "comment": "Cancelled due to urgent hotfix"
      }
    }
  }'
```

### 3. pin_build
Pin or unpin a build to prevent it from being cleaned up.

**Parameters:**
- `buildId` (required): Build ID to pin/unpin
- `pin` (required): true to pin, false to unpin
- `comment` (optional): Pin comment

**Example:**
```bash
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret" \
  -d '{
    "jsonrpc": "2.0",
    "id": 6,
    "method": "tools/call",
    "params": {
      "name": "pin_build",
      "arguments": {
        "buildId": "12345",
        "pin": true,
        "comment": "Release candidate build"
      }
    }
  }'
```

### 4. set_build_tag
Add or remove tags from a build.

**Parameters:**
- `buildId` (required): Build ID
- `tags` (optional): Array of tags to add
- `removeTags` (optional): Array of tags to remove

**Example:**
```bash
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret" \
  -d '{
    "jsonrpc": "2.0",
    "id": 7,
    "method": "tools/call",
    "params": {
      "name": "set_build_tag",
      "arguments": {
        "buildId": "12345",
        "tags": ["release", "v1.2.3"],
        "removeTags": ["beta"]
      }
    }
  }'
```

### 5. download_artifact
Download build artifacts.

**Parameters:**
- `buildId` (required): Build ID
- `artifactPath` (required): Path to the artifact

**Example:**
```bash
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret" \
  -d '{
    "jsonrpc": "2.0",
    "id": 8,
    "method": "tools/call",
    "params": {
      "name": "download_artifact",
      "arguments": {
        "buildId": "12345",
        "artifactPath": "dist/app.zip"
      }
    }
  }'
```

### 6. search_builds
Search for builds with comprehensive filtering options.

**Parameters (all optional):**
- `buildTypeId`: Filter by build configuration ID
- `status`: Filter by build status (SUCCESS, FAILURE, ERROR, UNKNOWN)
- `state`: Filter by build state (queued, running, finished)
- `branch`: Filter by branch name
- `agent`: Filter by agent name
- `user`: Filter by user who triggered the build
- `sinceBuild`: Search builds since this build ID
- `sinceDate`: Search builds since this date (YYYYMMDDTHHMMSS+HHMM)
- `untilDate`: Search builds until this date (YYYYMMDDTHHMMSS+HHMM)
- `tags`: Array of tags to filter by
- `personal`: Include personal builds (boolean)
- `pinned`: Filter by pinned status (boolean)
- `count`: Maximum number of builds to return (1-1000, default: 100)

**Examples:**

Search for failed builds:
```bash
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret" \
  -d '{
    "jsonrpc": "2.0",
    "id": 9,
    "method": "tools/call",
    "params": {
      "name": "search_builds",
      "arguments": {
        "status": "FAILURE",
        "count": 10
      }
    }
  }'
```

Search for recent builds on main branch:
```bash
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret" \
  -d '{
    "jsonrpc": "2.0",
    "id": 10,
    "method": "tools/call",
    "params": {
      "name": "search_builds",
      "arguments": {
        "branch": "main",
        "state": "finished",
        "count": 20
      }
    }
  }'
```

Search for builds with specific tags:
```bash
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret" \
  -d '{
    "jsonrpc": "2.0",
    "id": 11,
    "method": "tools/call",
    "params": {
      "name": "search_builds",
      "arguments": {
        "tags": ["release", "production"],
        "pinned": true
      }
    }
  }'
```

## IDE Integration (Cursor)

The TeamCity MCP server is designed to work seamlessly with AI-powered IDEs like Cursor. Here's how to configure it:

### Cursor Configuration

Add this to your Cursor MCP settings:

```json
{
  "teamcity": {
    "command": "docker",
    "args": [
      "run",
      "--rm",
      "-i",
      "-e",
      "TC_URL",
      "-e",
      "TC_TOKEN",
      "teamcity-mcp:latest",
      "--transport",
      "stdio"
    ],
    "env": {
      "TC_URL": "https://your-teamcity-server.com",
      "TC_TOKEN": "your-teamcity-api-token"
    }
  }
}
```

### Local Binary Configuration

If you prefer to use the local binary instead of Docker:

```json
{
  "teamcity": {
    "command": "/path/to/teamcity-mcp",
    "args": ["--transport", "stdio"],
    "env": {
      "TC_URL": "https://your-teamcity-server.com",
      "TC_TOKEN": "your-teamcity-api-token"
    }
  }
}
```

### Usage in Cursor

Once configured, you can use natural language commands like:

- **"Search for failed builds in the last week"**
- **"Trigger a build for the main branch"**
- **"Show me recent builds for project X"**
- **"Pin the latest successful build"**
- **"Cancel the running build 12345"**
- **"Add a release tag to build 12345"**

The AI will automatically use the appropriate TeamCity tools to fulfill your requests.

## Available Resources

The server exposes TeamCity data as MCP resources:

- **`teamcity://projects`** - List all projects
- **`teamcity://buildTypes`** - List all build configurations
- **`teamcity://builds`** - List recent builds
- **`teamcity://agents`** - List build agents

## Troubleshooting

### Common Issues

1. **Missing required environment variables**
   ```
   Error: TC_URL environment variable is required
   ```
   **Solution**: Set all required environment variables

2. **Authentication failures**
   ```
   Error: either TC_TOKEN or both TC_USERNAME and TC_PASSWORD environment variables are required
   ```
   **Solution**: Set either `TC_TOKEN` or both `TC_USERNAME` and `TC_PASSWORD`

3. **Invalid timeout format**
   ```
   Error: invalid TC_TIMEOUT format
   ```
   **Solution**: Use valid duration format like `30s`, `1m`, `2h`

4. **Port already in use**
   ```
   Error: listen tcp :8123: bind: address already in use
   ```
   **Solution**: Set `LISTEN_ADDR` to a different port or stop the conflicting service

### Debug Mode

Enable debug logging:

```bash
export LOG_LEVEL=debug
export LOG_FORMAT=console
./server
```

### Health Check

The server provides a health endpoint:

```bash
curl http://localhost:8123/healthz
# Expected: {"service":"teamcity-mcp","status":"ok","timestamp":"..."}
```

### Metrics

Prometheus metrics are available:

```bash
curl http://localhost:8123/metrics
```

### TeamCity Integration Testing

Verify TeamCity connectivity:

```bash
# Check TeamCity server accessibility
curl -H "Authorization: Bearer your-token" \
  http://your-teamcity-url/app/rest/projects

# Verify authentication
curl -H "Authorization: Bearer your-token" \
  http://your-teamcity-url/app/rest/server
```

## Security Best Practices

1. **Never hardcode secrets in scripts or Dockerfiles**
2. **Use secret management systems in production**
3. **Rotate API tokens regularly**
4. **Use TLS in production environments**
5. **Restrict network access to the MCP server**

### Secret Management Examples

```bash
# AWS Secrets Manager
export TC_TOKEN=$(aws secretsmanager get-secret-value --secret-id teamcity/api-token --query SecretString --output text)
export SERVER_SECRET=$(aws secretsmanager get-secret-value --secret-id mcp/server-secret --query SecretString --output text)

# HashiCorp Vault
export TC_TOKEN=$(vault kv get -field=token secret/teamcity)
export SERVER_SECRET=$(vault kv get -field=secret secret/mcp)

# Kubernetes secrets (automatically mounted)
export TC_TOKEN=$(cat /var/secrets/teamcity-token)
export SERVER_SECRET=$(cat /var/secrets/server-secret)
```

## Expected Behaviors

### ✅ Healthy Server
- Health endpoint returns `{"status":"ok"}`
- Metrics endpoint returns Prometheus format
- Server logs show successful startup

### ✅ MCP Protocol
- Initialize returns protocol version `2025-03-26`
- Resources list includes projects, buildTypes, builds, agents
- Tools list includes all 5 build management tools

### ✅ TeamCity Integration
- Can fetch projects, build types, builds, agents
- Authentication works with token/basic auth
- Build operations execute successfully

### ✅ Performance
- Health check responds in <100ms
- Resource queries cached for 10s TTL
- Concurrent requests handled properly

### ❌ Error Indicators

**Server Issues**:
- Health endpoint returns 500 or connection refused
- Logs show startup errors or panics
- Metrics endpoint not accessible

**MCP Protocol Issues**:
- JSON-RPC errors in responses
- Missing required fields in protocol messages
- Unsupported protocol version

**TeamCity Issues**:
- Authentication failures (401/403)
- Network connectivity problems
- Invalid TeamCity URL or API responses

**Configuration Issues**:
- Missing required environment variables (TC_URL)
- Invalid environment variable formats (timeouts, durations)

## Protocol Reference

See [Protocol.md](Protocol.md) for detailed MCP protocol implementation and TeamCity API mapping.

## License

MIT License - see [LICENSE](LICENSE) for details. 
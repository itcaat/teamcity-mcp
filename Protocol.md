# TeamCity MCP Protocol Documentation

This document describes how the TeamCity MCP server maps Model Context Protocol (MCP) resources and tools to TeamCity REST API endpoints.

## Protocol Version

The TeamCity MCP server implements MCP protocol version `2024-11-05`.

## Resources

Resources are read-only entities that provide structured access to TeamCity data.

### Projects

**MCP URI**: `teamcity://projects`

**TeamCity Endpoint**: `GET /app/rest/projects`

**Description**: Lists all projects accessible to the authenticated user.

**Example Response**:
```json
{
  "resources": [
    {
      "uri": "teamcity://projects/MyProject",
      "name": "My Project",
      "description": "Sample project description",
      "mimeType": "application/json"
    }
  ]
}
```

**Locator Examples**:
- All projects: `teamcity://projects`
- Specific project: `teamcity://projects/MyProject`
- Projects by name: `teamcity://projects?locator=name:MyProject`

### Build Types

**MCP URI**: `teamcity://buildTypes`

**TeamCity Endpoint**: `GET /app/rest/buildTypes`

**Description**: Lists all build configurations.

**Example Response**:
```json
{
  "resources": [
    {
      "uri": "teamcity://buildTypes/MyProject_Build",
      "name": "Build Configuration",
      "description": "Build configuration for My Project",
      "mimeType": "application/json"
    }
  ]
}
```

**Locator Examples**:
- All build types: `teamcity://buildTypes`
- By project: `teamcity://buildTypes?locator=project:MyProject`
- By name: `teamcity://buildTypes?locator=name:Build`

### Builds

**MCP URI**: `teamcity://builds`

**TeamCity Endpoint**: `GET /app/rest/builds`

**Description**: Lists build instances with their status and results.

**Example Response**:
```json
{
  "resources": [
    {
      "uri": "teamcity://builds/12345",
      "name": "Build #123",
      "description": "Status: SUCCESS",
      "mimeType": "application/json"
    }
  ]
}
```

**Locator Examples**:
- Recent builds: `teamcity://builds?locator=count:10`
- By build type: `teamcity://builds?locator=buildType:MyProject_Build`
- By status: `teamcity://builds?locator=status:SUCCESS`
- By branch: `teamcity://builds?locator=branch:main`

### Agents

**MCP URI**: `teamcity://agents`

**TeamCity Endpoint**: `GET /app/rest/agents`

**Description**: Lists build agents and their capabilities.

**Example Response**:
```json
{
  "resources": [
    {
      "uri": "teamcity://agents/1",
      "name": "Agent-01",
      "description": "Connected: true",
      "mimeType": "application/json"
    }
  ]
}
```

**Locator Examples**:
- All agents: `teamcity://agents`
- Connected only: `teamcity://agents?locator=connected:true`
- By name: `teamcity://agents?locator=name:Agent-01`

### Artifacts

**MCP URI**: `teamcity://artifacts`

**TeamCity Endpoint**: `GET /app/rest/builds/{buildId}/artifacts`

**Description**: Lists build artifacts for a specific build.

**Example**: `teamcity://artifacts?locator=build:12345`

## Tools

Tools provide write operations and actions on TeamCity entities.

### trigger_build

**Description**: Triggers a new build for a specified build configuration.

**TeamCity Endpoint**: `POST /app/rest/buildQueue`

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "buildTypeId": {
      "type": "string",
      "description": "Build configuration ID (required)"
    },
    "branchName": {
      "type": "string",
      "description": "Branch name (optional, defaults to default branch)"
    },
    "properties": {
      "type": "object",
      "description": "Build properties as key-value pairs (optional)"
    },
    "comment": {
      "type": "string",
      "description": "Build trigger comment (optional)"
    }
  },
  "required": ["buildTypeId"]
}
```

**Example Usage**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "trigger_build",
    "arguments": {
      "buildTypeId": "MyProject_Build",
      "branchName": "feature/new-feature",
      "properties": {
        "env.VERSION": "1.0.0",
        "system.deployment.target": "staging"
      },
      "comment": "Triggered via MCP"
    }
  }
}
```

### cancel_build

**Description**: Cancels a running or queued build.

**TeamCity Endpoint**: `POST /app/rest/builds/{buildId}/cancelRequest`

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "buildId": {
      "type": "string",
      "description": "Build ID to cancel (required)"
    },
    "comment": {
      "type": "string",
      "description": "Cancellation reason (optional)"
    },
    "readdToQueue": {
      "type": "boolean",
      "description": "Re-add to queue after cancellation (optional)"
    }
  },
  "required": ["buildId"]
}
```

### pin_build

**Description**: Pins or unpins a build to prevent it from being cleaned up.

**TeamCity Endpoint**: `PUT /app/rest/builds/{buildId}/pin`

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "buildId": {
      "type": "string",
      "description": "Build ID to pin/unpin (required)"
    },
    "pin": {
      "type": "boolean",
      "description": "True to pin, false to unpin (required)"
    },
    "comment": {
      "type": "string",
      "description": "Pin comment (optional)"
    }
  },
  "required": ["buildId", "pin"]
}
```

### set_build_tag

**Description**: Adds or removes tags from a build.

**TeamCity Endpoint**: 
- Add: `POST /app/rest/builds/{buildId}/tags`
- Remove: `DELETE /app/rest/builds/{buildId}/tags/{tag}`

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "buildId": {
      "type": "string",
      "description": "Build ID (required)"
    },
    "tags": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Tags to add (optional)"
    },
    "removeTags": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Tags to remove (optional)"
    }
  },
  "required": ["buildId"]
}
```

### download_artifact

**Description**: Downloads build artifacts.

**TeamCity Endpoint**: `GET /app/rest/builds/{buildId}/artifacts/content/{path}`

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "buildId": {
      "type": "string",
      "description": "Build ID (required)"
    },
    "artifactPath": {
      "type": "string",
      "description": "Path to artifact (required)"
    },
    "outputPath": {
      "type": "string",
      "description": "Local output path (optional)"
    }
  },
  "required": ["buildId", "artifactPath"]
}
```

## Authentication

### Client to MCP Server

Uses HMAC-SHA256 signed Bearer tokens:

```http
Authorization: Bearer <hmac_token>
```

The token is generated using:
```
HMAC-SHA256(message="teamcity-mcp", secret=server_secret)
```

### MCP Server to TeamCity

Supports both token-based and basic authentication:

**Token Authentication** (Preferred):
```http
Authorization: Bearer <teamcity_token>
```

**Basic Authentication** (Fallback):
```http
Authorization: Basic <base64(username:password)>
```

## Error Handling

The server follows JSON-RPC 2.0 error response format:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32603,
    "message": "Internal error",
    "data": "Detailed error description"
  }
}
```

**Common Error Codes**:
- `-32700`: Parse error (invalid JSON)
- `-32600`: Invalid request (malformed JSON-RPC)
- `-32601`: Method not found
- `-32602`: Invalid params
- `-32603`: Internal error (TeamCity API error)

## Rate Limiting

The server respects TeamCity's rate limiting:

- Honors `X-RateLimit-*` headers from TeamCity
- Implements exponential backoff on 429 responses
- Caches responses for 10 seconds by default

## Locator Syntax

TeamCity locators are supported in resource URIs using query parameters:

```
teamcity://builds?locator=buildType:MyProject_Build,status:SUCCESS,count:10
```

**Common Locator Dimensions**:
- `id`: Entity ID
- `name`: Entity name
- `project`: Project ID/name
- `buildType`: Build configuration ID
- `status`: Build status (SUCCESS, FAILURE, etc.)
- `branch`: Branch name
- `count`: Maximum number of results
- `start`: Start index for pagination

## Examples

### Complete MCP Session

1. **Initialize**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {"name": "my-client", "version": "1.0.0"}
  }
}
```

2. **List Projects**:
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "resources/list",
  "params": {"uri": "teamcity://projects"}
}
```

3. **Trigger Build**:
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "trigger_build",
    "arguments": {
      "buildTypeId": "MyProject_Build",
      "branchName": "main"
    }
  }
}
```

This protocol mapping enables AI agents to interact with TeamCity through a standardized MCP interface while leveraging the full power of TeamCity's REST API. 
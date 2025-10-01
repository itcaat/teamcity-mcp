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

### Runtime Information

**MCP URI**: `teamcity://runtime`

**Description**: Provides current server date, time, and runtime information to ensure AI models use real current time instead of training data dates.

**Example Response**:
```json
{
  "type": "runtime-info",
  "currentTime": "2024-12-26T14:30:22+03:00",
  "currentDate": "2024-12-26",
  "currentTimestamp": 1735207822,
  "timezone": "Local",
  "utcTime": "2024-12-26T11:30:22Z",
  "utcDate": "2024-12-26",
  "utcTimestamp": 1735207822,
  "serverInfo": {
    "name": "teamcity-mcp",
    "version": "1.0.0"
  },
  "note": "This is the REAL current date and time. Do not use any training data dates. Use this information for all time-based queries and operations."
}
```

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

### fetch_build_log

**Description**: Fetches the complete build log for a specific build.

**TeamCity Endpoint**: `GET /downloadBuildLog.html?buildId={buildId}`

**Input Schema**:
```json
{
  "type": "object", 
  "properties": {
    "buildId": {
      "type": "string",
      "description": "Build ID to fetch log for (required)"
    },
    "plain": {
      "type": "boolean",
      "description": "Return log as plain text (optional, default: true)"
    },
    "archived": {
      "type": "boolean", 
      "description": "Return log as zip archive (optional, default: false)"
    },
    "dateFormat": {
      "type": "string",
      "description": "Custom timestamp format following Java SimpleDateFormat (optional)"
    }
  },
  "required": ["buildId"]
}
```

**Additional Parameters**:
- `plain=true`: Returns the log content as plain text in the browser/response body
- `archived=true`: Returns the log file as a `.zip` archive
- `dateFormat=<pattern>`: Customizes timestamp format (e.g., "yyyy-MM-dd HH:mm:ss")

**Example Usage**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "fetch_build_log",
    "arguments": {
      "buildId": "12345",
      "plain": true,
      "dateFormat": "yyyy-MM-dd HH:mm:ss"
    }
  }
}
```

### search_build_configurations

**Description**: Searches for build configurations with comprehensive filters including basic filters, parameters, steps, and VCS roots.

**TeamCity Endpoints**: 
- `GET /app/rest/buildTypes` (basic search)
- `GET /app/rest/buildTypes/id:{buildTypeId}/parameters` (parameters)
- `GET /app/rest/buildTypes/id:{buildTypeId}/steps` (build steps)
- `GET /app/rest/buildTypes/id:{buildTypeId}/vcs-root-entries` (VCS roots)

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "projectId": {
      "type": "string",
      "description": "Filter by project ID (optional)"
    },
    "name": {
      "type": "string", 
      "description": "Search by configuration name with partial matching (optional)"
    },
    "enabled": {
      "type": "boolean",
      "description": "Filter by enabled status (optional)"
    },
    "paused": {
      "type": "boolean",
      "description": "Filter by paused status (optional)"
    },
    "template": {
      "type": "boolean",
      "description": "Filter templates (true) or regular configurations (false) (optional)"
    },
    "parameterName": {
      "type": "string",
      "description": "Search by parameter name with partial matching (optional)"
    },
    "parameterValue": {
      "type": "string",
      "description": "Search by parameter value with partial matching (optional)"
    },
    "stepType": {
      "type": "string",
      "description": "Search by build step type (e.g., 'gradle', 'docker', 'powershell') (optional)"
    },
    "stepName": {
      "type": "string",
      "description": "Search by build step name with partial matching (optional)"
    },
    "vcsType": {
      "type": "string",
      "description": "Search by VCS type (e.g., 'git', 'subversion') (optional)"
    },
    "includeDetails": {
      "type": "boolean",
      "description": "Include detailed information (parameters, steps, VCS) in results (optional, default: false)"
    },
    "count": {
      "type": "integer",
      "description": "Maximum number of configurations to return (optional, default: 100)"
    }
  }
}
```

**Example Usage**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "search_build_configurations",
    "arguments": {
      "stepType": "gradle",
      "parameterName": "env.DEPLOY_TARGET",
      "parameterValue": "production",
      "includeDetails": true,
      "count": 50
    }
  }
}
```

### get_current_time

**Description**: Gets the current server date and time to ensure AI models use real current time instead of training data dates.

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "format": {
      "type": "string",
      "description": "Date format (rfc3339, date, timestamp, or custom Go format) (optional, default: rfc3339)"
    },
    "timezone": {
      "type": "string",
      "description": "Timezone (e.g., 'UTC', 'Local', 'America/New_York') (optional, default: Local)"
    }
  }
}
```

### get_test_results

**Description**: Get test results for a specific build with optional filtering by test status and detailed information.

**TeamCity Endpoint**: `GET /app/rest/testOccurrences?locator=build:(id:{buildId})[,status:{status}]&fields=testOccurrence(id,name,status,duration,href[,details])`

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "buildId": {
      "type": "string",
      "description": "Build ID to get test results for (required)"
    },
    "status": {
      "type": "string",
      "description": "Filter by test status: SUCCESS, FAILURE, UNKNOWN, IGNORED (optional)"
    },
    "includeDetails": {
      "type": "boolean",
      "description": "Include test details like stack traces (optional, default: false)"
    },
    "count": {
      "type": "integer",
      "description": "Maximum number of tests to return (optional, default: 100, max: 1000)"
    }
  },
  "required": ["buildId"]
}
```

**Example Usage**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "get_test_results",
    "arguments": {
      "buildId": "12345"
    }
  }
}
```

**Example with Filters**:
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "get_test_results",
    "arguments": {
      "buildId": "12345",
      "status": "FAILURE",
      "includeDetails": true,
      "count": 50
    }
  }
}
```

**Example Usage**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "get_current_time",
    "arguments": {
      "format": "rfc3339",
      "timezone": "UTC"
    }
  }
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

Uses TeamCity API token authentication:

```http
Authorization: Bearer <teamcity_api_token>
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
# Runtime Date & Time Examples

This document demonstrates how to use the new runtime date/time functionality in TeamCity MCP server to ensure AI models use real current dates instead of training data dates.

## Problem Solved

AI models sometimes use their training data cutoff date as "today" which can lead to incorrect results when working with TeamCity builds, logs, and time-based queries. This functionality provides several ways for AI models to access the real current date and time.

## Available Solutions

### 1. Server Info During Initialization

When the MCP session initializes, current date/time information is automatically included:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {}
  }
}
```

**Response includes:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {...},
    "serverInfo": {
      "name": "teamcity-mcp",
      "version": "1.0.0",
      "currentTime": "2024-12-26T14:30:22+03:00",
      "currentDate": "2024-12-26",
      "timezone": "Local"
    }
  }
}
```

### 2. Runtime Resource

Access real-time information via the `teamcity://runtime` resource:

```bash
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "resources/read",
    "params": {
      "uri": "teamcity://runtime"
    }
  }'
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "contents": [{
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
    }]
  }
}
```

### 3. get_current_time Tool

Get current time with flexible formatting options:

#### Basic Usage (RFC3339 format)
```bash
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "get_current_time",
      "arguments": {}
    }
  }'
```

#### Date Only Format
```bash
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "get_current_time",
      "arguments": {
        "format": "date",
        "timezone": "UTC"
      }
    }
  }'
```

#### Timestamp Format
```bash
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret" \
  -d '{
    "jsonrpc": "2.0",
    "id": 5,
    "method": "tools/call",
    "params": {
      "name": "get_current_time",
      "arguments": {
        "format": "timestamp",
        "timezone": "America/New_York"
      }
    }
  }'
```

#### Custom Format
```bash
curl -X POST http://localhost:8123/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret" \
  -d '{
    "jsonrpc": "2.0",
    "id": 6,
    "method": "tools/call",
    "params": {
      "name": "get_current_time",
      "arguments": {
        "format": "2006-01-02 15:04:05",
        "timezone": "UTC"
      }
    }
  }'
```

## Usage in Cursor IDE

Once configured, you can use natural language with AI models:

### Examples:

- **"What's today's date?"** - AI will use `get_current_time` tool
- **"Show me builds from the last 24 hours"** - AI will get current time and calculate the correct date range
- **"Find builds that failed yesterday"** - AI will determine the correct date for "yesterday"
- **"List builds from this week"** - AI will calculate the current week boundaries

### Example Conversation:

**User:** "Show me all failed builds from yesterday"

**AI Response:** "Let me first get the current date to determine what 'yesterday' means, then search for failed builds."

*AI calls `get_current_time` tool*
*AI gets "2024-12-26" as today's date*
*AI calculates yesterday as "2024-12-25"*
*AI calls `search_builds` with appropriate date filters*

## Available Formats

The `get_current_time` tool supports these format options:

| Format | Description | Example Output |
|--------|-------------|----------------|
| `rfc3339` | RFC3339 timestamp (default) | `2024-12-26T14:30:22+03:00` |
| `date` | Date only | `2024-12-26` |
| `timestamp` | Unix timestamp | `1735207822` |
| Custom | Go time format | `2006-01-02 15:04:05` → `2024-12-26 14:30:22` |

## Available Timezones

| Timezone | Description |
|----------|-------------|
| `Local` | Server's local timezone (default) |
| `UTC` | Coordinated Universal Time |
| `America/New_York` | US Eastern Time |
| `Europe/London` | UK Time |
| `Asia/Tokyo` | Japan Time |
| Any valid IANA timezone | See [IANA Time Zone Database](https://www.iana.org/time-zones) |

## Integration Tips

1. **Always use current time tools** when dealing with relative dates like "today", "yesterday", "this week"
2. **Check serverInfo** during initialization to get baseline time information
3. **Use the runtime resource** for comprehensive time information including both local and UTC times
4. **Prefer UTC timezone** for consistent cross-timezone operations
5. **Use appropriate formats** - date format for day-based queries, timestamps for precise comparisons

## Benefits

- ✅ **Accurate date calculations** - AI models get real current dates
- ✅ **Timezone awareness** - Support for multiple timezones
- ✅ **Flexible formatting** - Multiple output formats for different use cases
- ✅ **Explicit messaging** - Clear notes that this is real current time
- ✅ **Multiple access methods** - Resource, tool, and initialization data
- ✅ **Production ready** - Integrated into the existing MCP protocol



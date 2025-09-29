# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Runtime Date/Time Support**: Added comprehensive current date/time functionality to prevent AI models from using training data dates
  - New `teamcity://runtime` resource providing current server date, time, and timezone information
  - New `get_current_time` tool with flexible formatting and timezone support
  - Current time information included in server initialization response
  - Support for multiple date formats: RFC3339, date-only, timestamp, and custom Go formats
  - Support for multiple timezones including UTC, Local, and IANA timezone names
  - Comprehensive documentation with examples in `docs/RUNTIME_DATE_EXAMPLES.md`
  - Unit tests covering all new functionality

### Changed
- Updated Protocol.md with documentation for new runtime resource and get_current_time tool
- Updated README.md to include new tool in the count (9 tools total) and usage examples
- Enhanced server initialization to include current time information in serverInfo

### Technical Details
- Added `listRuntimeInfo()` and `getRuntimeInfo()` methods to MCP handler
- Added `getCurrentTime()` tool implementation with timezone and format support
- Enhanced `handleInitialize()` to include current time in serverInfo
- Added comprehensive test suite in `tests/unit/runtime_test.go`

## [1.0.0] - Previous Release

### Added
- Initial release of TeamCity MCP Server
- Full MCP protocol support with JSON-RPC 2.0
- 8 powerful tools for TeamCity management:
  - `trigger_build` - Trigger new builds
  - `cancel_build` - Cancel running builds  
  - `pin_build` - Pin/unpin builds
  - `set_build_tag` - Add/remove build tags
  - `download_artifact` - Download build artifacts
  - `search_builds` - Advanced build search with filters
  - `fetch_build_log` - Get build logs (plain text or archived)
  - `search_build_configurations` - Search build configurations with detailed filters
- Resource access for projects, build types, builds, and agents
- HTTP/WebSocket and STDIO transport support
- HMAC authentication support
- Production-ready features:
  - Docker and Kubernetes deployment
  - Prometheus metrics
  - Health checks
  - Structured logging
  - Caching with configurable TTL
  - TLS support
- Comprehensive documentation and examples
- Full test coverage
- CI/CD pipeline with GitHub Actions
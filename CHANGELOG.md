# Changelog

All notable changes to the TeamCity MCP Server project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial implementation of TeamCity MCP Server
- MCP protocol version 2024-11-05 support
- HTTP and STDIO transport modes
- WebSocket upgrade support
- HMAC-based client authentication
- TeamCity token and basic authentication support
- Resource endpoints for projects, buildTypes, builds, and agents
- Tool implementations for build operations
- In-memory caching with configurable TTL
- Prometheus metrics integration
- Health check endpoints (liveness and readiness)
- Structured logging with Zap
- Configuration via YAML files and environment variables
- Docker containerization with multi-stage builds
- Docker Compose setup with TeamCity server
- Comprehensive test suite (unit and integration)
- Makefile for build automation
- Documentation for protocol mapping and API usage

### Resources
- `teamcity://projects` - List and access TeamCity projects
- `teamcity://buildTypes` - List and access build configurations
- `teamcity://builds` - List and access build instances
- `teamcity://agents` - List and access build agents

### Tools
- `trigger_build` - Trigger new builds with parameters
- `cancel_build` - Cancel running or queued builds
- `pin_build` - Pin/unpin builds to prevent cleanup
- `set_build_tag` - Add or remove build tags
- `download_artifact` - Download build artifacts
- `search_builds` - Search builds with comprehensive filters (status, branch, user, dates, tags, etc.)

### Technical Features
- Graceful shutdown with SIGTERM/SIGINT handling
- Configuration hot-reload with SIGHUP
- Rate limiting and retry logic for TeamCity API calls
- OpenTelemetry tracing support (planned)
- Horizontal scaling capability (stateless design)
- TLS 1.3 support with configurable certificates

## [1.0.0] - TBD

### Added
- First stable release of TeamCity MCP Server
- Full MCP protocol compliance
- Production-ready deployment artifacts
- Complete documentation and examples

## Security

### Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |

### Reporting Security Vulnerabilities

Please report security vulnerabilities by emailing the maintainers. Do not create public GitHub issues for security vulnerabilities. 
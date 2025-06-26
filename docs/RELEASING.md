# Releasing TeamCity MCP Server

This document describes the release process for TeamCity MCP Server using GoReleaser.

## Prerequisites

1. **GitHub Repository**: Ensure you have push access to the repository
2. **Docker Hub**: Ensure you have push access to `itcaat/teamcity-mcp`
3. **GitHub Secrets**: Configure the following secrets in your repository:
   - `DOCKER_USERNAME`: Your Docker Hub username
   - `DOCKER_PASSWORD`: Your Docker Hub password or access token

## Release Process

### 1. Prepare for Release

Ensure all changes are merged to the `main` branch and tests are passing:

```bash
# Run all checks
make check

# Test GoReleaser configuration
make release-check

# Build a snapshot to test locally
make release-snapshot
```

### 2. Create a Release

Create and push a semantic version tag:

```bash
# Create a new tag (replace with actual version)
git tag -a v1.0.0 -m "Release v1.0.0"

# Push the tag to trigger the release
git push origin v1.0.0
```

### 3. Automated Release Process

When you push a tag, GitHub Actions will automatically:

1. **Run Tests**: Execute all unit and integration tests
2. **Build Binaries**: Create cross-platform binaries for:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64, arm64)
3. **Build Docker Images**: Create multi-platform Docker images
4. **Create GitHub Release**: Generate release notes and upload artifacts
5. **Push to Docker Hub**: Publish images with proper tags
6. **Security Scan**: Run Trivy vulnerability scanner

### 4. Release Artifacts

Each release creates:

- **GitHub Release** with:
  - Compiled binaries for all platforms
  - Checksums file
  - Automated changelog
  - Docker pull commands

- **Docker Images** with tags:
  - `itcaat/teamcity-mcp:latest`
  - `itcaat/teamcity-mcp:v1.0.0`
  - `itcaat/teamcity-mcp:v1.0`
  - `itcaat/teamcity-mcp:v1`

## Version Strategy

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (v1.0.0 → v2.0.0): Breaking changes
- **MINOR** (v1.0.0 → v1.1.0): New features, backward compatible
- **PATCH** (v1.0.0 → v1.0.1): Bug fixes, backward compatible

## Pre-release Testing

Before creating a release tag, test locally:

```bash
# Build snapshot release
make release-snapshot

# Test the binary
./dist/teamcity-mcp_linux_amd64_v1/teamcity-mcp --version

# Test Docker image
docker run --rm itcaat/teamcity-mcp:latest-SNAPSHOT-<commit> --version
```

## Rollback Process

If a release has issues:

1. **Delete the tag** (if not yet widely distributed):
   ```bash
   git tag -d v1.0.0
   git push origin :refs/tags/v1.0.0
   ```

2. **Create a patch release** with fixes:
   ```bash
   git tag -a v1.0.1 -m "Hotfix v1.0.1"
   git push origin v1.0.1
   ```

## Manual Release

For manual releases or testing:

```bash
# Set required environment variables
export GITHUB_TOKEN="your-github-token"

# Run GoReleaser manually
goreleaser release --clean
```

## Troubleshooting

### Common Issues

1. **Docker login fails**: Check `DOCKER_USERNAME` and `DOCKER_PASSWORD` secrets
2. **GitHub release fails**: Ensure `GITHUB_TOKEN` has proper permissions
3. **Build fails**: Check Go version compatibility and dependencies

### Debug Commands

```bash
# Check GoReleaser configuration
goreleaser check

# Build without releasing
goreleaser build --snapshot --clean

# Release in dry-run mode
goreleaser release --skip-publish --clean
``` 
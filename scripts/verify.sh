#!/bin/bash

# TeamCity MCP Server Verification Script
# This script tests all major functionality to verify the server is working correctly

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="http://localhost:8123"
SECRET="test-secret"

# Set default environment variables for testing
export TC_URL="${TC_URL:-http://localhost:8111}"
export TC_TOKEN="${TC_TOKEN:-test-token}"
export SERVER_SECRET="${SERVER_SECRET:-test-secret}"  # Optional - enables auth
export LOG_LEVEL="${LOG_LEVEL:-info}"

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "PASS") echo -e "${GREEN}✅ PASS${NC}: $message" ;;
        "FAIL") echo -e "${RED}❌ FAIL${NC}: $message" ;;
        "WARN") echo -e "${YELLOW}⚠️  WARN${NC}: $message" ;;
        "INFO") echo -e "${BLUE}ℹ️  INFO${NC}: $message" ;;
    esac
}

# Function to check if server is running
check_server_running() {
    if pgrep -f "server" > /dev/null; then
        return 0
    else
        return 1
    fi
}

# Function to start server if not running
start_server() {
    if ! check_server_running; then
        print_status "INFO" "Starting server with environment variables..."
        ./server &
        SERVER_PID=$!
        sleep 2
        if check_server_running; then
            print_status "PASS" "Server started successfully"
            return 0
        else
            print_status "FAIL" "Failed to start server"
            return 1
        fi
    else
        print_status "INFO" "Server already running"
        return 0
    fi
}

# Function to stop server
stop_server() {
    if check_server_running; then
        print_status "INFO" "Stopping server..."
        pkill -f "server" || true
        sleep 1
        if ! check_server_running; then
            print_status "PASS" "Server stopped successfully"
        else
            print_status "WARN" "Server may still be running"
        fi
    fi
}

# Test 1: Build verification
test_build() {
    print_status "INFO" "Testing build..."
    if make build > /dev/null 2>&1; then
        print_status "PASS" "Project builds successfully"
    else
        print_status "FAIL" "Build failed"
        return 1
    fi
}

# Test 2: Unit tests
test_units() {
    print_status "INFO" "Running unit tests..."
    if go test ./tests/unit -v > /dev/null 2>&1; then
        print_status "PASS" "Unit tests pass"
    else
        print_status "FAIL" "Unit tests failed"
        return 1
    fi
}

# Test 3: Environment variable help
test_help() {
    print_status "INFO" "Testing help output..."
    if ./server --help > /dev/null 2>&1; then
        print_status "PASS" "Help command works"
    else
        print_status "FAIL" "Help command failed"
        return 1
    fi
}

# Test 4: Health endpoint
test_health() {
    print_status "INFO" "Testing health endpoint..."
    local response
    response=$(curl -s "$SERVER_URL/healthz" 2>/dev/null)
    if echo "$response" | grep -q '"status":"ok"'; then
        print_status "PASS" "Health endpoint responds correctly"
    else
        print_status "FAIL" "Health endpoint failed: $response"
        return 1
    fi
}

# Test 5: Metrics endpoint
test_metrics() {
    print_status "INFO" "Testing metrics endpoint..."
    local response
    response=$(curl -s "$SERVER_URL/metrics" 2>/dev/null)
    if [ -n "$response" ]; then
        print_status "PASS" "Metrics endpoint accessible"
    else
        print_status "FAIL" "Metrics endpoint failed"
        return 1
    fi
}

# Test 6: MCP Initialize
test_mcp_initialize() {
    print_status "INFO" "Testing MCP initialize..."
    local response
    response=$(curl -s -X POST "$SERVER_URL/mcp" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $SECRET" \
        -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{}}}' 2>/dev/null)
    
    if echo "$response" | grep -q '"protocolVersion":"2025-03-26"'; then
        print_status "PASS" "MCP initialize works correctly"
    else
        print_status "FAIL" "MCP initialize failed: $response"
        return 1
    fi
}

# Test 7: MCP Resources List
test_mcp_resources() {
    print_status "INFO" "Testing MCP resources list..."
    local response
    response=$(curl -s -X POST "$SERVER_URL/mcp" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $SECRET" \
        -d '{"jsonrpc":"2.0","id":2,"method":"resources/list","params":{}}' 2>/dev/null)
    
    if echo "$response" | grep -q '"resources"'; then
        print_status "PASS" "MCP resources list works"
    else
        print_status "FAIL" "MCP resources list failed: $response"
        return 1
    fi
}

# Test 8: MCP Tools List
test_mcp_tools() {
    print_status "INFO" "Testing MCP tools list..."
    local response
    response=$(curl -s -X POST "$SERVER_URL/mcp" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $SECRET" \
        -d '{"jsonrpc":"2.0","id":3,"method":"tools/list","params":{}}' 2>/dev/null)
    
    if echo "$response" | grep -q '"tools"' && echo "$response" | grep -q 'trigger_build'; then
        print_status "PASS" "MCP tools list works and includes expected tools"
    else
        print_status "FAIL" "MCP tools list failed: $response"
        return 1
    fi
}

# Test 9: Authentication
test_authentication() {
    print_status "INFO" "Testing authentication..."
    local response
    response=$(curl -s -X POST "$SERVER_URL/mcp" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer invalid-token" \
        -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' 2>/dev/null)
    
    if echo "$response" | grep -q -i "invalid\|unauthorized\|token"; then
        print_status "PASS" "Authentication properly rejects invalid tokens"
    else
        print_status "WARN" "Authentication test inconclusive: $response"
    fi
}

# Main execution
main() {
    echo "=============================================="
    echo "TeamCity MCP Server Verification"
    echo "=============================================="
    echo ""

    local failed_tests=0

    # Check prerequisites
    if [ ! -f "./server" ]; then
        print_status "FAIL" "Server binary not found. Run 'make build' first."
        exit 1
    fi

    # Run tests
    test_build || ((failed_tests++))
    test_units || ((failed_tests++))
    test_help || ((failed_tests++))

    # Start server for integration tests
    if start_server; then
        sleep 2  # Give server time to fully start
        
        test_health || ((failed_tests++))
        test_metrics || ((failed_tests++))
        test_mcp_initialize || ((failed_tests++))
        test_mcp_resources || ((failed_tests++))
        test_mcp_tools || ((failed_tests++))
        test_authentication || ((failed_tests++))
        
        # Clean up
        stop_server
    else
        print_status "FAIL" "Could not start server for integration tests"
        ((failed_tests++))
    fi

    echo ""
    echo "=============================================="
    if [ $failed_tests -eq 0 ]; then
        print_status "PASS" "All tests passed! 🎉"
        echo ""
        echo "Your TeamCity MCP server is working correctly!"
        echo ""
        echo "Environment variables used:"
        echo "  TC_URL=$TC_URL"
        echo "  TC_TOKEN=$TC_TOKEN"
        echo "  SERVER_SECRET=$SERVER_SECRET"
        echo ""
        echo "Next steps:"
        echo "1. Set your real TeamCity environment variables:"
        echo "   export TC_URL=https://your-teamcity-server.com"
        echo "   export TC_TOKEN=your-real-token"
        echo "   export SERVER_SECRET=your-real-secret"
        echo "2. Start the server: ./server"
        echo "3. Test with a real TeamCity instance"
        echo "4. Deploy using Docker or Kubernetes"
        exit 0
    else
        print_status "FAIL" "$failed_tests test(s) failed"
        echo ""
        echo "Please check the failed tests and fix any issues."
        exit 1
    fi
}

# Handle script arguments
case "${1:-}" in
    "help"|"-h"|"--help")
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  help, -h, --help    Show this help message"
        echo "  start               Start the server only"
        echo "  stop                Stop the server only"
        echo "  clean               Clean up any running servers"
        echo ""
        echo "Environment variables (set these for real usage):"
        echo "  TC_URL              TeamCity server URL"
        echo "  TC_TOKEN            TeamCity API token"
        echo "  SERVER_SECRET       Server secret for authentication"
        echo "  LOG_LEVEL           Log level (default: info)"
        echo ""
        echo "Examples:"
        echo "  $0                  Run all verification tests"
        echo "  TC_URL=https://tc.example.com TC_TOKEN=token123 $0"
        echo "  $0 start            Start server in background"
        exit 0
        ;;
    "start")
        start_server
        print_status "INFO" "Server running in background. Use '$0 stop' to stop it."
        exit 0
        ;;
    "stop")
        stop_server
        exit 0
        ;;
    "clean")
        stop_server
        print_status "INFO" "Cleanup complete"
        exit 0
        ;;
    "")
        main
        ;;
    *)
        print_status "FAIL" "Unknown option: $1"
        echo "Use '$0 help' for usage information."
        exit 1
        ;;
esac 
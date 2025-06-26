import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
export const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 100 }, // Ramp up to 100 users
    { duration: '5m', target: 100 }, // Stay at 100 users
    { duration: '2m', target: 200 }, // Ramp up to 200 users
    { duration: '5m', target: 200 }, // Stay at 200 users
    { duration: '2m', target: 0 },   // Ramp down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<200'], // 95% of requests must complete below 200ms
    'errors': ['rate<0.01'], // Error rate must be below 1%
  },
};

const BASE_URL = 'http://localhost:8123';
const AUTH_TOKEN = 'test-token';

const headers = {
  'Content-Type': 'application/json',
  'Authorization': `Bearer ${AUTH_TOKEN}`,
};

export default function () {
  const scenarios = [
    testHealthCheck,
    testMCPInitialize,
    testToolsList,
    testResourcesList,
  ];

  // Randomly select a scenario
  const scenario = scenarios[Math.floor(Math.random() * scenarios.length)];
  scenario();

  sleep(1);
}

function testHealthCheck() {
  const response = http.get(`${BASE_URL}/healthz`);
  
  const result = check(response, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 50ms': (r) => r.timings.duration < 50,
  });

  errorRate.add(!result);
}

function testMCPInitialize() {
  const payload = JSON.stringify({
    jsonrpc: '2.0',
    id: 1,
    method: 'initialize',
    params: {
      protocolVersion: '2024-11-05',
      capabilities: {},
      clientInfo: {
        name: 'k6-load-test',
        version: '1.0.0',
      },
    },
  });

  const response = http.post(`${BASE_URL}/mcp`, payload, { headers });

  const result = check(response, {
    'initialize status is 200': (r) => r.status === 200,
    'initialize has jsonrpc field': (r) => {
      try {
        return JSON.parse(r.body).jsonrpc === '2.0';
      } catch {
        return false;
      }
    },
    'initialize response time < 100ms': (r) => r.timings.duration < 100,
  });

  errorRate.add(!result);
}

function testToolsList() {
  const payload = JSON.stringify({
    jsonrpc: '2.0',
    id: 2,
    method: 'tools/list',
  });

  const response = http.post(`${BASE_URL}/mcp`, payload, { headers });

  const result = check(response, {
    'tools/list status is 200': (r) => r.status === 200,
    'tools/list has tools': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.result && Array.isArray(body.result.tools);
      } catch {
        return false;
      }
    },
    'tools/list response time < 100ms': (r) => r.timings.duration < 100,
  });

  errorRate.add(!result);
}

function testResourcesList() {
  const payload = JSON.stringify({
    jsonrpc: '2.0',
    id: 3,
    method: 'resources/list',
    params: {
      uri: 'teamcity://projects',
    },
  });

  const response = http.post(`${BASE_URL}/mcp`, payload, { headers });

  const result = check(response, {
    'resources/list status is 200': (r) => r.status === 200,
    'resources/list has resources': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.result && Array.isArray(body.result.resources);
      } catch {
        return false;
      }
    },
    'resources/list response time < 200ms': (r) => r.timings.duration < 200,
  });

  errorRate.add(!result);
}

export function setup() {
  console.log('Starting load test...');
  // Could add setup code here (e.g., warming up the server)
}

export function teardown() {
  console.log('Load test completed.');
  // Could add cleanup code here
} 
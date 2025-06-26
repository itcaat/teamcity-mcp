package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// MCP request metrics
	MCPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_requests_total",
			Help: "Total number of MCP requests",
		},
		[]string{"method", "status"},
	)

	MCPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_request_duration_seconds",
			Help:    "MCP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	// TeamCity API metrics
	TeamCityRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "teamcity_requests_total",
			Help: "Total number of TeamCity API requests",
		},
		[]string{"endpoint", "status"},
	)

	TeamCityRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "teamcity_request_duration_seconds",
			Help:    "TeamCity API request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)

	// Cache metrics
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"resource_type"},
	)

	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"resource_type"},
	)

	// Server health metrics
	ServerConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "server_connections_active",
			Help: "Number of active server connections",
		},
		[]string{"transport"},
	)

	ServerUptime = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "server_uptime_seconds_total",
			Help: "Total server uptime in seconds",
		},
	)
)

// Init initializes metrics collection
func Init() {
	// Register custom collectors if needed
}

// RecordMCPRequest records an MCP request metric
func RecordMCPRequest(method, status string, duration float64) {
	MCPRequestsTotal.WithLabelValues(method, status).Inc()
	MCPRequestDuration.WithLabelValues(method).Observe(duration)
}

// RecordTeamCityRequest records a TeamCity API request metric
func RecordTeamCityRequest(endpoint, status string, duration float64) {
	TeamCityRequestsTotal.WithLabelValues(endpoint, status).Inc()
	TeamCityRequestDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordCacheHit records a cache hit
func RecordCacheHit(resourceType string) {
	CacheHitsTotal.WithLabelValues(resourceType).Inc()
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss(resourceType string) {
	CacheMissesTotal.WithLabelValues(resourceType).Inc()
}

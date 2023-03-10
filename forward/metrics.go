package forward

import (
	"github.com/coredns/coredns/plugin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Variables declared for monitoring.
var (
	RequestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "pridns",
		Name:      "requests_total",
		Help:      "Counter of requests made per upstream.",
	}, []string{"to"})
	RcodeCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "pridns",
		Name:      "responses_total",
		Help:      "Counter of responses received per upstream.",
	}, []string{"rcode", "to"})
	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: plugin.Namespace,
		Subsystem: "pridns",
		Name:      "request_duration_seconds",
		Buckets:   plugin.TimeBuckets,
		Help:      "Histogram of the time each request took.",
	}, []string{"to", "rcode"})
	HealthcheckFailureCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "pridns",
		Name:      "healthcheck_failures_total",
		Help:      "Counter of the number of failed healthchecks.",
	}, []string{"to"})
	HealthcheckBrokenCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "pridns",
		Name:      "healthcheck_broken_total",
		Help:      "Counter of the number of complete failures of the healthchecks.",
	})
	MaxConcurrentRejectCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "pridns",
		Name:      "max_concurrent_rejects_total",
		Help:      "Counter of the number of queries rejected because the concurrent queries were at maximum.",
	})
	ConnCacheHitsCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "pridns",
		Name:      "conn_cache_hits_total",
		Help:      "Counter of connection cache hits per upstream and protocol.",
	}, []string{"to", "proto"})
	ConnCacheMissesCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "pridns",
		Name:      "conn_cache_misses_total",
		Help:      "Counter of connection cache misses per upstream and protocol.",
	}, []string{"to", "proto"})
)

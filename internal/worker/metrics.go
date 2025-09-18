package worker

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"time"
)

var (
	URLChecksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_url_checks_total",
			Help: "Nombre total de vérifications d'URL effectuées.",
		},
		[]string{"status"},
	)

	URLCheckDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "app_url_check_duration_seconds",
			Help:    "Durée des vérifications d'URL en secondes.",
			Buckets: prometheus.LinearBuckets(0.1, 0.1, 10),
		},
	)

	CacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_cache_hits_total",
		Help: "Nombre total de requêtes servies depuis le cache.",
	})

	CacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_cache_misses_total",
		Help: "Nombre total de requêtes non trouvées dans le cache.",
	})
)

type PrometheusMetricsProvider struct{}

func NewPrometheusMetricsProvider() *PrometheusMetricsProvider {
	return &PrometheusMetricsProvider{}
}

func (p *PrometheusMetricsProvider) IncChecksTotal(status string) {
	URLChecksTotal.WithLabelValues(status).Inc()
}

func (p *PrometheusMetricsProvider) ObserveCheckDuration(duration time.Duration) {
	URLCheckDuration.Observe(duration.Seconds())
}

func (p *PrometheusMetricsProvider) IncCacheHit() {
	CacheHits.Inc()
}

func (p *PrometheusMetricsProvider) IncCacheMiss() {
	CacheMisses.Inc()
}

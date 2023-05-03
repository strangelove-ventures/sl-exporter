package metrics

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func metricsHandler(reg *prometheus.Registry) http.Handler {
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{Timeout: 5 * time.Second})
}

var stubRequest = httptest.NewRequest("GET", "/metrics", nil)

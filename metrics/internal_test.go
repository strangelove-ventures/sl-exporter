package metrics

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestInternal_IncAPIError(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	metrics := NewInternal()
	reg.MustRegister(metrics.Metrics()[0])

	u, err := url.Parse("http://test.example/should/not/be/used")
	require.NoError(t, err)

	metrics.IncAPIError(*u, "timeout")

	h := metricsHandler(reg)
	r := httptest.NewRecorder()
	h.ServeHTTP(r, stubRequest)

	require.Contains(t, r.Body.String(), `sl_exporter_reference_api_error_total{host="test.example",reason="timeout"} 1`)
}

func TestInternal_IncFailedTask(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	metrics := NewInternal()
	reg.MustRegister(metrics.Metrics()[1])

	// Increment twice
	metrics.IncFailedTask("test_group")
	metrics.IncFailedTask("test_group")

	h := metricsHandler(reg)
	r := httptest.NewRecorder()
	h.ServeHTTP(r, stubRequest)

	require.Contains(t, r.Body.String(), `sl_exporter_task_error_total{group="test_group"} 2`)
}

package metrics

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestReferenceRPC_IncClientError(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	metrics := NewReferenceRPC()
	reg.MustRegister(metrics.Metrics()[0])

	metrics.IncClientError("cosmos-lcd", url.URL{Host: "test.example"}, "timeout")

	h := metricsHandler(reg)
	r := httptest.NewRecorder()
	h.ServeHTTP(r, stubRequest)

	const want = `# HELP sl_exporter_reference_rpc_error_count Number of errors encountered while making external RPC calls.
# TYPE sl_exporter_reference_rpc_error_count counter
sl_exporter_reference_rpc_error_count{host="test.example",reason="timeout",type="cosmos-lcd"} 1`
	require.Equal(t, strings.TrimSpace(want), strings.TrimSpace(r.Body.String()))
}

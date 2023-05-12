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

	u, err := url.Parse("http://test.example/should/not/be/used")
	require.NoError(t, err)

	metrics.IncClientError("cosmos-lcd", *u, "timeout")

	h := metricsHandler(reg)
	r := httptest.NewRecorder()
	h.ServeHTTP(r, stubRequest)

	const want = `# HELP sl_exporter_reference_rpc_error_count Number of errors encountered while making external RPC, API, or GRPC calls.
# TYPE sl_exporter_reference_rpc_error_count counter
sl_exporter_reference_rpc_error_count{host="test.example",reason="timeout",type="cosmos-lcd"} 1`
	require.Equal(t, strings.TrimSpace(want), strings.TrimSpace(r.Body.String()))
}

package metrics

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestCosmos(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		cosmos := NewCosmos()
		reg.MustRegister(cosmos.Metrics()...)

		u, err := url.Parse("http://cosmos.api.example.com:26657?timeout=10s")
		require.NoError(t, err)

		cosmos.SetNodeHeight("cosmoshub-4", *u, 12345)

		h := metricsHandler(reg)
		r := httptest.NewRecorder()
		h.ServeHTTP(r, stubRequest)

		const want = `# HELP sl_exporter_cosmos_block_height Latest block height of a cosmos node.
# TYPE sl_exporter_cosmos_block_height gauge
sl_exporter_cosmos_block_height{chain_id="cosmoshub-4",source="cosmos.api.example.com"} 12345
`
		require.Equal(t, strings.TrimSpace(want), strings.TrimSpace(r.Body.String()))
	})

	t.Run("url with path", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		cosmos := NewCosmos()
		reg.MustRegister(cosmos.Metrics()...)

		// Some nodes, like Strangelove's Voyager API, use one hostname and different paths for different chains.
		u, err := url.Parse("http://api.example.com:26657/v1/cosmos")
		require.NoError(t, err)

		cosmos.SetNodeHeight("cosmoshub-4", *u, 12345)

		h := metricsHandler(reg)
		r := httptest.NewRecorder()
		h.ServeHTTP(r, stubRequest)

		const want = `sl_exporter_cosmos_block_height{chain_id="cosmoshub-4",source="api.example.com/v1/cosmos"} 12345`
		require.Contains(t, r.Body.String(), want)
	})
}

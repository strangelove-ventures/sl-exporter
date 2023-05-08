package metrics

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestComet(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		comet := NewComet()
		reg.MustRegister(comet.Metrics()...)

		u, err := url.Parse("http://cosmos.rpc.example.com:26657?timeout=10s")
		require.NoError(t, err)

		comet.SetNodeHeight("cosmoshub-4", *u, 12345)

		h := metricsHandler(reg)
		r := httptest.NewRecorder()
		h.ServeHTTP(r, stubRequest)

		const want = `# HELP public_rpc_node_height Node height of a public RPC node
# TYPE public_rpc_node_height gauge
public_rpc_node_height{chain_id="cosmoshub-4",source="cosmos.rpc.example.com"} 12345
`
		require.Equal(t, strings.TrimSpace(want), strings.TrimSpace(r.Body.String()))
	})

	t.Run("rpc url with path", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		comet := NewComet()
		reg.MustRegister(comet.Metrics()...)

		// Some RPC nodes, like Strangelove's Voyager API, use one hostname and different paths for different chains.
		u, err := url.Parse("http://rpc.example.com:26657/v1/cosmos")
		require.NoError(t, err)

		comet.SetNodeHeight("cosmoshub-4", *u, 12345)

		h := metricsHandler(reg)
		r := httptest.NewRecorder()
		h.ServeHTTP(r, stubRequest)

		const want = `public_rpc_node_height{chain_id="cosmoshub-4",source="rpc.example.com/v1/cosmos"} 12345`
		require.Contains(t, r.Body.String(), want)
	})
}

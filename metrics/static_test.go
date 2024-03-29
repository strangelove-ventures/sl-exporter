package metrics

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestBuildStatic(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		gauges := []StaticGauge{
			{
				Name:        "gauge_1",
				Description: "desc_1",
				Labels:      []string{"chain", "denom"},
				Samples: []StaticSample{
					{Labels: []string{"agoric-1", "ubld"}, Value: 1},
					{Labels: []string{"cosmoshub-4", "uatom"}, Value: 2},
				},
			},
			{
				Name:        "gauge_2",
				Description: "desc_2",
				Samples: []StaticSample{
					{Value: 3},
				},
			},
		}

		got := BuildStatic(gauges)
		require.Len(t, got, 2)

		reg := prometheus.NewRegistry()
		reg.MustRegister(got...)
		h := metricsHandler(reg)
		r := httptest.NewRecorder()
		h.ServeHTTP(r, stubRequest)

		const want = `# HELP sl_exporter_static_gauge_1 desc_1
# TYPE sl_exporter_static_gauge_1 gauge
sl_exporter_static_gauge_1{chain="agoric-1",denom="ubld"} 1
sl_exporter_static_gauge_1{chain="cosmoshub-4",denom="uatom"} 2
# HELP sl_exporter_static_gauge_2 desc_2
# TYPE sl_exporter_static_gauge_2 gauge
sl_exporter_static_gauge_2 3`
		require.Equal(t, want, strings.TrimSpace(r.Body.String()))
	})

	t.Run("zero state", func(t *testing.T) {
		require.Empty(t, BuildStatic(nil))
	})

	t.Run("invalid labels", func(t *testing.T) {
		gauges := []StaticGauge{
			{
				Name:        "sl_exporter_static_gauge_1",
				Description: "desc_1",
				Labels:      []string{"chain", "denom"},
				Samples: []StaticSample{
					{Labels: []string{"agoric-1"}, Value: 1}, // 1 label, should be 2
					{Labels: []string{"cosmoshub-4", "uatom"}, Value: 2},
				},
			},
		}

		require.Panics(t, func() {
			_ = BuildStatic(gauges)
		})
	})
}

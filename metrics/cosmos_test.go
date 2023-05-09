package metrics

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestCosmos_SetNodeHeight(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		cosmos := NewCosmos()
		reg.MustRegister(cosmos.Metrics()[0])

		u, err := url.Parse("http://cosmos.api.example.com:26657?timeout=10s")
		require.NoError(t, err)

		cosmos.SetNodeHeight("cosmoshub-4", *u, 12345)

		h := metricsHandler(reg)
		r := httptest.NewRecorder()
		h.ServeHTTP(r, stubRequest)

		const want = `# HELP sl_exporter_cosmos_latest_block_height Latest block height of a cosmos node.
# TYPE sl_exporter_cosmos_latest_block_height gauge
sl_exporter_cosmos_latest_block_height{chain_id="cosmoshub-4",source="cosmos.api.example.com"} 12345
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

		const want = `sl_exporter_cosmos_latest_block_height{chain_id="cosmoshub-4",source="api.example.com/v1/cosmos"} 12345`
		require.Contains(t, r.Body.String(), want)
	})
}

func TestCosmos_SetValidatorStatus(t *testing.T) {
	t.Parallel()

	cosmos := NewCosmos()
	reg := prometheus.NewRegistry()
	reg.MustRegister(cosmos.Metrics()[1])
	h := metricsHandler(reg)
	u, err := url.Parse("http://cosmos.api.example.com:26657/v1/cosmos?timeout=10s")
	require.NoError(t, err)

	for _, tt := range []struct {
		Status    ValStatus
		WantValue int
	}{
		{Status: ValStatusActive, WantValue: 0},
		{Status: ValStatusJailed, WantValue: 1},
		{Status: ValStatusTombstoned, WantValue: 1},
	} {
		cosmos.SetValJailStatus("cosmoshub-4", "cosmosvalcons123", *u, tt.Status)
		const wantInfo = `
# HELP sl_exporter_cosmos_val_jailed_status Whether a validator is active, jailed, or tombstoned.
# TYPE sl_exporter_cosmos_val_jailed_status gauge
`
		r := httptest.NewRecorder()
		h.ServeHTTP(r, stubRequest)

		require.Contains(t, strings.TrimSpace(r.Body.String()), strings.TrimSpace(wantInfo), tt)

		want := fmt.Sprintf(`
sl_exporter_cosmos_val_jailed_status{address="cosmosvalcons123",chain_id="cosmoshub-4",source="cosmos.api.example.com/v1/cosmos",status="%s"} %d`,
			tt.Status, tt.WantValue)

		require.Contains(t, strings.TrimSpace(r.Body.String()), strings.TrimSpace(want), tt)
	}
}

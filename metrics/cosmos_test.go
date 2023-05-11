package metrics

import (
	"fmt"
	"net/http/httptest"
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

		cosmos.SetNodeHeight("cosmoshub-4", 12345)

		h := metricsHandler(reg)
		r := httptest.NewRecorder()
		h.ServeHTTP(r, stubRequest)

		const want = `# HELP sl_exporter_cosmos_latest_block_height Latest block height of a cosmos node.
# TYPE sl_exporter_cosmos_latest_block_height gauge
sl_exporter_cosmos_latest_block_height{chain_id="cosmoshub-4"} 12345
`
		require.Equal(t, strings.TrimSpace(want), strings.TrimSpace(r.Body.String()))
	})

	t.Run("url with path", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		cosmos := NewCosmos()
		reg.MustRegister(cosmos.Metrics()...)

		cosmos.SetNodeHeight("cosmoshub-4", 12345)

		h := metricsHandler(reg)
		r := httptest.NewRecorder()
		h.ServeHTTP(r, stubRequest)

		const want = `sl_exporter_cosmos_latest_block_height{chain_id="cosmoshub-4"} 12345`
		require.Contains(t, r.Body.String(), want)
	})
}

func TestCosmos_SetValJailStatus(t *testing.T) {
	t.Parallel()

	cosmos := NewCosmos()
	reg := prometheus.NewRegistry()
	reg.MustRegister(cosmos.Metrics()[1])
	h := metricsHandler(reg)

	for _, tt := range []struct {
		Status    JailStatus
		WantValue int
	}{
		{Status: JailStatusActive, WantValue: 0},
		{Status: JailStatusJailed, WantValue: 1},
		{Status: JailStatusTombstoned, WantValue: 2},
	} {
		cosmos.SetValJailStatus("cosmoshub-4", "cosmosvalcons123", tt.Status)
		r := httptest.NewRecorder()
		h.ServeHTTP(r, stubRequest)

		const wantInfo = `
# HELP sl_exporter_cosmos_val_latest_jailed_status 0 if the validator is not jailed. 1 if the validator is jailed. 2 if the validator is tombstoned.
# TYPE sl_exporter_cosmos_val_latest_jailed_status gauge
`
		require.Contains(t, strings.TrimSpace(r.Body.String()), strings.TrimSpace(wantInfo), tt)

		want := fmt.Sprintf(`
sl_exporter_cosmos_val_latest_jailed_status{address="cosmosvalcons123",chain_id="cosmoshub-4"} %d`,
			tt.WantValue)

		require.Contains(t, strings.TrimSpace(r.Body.String()), strings.TrimSpace(want), tt)
	}
}

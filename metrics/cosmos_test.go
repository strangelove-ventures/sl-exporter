package metrics

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/strangelove-ventures/sl-exporter/cosmos"
	"github.com/stretchr/testify/require"
)

func TestCosmos_SetNodeHeight(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	metrics := NewCosmos()
	reg.MustRegister(metrics.Metrics()[0])

	metrics.SetNodeHeight("cosmoshub-4", 12345)

	h := metricsHandler(reg)
	r := httptest.NewRecorder()
	h.ServeHTTP(r, stubRequest)

	const want = `sl_exporter_cosmos_latest_block_height{chain_id="cosmoshub-4"} 12345`
	require.Contains(t, r.Body.String(), want)
}

func TestCosmos_SetValJailStatus(t *testing.T) {
	t.Parallel()

	metrics := NewCosmos()
	reg := prometheus.NewRegistry()
	reg.MustRegister(metrics.Metrics()[1])
	h := metricsHandler(reg)

	for _, tt := range []struct {
		Status    cosmos.JailStatus
		WantValue int
	}{
		{Status: cosmos.JailStatusActive, WantValue: 0},
		{Status: cosmos.JailStatusJailed, WantValue: 1},
		{Status: cosmos.JailStatusTombstoned, WantValue: 2},
	} {
		metrics.SetValJailStatus("cosmoshub-4", "cosmosvalcons123", tt.Status)
		r := httptest.NewRecorder()
		h.ServeHTTP(r, stubRequest)

		want := fmt.Sprintf(`
sl_exporter_cosmos_val_latest_jailed_status{address="cosmosvalcons123",chain_id="cosmoshub-4"} %d`,
			tt.WantValue)

		require.Contains(t, r.Body.String(), want, tt)
	}
}

func TestCosmos_IncValSignedBlocks(t *testing.T) {
	t.Parallel()

	metrics := NewCosmos()
	reg := prometheus.NewRegistry()
	reg.MustRegister(metrics.Metrics()[2])
	h := metricsHandler(reg)

	// Purposefully calling twice
	metrics.IncValSignedBlocks("cosmoshub-4", "cosmosvalcons123")
	metrics.IncValSignedBlocks("cosmoshub-4", "cosmosvalcons123")

	r := httptest.NewRecorder()
	h.ServeHTTP(r, stubRequest)

	const want = `sl_exporter_cosmos_val_signed_blocks_total{address="cosmosvalcons123",chain_id="cosmoshub-4"} 2`
	require.Contains(t, r.Body.String(), want)
}

func TestCosmos_SetValSignedBlock(t *testing.T) {
	t.Parallel()

	metrics := NewCosmos()
	reg := prometheus.NewRegistry()
	reg.MustRegister(metrics.Metrics()[3])
	h := metricsHandler(reg)

	metrics.SetValSignedBlock("cosmoshub-4", "cosmosvalcons123", 12345)

	r := httptest.NewRecorder()
	h.ServeHTTP(r, stubRequest)

	const want = `sl_exporter_cosmos_val_latest_signed_block_height{address="cosmosvalcons123",chain_id="cosmoshub-4"} 12345`
	require.Contains(t, r.Body.String(), want)
}

func TestCosmos_SetValMissedBlocks(t *testing.T) {
	t.Parallel()

	metrics := NewCosmos()
	reg := prometheus.NewRegistry()
	reg.MustRegister(metrics.Metrics()[4])
	h := metricsHandler(reg)

	metrics.SetValMissedBlocks("cosmoshub-4", "cosmosvalcons123", 9)

	r := httptest.NewRecorder()
	h.ServeHTTP(r, stubRequest)

	const want = `sl_exporter_cosmos_val_latest_missed_blocks{address="cosmosvalcons123",chain_id="cosmoshub-4"} 9`
	require.Contains(t, r.Body.String(), want)
}

func TestCosmos_SetValSlashingParams(t *testing.T) {
	t.Parallel()

	metrics := NewCosmos()
	reg := prometheus.NewRegistry()
	reg.MustRegister(metrics.Metrics()[5])
	h := metricsHandler(reg)

	metrics.SetValSlashingParams("cosmoshub-4", 100)

	r := httptest.NewRecorder()
	h.ServeHTTP(r, stubRequest)

	const want = `sl_exporter_cosmos_val_slashing_window_blocks{chain_id="cosmoshub-4"} 100`
	require.Contains(t, r.Body.String(), want)
}

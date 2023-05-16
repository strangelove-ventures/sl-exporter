package metrics

import (
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

// Internal records metrics that represent the health of sl-exporter itself.
type Internal struct {
	refAPIErrors *prometheus.CounterVec
	// TODO(nix): Reference API requests and histogram of latency.
	failedTasks *prometheus.CounterVec
}

func NewInternal() *Internal {
	return &Internal{
		refAPIErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: prometheus.BuildFQName(namespace, "", "reference_api_error_total"),
				Help: "Number of errors encountered while making external calls to an API to gather reference data.",
			},
			[]string{"host", "reason"},
		),
		failedTasks: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: prometheus.BuildFQName(namespace, "", "task_error_total"),
				Help: "Number of failed sl-exporter tasks. A task is a generalized unit of work such as querying validator signing status.",
			},
			[]string{"group"},
		),
	}
}

// IncAPIError increments the number of errors encountered while making external API calls.
func (c Internal) IncAPIError(host url.URL, reason string) {
	c.refAPIErrors.WithLabelValues(host.Hostname(), reason).Inc()
}

// IncFailedTask increments the number of failed sl-exporter tasks.
func (c Internal) IncFailedTask(group string) {
	c.failedTasks.WithLabelValues(group).Inc()
}

func (c Internal) Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		c.refAPIErrors,
		c.failedTasks,
	}
}

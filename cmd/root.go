package cmd

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/strangelove-ventures/sl-exporter/cosmos"
	"github.com/strangelove-ventures/sl-exporter/metrics"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

const collector = "sl_exporter"

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

func Execute() {
	var cfg Config

	flag.StringVar(&cfg.File, "config", "config.yaml", "Path to configuration file")
	flag.StringVar(&cfg.BindAddr, "bind", ":9100", "Address to bind")
	flag.IntVar(&cfg.NumWorkers, "workers", runtime.NumCPU()*25, "Number of background workers that poll for data")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.StringVar(&cfg.LogFormat, "log-format", "text", "Log format (text, json)")
	flag.Parse()

	// Setup logging
	var programLevel = new(slog.LevelVar)
	if err := programLevel.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		logFatal("Failed to parse log level", err)
	}
	if cfg.LogFormat == "json" {
		slog.SetDefault(slog.New(slog.HandlerOptions{Level: programLevel}.NewJSONHandler(os.Stderr)))
	} else {
		slog.SetDefault(slog.New(slog.HandlerOptions{Level: programLevel}.NewTextHandler(os.Stderr)))
	}

	// Parse config
	if err := parseConfig(&cfg); err != nil {
		logFatal("Failed to parse config", err)
	}

	// Initialize prometheus registry
	registry := prometheus.NewRegistry()
	registry.MustRegister(version.NewCollector(collector))

	// Register static metrics
	registry.MustRegister(metrics.BuildStatic(cfg.Static.Gauges)...)

	// Register cosmos chain metrics
	cosmosMets := metrics.NewCosmos()
	registry.MustRegister(cosmosMets.Metrics()...)

	var jobs []metrics.Job

	// Initialize Endpoint jobs
	restClient := cosmos.NewRestClient(httpClient)
	restJobs, err := metrics.BuildCosmosRestJobs(cosmosMets, restClient, cfg.Cosmos)
	if err != nil {
		logFatal("Failed to create cosmos rest jobs", err)
	}
	jobs = append(jobs, toJobs(restJobs)...)

	// Configure error group with signal handling.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	eg, ctx := errgroup.WithContext(ctx)

	// Add all jobs to worker pool
	pool, err := metrics.NewWorkerPool(jobs, cfg.NumWorkers)
	if err != nil {
		logFatal("Failed to create worker pool", err)
	}

	// Configure server
	const timeout = 60 * time.Second
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{Timeout: timeout}))
	server := &http.Server{
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		Addr:         cfg.BindAddr,
		Handler:      mux,
	}

	// Start goroutines
	eg.Go(func() error {
		slog.Info("Starting Prometheus metrics server", "addr", cfg.BindAddr)
		return server.ListenAndServe()
	})
	eg.Go(func() error {
		<-ctx.Done()
		// Give server 5 seconds to shutdown gracefully.
		cctx, ccancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer ccancel()
		return server.Shutdown(cctx)
	})
	eg.Go(func() error {
		pool.Start(ctx)
		return nil
	})

	err = eg.Wait()
	switch {
	case errors.Is(err, http.ErrServerClosed):
		slog.Info("Server shutdown")
	case err != nil:
		logFatal("Fatal error", err)
	}
}

func logFatal(msg string, err error) {
	slog.Error(msg, "error", err)
	os.Exit(1)
}

func toJobs[T metrics.Job](jobs []T) []metrics.Job {
	result := make([]metrics.Job, len(jobs))
	for i := range jobs {
		result[i] = jobs[i]
	}
	return result
}

package cmd

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
	"github.com/strangelove-ventures/sl-exporter/metrics"
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
	flag.IntVar(&cfg.NumWorkers, "workers", runtime.NumCPU()*25, "Number of background workers that poll data")
	flag.Parse()

	if err := parseConfig(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	// Initialize prometheus registry
	registry := prometheus.NewRegistry()
	registry.MustRegister(version.NewCollector(collector))

	// Register static metrics
	registry.MustRegister(metrics.BuildStatic(cfg.Static.Gauges)...)

	// Register cosmos chain metrics
	cosmos := metrics.NewCosmos()
	registry.MustRegister(cosmos.Metrics()...)

	var jobs []metrics.Job

	// Initialize RPC jobs
	cometClient := metrics.NewCometClient(httpClient)
	rpcJobs, err := metrics.NewRPCJobs(cosmos, cometClient, cfg.Cosmos)
	if err != nil {
		log.Fatalf("Failed to create RPC jobs: %v", err)
	}
	jobs = append(jobs, metrics.ToJobs(rpcJobs)...)

	// Configure error group with signal handling.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	eg, ctx := errgroup.WithContext(ctx)

	// Create worker pool to poll data
	pool := metrics.NewWorkerPool(jobs, cfg.NumWorkers)

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
		log.Infof("Starting Prometheus metrics server - %s", cfg.BindAddr)
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
	eg.Go(func() error {
		pool.Wait()
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}

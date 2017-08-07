package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dparrish/dm200stats/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	username = flag.String("user", "", "Username for HTTP authentication")
	password = flag.String("pass", "", "Password for HTTP authentication")
	port     = flag.Int("port", 8080, "Port to listen for Prometheus requests")

	timer = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "dm200_collect_duration_seconds",
		Help:    "Time taken to poll the device for statistics",
		Buckets: prometheus.LinearBuckets(0.01, 0.01, 10),
	})
	syncSpeedDown = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "sync_speed_down",
		Help: "Downstream sync speed in bits per second",
	})
	syncSpeedUp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "sync_speed_up",
		Help: "Upstream sync speed in bits per second",
	})
	noiseDown = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "noise_down",
		Help: "Downstream noise in dB",
	})
	noiseUp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "noise_up",
		Help: "Upstream noise in dB",
	})
	attenuationDown = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "attenuation_down",
		Help: "Downstream attenuation in dB",
	})
	attenuationUp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "attenuation_up",
		Help: "Upstream attenuation in dB",
	})
	bytesDown = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bytes_down",
		Help: "Downstream bytes",
	})
	bytesUp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bytes_up",
		Help: "Upstream bytes",
	})
)

func init() {
	prometheus.MustRegister(timer)
	prometheus.MustRegister(syncSpeedDown)
	prometheus.MustRegister(syncSpeedUp)
	prometheus.MustRegister(noiseDown)
	prometheus.MustRegister(noiseUp)
	prometheus.MustRegister(attenuationDown)
	prometheus.MustRegister(attenuationUp)
	prometheus.MustRegister(bytesDown)
	prometheus.MustRegister(bytesUp)
}

func main() {
	flag.Parse()

	if *username == "" {
		*username = os.Getenv("DM200_USER")
	}
	if *username == "" {
		log.Fatalf("--user or DM200_USER must be supplied")
	}
	if *password == "" {
		*password = os.Getenv("DM200_PASS")
	}
	if *password == "" {
		log.Fatalf("--pass or DM200_PASS must be supplied")
	}

	ip := flag.Arg(0)
	if ip == "" {
		ip = os.Getenv("DM200_IP")
	}
	if ip == "" {
		log.Fatalf("An IP address must be supplied on the command line or in a DM200_IP environment variable")
	}

	ctx := context.Background()

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.Handler())
	go func() { log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)) }()

	for {
		t := prometheus.NewTimer(timer)
		// Set a 5 second timeout collecting stats.
		tctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		stats, err := collector.Collect(tctx, ip, *username, *password)
		cancel()
		if err != nil {
			log.Print(err)
			time.Sleep(time.Second)
			continue
		}
		t.ObserveDuration()

		// Record the latest statistics for later polling by Prometheus.
		syncSpeedDown.Set(stats.SyncSpeedDown)
		syncSpeedUp.Set(stats.SyncSpeedUp)
		noiseDown.Set(stats.NoiseDown)
		noiseUp.Set(stats.NoiseUp)
		attenuationDown.Set(stats.AttenuationDown)
		attenuationUp.Set(stats.AttenuationUp)
		bytesDown.Set(stats.BytesDown)
		bytesUp.Set(stats.BytesUp)

		// Poll again in 10 seconds.
		time.Sleep(10 * time.Second)
	}
}

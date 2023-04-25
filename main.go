package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gartnera/arris-surfboard-exporter/lib"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metrics struct {
	upstreamPower *prometheus.GaugeVec

	downstreamPower          *prometheus.GaugeVec
	downstreamSNR            *prometheus.GaugeVec
	downstreamCorrected      *prometheus.GaugeVec
	downstreamUncorrectables *prometheus.GaugeVec
}

var upstreamLabels = []string{
	"channel",
	"channel_id",
	"us_channel_type",
}

var downstreamLabels = []string{
	"channel_id",
	"modulation",
}

func NewMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		upstreamPower: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "arris_surfboard",
				Subsystem: "upstream",
				Name:      "power",
			},
			upstreamLabels,
		),
		downstreamPower: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "arris_surfboard",
				Subsystem: "downstream",
				Name:      "power",
			},
			downstreamLabels,
		),
		downstreamSNR: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "arris_surfboard",
				Subsystem: "downstream",
				Name:      "snr",
			},
			downstreamLabels,
		),
		downstreamCorrected: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "arris_surfboard",
				Subsystem: "downstream",
				Name:      "corrected",
			},
			downstreamLabels,
		),
		downstreamUncorrectables: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "arris_surfboard",
				Subsystem: "downstream",
				Name:      "uncorrectables",
			},
			downstreamLabels,
		),
	}
	reg.MustRegister(m.upstreamPower)
	reg.MustRegister(m.downstreamPower)
	reg.MustRegister(m.downstreamSNR)
	reg.MustRegister(m.downstreamCorrected)
	reg.MustRegister(m.downstreamUncorrectables)
	return m
}

func main() {
	// Create a non-global registry.
	reg := prometheus.NewRegistry()

	// Create new metrics and register them using the custom registry.
	m := NewMetrics(reg)

	upstreamBondedChannelCb := func(data *lib.UpstreamBondedChannel) {
		labels := prometheus.Labels(data.Labels())
		m.upstreamPower.With(labels).Set(data.Power)
	}

	downstreamBondedChannelCb := func(data *lib.DownstreamBondedChannel) {
		labels := prometheus.Labels(data.Labels())
		m.downstreamPower.With(labels).Set(data.Power)
		m.downstreamSNR.With(labels).Set(data.SNR)
		m.downstreamCorrected.With(labels).Set(float64(data.Corrected))
		m.downstreamUncorrectables.With(labels).Set(float64(data.Uncorrectables))
	}

	scraper, err := lib.NewScraper("https://192.168.100.1", os.Getenv("CREDS"))
	if err != nil {
		log.Fatal(err)
	}
	go scraper.Run(upstreamBondedChannelCb, downstreamBondedChannelCb)

	// Expose metrics and custom registry via an HTTP server
	// using the HandleFor function. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	log.Fatal(http.ListenAndServe(":29116", nil))
}

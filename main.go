package main

import (
	"flag"
	"net"
	"net/http"

	"github.com/hatamiarash7/netflow-exporter/collector"
	"github.com/hatamiarash7/netflow-exporter/config"

	"github.com/itzg/go-flagsfiller"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	cfg     config.Config
	version = "dev"
)

func configureLog() {
	ll, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		panic(err)
	}

	log.SetLevel(ll)

	if cfg.LogFormat == "text" {
		log.SetFormatter(&log.TextFormatter{
			ForceColors:      true,
			ForceQuote:       true,
			FullTimestamp:    true,
			DisableTimestamp: false,
			TimestampFormat:  "2006-01-02 15:04:05",
		})
	} else {
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat:  "2006-01-02 15:04:05",
			DisableTimestamp: false,
			DataKey:          "",
			PrettyPrint:      false,
		})
	}
}

func init() {
	filler := flagsfiller.New()
	err := filler.Fill(flag.CommandLine, &cfg)
	if err != nil {
		log.Fatal(err)
	}
	flag.Parse()

	configureLog()

	log.Infof("Starting NetFlow Exporter %s", version)

	http.Handle(cfg.MetricPath, promhttp.Handler())
}

func main() {
	c := collector.NewCollector(cfg)
	prometheus.MustRegister(c)

	udpAddress, err := net.ResolveUDPAddr("udp", cfg.ListenAddress)
	if err != nil {
		log.Fatalf("Error resolving UDP address: %s", err)
	}

	udpSocket, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		log.Fatalf("Error listening to UDP address: %s", err)
	}

	log.Infof("Include: %s", cfg.Include)

	if len(cfg.Exclude) > 0 {
		log.Infof("Exclude: %s", cfg.Exclude)
	}

	go c.Reader(udpSocket)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
        <head><title>NetFlow Exporter</title></head>
        <body>
        <h1>NetFlow Exporter</h1>
        <p><a href='` + cfg.MetricPath + `'>Metrics</a></p>
        </body>
        </html>`))
	})

	log.Infof("Listening NetFlow on %s", cfg.ListenAddress)
	log.Infof("Listening metrics on %s", cfg.MetricAddress)

	log.Fatal(http.ListenAndServe(cfg.MetricAddress, nil))
}

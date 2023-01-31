package main

import (
	"flag"
	"net/http"
	"saltstack_exporter/build"
	"saltstack_exporter/environment"
	"saltstack_exporter/exporter"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	listenAddress = flag.String("web.listen-address", ":9142", "Address to listen on for telemetry")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics")
	askVersion    = flag.Bool("version", false, "Display version of this binary")
	hasDebug      = flag.Bool("debug", false, "active debug in log")
)

/*
 * Main function, starting point of the package
 */
func main() {
	flag.Parse()

	if *hasDebug {
		log.SetLevel(log.DebugLevel)
		//log.SetReportCaller(true)
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if *askVersion {
		build.VersionInformation()
	}

	saltUrl, saltUser, saltPassword := environment.LoadEnv()

	if saltUrl == "" || saltUser == "" || saltPassword == "" {
		log.WithFields(log.Fields{
			"saltUrl":      saltUrl,
			"saltUser":     saltUser,
			"saltPassword": saltPassword,
		}).Fatalln("You must filled the SALTSTACK_API_URL, SALTSTACK_API_USER and SALTSTACK_API_PASSWORD viariables. Either via local .env file or loaded globaly in your environment.")
	}

	exporter := exporter.NewExporter(saltUrl, saltUser, saltPassword)
	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

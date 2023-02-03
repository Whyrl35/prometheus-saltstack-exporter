package main

import (
	"net/http"
	"os"

	"github.com/Whyrl35/prometheus-saltstack-exporter/exporter"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

var (
	configFile    = kingpin.Flag("config.file", "Exporter configuration file.").Default("config.yaml").String()
	listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for telemetry").Default(":19142").String()
	metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics").Default("/metrics").String()
	hasDebug      = kingpin.Flag("debug", "Active debug in log").Bool()
)

/*
 * Config of the exporter, loading for YAML file
 */
type Config struct {
	Saltstack struct {
		Url      string `yaml:"url"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	}
}

/*
 * Init helper to pre-configure some metrics
 */
func init() {
	// Don't initialized the default version metrics as version don't report correctly, need to check LD_FLAGS
	// prometheus.MustRegister(version.NewCollector("prometheus-saltstack-exporter"))
}

/*
 * Main function, starting point of the package
 */
func main() {
	kingpin.Version(version.Print("prometheus-saltstack-exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	// Configuring the logger
	if *hasDebug {
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	log.WithFields(log.Fields{
		"config-file": *configFile,
		"version":     version.Info(),
		"build":       version.BuildContext(),
	}).Info("starting exporter")

	var config = loadConfig(*configFile)
	log.WithFields(log.Fields{
		"url":      config.Saltstack.Url,
		"username": config.Saltstack.Username,
		"password": config.Saltstack.Password,
	}).Debug("Configuration file loaded")

	exporter := exporter.NewExporter(config.Saltstack.Url, config.Saltstack.Username, config.Saltstack.Password)
	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
			<head><title>prometheus-saltstack-exporter</title></head>
			<body>
			<h1>prometheus-saltstack-exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
		if err != nil {
			log.Fatal("Error writing default message")
		}
	})

	log.Info("Beginning to serve on address ", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func loadConfig(configFile string) Config {
	config := Config{}

	// Load the config from the file
	configData, err := os.ReadFile(configFile)
	if err != nil {
		log.WithFields(log.Fields{
			"config-file": configFile,
		}).Fatalf("Error: %v", err)
	}

	errYAML := yaml.Unmarshal([]byte(configData), &config)
	if errYAML != nil {
		log.WithFields(log.Fields{
			"config-file": configFile,
		}).Fatalf("Error: %v", errYAML)
	}

	return config
}

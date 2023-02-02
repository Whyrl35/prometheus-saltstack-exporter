package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type Exporter struct {
	saltUrl, saltUser, saltPassword string
}

func NewExporter(saltUrl string, saltUser string, saltPassword string) *Exporter {
	return &Exporter{
		saltUrl:      saltUrl,
		saltUser:     saltUser,
		saltPassword: saltPassword,
	}
}

var masterUp = prometheus.NewDesc(prometheus.BuildFQName("saltstack", "", "master_up"), "Master in up(1) or down(0)", []string{"master"}, nil)
var minionsCount = prometheus.NewDesc(prometheus.BuildFQName("saltstack", "", "minions_count"), "Number of minions declared in salt", nil, nil)

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- masterUp
	ch <- minionsCount
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	f := NewFetcher(e.saltUrl, e.saltUser, e.saltPassword)

	err := f.Login()
	if err != nil {
		log.WithFields(log.Fields{
			"saltUrl":      e.saltUrl,
			"saltUser":     e.saltPassword,
			"saltPassword": "***",
		}).Fatal(err)
	}

	// Check master status
	masters, err := f.Masters()
	if err != nil {
		log.WithFields(log.Fields{
			"saltUrl":      e.saltUrl,
			"saltUser":     e.saltPassword,
			"saltPassword": "***",
		}).Error(err)
	}

	for k, v := range masters.status {
		ch <- prometheus.MustNewConstMetric(masterUp, prometheus.GaugeValue, map[bool]float64{true: 1, false: 0}[v], k)
	}

	// Check minions status
	minions, err := f.Minions()
	if err != nil {
		log.WithFields(log.Fields{
			"saltUrl":      e.saltUrl,
			"saltUser":     e.saltPassword,
			"saltPassword": "***",
		}).Error(err)
	}
	ch <- prometheus.MustNewConstMetric(minionsCount, prometheus.GaugeValue, float64(minions.count))

	// Check jobs status
	// Status to be checked : state.highstate / state.apply with Arguments = []
	// For one given Job check the Minions list and the Result for this minion
	jobs, err := f.Jobs()
	if err != nil {
		log.WithFields(log.Fields{
			"saltUrl":      e.saltUrl,
			"saltUser":     e.saltPassword,
			"saltPassword": "***",
		}).Error(err)
	}

	for _, job := range *jobs {
		if job.function == "state.highstate" || job.function == "state.apply" {
			// call a function that get the result for the job (success or not) then compute a metrics on it
			log.Debug()
		}
	}
}

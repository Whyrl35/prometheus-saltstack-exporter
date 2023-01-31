package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type Exporter struct {
	saltUrl, saltUser, saltPassword string
}

func NewExporter(saltUrl string, saltUser, saltPassword string) *Exporter {
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
	f.Login()

	// Check master status
	masters, err := f.Masters()
	if err != nil {
		log.WithFields(log.Fields{
			"saltUrl":      e.saltUrl,
			"saltUser":     e.saltPassword,
			"saltPassword": "***",
		}).Error("Unable to retreive Masters information: %v", err)
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
		}).Error("Unable to retreive Minions information: %v", err)
	}
	ch <- prometheus.MustNewConstMetric(minionsCount, prometheus.GaugeValue, float64(minions.count))
}

package exporter

import (
	"strconv"
	"sync"

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

var wg sync.WaitGroup
var masterUp = prometheus.NewDesc(prometheus.BuildFQName("saltstack", "", "master_up"), "Master in up(1) or down(0)", []string{"master"}, nil)
var minionsCount = prometheus.NewDesc(prometheus.BuildFQName("saltstack", "", "minions_count"), "Number of minions declared in salt", nil, nil)
var jobsStatus = prometheus.NewDesc(prometheus.BuildFQName("saltstack", "", "job_status"), "Job status", []string{"minion", "function"}, nil)

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- masterUp
	ch <- minionsCount
	ch <- jobsStatus
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

	// Create all chanels needed for the go routines
	masterChan := make(chan Masters)
	minionsChan := make(chan Minions)
	jobsChan := make(chan []Job)

	// Go routine on all simple data fetcher
	go f.Masters(masterChan)
	go f.Minions(minionsChan)
	go f.Jobs(jobsChan)

	// Treat data from channels
	for k, v := range (<-masterChan).status {
		ch <- prometheus.MustNewConstMetric(masterUp, prometheus.GaugeValue, map[bool]float64{true: 1, false: 0}[v], k)
	}

	ch <- prometheus.MustNewConstMetric(minionsCount, prometheus.GaugeValue, float64((<-minionsChan).count))

	// Check jobs status
	// Status to be checked : state.highstate / state.apply with Arguments = []
	// For one given Job check the Minions list and the Result for this minion
	var jobs_details []JobStatus
	jobsStatusChan := make(chan JobStatus, 1000)
	minion_last_event := make(map[string]uint64)

	for _, job := range <-jobsChan {
		if job.function == "state.highstate" || job.function == "state.apply" {
			wg.Add(1)
			go f.JobStatus(job.id, jobsStatusChan)
		}
	}
	wg.Wait()
	close(jobsStatusChan)

	for elem := range jobsStatusChan {
		jobs_details = append(jobs_details, elem)
	}

	for _, detail := range jobs_details {
		for _, minion := range detail.minions {
			id, _ := strconv.ParseUint(detail.id[0:14], 10, 64)
			if minion_last_event[minion] < id {
				minion_last_event[minion] = id
			}
		}
	}

	for _, detail := range jobs_details {
		for _, minion := range detail.minions {
			id, _ := strconv.ParseUint(detail.id[0:14], 10, 64)
			if minion_last_event[minion] == id {
				var status bool = detail.status[minion]

				ch <- prometheus.NewMetricWithTimestamp(detail.start_time,
					prometheus.MustNewConstMetric(jobsStatus, prometheus.GaugeValue, map[bool]float64{true: 1, false: 0}[status], minion, detail.function))
			}
		}
	}
}

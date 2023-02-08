package exporter

import (
	"strconv"

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

	var jobs_details []*JobStatus
	for _, job := range *jobs {
		if job.function == "state.highstate" || job.function == "state.apply" {
			// store the job ID if job.function and job.id and job.target
			job_status, err := f.JobStatus(job.id)
			if err != nil {
				log.WithFields(log.Fields{
					"saltUrl":      e.saltUrl,
					"saltUser":     e.saltPassword,
					"saltPassword": "***",
				}).Error(err)
			}

			jobs_details = append(jobs_details, job_status)

			/*for _, minion := range job_status.minions {
				var status bool = job_status.status[minion]
				// var retcode float64 = job_status.errors[minion]

				ch <- prometheus.NewMetricWithTimestamp(job_status.start_time,
						prometheus.MustNewConstMetric(jobsStatus, prometheus.GaugeValue, map[bool]float64{true: 1, false: 0}[status], minion, job_status.function))

			}*/
		}
	}

	minion_last_event := make(map[string]uint64)
	for _, detail := range jobs_details {
		for _, minion := range detail.minions {
			id, _ := strconv.ParseUint(detail.id[0:14], 10, 64)
			if minion_last_event[minion] < id {
				minion_last_event[minion] = id
			}
		}
	}

	log.Debug(minion_last_event)

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

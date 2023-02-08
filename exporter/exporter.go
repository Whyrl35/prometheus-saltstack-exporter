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

var wgGlobal sync.WaitGroup
var wg sync.WaitGroup
var masterUp = prometheus.NewDesc(prometheus.BuildFQName("saltstack", "", "master_up"), "Master in up(1) or down(0)", []string{"master"}, nil)
var minionsCount = prometheus.NewDesc(prometheus.BuildFQName("saltstack", "", "minions_count"), "Number of minions declared in salt", nil, nil)
var jobsStatus = prometheus.NewDesc(prometheus.BuildFQName("saltstack", "", "job_status"), "Job status", []string{"minion", "function"}, nil)

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- masterUp
	ch <- minionsCount
	ch <- jobsStatus
}

func CollectMaster(ch chan<- prometheus.Metric, f *Fetcher) {
	masterChan := make(chan Masters)

	go f.Masters(masterChan)

	for k, v := range (<-masterChan).status {
		ch <- prometheus.MustNewConstMetric(masterUp, prometheus.GaugeValue, map[bool]float64{true: 1, false: 0}[v], k)
	}

	defer wgGlobal.Done()
}

func CollectMinions(ch chan<- prometheus.Metric, f *Fetcher) {
	minionsChan := make(chan Minions)

	go f.Minions(minionsChan)

	ch <- prometheus.MustNewConstMetric(minionsCount, prometheus.GaugeValue, float64((<-minionsChan).count))

	defer wgGlobal.Done()
}

func CollectJobInfos(ch chan<- prometheus.Metric, f *Fetcher) {
	jobsChan := make(chan []Job)

	go f.Jobs(jobsChan)

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

	defer wgGlobal.Done()
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
	wgGlobal.Add(3)
	go CollectMaster(ch, f)
	go CollectMinions(ch, f)
	go CollectJobInfos(ch, f)
	wgGlobal.Wait()
}

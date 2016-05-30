package resqueExporter

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/redis.v3"
)

const namespace = "resque"

type exporter struct {
	config         *Config
	mut            *sync.Mutex
	scrapeFailures prometheus.Counter
	processed      prometheus.Gauge
	failed         prometheus.Gauge
	queueStatus    *prometheus.GaugeVec
	totalWorkers   prometheus.Gauge
	activeWorkers  prometheus.Gauge
	idleWorkers    prometheus.Gauge
	gate           bool
}

func newExporter(config *Config) (*exporter, error) {
	e := &exporter{
		mut:    new(sync.Mutex),
		config: config,
		gate:   true,
		queueStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "jobs_in_queue",
				Help:      "Number of remained jobs in queue",
			},
			[]string{"queue_name"},
		),
		processed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "processed",
			Help:      "Number of processed jobs",
		}),
		failed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "failed",
			Help:      "Number of failed jobs",
		}),
		scrapeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrape_failures_total",
			Help:      "Number of errors while scraping resque.",
		}),
		totalWorkers: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "total_workers",
			Help:      "Number of workers",
		}),
		activeWorkers: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "active_workers",
			Help:      "Number of active workers",
		}),
		idleWorkers: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "idle_workers",
			Help:      "Number of idle workers",
		}),
	}
	go e.runTicker()

	return e, nil
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	e.scrapeFailures.Describe(ch)
	e.queueStatus.Describe(ch)
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	resqueNamespace := e.config.ResqueNamespace

	if e.gate {
		e.mut.Lock() // To protect metrics from concurrent collects.
		defer e.mut.Unlock()

		redisConfig := e.config.Redis
		redisOpt := &redis.Options{
			Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
			Password: redisConfig.Password,
			DB:       redisConfig.DB,
		}
		redis := redis.NewClient(redisOpt)
		defer redis.Close()

		queues, err := redis.SMembers(fmt.Sprintf("%s:queues", resqueNamespace)).Result()
		if err != nil {
			e.incrementFailures(ch)
			return
		}

		for _, q := range queues {
			n, err := redis.LLen(fmt.Sprintf("%s:queue:%s", resqueNamespace, q)).Result()
			if err != nil {
				e.incrementFailures(ch)
				return
			}
			e.queueStatus.WithLabelValues(q).Set(float64(n))
		}

		processed, err := redis.Get(fmt.Sprintf("%s:stat:processed", resqueNamespace)).Result()
		if err != nil {
			e.incrementFailures(ch)
			return
		}
		processedCnt, _ := strconv.ParseFloat(processed, 64)
		e.processed.Set(processedCnt)

		failed, err := redis.Get(fmt.Sprintf("%s:stat:failed", resqueNamespace)).Result()
		if err != nil {
			e.incrementFailures(ch)
			return
		}
		failedCnt, _ := strconv.ParseFloat(failed, 64)
		e.failed.Set(failedCnt)

		workers, err := redis.SMembers(fmt.Sprintf("%s:workers", resqueNamespace)).Result()
		if err != nil {
			e.incrementFailures(ch)
			return
		}
		e.totalWorkers.Set(float64(len(workers)))

		activeWorkers := 0
		idleWorkers := 0
		for _, w := range workers {
			_, err := redis.Get(fmt.Sprintf("%s:worker:%s", resqueNamespace, w)).Result()
			if err == nil {
				activeWorkers++
			} else {
				idleWorkers++
			}
		}
		e.activeWorkers.Set(float64(activeWorkers))
		e.idleWorkers.Set(float64(idleWorkers))

		if e.config.GuardIntervalMillis > 0 {
			e.gate = false
		}
	}

	e.queueStatus.Collect(ch)
	e.processed.Collect(ch)
	e.failed.Collect(ch)
	e.totalWorkers.Collect(ch)
	e.activeWorkers.Collect(ch)
	e.idleWorkers.Collect(ch)
}

func (e *exporter) incrementFailures(ch chan<- prometheus.Metric) {
	e.scrapeFailures.Inc()
	e.scrapeFailures.Collect(ch)
}

func (e *exporter) runTicker() {
	guardIntervalMillis := e.config.GuardIntervalMillis
	if guardIntervalMillis > 0 {
		guardIntervalDuration := time.Duration(guardIntervalMillis) * time.Millisecond
		for _ = range time.Tick(guardIntervalDuration) {
			e.gate = true
		}
	}
}

package resqueExporter

import (
	"fmt"
	"strconv"
	"sync"

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
}

func newExporter(config *Config) (*exporter, error) {
	return &exporter{
		mut:    new(sync.Mutex),
		config: config,
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
	}, nil
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	e.scrapeFailures.Describe(ch)
	e.queueStatus.Describe(ch)
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
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

	queues, err := redis.SMembers(fmt.Sprintf("%s:queues", e.config.ResqueNamespace)).Result()
	if err != nil {
		e.incrementFailures(ch)
		return
	}

	for _, q := range queues {
		n, err := redis.LLen(fmt.Sprintf("%s:queue:%s", e.config.ResqueNamespace, q)).Result()
		if err != nil {
			e.incrementFailures(ch)
			return
		}
		e.queueStatus.WithLabelValues(q).Set(float64(n))
	}

	processed, err := redis.Get(fmt.Sprintf("%s:stat:processed", e.config.ResqueNamespace)).Result()
	if err != nil {
		e.incrementFailures(ch)
		return
	}
	processedCnt, _ := strconv.ParseFloat(processed, 64)
	e.processed.Set(processedCnt)

	failed, err := redis.Get(fmt.Sprintf("%s:stat:failed", e.config.ResqueNamespace)).Result()
	if err != nil {
		e.incrementFailures(ch)
		return
	}
	failedCnt, _ := strconv.ParseFloat(failed, 64)
	e.failed.Set(failedCnt)

	e.queueStatus.Collect(ch)
	e.processed.Collect(ch)
	e.failed.Collect(ch)
}

func (e *exporter) incrementFailures(ch chan<- prometheus.Metric) {
	e.scrapeFailures.Inc()
	e.scrapeFailures.Collect(ch)
}

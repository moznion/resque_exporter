package resqueExporter

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/redis.v3"
)

const namespace = "resque"

type exporter struct {
	config         *Config
	mut            *sync.Mutex
	scrapeFailures prometheus.Counter
	counters       map[string]prometheus.Counter
}

func newExporter(config *Config) (*exporter, error) {
	return &exporter{
		mut:      new(sync.Mutex),
		config:   config,
		counters: make(map[string]prometheus.Counter),
		scrapeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrape_failures_total",
			Help:      "Number of errors while scraping resque.",
		}),
	}, nil
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	e.scrapeFailures.Describe(ch)
	for _, counter := range e.counters {
		counter.Describe(ch)
	}
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

	queues, err := redis.SMembers(fmt.Sprintf("%s:queues", e.config.resqueNamespace)).Result()
	if err != nil {
		e.incrementFailures(ch)
		return
	}

	for _, q := range queues {
		n, err := redis.LLen(fmt.Sprintf("%s:queue:%s", e.config.resqueNamespace, q)).Result()
		if err != nil {
			e.incrementFailures(ch)
			return
		}

		if _, existed := e.counters[q]; !existed {
			e.counters[q] = prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: namespace,
				Name:      fmt.Sprintf(`jobs_in_queue{queue_name="%s"}`, q),
				Help:      fmt.Sprintf("Number of remained jobs of %s", q),
			})
		}
		counter := e.counters[q]
		counter.Set(float64(n))
		counter.Collect(ch)
	}
}

func (e *exporter) incrementFailures(ch chan<- prometheus.Metric) {
	e.scrapeFailures.Inc()
	e.scrapeFailures.Collect(ch)
}

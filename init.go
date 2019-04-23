package microbot

import (
	"container/list"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	duration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "microbot_http_request_duration_milliseconds",
			Help: "Summary of http request duration in milliseconds.",
		},
		[]string{"handler", "status", "method", "ip_type"},
	)

	requests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "microbot_http_request_total",
			Help: "Total number of http requests.",
		},
		[]string{"handler", "status", "method", "ip_type"},
	)

	panics = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "microbot_panic_total",
			Help: "Total number of panic.",
		})

	accessibility = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "microbot_db_accessibility_total",
			Help: "Total number of DB accessibility.",
		},
		[]string{"status"},
	)
)

func init() {
	keyEventList = KeyEventList{
		data: list.New(),
		max:  DefaultListMax,
	}

	prometheus.MustRegister(duration)
	prometheus.MustRegister(requests)
	prometheus.MustRegister(panics)
	prometheus.MustRegister(accessibility)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for _, r := range PingDB() {
					status := "ok"
					if r.err != nil {
						status = "error"
					}
					accessibility.WithLabelValues(status).Inc()
				}
			}
		}
	}()
}

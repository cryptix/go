/*
Package metrics implements a negroni middleware to caputre http request/response metrics.

Three things are collected:

	A counter vector (partitioned by status code, method and path)
	A summary of request latency
	A summary of request size


remember to expose it:

	http.ListenAndServe("localhost:9999", prometheus.Handler())
*/
package metrics

import (
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/cryptix/go/logging"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/errgo.v1"
)

var log = logging.Logger("")

type HTTPMetric struct {
	reqs    *prometheus.CounterVec
	latency prometheus.Summary
	size    prometheus.Summary
}

func NewNegroni(name string) *HTTPMetric {
	var m HTTPMetric
	m.reqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "negroni_requests_total",
			Help:        "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(m.reqs)
	m.latency = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:        "negroni_request_latency",
		Help:        "How long it took to process the request",
		ConstLabels: prometheus.Labels{"service": name},
	})
	prometheus.MustRegister(m.latency)
	m.size = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:        "negroni_request_size",
		Help:        "How big the response was",
		ConstLabels: prometheus.Labels{"service": name},
	})
	prometheus.MustRegister(m.size)
	return &m
}

func (m *HTTPMetric) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	next(rw, r)
	res, ok := rw.(negroni.ResponseWriter)
	if !ok {
		log.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"error":  errgo.New("handler type assertion"),
		}).Error("HTTPMetric failed")
		return
	}
	m.reqs.WithLabelValues(http.StatusText(res.Status()), r.Method, r.URL.Path).Inc()
	m.latency.Observe(time.Since(start).Seconds())
	m.size.Observe(float64(res.Size()))
}

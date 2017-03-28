package server

import (
	"net/http"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/metrics"
)

func NewMetricsServer(config config.Config, metrics []metrics.Metrics) *http.Server {
	return &http.Server{
		Handler: newMetricsHandler(config, metrics),
	}

}

type metricsHandler struct {
	config  config.Config
	metrics []metrics.Metrics
}

func newMetricsHandler(config config.Config, metrics []metrics.Metrics) metricsHandler {
	return metricsHandler{
		config:  config,
		metrics: metrics,
	}
}

func (mh metricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, m := range mh.metrics {
		m.WriteJSON(w)
	}
	w.WriteHeader(http.StatusOK)
	return
}

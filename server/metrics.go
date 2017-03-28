package server

import (
	"encoding/json"
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

type metricsResponse struct {
	Clients int64 `json:"clients"`
}

func (mh metricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics := metricsResponse{}
	for _, m := range mh.metrics {
		metrics.Clients += m.GetClientCount()
	}
	json.NewEncoder(w).Encode(metrics)
	w.WriteHeader(http.StatusOK)
	return
}

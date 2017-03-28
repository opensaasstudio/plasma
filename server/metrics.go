package server

import (
	"encoding/json"
	"net/http"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/metrics"
)

func NewMetricsServer(config config.Config, registry metrics.Registry) *http.Server {
	return &http.Server{
		Handler: newMetricsHandler(config, registry),
	}

}

type metricsHandler struct {
	config   config.Config
	registry metrics.Registry
}

func newMetricsHandler(config config.Config, registry metrics.Registry) metricsHandler {
	return metricsHandler{
		config:   config,
		registry: registry,
	}
}

func (mh metricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(mh.registry)
	return
}

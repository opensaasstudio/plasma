package server

import (
	"net/http"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/metrics"
	"go.uber.org/zap"
)

type metricsHandler struct {
	accessLogger *zap.Logger
	errorLogger  *zap.Logger
	config       config.Config
	mux          *http.ServeMux
}

func (h metricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func NewMetricsHandler(opt Option) metricsHandler {

	h := metricsHandler{
		accessLogger: opt.AccessLogger,
		errorLogger:  opt.ErrorLogger,
		config:       opt.Config,
		mux:          http.NewServeMux(),
	}

	h.mux.HandleFunc("/metrics/go", h.metricsGo)
	h.mux.HandleFunc("/metrics/plasma", h.metricsPlasma)
	return h
}

func (h *metricsHandler) metricsGo(w http.ResponseWriter, r *http.Request) {
	metrics.GoStatsHandler(w, r)
}

func (h *metricsHandler) metricsPlasma(w http.ResponseWriter, r *http.Request) {
	metrics.PlasmaStatsHandler(w, r)
}

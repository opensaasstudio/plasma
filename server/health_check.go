package server

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/log"
	"gopkg.in/redis.v5"
)

func NewHealthCheckServer(accessLogger, errorLogger *zap.Logger, config config.Config) *http.Server {
	return &http.Server{
		Handler: newHealthCheckHandler(accessLogger, errorLogger, config),
	}
}

func newHealthCheckHandler(accessLogger, errorLogger *zap.Logger, config config.Config) healthCheckHandler {
	return healthCheckHandler{
		config:       config,
		accessLogger: accessLogger,
		errorLogger:  errorLogger,
	}
}

type healthCheckHandler struct {
	config       config.Config
	accessLogger *zap.Logger
	errorLogger  *zap.Logger
}

func checkRedis(config config.Redis) error {
	addr := config.Addr
	opt := &redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       config.DB,
	}
	client := redis.NewClient(opt)
	if err := client.Ping().Err(); err != nil {
		return err
	}
	return nil
}

func (h healthCheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.accessLogger.Info("healthCheck", log.HttpRequestToLogFields(r)...)
	if h.config.Subscriber.Type == "redis" {
		if err := checkRedis(h.config.Subscriber.Redis); err != nil {
			h.errorLogger.Error("failed to connect redis",
				zap.Error(err),
				zap.Object("redis", h.config.Subscriber.Redis),
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	return
}

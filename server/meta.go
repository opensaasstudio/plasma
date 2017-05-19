package server

import (
	"encoding/json"
	"net/http"

	redis "gopkg.in/redis.v5"

	"go.uber.org/zap"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/event"
	"github.com/openfresh/plasma/log"
	"github.com/openfresh/plasma/metrics"
)

type metaHandler struct {
	accessLogger *zap.Logger
	errorLogger  *zap.Logger
	config       config.Config
	mux          *http.ServeMux
}

func (h metaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func NewMetaHandler(opt Option) metaHandler {
	h := metaHandler{
		accessLogger: opt.AccessLogger,
		errorLogger:  opt.ErrorLogger,
		config:       opt.Config,
		mux:          http.NewServeMux(),
	}

	if h.config.Debug {
		h.mux.HandleFunc("/debug", h.debug)
	}
	h.mux.HandleFunc("/hc", h.healthCheck)
	h.mux.HandleFunc("/metrics", h.metrics)

	return h
}

func checkRedis(config config.Redis) error {
	addr := config.Addr
	opt := &redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       config.DB,
	}
	client := redis.NewClient(opt)
	return client.Ping().Err()
}

func (h *metaHandler) debug(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "template/debug.html")
		return
	case http.MethodPost:
		p := event.Payload{}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			h.errorLogger.Info("failed to decode json in debug endpoint",
				zap.Error(err),
			)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// publish to Redis for testing
		redisConf := h.config.Subscriber.Redis
		opt := &redis.Options{
			Addr:     redisConf.Addr,
			Password: redisConf.Password,
			DB:       redisConf.DB,
		}
		b, err := json.Marshal(p)
		if err != nil {
			h.errorLogger.Error("failed to marshal json in debug endpoint",
				zap.Error(err),
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		client := redis.NewClient(opt)
		channel := h.config.Subscriber.Redis.Channels[0]
		if err := client.Publish(channel, string(b)).Err(); err != nil {
			h.errorLogger.Error("failed to publlish to redis",
				zap.Object("redis", redisConf),
				zap.Error(err),
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	return
}

func (h *metaHandler) healthCheck(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	if h.config.Subscriber.Type == "redis" {
		if err := checkRedis(h.config.Subscriber.Redis); err != nil {
			h.errorLogger.Error("failed to connect redis",
				zap.Error(err),
				zap.Object("redis", h.config.Subscriber.Redis),
			)
			status = http.StatusInternalServerError
		}
	}

	w.WriteHeader(status)

	fields := append(log.HTTPRequestToLogFields(r), zap.Int("status", status))
	h.accessLogger.Info("healthCheck", fields...)
}

func (h *metaHandler) metrics(w http.ResponseWriter, r *http.Request) {
	metrics.HTTPHandler(w, r)
}

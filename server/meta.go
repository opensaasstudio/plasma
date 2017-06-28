package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-redis/redis"

	"go.uber.org/zap"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/event"
	"github.com/openfresh/plasma/log"
)

type metaHandler struct {
	accessLogger *zap.Logger
	errorLogger  *zap.Logger
	config       config.Config
	mux          *http.ServeMux
	redisClient  *redis.Client
}

func (h metaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func NewMetaHandler(opt Option) metaHandler {

	redisConf := opt.Config.Subscriber.Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisConf.Addr,
		Password: redisConf.Password,
		DB:       redisConf.DB,
	})

	h := metaHandler{
		accessLogger: opt.AccessLogger,
		errorLogger:  opt.ErrorLogger,
		config:       opt.Config,
		mux:          http.NewServeMux(),
		redisClient:  redisClient,
	}

	if h.config.Debug {
		h.mux.HandleFunc("/debug", h.debug)
	}
	h.mux.HandleFunc("/hc", h.healthCheck)

	return h
}

func (h *metaHandler) checkRedis() error {
	return h.redisClient.Ping().Err()
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
		b, err := json.Marshal(p)
		if err != nil {
			h.errorLogger.Error("failed to marshal json in debug endpoint",
				zap.Error(err),
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		channel := h.config.Subscriber.Redis.Channels[0]
		if err := h.redisClient.Publish(channel, string(b)).Err(); err != nil {
			h.errorLogger.Error("failed to publlish to redis",
				zap.Object("redis", h.config.Subscriber.Redis),
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
		if err := h.checkRedis(); err != nil {
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

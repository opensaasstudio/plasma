package subscriber

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"go.uber.org/zap"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/event"
	"github.com/openfresh/plasma/pubsub"
	"gopkg.in/redis.v5"
)

type Redis struct {
	config      config.Redis
	pubsub      pubsub.PubSuber
	client      *redis.Client
	errorLogger *zap.Logger
}

func newRedis(pb pubsub.PubSuber, errorLogger *zap.Logger, config config.Config) (Subscriber, error) {
	redisConf := config.Subscriber.Redis
	addr := redisConf.Addr
	opt := &redis.Options{
		Addr:     addr,
		Password: redisConf.Password,
		DB:       redisConf.DB,
	}

	client := redis.NewClient(opt)
	return &Redis{
		client:      client,
		config:      redisConf,
		pubsub:      pb,
		errorLogger: errorLogger,
	}, nil
}

func (r *Redis) isNetworkError(err error) bool {
	// NOTE: https://github.com/go-redis/redis/blob/v5.2.9/internal/errors.go#L24-L30
	if err == io.EOF {
		return true
	}
	_, ok := err.(net.Error)
	return ok
}

func (r *Redis) receiveMessage(pb *redis.PubSub) (*redis.Message, error) {
	var errNum int
	for {
		msgi, err := pb.ReceiveTimeout(r.config.Timeout)
		if err != nil {
			if !r.isNetworkError(err) {
				return nil, err
			}

			errNum++
			if 1 < errNum {
				r.errorLogger.Info("failed to receive message from redis continuously",
					zap.Error(err),
					zap.Int("errorCount", errNum),
					zap.Int("maxErrorCount", r.config.MaxRetry),
				)
			}

			if errNum >= r.config.MaxRetry {
				return nil, err
			}

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				if err := pb.Ping(); err != nil {
					r.errorLogger.Info("redis ping error",
						zap.Int("errorCount", errNum),
						zap.Object("config", r.config),
						zap.Error(err),
					)
				}
			}
			time.Sleep(r.config.RetryInterval)
			continue
		}

		errNum = 0

		switch msg := msgi.(type) {
		case *redis.Subscription:
			// Ignore.
		case *redis.Pong:
			// Ignore.
		case *redis.Message:
			return msg, nil
		default:
			return nil, fmt.Errorf("redis: unknown message: %T", msgi)
		}
	}
}

func (r *Redis) Subscribe() error {
	ps, err := r.client.Subscribe(r.config.Channels...)
	if err != nil {
		return err
	}
	defer ps.Close()
	for {
		msg, err := r.receiveMessage(ps)
		if err != nil {
			switch r.config.OverMaxRetryBehavior.Type {
			case config.OverMaxRetryBehaviorAlive:
				r.errorLogger.Info("reset error count and retry",
					zap.Error(err),
					zap.Object("config", r.config),
				)
				continue
			case config.OverMaxRetryBehaviorDie:
				r.errorLogger.Fatal("over max retry count",
					zap.Error(err),
					zap.Object("config", r.config),
				)
			default:
				r.errorLogger.Fatal("unknown behavior",
					zap.Error(err),
					zap.Object("config", r.config),
				)
			}
		}

		var payload event.Payload
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			r.errorLogger.Info("failed to unmarhsal to json when subscribing redis",
				zap.Error(err),
				zap.String("payload", msg.Payload),
				zap.String("channel", msg.Channel),
			)
			continue
		}
		r.pubsub.Publish(payload)
	}
}

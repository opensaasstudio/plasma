package subscriber

import (
	"fmt"

	"go.uber.org/zap"

	conf "github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/pubsub"
	"github.com/pkg/errors"
)

type Subscriber interface {
	Subscribe() error
}

func New(pb pubsub.PubSuber, errorLogger *zap.Logger, config conf.Config) (Subscriber, error) {
	var subscriber Subscriber
	var err error

	var f func(pubsub.PubSuber, *zap.Logger, conf.Config) (Subscriber, error)
	switch config.Subscriber.Type {
	case "mock":
		f = newMock
	case "redis":
		f = newRedis
	case "api":
		f = newApi
	default:
		return subscriber, fmt.Errorf("can't get such %s type subscriber", config.Subscriber.Type)
	}

	subscriber, err = f(pb, errorLogger, config)
	if err != nil {
		return subscriber, errors.Wrap(err, "failed to create a new subscriber")
	}

	return subscriber, nil
}

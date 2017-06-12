package subscriber

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/go-redis/redis"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/event"
	"github.com/openfresh/plasma/log"
	"github.com/openfresh/plasma/pubsub"

	"github.com/stretchr/testify/assert"
)

func TestRedisReceiveMessage(t *testing.T) {
	assert := assert.New(t)

	pb := pubsub.NewPubSub()

	el, err := log.NewLogger(config.Log{
		Out: "discard",
	})
	assert.Nil(err)

	baseRedisConf := config.Redis{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Channels: config.Channels([]string{"plasma_test"}),
		OverMaxRetryBehavior: config.OverMaxRetryBehavior{
			Type: "alive",
		},
		MaxRetry:      3,
		Timeout:       100 * time.Millisecond,
		RetryInterval: 100 * time.Millisecond,
	}

	r, err := newRedis(pb, el, config.Config{
		Subscriber: config.Subscriber{
			Redis: baseRedisConf,
		},
	})
	assert.Nil(err)

	payload := event.Payload{
		Meta: event.MetaData{
			Type: "test",
		},
		Data: json.RawMessage([]byte(`{"data":"programId:1234"}`)),
	}

	pb.Subscribe(func(p event.Payload) {
		assert.Equal(payload, p)
	})
	go r.Subscribe()

	b, err := json.Marshal(payload)
	assert.Nil(err)
	opt := &redis.Options{
		Addr:     baseRedisConf.Addr,
		Password: baseRedisConf.Password,
		DB:       baseRedisConf.DB,
	}
	err = redis.NewClient(opt).Publish(baseRedisConf.Channels[0], string(b)).Err()
	assert.Nil(err)

}

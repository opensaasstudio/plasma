package subscriber

import (
	"math/rand"
	"time"

	"go.uber.org/zap"

	"encoding/json"
	"fmt"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/event"
	"github.com/openfresh/plasma/pubsub"
)

type Mock struct {
	config config.Config
	pubsub pubsub.PubSuber
}

var fakeEventTypes = []string{"poll", "program", "other"}

func genFakeEventPayload() event.Payload {
	eventType := fakeEventTypes[rand.Intn(len(fakeEventTypes))]

	var data json.RawMessage
	switch eventType {
	case "poll":
		data = json.RawMessage(`{"1": "One", "2":"Two", "3":"Three", "4":"Four"}`)
	case "program":
		data = json.RawMessage(fmt.Sprintf(`{"programId": %d}`, rand.Intn(100000)))
	}

	return event.Payload{
		Meta: event.MetaData{
			Type: eventType,
		},
		Data: data,
	}
}

func newMock(pb pubsub.PubSuber, errorLogger *zap.Logger, config config.Config) (Subscriber, error) {
	rand.Seed(time.Now().UnixNano())
	return &Mock{
		config: config,
		pubsub: pb,
	}, nil
}

func (m *Mock) Subscribe() error {
	t := time.NewTicker(m.config.Subscriber.Mock.Interval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			payload := genFakeEventPayload()
			m.pubsub.Publish(payload)
		}
	}

	return nil
}

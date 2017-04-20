package pubsub

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/openfresh/plasma/event"
)

func createPayload(t, data string) event.Payload {
	return event.Payload{
		Meta: event.MetaData{
			Type: t,
		},
		Data: json.RawMessage(data),
	}
}

func TestPubSub(t *testing.T) {
	p := createPayload("test", `{"dummy": "data"}`)

	pb := NewPubSub()

	f := func(payload event.Payload) {
		assert.Equal(t, p, payload, "should be equal")
	}
	assert.NoError(t, pb.Subscribe(f))
	pb.Publish(p)
}

package pubsub

import (
	"github.com/mattn/go-pubsub"
	"github.com/openfresh/plasma/event"
)

type PubSuber interface {
	Publish(payload event.Payload)
	Subscribe(f func(paylaod event.Payload)) error
}

type PubSub struct {
	pubsub *pubsub.PubSub
}

func NewPubSub() PubSuber {
	return &PubSub{
		pubsub: pubsub.New(),
	}
}

func (d *PubSub) Publish(payload event.Payload) {
	d.pubsub.Pub(payload)
}

func (d *PubSub) Subscribe(f func(payload event.Payload)) error {
	return d.pubsub.Sub(f)
}

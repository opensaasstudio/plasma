package server

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/event"
	"github.com/openfresh/plasma/manager"
	"github.com/openfresh/plasma/protobuf"
	"github.com/openfresh/plasma/pubsub"
	"github.com/pkg/errors"
)

type StreamServer struct {
	clientManager *manager.ClientManager
	newClients    chan manager.Client
	removeClients chan manager.Client
	payloads      chan event.Payload
	pubsub        pubsub.PubSuber
	accessLogger  *zap.Logger
	errorLogger   *zap.Logger
}

func NewStreamServer(pubsub pubsub.PubSuber, accessLogger *zap.Logger, errorLogger *zap.Logger, config config.Config) *StreamServer {
	ss := &StreamServer{
		clientManager: manager.NewClientManager(),
		newClients:    make(chan manager.Client),
		removeClients: make(chan manager.Client),
		payloads:      make(chan event.Payload),
		pubsub:        pubsub,
		accessLogger:  accessLogger,
		errorLogger:   errorLogger,
	}
	ss.pubsub.Subscribe(func(payload event.Payload) {
		ss.payloads <- payload
	})
	ss.Run()

	return ss
}

func (ss *StreamServer) Run() {
	go func() {
		for {
			select {
			case client := <-ss.newClients:
				ss.clientManager.AddClient(client)
			case client := <-ss.removeClients:
				ss.clientManager.RemoveClient(client)
			case payload := <-ss.payloads:
				ss.clientManager.SendPayload(payload)
			}
		}
	}()
}

func (ss *StreamServer) Events(request *proto.Request, es proto.StreamService_EventsServer) error {
	ss.accessLogger.Info("gRPC",
		zap.Array("request-events", zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
			for _, e := range request.Events {
				enc.AppendString(e.Type)
			}
			return nil
		})),
	)
	if request == nil || request.Events == nil {
		return errors.New("request can't be nil")
	}

	l := len(request.Events)
	events := make([]string, l)
	for i := 0; i < l; i++ {
		events[i] = request.Events[i].Type
	}

	client := manager.NewClient(events)
	ss.newClients <- client

	for {
		select {
		case pl, open := <-client.ReceivePayload():
			if !open {
				return nil
			}
			eventType := proto.EventType{pl.Meta.Type}
			p := &proto.Payload{
				EventType: &eventType,
				Data:      string(pl.Data),
			}
			if err := es.Send(p); err != nil {
				ss.errorLogger.Error("failed to send message",
					zap.Error(err),
					zap.Object("payload", pl),
				)
				ss.removeClients <- client
				return err
			}
		case <-es.Context().Done():
			ss.removeClients <- client
			return nil
		}

	}

	return nil
}

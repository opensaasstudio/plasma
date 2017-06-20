package server

import (
	"encoding/json"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/log"
	"github.com/openfresh/plasma/pubsub"

	"github.com/openfresh/plasma/event"
	"github.com/openfresh/plasma/protobuf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func eventType(et string) *proto.EventType {
	return &proto.EventType{Type: et}
}

func TestGRPCEvents(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	pb := pubsub.NewPubSub()

	logger, err := log.NewLogger(config.Log{
		Out:   "discard",
		Level: "error",
	})
	require.NoError(err)

	config := config.Config{Port: "8082"}
	grpcServer, err := NewGRPCServer(Option{
		PubSuber:     pb,
		AccessLogger: logger,
		ErrorLogger:  logger,
		Config:       config,
	})
	require.NoError(err)

	l, err := net.Listen("tcp", ":"+config.Port)
	require.NoError(err)
	go grpcServer.Serve(l)

	events := []event.Payload{
		{
			Meta: event.MetaData{
				Type: "program:1234:poll",
			},
			Data: json.RawMessage(`{"poll": {"1": "One", "2": "Two", "3": "Three"}}`),
		},
		{
			Meta: event.MetaData{
				Type: "program:1234:views",
			},
			Data: json.RawMessage(`{"views": 55301}`),
		},
		{
			Meta: event.MetaData{
				Type: "program:1234:annotation",
			},
			Data: json.RawMessage(`{"text": "hello world"}`),
		},
	}

	cases := []struct {
		req         proto.Request
		expectCount int
		actualCount int
	}{
		{
			req: proto.Request{
				Events: []*proto.EventType{
					eventType("program:1234:views"),
					eventType("program:1234:poll"),
				},
			},
			expectCount: 2,
		},
		{
			req: proto.Request{
				Events: []*proto.EventType{
					eventType("program:1234:annotation"),
				},
			},
			expectCount: 1,
		},
		{
			req: proto.Request{
				Events: []*proto.EventType{
					eventType("program:1234"),
				},
			},
			expectCount: 3,
		},
	}

	dummyEvent := event.Payload{
		Meta: event.MetaData{
			Type: "dummy",
		},
		Data: json.RawMessage(""),
	}

	// for test
	for i := range cases {
		cases[i].req.Events = append(cases[i].req.Events, eventType(dummyEvent.Meta.Type))
	}

	wg := sync.WaitGroup{}
	var readyClient int32
	for i := range cases {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			conn, err := grpc.Dial(":"+config.Port, grpc.WithInsecure(), grpc.WithTimeout(5*time.Second))
			require.NoError(err)
			defer conn.Close()
			client := proto.NewStreamServiceClient(conn)
			ctx := context.Background()
			ss, err := client.Events(ctx)
			if err := ss.Send(&cases[i].req); err != nil {
				require.NoError(err)
			}

			require.NoError(err)
			isFirst := true
			for cases[i].expectCount != cases[i].actualCount {
				resp, err := ss.Recv()
				require.NoError(err)

				// for checking ready to receive message
				if isFirst {
					isFirst = false
					atomic.AddInt32(&readyClient, 1)
				}
				if resp.GetEventType().GetType() == dummyEvent.Meta.Type {
					continue
				}

				flag := false
				for _, e := range cases[i].req.Events {
					if strings.HasPrefix(resp.GetEventType().GetType(), e.GetType()) {
						flag = true
						break
					}
				}
				assert.True(flag)
				cases[i].actualCount++
				js := make(map[string]interface{})
				isJSON := json.Unmarshal([]byte(resp.Data), &js) == nil
				assert.True(isJSON)
			}
		}(i)
	}

	// keep sending dummy messages until all clients are ready to receive messages
	for int(atomic.LoadInt32(&readyClient)) != len(cases) {
		pb.Publish(dummyEvent)
	}

	for _, e := range events {
		pb.Publish(e)
	}

	wg.Wait()
	grpcServer.GracefulStop()

	for _, c := range cases {
		assert.Equal(c.expectCount, c.actualCount)
	}
}

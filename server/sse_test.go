package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/event"
	"github.com/openfresh/plasma/log"
	"github.com/openfresh/plasma/pubsub"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsNotSupportSSE(t *testing.T) {
	cases := []struct {
		Browser         string
		UA              string
		IsNotSupportSSE bool
	}{
		{
			"Edge",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.79 Safari/537.36 Edge/14.14393",
			true,
		},
		{
			"IE",
			"Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; rv:11.0) like Gecko",
			true,
		},
		{
			"Chrome",
			"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 Safari/537.36",
			false,
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.IsNotSupportSSE, isNotSupportSSE(c.UA), c.Browser)
	}
}

func setUpSSEHandler(t *testing.T, pb pubsub.PubSuber, origin string) sseHandler {
	logger, err := log.NewLogger(config.Log{
		Out:   "discard",
		Level: "error",
	})
	require.NoError(t, err)

	config := config.Config{
		SSE: config.ServerSentEvent{
			EventQuery: "eventType",
			Retry:      2000,
		},
		Origin: origin,
	}

	opt := Option{
		PubSuber:     pb,
		AccessLogger: logger,
		ErrorLogger:  logger,
		Config:       config,
	}
	handler := newHandler(opt)

	return handler
}

func TestSSEHandler(t *testing.T) {
	assert := assert.New(t)
	pb := pubsub.NewPubSub()

	origin := "*.test.com"
	handler := setUpSSEHandler(t, pb, origin)
	server := httptest.NewServer(handler)
	defer server.Close()

	events := []event.Payload{
		{
			Meta: event.MetaData{
				Type: "program:1234:poll",
			},
			Data: json.RawMessage(`{"poll": {"1": "One", "2": "Two", "3": "Three"}}`),
		},
		{
			Meta: event.MetaData{
				Type: "program:1234:annotation",
			},
			Data: json.RawMessage(`{"text": "hello world"}`),
		},
		{
			Meta: event.MetaData{
				Type: "program:1234:views",
			},
			Data: json.RawMessage(`{"views": 55301}`),
		},
	}

	cases := []struct {
		events      []string
		expectCount int
		actualCount int
	}{
		{
			events: []string{
				"program:1234:views",
				"program:1234:poll",
			},
			expectCount: 2,
		},
		{
			events: []string{
				"program:1234:annotation",
			},
			expectCount: 1,
		},
		{
			events: []string{
				"program:1234",
			},
			expectCount: 3,
		},
	}

	dummyEvent := event.Payload{
		Meta: event.MetaData{
			Type: "dummy",
		},
		Data: json.RawMessage(`{"dummy": true}`),
	}

	// for test
	for i := range cases {
		cases[i].events = append(cases[i].events, dummyEvent.Meta.Type)
	}

	wg := sync.WaitGroup{}
	var readyClient int32
	for i := range cases {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// create request
			eventReq := strings.Join(cases[i].events, ",")
			url := fmt.Sprintf("%s/events?eventType=%s", server.URL, eventReq)
			resp, err := http.Get(url)
			require.NoError(t, err)
			defer resp.Body.Close()

			isFirst := true
			for cases[i].expectCount != cases[i].actualCount {
				data := readData(t, resp.Body)
				if len(data) == 0 {
					continue
				}

				var p event.Payload
				require.NoError(t, json.Unmarshal(data, &p))

				// for checking ready to receive message
				if isFirst {
					isFirst = false
					atomic.AddInt32(&readyClient, 1)
				}
				if p.Meta.Type == dummyEvent.Meta.Type {
					continue
				}

				flag := false
				for _, e := range cases[i].events {
					if strings.HasPrefix(p.Meta.Type, e) {
						flag = true
						break
					}
				}

				assert.True(flag)
				cases[i].actualCount++

				assert.Equal(resp.Header.Get("Access-Control-Allow-Origin"), origin)

				js := make(map[string]interface{})
				isJSON := json.Unmarshal([]byte(p.Data), &js) == nil
				assert.True(isJSON)
			}

		}(i)
	}

	// keep sending dummy messages until all clients are ready to receive messages
	for int(readyClient) != len(cases) {
		pb.Publish(dummyEvent)
		time.Sleep(10 * time.Millisecond)
	}

	for _, e := range events {
		pb.Publish(e)
		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()
	for _, c := range cases {
		assert.Equal(c.expectCount, c.actualCount)
	}
}

func readData(t *testing.T, r io.Reader) []byte {
	reader := bufio.NewReader(r)
	var data string
	for {
		l, _, err := reader.ReadLine()
		require.NoError(t, err)
		if len(l) != 0 {
			// handle only messages containing data
			test := string(l)
			if strings.HasPrefix(test, "data: ") {
				data = strings.TrimPrefix(test, "data: ")
				break
			}
		}
	}

	return []byte(data)
}

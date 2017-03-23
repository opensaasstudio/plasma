package manager

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/openfresh/plasma/event"
	"github.com/stretchr/testify/assert"
)

func TestAddClient(t *testing.T) {
	cases := []struct {
		Test Client
	}{
		{
			Test: NewClient([]string{"program:1234:poll"}),
		},
	}
	cm := NewClientManager()

	eventCnt := 0
	for _, c := range cases {
		eventCnt += len(c.Test.events)
		cm.AddClient(c.Test)
	}

	assert := assert.New(t)

	assert.Equal(eventCnt, len(cm.clientsTable), "should be equal")

	for _, c := range cases {
		for _, e := range c.Test.events {
			ch, ok := cm.clientsTable[e]
			assert.True(ok, "shuold be true")
			_, ok = ch.clients[c.Test.payloadChan]
			assert.True(ok, "should be true")
		}
	}
}

func TestRemoveClient(t *testing.T) {
	cases := []struct {
		Test Client
	}{
		{
			Test: NewClient([]string{"program:1234:poll"}),
		},
		{
			Test: NewClient([]string{
				"program:1234:poll",
				"program:1234:views",
			}),
		},
	}
	cm := NewClientManager()

	eventSet := make(map[string]struct{})
	for _, c := range cases {
		for _, e := range c.Test.events {
			eventSet[e] = struct{}{}
		}
		cm.AddClient(c.Test)
	}

	assert := assert.New(t)

	assert.Equal(len(eventSet), len(cm.clientsTable), "should be equal")

	for _, c := range cases {
		cm.RemoveClient(c.Test)
		_, ok := <-c.Test.payloadChan
		assert.False(ok, "should be close channel")
	}

	for _, c := range cases {
		for _, e := range c.Test.events {
			assert.Len(cm.clientsTable[e].clients, 0, "should be empty")
		}
	}

}

func TestCreateEvents(t *testing.T) {
	cases := []struct {
		Test   string
		Expect []string
	}{
		{
			Test: "program",
			Expect: []string{
				"program",
			},
		},
		{
			Test: "program:1234:views",
			Expect: []string{
				"program",
				"program:1234",
				"program:1234:views",
			},
		},
	}

	cm := NewClientManager()
	assert := assert.New(t)

	for _, c := range cases {
		actual := cm.createEvents(c.Test)
		assert.Equal(c.Expect, actual, "should be equal")
	}

}

func TestSendPayload(t *testing.T) {
	assert := assert.New(t)

	cm := NewClientManager()
	clients := []Client{
		NewClient([]string{"program:1234"}),
		NewClient([]string{"program:1234:views"}),
		NewClient([]string{"program:1234:poll"}),
		NewClient([]string{"program:1234:views", "program:1234:poll"}),
	}
	payload := event.Payload{
		Meta: event.MetaData{
			Type: "program:1234:views",
		},
		Data: json.RawMessage(`{"data": "Message"}`),
	}

	for _, c := range clients {
		cm.AddClient(c)
	}

	wg := &sync.WaitGroup{}
	for _, c := range clients {
		flag := false
		for _, e := range c.events {
			if strings.HasPrefix(payload.Meta.Type, e) {
				flag = true
				break
			}
		}
		if !flag {
			continue
		}
		wg.Add(1)
		go func(c Client) {
			defer wg.Done()
			p := <-c.payloadChan
			assert.Equal(payload, p)
		}(c)
	}

	cm.SendPayload(payload)

	wg.Wait()
}

func TestSendHeartBeat(t *testing.T) {
	assert := assert.New(t)

	cm := NewClientManager()
	clients := []Client{
		NewClient([]string{heartBeatEvent, "program:1234"}),
		NewClient([]string{heartBeatEvent, "program:1234:views"}),
	}

	for _, c := range clients {
		cm.AddClient(c)
	}

	wg := &sync.WaitGroup{}
	for _, c := range clients {
		wg.Add(1)
		go func(c Client) {
			defer wg.Done()
			p := <-c.payloadChan
			assert.Equal(heartBeatEvent, p.Meta.Type)
		}(c)
	}

	cm.SendHeartBeat()

	wg.Wait()
}

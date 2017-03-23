package manager

import (
	"strings"
	"sync"

	"github.com/openfresh/plasma/event"
)

type Client struct {
	events      []string
	payloadChan chan event.Payload
}

func (c *Client) ReceivePayload() <-chan event.Payload {
	return c.payloadChan
}

func NewClient(events []string) Client {
	return Client{
		events:      events,
		payloadChan: make(chan event.Payload),
	}
}

type Clients struct {
	clients map[chan event.Payload]struct{}
	mu      *sync.RWMutex
}

type ClientManager struct {
	clientsTable map[string]Clients
}

func (cm *ClientManager) AddClient(client Client) {
	for _, e := range client.events {
		if cm.clientsTable[e].clients == nil {
			cm.clientsTable[e] = Clients{
				clients: make(map[chan event.Payload]struct{}),
				mu:      &sync.RWMutex{},
			}
		}
		cm.clientsTable[e].mu.Lock()
		cm.clientsTable[e].clients[client.payloadChan] = struct{}{}
		cm.clientsTable[e].mu.Unlock()
	}
}

func (cm *ClientManager) RemoveClient(client Client) {
	for _, e := range client.events {
		clients, ok := cm.clientsTable[e]
		if !ok {
			continue
		}
		clients.mu.Lock()
		delete(clients.clients, client.payloadChan)
		clients.mu.Unlock()
	}
	close(client.payloadChan)
}

const eventSeparator = ":"

func (cm *ClientManager) createEvents(request string) []string {
	cnt := strings.Count(request, eventSeparator) + 1
	events := make([]string, cnt)

	idx := strings.Index(request, eventSeparator)
	for i := 0; i < cnt-1; i++ {
		events[i] = request[:idx]
		idx = len(request[:idx]) + strings.Index(request[idx+1:], eventSeparator) + 1
	}
	events[cnt-1] = request

	return events
}

func (cm *ClientManager) SendPayload(payload event.Payload) {
	for _, event := range cm.createEvents(payload.Meta.Type) {
		clientsTable, ok := cm.clientsTable[event]
		if !ok {
			continue
		}
		clientsTable.mu.RLock()
		clients := clientsTable.clients
		clientsTable.mu.RUnlock()
		for client := range clients {
			client <- payload
		}
	}
}

const heartBeatEvent = "heartbeat"

func (cm *ClientManager) SendHeartBeat() {
	clients, ok := cm.clientsTable[heartBeatEvent]
	if !ok {
		return
	}
	clients.mu.RLock()
	c := clients.clients
	clients.mu.RUnlock()
	for send, _ := range c {
		send <- event.Payload{Meta: event.MetaData{Type: heartBeatEvent}}
	}
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		clientsTable: make(map[string]Clients),
	}
}

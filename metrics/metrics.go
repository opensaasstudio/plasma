package metrics

import (
	"encoding/json"
	"io"

	metrics "github.com/rcrowley/go-metrics"
)

type Metrics struct {
	registry metrics.Registry
}

const (
	clientCounterName = "counter"
)

func New() (Metrics, error) {
	m := Metrics{
		registry: metrics.NewRegistry(),
	}

	clientCounter := metrics.NewCounter()
	if err := m.registry.Register(clientCounterName, clientCounter); err != nil {
		return m, err
	}

	return m, nil
}

func (m *Metrics) IncClientCount() {
	metrics.GetOrRegisterCounter(clientCounterName, m.registry).Inc(int64(1))
}

func (m *Metrics) DecClientCount() {
	metrics.GetOrRegisterCounter(clientCounterName, m.registry).Dec(int64(1))
}

func (m *Metrics) WriteJSON(w io.Writer) {
	json.NewEncoder(w).Encode(m.registry)
}

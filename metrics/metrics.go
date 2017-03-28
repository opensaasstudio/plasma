package metrics

import (
	"encoding/json"
	"io"
	"strings"

	metrics "github.com/rcrowley/go-metrics"
)

const (
	GRPC = "grpc"
	SSE  = "sse"
)

type Metrics struct {
	Type     string
	registry metrics.Registry
}

const (
	clientCounterName = "clients"
)

var registry metrics.Registry

func init() {
	registry = metrics.NewRegistry()
}

type Registry metrics.Registry

func New(metricsType string) (Metrics, error) {
	m := Metrics{
		registry: registry,
		Type:     metricsType,
	}
	return m, nil
}

func GetRegistry() metrics.Registry {
	return registry
}

func (m *Metrics) getFieldName(names []string) string {
	return strings.Join(append([]string{m.Type}, names...), ":")
}

func (m *Metrics) IncClientCount() {
	name := m.getFieldName([]string{clientCounterName})
	metrics.GetOrRegisterCounter(name, m.registry).Inc(int64(1))
}

func (m *Metrics) DecClientCount() {
	name := m.getFieldName([]string{clientCounterName})
	metrics.GetOrRegisterCounter(name, m.registry).Dec(int64(1))
}

func (m *Metrics) GetClientCount() int64 {
	name := m.getFieldName([]string{clientCounterName})
	return metrics.GetOrRegisterCounter(name, m.registry).Count()
}

func (m *Metrics) WriteJSON(w io.Writer) {
	json.NewEncoder(w).Encode(m.registry)
}

package metrics

import (
	"time"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/metrics/sender"
	metrics "github.com/rcrowley/go-metrics"
)

type Metrics struct {
	ticker *time.Ticker
	config config.Metrics
	sender sender.MetricsSender

	GcLast metrics.Gauge
	GcNext metrics.Gauge
	GcNum  metrics.Gauge

	GcPausePerSecond metrics.Gauge
	GoroutineNum     metrics.Gauge

	HeapAlloc   metrics.Gauge
	HeapIdle    metrics.Gauge
	HeapInuse   metrics.Gauge
	HeapObjects metrics.Gauge
	HeapSys     metrics.Gauge

	MemoryAlloc   metrics.Gauge
	MemoryFrees   metrics.Gauge
	MemoryLookups metrics.Gauge
	MemoryMallocs metrics.Gauge
	MemorySys     metrics.Gauge

	StackInUse metrics.Gauge

	Connections metrics.Gauge
}

func NewMetrics(config config.Config) (*Metrics, error) {
	m := &Metrics{
		config:           config.Metrics,
		GcLast:           metrics.NewGauge(),
		GcNext:           metrics.NewGauge(),
		GcNum:            metrics.NewGauge(),
		GcPausePerSecond: metrics.NewGauge(),
		GoroutineNum:     metrics.NewGauge(),
		HeapAlloc:        metrics.NewGauge(),
		HeapIdle:         metrics.NewGauge(),
		HeapInuse:        metrics.NewGauge(),
		HeapObjects:      metrics.NewGauge(),
		HeapSys:          metrics.NewGauge(),
		MemoryAlloc:      metrics.NewGauge(),
		MemoryFrees:      metrics.NewGauge(),
		MemoryLookups:    metrics.NewGauge(),
		MemoryMallocs:    metrics.NewGauge(),
		MemorySys:        metrics.NewGauge(),
		StackInUse:       metrics.NewGauge(),
		Connections:      metrics.NewGauge(),
	}

	if err := metrics.Register("GcLast", m.GcLast); err != nil {
		return m, err
	}
	if err := metrics.Register("GcNext", m.GcNext); err != nil {
		return m, err
	}
	if err := metrics.Register("GcNum", m.GcNum); err != nil {
		return m, err
	}
	if err := metrics.Register("GcPausePerSecond", m.GcPausePerSecond); err != nil {
		return m, err
	}
	if err := metrics.Register("GoroutineNum", m.GoroutineNum); err != nil {
		return m, err
	}
	if err := metrics.Register("HeapAlloc", m.HeapAlloc); err != nil {
		return m, err
	}
	if err := metrics.Register("HeapIdle", m.HeapIdle); err != nil {
		return m, err
	}
	if err := metrics.Register("HeapInuse", m.HeapInuse); err != nil {
		return m, err
	}
	if err := metrics.Register("HeapObjects", m.HeapObjects); err != nil {
		return m, err
	}
	if err := metrics.Register("HeapSys", m.HeapSys); err != nil {
		return m, err
	}
	if err := metrics.Register("MemoryAlloc", m.MemoryAlloc); err != nil {
		return m, err
	}
	if err := metrics.Register("MemoryFrees", m.MemoryFrees); err != nil {
		return m, err
	}
	if err := metrics.Register("MemoryLookups", m.MemoryLookups); err != nil {
		return m, err
	}
	if err := metrics.Register("MemoryMallocs", m.MemoryMallocs); err != nil {
		return m, err
	}
	if err := metrics.Register("MemorySys", m.MemorySys); err != nil {
		return m, err
	}
	if err := metrics.Register("StackInUse", m.StackInUse); err != nil {
		return m, err
	}
	if err := metrics.Register("Connections", m.Connections); err != nil {
		return m, err
	}

	sender, err := sender.NewMetricsSender(m.config)
	if err != nil {
		return m, err
	}
	m.sender = sender
	m.ticker = time.NewTicker(m.config.Interval)

	return m, nil
}

func (m *Metrics) Start() {
	go func() {
		for range m.ticker.C {
			s := GetStats()
			m.update(s)
		}
	}()

	go m.sender.Send()
}

func (m *Metrics) Stop() {
	m.ticker.Stop()
}

func (m *Metrics) update(s *Stats) {
	m.GcLast.Update(int64(s.GcLast))
	m.GcNext.Update(int64(s.GcNext))
	m.GcNum.Update(int64(s.GcNum))
	m.GcPausePerSecond.Update(int64(s.GcPausePerSecond))
	m.GoroutineNum.Update(int64(s.GoroutineNum))
	m.HeapAlloc.Update(int64(s.HeapAlloc))
	m.HeapIdle.Update(int64(s.HeapIdle))
	m.HeapInuse.Update(int64(s.HeapInuse))
	m.HeapObjects.Update(int64(s.HeapObjects))
	m.HeapSys.Update(int64(s.HeapSys))
	m.MemoryAlloc.Update(int64(s.MemoryAlloc))
	m.MemoryFrees.Update(int64(s.MemoryFrees))
	m.MemoryLookups.Update(int64(s.MemoryLookups))
	m.MemoryMallocs.Update(int64(s.MemoryMallocs))
	m.MemorySys.Update(int64(s.MemorySys))
	m.StackInUse.Update(int64(s.StackInUse))
	m.Connections.Update(s.Connections)
}

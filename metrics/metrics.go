package metrics

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/openfresh/plasma/config"
	"github.com/pkg/errors"
	metrics "github.com/rcrowley/go-metrics"
)

const (
	Log = "log"
)

type metricsSender interface {
	Send()
}

type logSender struct {
	logger metrics.Logger
	config config.LogMetrics
}

func newLogSender(config config.LogMetrics) (logSender, error) {
	sender := logSender{}

	var writer io.Writer

	switch config.Out {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		w, err := os.OpenFile(config.Out, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return sender, errors.Wrapf(err, "failed to open file: %s", config.Out)
		}
		writer = w
	}

	sender.logger = log.New(writer, config.Prefix, config.Flag)

	return sender, nil
}

func (s logSender) Send() {
	metrics.Log(metrics.DefaultRegistry, s.config.Interval, s.logger)
}

func newMetricsSender(config config.Metrics) (metricsSender, error) {
	var metricsSender metricsSender
	var err error

	switch config.Type {
	case Log:
		metricsSender, err = newLogSender(config.Log)
	default:
		err = errors.New(fmt.Sprintf("unkown metrics sender type: %s", config.Type))
	}

	return metricsSender, err
}

type Metrics struct {
	config config.Metrics
	sender metricsSender

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

	metrics.Register("GcLast", m.GcLast)
	metrics.Register("GcNext", m.GcNext)
	metrics.Register("GcNum", m.GcNum)
	metrics.Register("GcPausePerSecond", m.GcPausePerSecond)
	metrics.Register("GoroutineNum", m.GoroutineNum)
	metrics.Register("HeapAlloc", m.HeapAlloc)
	metrics.Register("HeapIdle", m.HeapIdle)
	metrics.Register("HeapInuse", m.HeapInuse)
	metrics.Register("HeapObjects", m.HeapObjects)
	metrics.Register("HeapSys", m.HeapSys)
	metrics.Register("MemoryAlloc", m.MemoryAlloc)
	metrics.Register("MemoryFrees", m.MemoryFrees)
	metrics.Register("MemoryLookups", m.MemoryLookups)
	metrics.Register("MemoryMallocs", m.MemoryMallocs)
	metrics.Register("MemorySys", m.MemorySys)
	metrics.Register("StackInUse", m.StackInUse)
	metrics.Register("Connections", m.Connections)

	sender, err := newMetricsSender(m.config)
	if err != nil {
		return m, err
	}
	m.sender = sender

	return m, nil
}

func (m *Metrics) Start() {
	ticker := time.NewTicker(m.config.Interval)
	defer ticker.Stop()

	go func() {
		<-ticker.C
		s := GetStats()
		m.update(s)
	}()

	go m.sender.Send()
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

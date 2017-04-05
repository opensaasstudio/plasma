package sender

import (
	"fmt"

	"github.com/openfresh/plasma/config"
)

type MetricsSender interface {
	Send()
}

func NewMetricsSender(config config.Metrics) (MetricsSender, error) {
	var metricsSender MetricsSender
	var err error

	switch config.Type {
	case Log:
		metricsSender, err = newLogSender(config.Log)
	default:
		err = fmt.Errorf("unkown metrics sender type: %s", config.Type)
	}

	return metricsSender, err
}

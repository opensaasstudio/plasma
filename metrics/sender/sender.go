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
	case Syslog:
		metricsSender, err = newSyslogSender(config.Syslog)
	default:
		err = fmt.Errorf("unknown metrics sender type: %s", config.Type)
	}

	return metricsSender, err
}

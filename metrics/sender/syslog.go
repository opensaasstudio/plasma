package sender

import (
	"log/syslog"

	"github.com/openfresh/plasma/config"
	metrics "github.com/rcrowley/go-metrics"
)

const Syslog = "syslog"

type syslogSender struct {
	writer *syslog.Writer
	config config.SyslogMetrics
}

func newSyslogSender(config config.SyslogMetrics) (syslogSender, error) {
	sender := syslogSender{
		config: config,
	}

	c := sender.config

	priority := syslog.Priority(c.Severity | c.Facility)
	writer, err := syslog.Dial(c.Network, c.Addr, priority, c.Tag)
	if err != nil {
		return sender, err
	}

	sender.writer = writer

	return sender, nil
}

func (s syslogSender) Send() {
	metrics.Syslog(metrics.DefaultRegistry, s.config.Interval, s.writer)
}

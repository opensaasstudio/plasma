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

	priority := syslog.Priority(sender.config.Priority)
	writer, err := syslog.New(priority, sender.config.Tag)
	if err != nil {
		return sender, err
	}

	sender.writer = writer

	return sender, nil
}

func (s syslogSender) Send() {
	metrics.Syslog(metrics.DefaultRegistry, s.config.Interval, s.writer)
}

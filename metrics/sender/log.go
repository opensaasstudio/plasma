package sender

import (
	"io"
	"log"
	"os"

	"github.com/openfresh/plasma/config"
	"github.com/pkg/errors"
	metrics "github.com/rcrowley/go-metrics"
)

const Log = "log"

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

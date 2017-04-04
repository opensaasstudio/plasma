package log

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/pkg/errors"

	"github.com/openfresh/plasma/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(config config.Log) (*zap.Logger, error) {
	var writer io.Writer
	switch config.Out {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	case "discard":
		writer = ioutil.Discard
	default:
		w, err := os.OpenFile(config.Out, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open file: %s", config.Out)
		}
		writer = w
	}

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal level %s", config.Level)
	}

	writerSyncer := zapcore.AddSync(writer)
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.Lock(writerSyncer),
		level,
	), zap.ErrorOutput(writerSyncer))

	return logger, nil
}

func HTTPRequestToLogFields(r *http.Request) []zapcore.Field {
	remoteAddr := r.RemoteAddr
	if addr := r.Header.Get("X-Forwarded-For"); addr != "" {
		remoteAddr = addr
	}
	return []zapcore.Field{
		zap.String("user-agent", r.UserAgent()),
		zap.String("referer", r.Referer()),
		zap.Int64("content-length", r.ContentLength),
		zap.String("host", r.Host),
		zap.String("method", r.Method),
		zap.String("remote-addr", remoteAddr),
	}
}

package config

import (
	"errors"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/kelseyhightower/envconfig"
)

func New() Config {
	config := Config{}
	envconfig.MustProcess("plasma", &config)

	if config.AccessLog.Out == "" {
		config.AccessLog.Out = "stdout"
	}
	if config.ErrorLog.Out == "" {
		config.ErrorLog.Out = "stderr"
	}
	if config.AccessLog.Level == "" {
		config.AccessLog.Level = "debug"
	}

	if config.ErrorLog.Level == "" {
		config.ErrorLog.Level = "debug"
	}

	return config
}

type Config struct {
	AccessLog  Log `envconfig:"ACCESS_LOG"`
	ErrorLog   Log `envconfig:"ERROR_LOG"`
	Debug      bool
	Origin     string
	Port       string `default:"8080"`
	SSE        ServerSideEvent
	Subscriber Subscriber
	TLS        Cert `envconfig:"TLS"`
}

type ServerSideEvent struct {
	Retry      int    `default:"2000"`
	EventQuery string `default:"eventType"`
}

type Subscriber struct {
	Type  string `default:"mock"`
	Redis Redis
	Mock  Mock
}

type Mock struct {
	Interval time.Duration `default:"1s"`
}

type Channels []string

func (cs Channels) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, c := range cs {
		enc.AppendString(c)
	}
	return nil
}

type OverMaxRetryBehavior struct {
	Type string
}

func (b *OverMaxRetryBehavior) UnmarshalText(text []byte) error {
	switch string(text) {
	case OverMaxRetryBehaviorAlive:
		b.Type = OverMaxRetryBehaviorAlive
	case OverMaxRetryBehaviorDie:
		b.Type = OverMaxRetryBehaviorDie
	default:
		return errors.New("unknown OverMaxRetryBehavior type: " + string(text))
	}

	return nil
}

var OverMaxRetryBehaviorAlive = "alive"
var OverMaxRetryBehaviorDie = "die"

type Redis struct {
	Addr                 string `default:"localhost:6379"`
	Password             string
	DB                   int
	Channels             Channels
	OverMaxRetryBehavior OverMaxRetryBehavior `envconfig:"OVER_MAX_RETRY_BEHAVIOR"`
	MaxRetry             int                  `default:"5"`
	Timeout              time.Duration        `default:"1s"`
	RetryInterval        time.Duration        `default:"5s" envconfig:"RETRY_INTERVAL"`
}

func (r Redis) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("Addr", r.Addr)
	enc.AddString("Password", r.Password)
	enc.AddInt("DB", r.DB)
	if err := enc.AddArray("Channels", r.Channels); err != nil {
		return err
	}
	return nil
}

type Log struct {
	Out   string
	Level string
}

type Cert struct {
	CertFile string `envconfig:"CERT_FILE"`
	KeyFile  string `envconfig:"KEY_FILE"`
}

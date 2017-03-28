package server

import (
	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/metrics"
	"github.com/openfresh/plasma/pubsub"
	"go.uber.org/zap"
)

type Option struct {
	PubSuber     pubsub.PubSuber
	AccessLogger *zap.Logger
	ErrorLogger  *zap.Logger
	Config       config.Config
	Metrics      metrics.Metrics
}

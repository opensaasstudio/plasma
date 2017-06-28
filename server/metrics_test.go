package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/log"
	"github.com/stretchr/testify/assert"
)

func TestMetricsGo(t *testing.T) {
	assert := assert.New(t)
	l, err := log.NewLogger(config.Log{
		Out: "discard",
	})
	assert.Nil(err)

	handler := NewMetricsHandler(Option{
		AccessLogger: l,
		ErrorLogger:  l,
		Config: config.Config{
			Subscriber: config.Subscriber{
				Type: "redis",
				Redis: config.Redis{
					Addr:     "localhost:6379",
					DB:       0,
					Password: "",
				},
			},
		},
	})

	req, err := http.NewRequest("GET", "/metrics/go", nil)
	assert.Nil(err)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
}

func TestMetricsPlasma(t *testing.T) {
	assert := assert.New(t)
	l, err := log.NewLogger(config.Log{
		Out: "discard",
	})
	assert.Nil(err)

	handler := NewMetricsHandler(Option{
		AccessLogger: l,
		ErrorLogger:  l,
		Config: config.Config{
			Subscriber: config.Subscriber{
				Type: "redis",
				Redis: config.Redis{
					Addr:     "localhost:6379",
					DB:       0,
					Password: "",
				},
			},
		},
	})

	req, err := http.NewRequest("GET", "/metrics/plasma", nil)
	assert.Nil(err)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
}

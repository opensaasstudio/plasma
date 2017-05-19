package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/log"

	"github.com/stretchr/testify/assert"
)

func TestCheckRedis(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		Redis config.Redis
		IsErr bool
	}{
		{
			Redis: config.Redis{
				Addr:     "localhost:6379",
				DB:       0,
				Password: "",
			},
			IsErr: false,
		},
		{
			Redis: config.Redis{
				Addr:     "fakehost:6379",
				DB:       0,
				Password: "",
			},
			IsErr: true,
		},
	}

	for _, c := range cases {
		err := checkRedis(c.Redis)
		if c.IsErr {
			assert.NotNil(err)
		} else {
			assert.Nil(err)
		}
	}
}

func TestHealthCheckHandler(t *testing.T) {
	assert := assert.New(t)
	l, err := log.NewLogger(config.Log{
		Out: "discard",
	})
	assert.Nil(err)

	cases := []struct {
		Redis config.Redis
		IsErr bool
	}{
		{
			Redis: config.Redis{
				Addr:     "localhost:6379",
				DB:       0,
				Password: "",
			},
			IsErr: false,
		},
		{
			Redis: config.Redis{
				Addr:     "fakehost:6379",
				DB:       0,
				Password: "",
			},
			IsErr: true,
		},
	}

	for _, c := range cases {
		handler := NewMetaHandler(Option{
			AccessLogger: l,
			ErrorLogger:  l,
			Config: config.Config{
				Subscriber: config.Subscriber{
					Type:  "redis",
					Redis: c.Redis,
				},
			},
		})

		req, err := http.NewRequest("GET", "/hc", nil)
		assert.Nil(err)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if c.IsErr {
			assert.Equal(http.StatusInternalServerError, rec.Code)
		} else {
			assert.Equal(http.StatusOK, rec.Code)
		}
	}
}

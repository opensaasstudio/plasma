package subscriber

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"

	"github.com/gorilla/mux"
	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/event"
	"github.com/openfresh/plasma/pubsub"
)

type Api struct {
	config  config.Config
	pubsub  pubsub.PubSuber
	payload chan event.Payload
}

func newApi(pb pubsub.PubSuber, errorLogger *zap.Logger, config config.Config) (Subscriber, error) {
	return &Api{
		config:  config,
		pubsub:  pb,
		payload: make(chan event.Payload),
	}, nil
}

func (r *Api) handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	var payload event.Payload
	b, _ := ioutil.ReadAll(req.Body)
	json.Unmarshal(b, &payload)
	r.payload <- payload
}

func (r *Api) Subscribe() error {
	l := mux.NewRouter()
	go func() {
		l.HandleFunc("/", r.handler).Methods(http.MethodPost)
		if err := http.ListenAndServe(":8090", l); err != nil {
			fmt.Println(err)
		}
	}()

	for {
		r.pubsub.Publish(<-r.payload)
	}
}

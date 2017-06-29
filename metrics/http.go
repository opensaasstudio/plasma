package metrics

import (
	"encoding/json"
	"net/http"
)

func GoStatsHandler(w http.ResponseWriter, r *http.Request) {
	s := GetGoStats()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func PlasmaStatsHandler(w http.ResponseWriter, r *http.Request) {
	s := GetPlasmaStats()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

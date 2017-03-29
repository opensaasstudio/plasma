package metrics

import (
	"encoding/json"
	"net/http"
)

func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	s := GetStats()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

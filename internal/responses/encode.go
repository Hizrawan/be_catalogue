package responses

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func Upsert(w http.ResponseWriter, status int, ok bool, data any) error {
	return JSON(w, status, struct {
		OK   bool `json:"ok"`
		Data any  `json:"data"`
	}{
		OK:   ok,
		Data: data,
	})
}

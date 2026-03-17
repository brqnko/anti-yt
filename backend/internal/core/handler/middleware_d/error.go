package middleware_d

import (
	"encoding/json"
	"net/http"
	"time"
)

func writeErrorJSON(w http.ResponseWriter, statusCode int, title, detail string) (interface{}, error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(struct {
		Title  string `json:"title"`
		Detail string `json:"detail"`
	}{
		Title:  title,
		Detail: detail,
	})
	return nil, nil
}

func writeForbiddenJSON(w http.ResponseWriter, code, title, detail string, t *time.Time) (interface{}, error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(struct {
		Code   string     `json:"code"`
		Title  string     `json:"title"`
		Detail string     `json:"detail"`
		Time   *time.Time `json:"time,omitempty"`
	}{
		Code:   code,
		Title:  title,
		Detail: detail,
		Time:   t,
	})
	return nil, nil
}

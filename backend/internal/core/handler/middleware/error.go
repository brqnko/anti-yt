package middleware

import (
	"encoding/json"
	"net/http"
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

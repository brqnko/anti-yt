package middleware_d

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecureHeaders(t *testing.T) {
	t.Parallel()

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecureHeaders(inner)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	expected := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":       "SAMEORIGIN",
		"X-Xss-Protection":      "1; mode=block",
	}
	for key, want := range expected {
		got := w.Header().Get(key)
		if got != want {
			t.Fatalf("header %q: expected %q, got %q", key, want, got)
		}
	}
}

package http

import (
	nethttp "net/http"
	"strings"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	rec := serve(nethttp.MethodGet, "/health", "")

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != `{"status":"ok"}` {
		t.Fatalf("unexpected body: %s", body)
	}
}

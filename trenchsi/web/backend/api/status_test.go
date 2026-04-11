package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandleStatusReturnsOK(t *testing.T) {
	h := NewHandler("")
	h.startTime = time.Now().Add(-5 * time.Second)

	mux := http.NewServeMux()
	h.registerStatusRoutes(mux)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp appStatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if resp.Status != "ok" {
		t.Fatalf("resp.Status = %q, want ok", resp.Status)
	}
	if resp.Version == "" {
		t.Fatal("resp.Version should not be empty")
	}
	if resp.Uptime == "" {
		t.Fatal("resp.Uptime should not be empty")
	}
}

func TestNormalizeCommit(t *testing.T) {
	if got := normalizeCommit("acf38242b7d1\n"); got != "acf38242" {
		t.Fatalf("normalizeCommit() = %q, want %q", got, "acf38242")
	}
}

package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestHandler() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(NewBuildInfo("v1.2.3", "abc1234", "2026-01-01T00:00:00Z"), logger)
}

func do(t *testing.T, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	newTestHandler().ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), method, path, nil))
	return rec
}

func TestHealthEndpoints(t *testing.T) {
	for _, path := range []string{"/healthz", "/readyz"} {
		t.Run(path, func(t *testing.T) {
			rec := do(t, http.MethodGet, path)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			var body map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}
			if body["status"] != "ok" {
				t.Fatalf("status field = %q, want ok", body["status"])
			}
		})
	}
}

func TestVersionReturnsBuildInfo(t *testing.T) {
	rec := do(t, http.MethodGet, "/version")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var info BuildInfo
	if err := json.Unmarshal(rec.Body.Bytes(), &info); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if info.Version != "v1.2.3" || info.Commit != "abc1234" {
		t.Fatalf("unexpected build info: %+v", info)
	}
	if info.GoVersion == "" {
		t.Fatal("goVersion must be set")
	}
}

func TestNewBuildInfoDefaultsToUnknown(t *testing.T) {
	info := NewBuildInfo("", "", "")
	if info.Version != "unknown" || info.Commit != "unknown" || info.BuildDate != "unknown" {
		t.Fatalf("empty values must become unknown, got %+v", info)
	}
}

func TestSecurityHeadersOnEveryResponse(t *testing.T) {
	want := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "no-referrer",
		"Cache-Control":          "no-store",
	}
	for _, path := range []string{"/healthz", "/version", "/does-not-exist"} {
		rec := do(t, http.MethodGet, path)
		for k, v := range want {
			if got := rec.Header().Get(k); got != v {
				t.Errorf("%s: header %s = %q, want %q", path, k, got, v)
			}
		}
	}
}

func TestUnknownRouteIs404(t *testing.T) {
	if rec := do(t, http.MethodGet, "/admin"); rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestWrongMethodIsRejected(t *testing.T) {
	if rec := do(t, http.MethodPost, "/healthz"); rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

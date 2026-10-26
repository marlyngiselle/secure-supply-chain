// Package server exposes the HTTP handlers of the demo service.
//
// The service is deliberately small. Its job is to be the payload that the
// supply chain builds, signs, attests and admits, so every handler is pure,
// dependency free and covered by unit tests.
package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime"
	"time"
)

// BuildInfo is stamped at link time by the release pipeline with -ldflags.
type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
}

// NewBuildInfo returns build metadata, replacing empty values with "unknown".
func NewBuildInfo(version, commit, date string) BuildInfo {
	orUnknown := func(s string) string {
		if s == "" {
			return "unknown"
		}
		return s
	}
	return BuildInfo{
		Version:   orUnknown(version),
		Commit:    orUnknown(commit),
		BuildDate: orUnknown(date),
		GoVersion: runtime.Version(),
	}
}

// New returns the HTTP handler with every route and the security headers applied.
func New(info BuildInfo, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /readyz", healthz)
	mux.HandleFunc("GET /version", version(info))
	return securityHeaders(accessLog(logger, mux))
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func version(info BuildInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, info)
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// securityHeaders sets conservative defaults for an API that serves no HTML.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func accessLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

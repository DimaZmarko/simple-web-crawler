package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	"github.com/DimaZmarko/simple-web-crawler/apps/api/internal/api/types"
)

// fakePinger implements Pinger, returning a fixed error (nil means healthy).
type fakePinger struct {
	err error
}

func (f fakePinger) Ping(context.Context) error { return f.err }

func TestHealthEndpoints(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		pinger         Pinger
		wantStatus     int
		wantBodyStatus string
		wantDBCheck    string // empty when the response has no checks.db field
	}{
		{
			name:           "healthz always ok",
			method:         http.MethodGet,
			path:           "/healthz",
			pinger:         fakePinger{err: nil},
			wantStatus:     http.StatusOK,
			wantBodyStatus: "ok",
		},
		{
			name:           "readyz ok when db ping succeeds",
			method:         http.MethodGet,
			path:           "/readyz",
			pinger:         fakePinger{err: nil},
			wantStatus:     http.StatusOK,
			wantBodyStatus: "ok",
			wantDBCheck:    "ok",
		},
		{
			name:           "readyz degraded when db ping fails",
			method:         http.MethodGet,
			path:           "/readyz",
			pinger:         fakePinger{err: errors.New("connection refused")},
			wantStatus:     http.StatusServiceUnavailable,
			wantBodyStatus: "degraded",
			wantDBCheck:    "down",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := NewRouter(tc.pinger, nil, zerolog.Nop())

			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			res := rec.Result()
			defer func() { _ = res.Body.Close() }()

			if res.StatusCode != tc.wantStatus {
				body, _ := io.ReadAll(res.Body)
				t.Fatalf("status = %d, want %d (body: %s)", res.StatusCode, tc.wantStatus, body)
			}

			if ct := res.Header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}

			// Decode into a permissive shape so both HealthStatus and
			// ReadinessStatus bodies are verifiable from one struct.
			var body struct {
				Status string `json:"status"`
				Checks *struct {
					Db string `json:"db"`
				} `json:"checks"`
			}
			if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}

			if body.Status != tc.wantBodyStatus {
				t.Errorf("status field = %q, want %q", body.Status, tc.wantBodyStatus)
			}

			if tc.wantDBCheck != "" {
				if body.Checks == nil {
					t.Fatalf("checks missing, want checks.db = %q", tc.wantDBCheck)
				}
				if body.Checks.Db != tc.wantDBCheck {
					t.Errorf("checks.db = %q, want %q", body.Checks.Db, tc.wantDBCheck)
				}
			}
		})
	}
}

// TestCORSAllowsConfiguredOrigin verifies the browser origin gets an
// Access-Control-Allow-Origin header, so the web app can read /readyz cross-origin.
func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	const origin = "http://localhost:3000"
	router := NewRouter(fakePinger{err: nil}, nil, zerolog.Nop(), origin)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	req.Header.Set("Origin", origin)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, origin)
	}
}

// TestReadyzMatchesContractTypes guards that the readiness handler emits the
// exact enum values defined by the generated contract types.
func TestReadyzMatchesContractTypes(t *testing.T) {
	router := NewRouter(fakePinger{err: nil}, nil, zerolog.Nop())

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var got types.ReadinessStatus
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode ReadinessStatus: %v", err)
	}

	if got.Status != types.Ok {
		t.Errorf("status = %q, want %q", got.Status, types.Ok)
	}
	if got.Checks.Db != types.ReadinessStatusChecksDbOk {
		t.Errorf("checks.db = %q, want %q", got.Checks.Db, types.ReadinessStatusChecksDbOk)
	}
}

package crawl

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateCreate(t *testing.T) {
	tests := []struct {
		name       string
		in         CreateInput
		wantFields []string // field names expected in the ValidationError, nil = valid
	}{
		{
			name: "valid input",
			in:   CreateInput{SeedURL: "https://example.com", MaxDepth: 2, MaxPages: 500},
		},
		{
			name: "valid http and zero depth",
			in:   CreateInput{SeedURL: "http://example.com/path", MaxDepth: 0, MaxPages: 1},
		},
		{
			name:       "relative url",
			in:         CreateInput{SeedURL: "/just/a/path", MaxDepth: 1, MaxPages: 1},
			wantFields: []string{"seedUrl"},
		},
		{
			name:       "non-http scheme",
			in:         CreateInput{SeedURL: "ftp://example.com", MaxDepth: 1, MaxPages: 1},
			wantFields: []string{"seedUrl"},
		},
		{
			name:       "missing host",
			in:         CreateInput{SeedURL: "https://", MaxDepth: 1, MaxPages: 1},
			wantFields: []string{"seedUrl"},
		},
		{
			name:       "negative depth",
			in:         CreateInput{SeedURL: "https://example.com", MaxDepth: -1, MaxPages: 1},
			wantFields: []string{"maxDepth"},
		},
		{
			name:       "zero pages",
			in:         CreateInput{SeedURL: "https://example.com", MaxDepth: 1, MaxPages: 0},
			wantFields: []string{"maxPages"},
		},
		{
			name:       "pages above ceiling",
			in:         CreateInput{SeedURL: "https://example.com", MaxDepth: 1, MaxPages: maxPagesCeiling + 1},
			wantFields: []string{"maxPages"},
		},
		{
			name:       "all invalid",
			in:         CreateInput{SeedURL: "nope", MaxDepth: -3, MaxPages: 0},
			wantFields: []string{"seedUrl", "maxDepth", "maxPages"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			verr := validateCreate(tc.in)

			if tc.wantFields == nil {
				if verr != nil {
					t.Fatalf("got error %v, want valid", verr)
				}
				return
			}

			if verr == nil {
				t.Fatalf("got valid, want errors on %v", tc.wantFields)
			}

			got := map[string]bool{}
			for _, fe := range verr.Errors {
				got[fe.Field] = true
			}
			for _, f := range tc.wantFields {
				if !got[f] {
					t.Errorf("missing field error for %q (got %v)", f, verr.Errors)
				}
			}
			if len(verr.Errors) != len(tc.wantFields) {
				t.Errorf("error count = %d, want %d", len(verr.Errors), len(tc.wantFields))
			}
		})
	}
}

func TestClampLimit(t *testing.T) {
	tests := []struct {
		in, want int
	}{
		{0, defaultLimit},
		{-5, defaultLimit},
		{1, 1},
		{50, 50},
		{MaxLimit, MaxLimit},
		{MaxLimit + 1, MaxLimit},
		{10000, MaxLimit},
	}
	for _, tc := range tests {
		if got := clampLimit(tc.in); got != tc.want {
			t.Errorf("clampLimit(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestCursorRoundTrip(t *testing.T) {
	orig := cursor{
		CreatedAt: time.Date(2026, 6, 14, 9, 45, 0, 0, time.UTC),
		ID:        uuid.MustParse("1a2b3c4d-5e6f-4071-8293-a4b5c6d7e8f9"),
	}

	token := encodeCursor(orig)
	if token == "" {
		t.Fatal("encodeCursor returned empty token")
	}

	got, err := decodeCursor(token)
	if err != nil {
		t.Fatalf("decodeCursor: %v", err)
	}
	if !got.CreatedAt.Equal(orig.CreatedAt) {
		t.Errorf("createdAt = %v, want %v", got.CreatedAt, orig.CreatedAt)
	}
	if got.ID != orig.ID {
		t.Errorf("id = %v, want %v", got.ID, orig.ID)
	}
}

func TestDecodeCursorRejectsGarbage(t *testing.T) {
	tests := []string{
		"!!!not base64!!!",
		"",
		"e30",        // base64 of "{}" — decodes but has zero-value fields
		"bm90anNvbg", // base64 of "notjson"
	}
	for _, token := range tests {
		t.Run(token, func(t *testing.T) {
			if _, err := decodeCursor(token); err == nil {
				t.Errorf("decodeCursor(%q) = nil error, want ErrInvalidCursor", token)
			}
		})
	}
}
